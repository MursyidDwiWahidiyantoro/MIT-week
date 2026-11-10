package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func analyze(t *testing.T, c *Client, inputType, content string) (Findings, error) {
	t.Helper()
	return c.Analyze(context.Background(), Request{InputType: inputType, Content: content})
}

func TestMockIsDeterministic(t *testing.T) {
	c := New("http://unused", "", "", true)
	a, err := analyze(t, c, "TEXT", "pesan yang sama")
	if err != nil {
		t.Fatal(err)
	}
	b, err := analyze(t, c, "TEXT", "pesan yang sama")
	if err != nil {
		t.Fatal(err)
	}
	if a.AIScore != b.AIScore || a.SearchQuery != b.SearchQuery {
		t.Fatalf("mock is not deterministic: %+v vs %+v", a, b)
	}
}

func TestMockVariesByInput(t *testing.T) {
	c := New("http://unused", "", "", true)
	seen := map[int]bool{}
	for _, input := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		f, err := analyze(t, c, "TEXT", input)
		if err != nil {
			t.Fatal(err)
		}
		seen[f.AIScore] = true
	}
	if len(seen) < 2 {
		t.Fatalf("mock returned a single value across inputs: %v", seen)
	}
}

func TestMockNeverClaimsCertainty(t *testing.T) {
	c := New("http://unused", "", "", true)
	for _, input := range []string{"a", "b", "c", "d", "e", "f"} {
		f, _ := analyze(t, c, "TEXT", input)
		blob := strings.ToLower(strings.Join(f.AIIndicators, " ") + f.Claim)
		for _, banned := range []string{"pasti", "terbukti", "100%", "dijamin"} {
			if strings.Contains(blob, banned) {
				t.Errorf("mock copy contains %q: %s", banned, blob)
			}
		}
	}
}

func TestSanitizeClampsScore(t *testing.T) {
	if got := sanitize(Findings{AIScore: 500}); got.AIScore != 100 {
		t.Errorf("high score = %d, want 100", got.AIScore)
	}
	if got := sanitize(Findings{AIScore: -50}); got.AIScore != 0 {
		t.Errorf("low score = %d, want 0", got.AIScore)
	}
}

// The gateway must receive an explicit stream:false; without it 9router
// answers with SSE and the old client failed every live call.
func TestRequestAlwaysAsksForNonStreaming(t *testing.T) {
	var seen map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&seen)
		writeCompletion(w, `{"ai_score":10,"claim":"x","ai_indicators":[],"ocr_text":"","search_query":"q","tactics":[],"injection_suspected":false}`)
	}))
	defer srv.Close()

	c := New(srv.URL, "k", "Maru", false)
	if _, err := analyze(t, c, "TEXT", "halo"); err != nil {
		t.Fatal(err)
	}
	if stream, ok := seen["stream"].(bool); !ok || stream {
		t.Fatalf("stream = %v, want explicit false", seen["stream"])
	}
	if seen["reasoning_effort"] != "low" {
		t.Errorf("reasoning_effort = %v, want low", seen["reasoning_effort"])
	}
}

// A gateway that ignores stream:false must still be understood.
func TestSSEBodyIsDecoded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		chunk := func(delta string) string {
			return `data: {"choices":[{"delta":{"content":` + strconvQuote(delta) + `},"finish_reason":null}]}` + "\n\n"
		}
		_, _ = w.Write([]byte(chunk(`{"ai_score":77,"claim":"x","ai_ind`)))
		_, _ = w.Write([]byte(chunk(`icators":[],"ocr_text":"","search_query":"q","tactics":[],"injection_suspected":false}`)))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	c := New(srv.URL, "k", "Maru", false)
	f, err := analyze(t, c, "TEXT", "halo")
	if err != nil {
		t.Fatalf("SSE body should decode: %v", err)
	}
	if f.AIScore != 77 {
		t.Fatalf("ai_score = %d, want 77", f.AIScore)
	}
}

func TestFencedJSONIsUnwrapped(t *testing.T) {
	for _, body := range []string{
		"```json\n{\"ai_score\":42}\n```",
		"Here is the JSON:\n```json\n{\"ai_score\":42}\n```\nDone.",
		"{\"ai_score\":42}",
	} {
		got := stripFences(body)
		var v struct {
			AIScore int `json:"ai_score"`
		}
		if err := json.Unmarshal([]byte(got), &v); err != nil {
			t.Errorf("stripFences(%q) = %q, not JSON: %v", body, got, err)
		}
	}
}

func TestSchemaViolationReturnsUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeCompletion(w, "not json at all")
	}))
	defer srv.Close()

	c := New(srv.URL, "k", "m", false)
	if _, err := analyze(t, c, "TEXT", "x"); err == nil {
		t.Fatal("expected ErrUnavailable on a schema violation")
	}
}

func TestNon200ReturnsUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(srv.URL, "k", "m", false)
	if _, err := analyze(t, c, "TEXT", "x"); err == nil {
		t.Fatal("expected an error on HTTP 500")
	}
}

