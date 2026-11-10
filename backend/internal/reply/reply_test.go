package reply

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/amba/amba/backend/internal/llm"
	"github.com/amba/amba/backend/internal/token"
)

func TestTemplatesCoverEveryStatusAndTone(t *testing.T) {
	for _, status := range []string{"TAMPERED", "SUSPICIOUS", "SAFE", "INSUFFICIENT"} {
		for _, tone := range Tones {
			text := Template(status, tone)
			if text == "" || len([]rune(text)) > MaxLength {
				t.Errorf("%s/%s invalid: %q", status, tone, text)
			}
			if _, ok := Validate(text, nil); !ok {
				t.Errorf("template %s/%s fails its own validation", status, tone)
			}
		}
	}
	if Template("WHATEVER", "santai") != Template("INSUFFICIENT", "santai") {
		t.Error("unknown status must fall back to INSUFFICIENT")
	}
}

// The embedded copy must equal the reviewed config (the frontend copy is
// diffed by tools/check-demo-diff.sh).
func TestEmbeddedTemplatesMatchConfig(t *testing.T) {
	config, err := os.ReadFile("../../../configs/reply_templates.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(config) != string(templatesJSON) {
		t.Fatal("backend/internal/reply/reply_templates.json drifted from configs/reply_templates.json")
	}
}

func TestValidate(t *testing.T) {
	allowed := map[string]bool{"https://www.kompas.com/cekfakta/read/1": true}
	tests := []struct {
		name, in, want string
		ok             bool
	}{
		{"plain", "Makasih infonya, aku cek dulu ya sebelum ikut sebar.", "Makasih infonya, aku cek dulu ya sebelum ikut sebar.", true},
		{"allowed url kept", "Ternyata sudah dicek: https://www.kompas.com/cekfakta/read/1.", "Ternyata sudah dicek: https://www.kompas.com/cekfakta/read/1.", true},
		{"foreign url stripped", "Cek dulu ya https://evil.example/x sebelum disebar.", "Cek dulu ya sebelum disebar.", true},
		{"only a url", "https://evil.example/x", "", false},
		{"certainty", "Ini pasti hoaks, jangan disebar.", "", false},
		{"percent", "Kabar ini 100% tidak benar.", "", false},
		{"score", "Skornya 92/100, hati-hati ya.", "", false},
		{"shaming", "Kamu salah lagi, jangan bodoh.", "", false},
		{"markup", "Terima kasih <b>banyak</b> infonya ya.", "", false},
		{"too long", strings.Repeat("Terima kasih. ", 30), "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Validate(tt.in, allowed)
			if ok != tt.ok || (ok && got != tt.want) {
				t.Fatalf("Validate(%q) = %q, %v; want %q, %v", tt.in, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func gateway(t *testing.T, text string, calls *int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(calls, 1)
		inner, _ := json.Marshal(map[string]string{"text": text})
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"message": map[string]string{"content": string(inner)}}}})
	}))
}

func TestDraftUsesValidModelTextAndCaches(t *testing.T) {
	var calls int32
	srv := gateway(t, "Makasih infonya ya, kata cek fakta Kompas ini keliru: https://www.kompas.com/cekfakta/read/1", &calls)
	defer srv.Close()
	s := New(llm.New(srv.URL, "k", "m", false))
	p := token.Payload{Status: "TAMPERED", Title: "Terindikasi Hoax", Srcs: []token.Source{{Pub: "Kompas", URL: "https://www.kompas.com/cekfakta/read/1", M: "CONTRADICTS"}}}

	text, source := s.Draft(context.Background(), "tok", p, "santai")
	if source != "llm" || !strings.Contains(text, "kompas.com") {
		t.Fatalf("got %q from %s", text, source)
	}
	again, _ := s.Draft(context.Background(), "tok", p, "santai")
	if again != text || atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("second call should hit the cache (calls=%d)", calls)
	}
}

func TestDraftFallsBackToTemplate(t *testing.T) {
	var calls int32
	srv := gateway(t, "Ini 100% hoaks!", &calls)
	defer srv.Close()
	p := token.Payload{Status: "SUSPICIOUS"}

	text, source := New(llm.New(srv.URL, "k", "m", false)).Draft(context.Background(), "tok", p, "hormat")
	if source != "template" || text != Template("SUSPICIOUS", "hormat") {
		t.Fatalf("invalid draft must fall back: %q (%s)", text, source)
	}
	text, source = New(llm.New("", "", "", true)).Draft(context.Background(), "tok", p, "singkat")
	if source != "template" || text != Template("SUSPICIOUS", "singkat") {
		t.Fatalf("mock mode must use templates: %q (%s)", text, source)
	}
}