func Test429IsDistinct(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := New(srv.URL, "k", "m", false)
	if _, err := analyze(t, c, "TEXT", "x"); err == nil || err.Error() != ErrRateLimited.Error() {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}
}

func TestTransportErrorRetriesOnce(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			// Abort the connection mid-response: a transport-level failure.
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Error("cannot hijack")
				return
			}
			conn, _, _ := hj.Hijack()
			_ = conn.Close()
			return
		}
		writeCompletion(w, `{"ai_score":5,"claim":"x","ai_indicators":[],"ocr_text":"","search_query":"q","tactics":[],"injection_suspected":false}`)
	}))
	defer srv.Close()

	c := New(srv.URL, "k", "m", false)
	if _, err := analyze(t, c, "TEXT", "x"); err != nil {
		t.Fatalf("second attempt should succeed: %v", err)
	}
	if hits.Load() != 2 {
		t.Fatalf("hits = %d, want 2", hits.Load())
	}
}

func TestMissingCredentialsReturnsUnavailable(t *testing.T) {
	c := New("http://localhost:1", "", "", false)
	if _, err := analyze(t, c, "TEXT", "x"); err == nil {
		t.Fatal("expected an error without credentials")
	}
}

func TestDescribeNeverLeaksKeyOrURL(t *testing.T) {
	c := New("http://secret-host:9999/v1", "super-secret", "m", false)
	d := c.Describe()
	if strings.Contains(d, "super-secret") || strings.Contains(d, "secret-host") {
		t.Fatalf("Describe leaked sensitive data: %q", d)
	}
}

func TestImageIsSentAsDataURL(t *testing.T) {
	var seen map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&seen)
		writeCompletion(w, `{"ai_score":50,"claim":"x","ai_indicators":[],"ocr_text":"","search_query":"q","tactics":[],"injection_suspected":false}`)
	}))
	defer srv.Close()

	c := New(srv.URL, "k", "m", false)
	_, err := c.Analyze(context.Background(), Request{
		InputType: "FILE", Image: []byte{0xFF, 0xD8, 0xFF}, ImageMIME: "image/jpeg",
	})
	if err != nil {
		t.Fatal(err)
	}
	msgs := seen["messages"].([]any)
	user := msgs[1].(map[string]any)
	parts, ok := user["content"].([]any)
	if !ok || len(parts) != 2 {
		t.Fatalf("vision content = %#v, want text+image parts", user["content"])
	}
	img := parts[1].(map[string]any)["image_url"].(map[string]any)["url"].(string)
	if !strings.HasPrefix(img, "data:image/jpeg;base64,") {
		t.Fatalf("image url = %q", img)
	}
}

func TestHealthCheckFindsModel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"Maru","capabilities":{"vision":true,"tools":true}}]}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "k", "Maru", false)
	info := c.HealthCheck(context.Background())
	if !info.Found || info.Count != 1 {
		t.Fatalf("info = %+v", info)
	}
	if len(info.Capabilities) != 2 {
		t.Fatalf("capabilities = %v", info.Capabilities)
	}
}

func writeCompletion(w http.ResponseWriter, inner string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"choices": []map[string]any{{"message": map[string]string{"content": inner}}},
	})
}

func strconvQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// Maru ignores response_format and returns its own keys. That response is
// still a usable assessment; the decoder must not throw it away.
func TestDecodeFindingsAcceptsMaruShape(t *testing.T) {
	raw := `{
		"ai_score": 85,
		"patterns": ["urgensi palsu", "rantai pesan", "tautan mencurigakan"],
		"reasoning": "Konten menciptakan urgensi palsu dan meminta pesan disebar.",
		"search_query": "bantuan 500 ribu bagikan grup verifikasi bitly"
	}`
	f, err := decodeFindings(raw)
	if err != nil {
		t.Fatal(err)
	}
	if f.AIScore != 85 {
		t.Errorf("ai_score = %d, want 85", f.AIScore)
	}
	if len(f.AIIndicators) != 3 {
		t.Errorf("indicators = %v", f.AIIndicators)
	}
	if f.SearchQuery == "" {
		t.Error("search_query was dropped")
	}
	if f.Claim == "" {
		t.Error("reasoning should fill the claim when none is given")
	}
}

func TestDecodeFindingsNestedAndAliased(t *testing.T) {
	raw := `{"findings":{"skor_ai":42,"indikator":["tekanan"],"kueri":"cek fakta","taktik":[{"kode":"FEAR","frasa":"diblokir"}]}}`
	f, err := decodeFindings(raw)
	if err != nil {
		t.Fatal(err)
	}
	if f.AIScore != 42 || len(f.AIIndicators) != 1 || f.SearchQuery != "cek fakta" {
		t.Fatalf("decoded %+v", f)
	}
	if len(f.Tactics) != 1 || f.Tactics[0].Code != "FEAR" || f.Tactics[0].Phrase != "diblokir" {
		t.Fatalf("tactics = %+v", f.Tactics)
	}
}

func TestDecodeFindingsRequiresScore(t *testing.T) {
	if _, err := decodeFindings(`{"patterns":["x"]}`); err == nil {
		t.Fatal("a response with no score must fail the pillar")
	}
}
