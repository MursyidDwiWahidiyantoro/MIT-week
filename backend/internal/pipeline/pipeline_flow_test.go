package pipeline

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/amba/amba/backend/internal/grounding"
	"github.com/amba/amba/backend/internal/linkcheck"
	"github.com/amba/amba/backend/internal/llm"
	"github.com/amba/amba/backend/internal/models"
)

// fakeSearcher returns fixed results; URLs are what verification checks.
type fakeSearcher struct {
	results []grounding.Result
	err     error
	queries []string
	mu      sync.Mutex
}

func (f *fakeSearcher) Enabled() bool { return true }
func (f *fakeSearcher) Search(_ context.Context, q string) ([]grounding.Result, error) {
	f.mu.Lock()
	f.queries = append(f.queries, q)
	f.mu.Unlock()
	return f.results, f.err
}

// claimRequest is an image the deterministic mock answers with a claim: the
// fact pillar only runs when the model found one.
func claimRequest(t *testing.T) Request {
	t.Helper()
	for i := 0; i < 60; i++ {
		req := imageRequest(t, 100+i, "Bantuan cair hari ini")
		if newTestPipeline(true).Run(context.Background(), req).Pillars.Fact.Available {
			return req
		}
	}
	t.Fatal("no mock input with a claim")
	return Request{}
}

func TestFactCheckUsesOnlySearchedSources(t *testing.T) {
	p := newTestPipeline(true)
	s := &fakeSearcher{results: []grounding.Result{
		{Title: "[HOAKS] Bantuan", Publisher: "Kompas.com", URL: "https://www.kompas.com/cekfakta/read/1", Rating: "Hoaks"},
		{Title: "Lain", Publisher: "Tempo", URL: "https://cekfakta.tempo.co/2", Rating: "Perlu konteks"},
	}}
	p.Searcher = s

	res := p.Run(context.Background(), claimRequest(t))
	fact := res.Pillars.Fact
	if !fact.Available || fact.Status != models.FactContradicts || len(fact.Sources) != 2 {
		t.Fatalf("fact pillar = %+v", fact)
	}
	if fact.Sources[0].Match != models.FactContradicts || fact.Sources[1].Match != models.FactUnverified {
		t.Fatalf("matches = %+v", fact.Sources)
	}
	if len(s.queries) != 1 || s.queries[0] == "" {
		t.Fatalf("search queries = %v", s.queries)
	}
	v := Verdict(res, false)
	if v.Status != models.StatusTampered || len(v.FloorsApplied) == 0 || v.FloorsApplied[0] != "FACT_CONTRADICTED" {
		t.Fatalf("verdict = %+v", v)
	}
}

func TestFactCheckWithoutSearchIsUnverifiedNotDegraded(t *testing.T) {
	p := newTestPipeline(true)
	res := p.Run(context.Background(), claimRequest(t))
	if res.Pillars.Fact.Status != models.FactUnverified || len(res.Pillars.Fact.Sources) != 0 {
		t.Fatalf("fact = %+v", res.Pillars.Fact)
	}
	if res.Degraded {
		t.Fatal("no search configured is a deployment choice, not a degraded run")
	}
}

func TestSearchOutageDegrades(t *testing.T) {
	p := newTestPipeline(true)
	p.Searcher = &fakeSearcher{err: grounding.ErrSearchFailed}
	res := p.Run(context.Background(), claimRequest(t))
	if !res.Degraded || res.Pillars.Fact.Status != models.FactUnverified {
		t.Fatalf("degraded=%v fact=%+v", res.Degraded, res.Pillars.Fact)
	}
}

func TestEventsArriveAsPillarsFinish(t *testing.T) {
	p := newTestPipeline(true)
	var mu sync.Mutex
	var events []Event
	emit := func(e Event) {
		mu.Lock()
		events = append(events, e)
		mu.Unlock()
	}
	p.RunStream(context.Background(), Request{InputType: models.InputURL, URL: "https://bri-verifikasi.xyz/login"}, emit)

	terminal := map[string]string{}
	running := map[string]int{}
	for i, e := range events {
		switch e.Status {
		case "running":
			running[e.Pillar] = i
		default:
			if _, dup := terminal[e.Pillar]; dup {
				t.Fatalf("pillar %s ended twice: %+v", e.Pillar, events)
			}
			if start, ok := running[e.Pillar]; ok && start > i {
				t.Fatalf("pillar %s ended before it started", e.Pillar)
			}
			terminal[e.Pillar] = e.Status
		}
	}
	want := map[string]string{"ela": "skipped", "link": "done"}
	for pillar, status := range want {
		if terminal[pillar] != status {
			t.Fatalf("%s = %q, want %q (events %+v)", pillar, terminal[pillar], status, events)
		}
	}
	if _, ok := terminal["fact"]; !ok {
		t.Fatal("fact pillar never settled")
	}
}

func TestCachedRunReplaysEvents(t *testing.T) {
	p := newTestPipeline(true)
	req := imageRequest(t, 7, "")
	p.Run(context.Background(), req)
	var events []Event
	p.RunStream(context.Background(), req, func(e Event) { events = append(events, e) })
	done := 0
	for _, e := range events {
		if e.Status == "done" {
			done++
		}
	}
	if done == 0 || len(events) < 4 {
		t.Fatalf("cache hit must still settle every pillar: %+v", events)
	}
}

func TestCacheIsBounded(t *testing.T) {
	p := newTestPipeline(true)
	p.CacheMax = 3
	for i := 0; i < 10; i++ {
		p.toCache(strings.Repeat("k", i+1), Result{})
	}
	if n := p.CacheLen(); n != 3 {
		t.Fatalf("cache length = %d, want 3", n)
	}
	if _, ok := p.fromCache("k"); ok {
		t.Fatal("oldest entry should have been evicted")
	}
	if _, ok := p.fromCache(strings.Repeat("k", 10)); !ok {
		t.Fatal("newest entry missing")
	}

	p.CacheTTL = time.Millisecond
	p.toCache("short", Result{})
	time.Sleep(5 * time.Millisecond)
	if _, ok := p.fromCache("short"); ok {
		t.Fatal("expired entry returned")
	}
}

// metadataServer impersonates a page with OpenGraph tags; the pipeline's
// fetch client is pointed at it so no real network is touched.
func pipelineWithFetch(t *testing.T, mock bool, handler http.HandlerFunc) (*Pipeline, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	p := newTestPipeline(mock)
	p.NewFetchClient = func(timeout time.Duration) *http.Client {
		return &http.Client{Timeout: timeout, Transport: rewrite(srv)}
	}
	return p, srv
}

func TestURLAuditReadsMetadataAndVerifiesTacticsAgainstCaption(t *testing.T) {
	p, srv := pipelineWithFetch(t, true, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><meta property="og:title" content="Segera klaim hadiah hari ini"></head></html>`))
	})
	defer srv.Close()

	res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: "https://promo.example.com/post"})
	if res.Metadata == nil || res.Metadata.Status != models.MetadataOK || res.Metadata.Title != "Segera klaim hadiah hari ini" {
		t.Fatalf("metadata = %+v", res.Metadata)
	}
	if len(res.Tactics) == 0 {
		t.Fatal("tactics from the caption should survive verification")
	}
	for _, tc := range res.Tactics {
		if !strings.Contains(strings.ToLower(res.Metadata.Title), strings.ToLower(tc.Phrase)) {
			t.Fatalf("tactic %q is not in the caption", tc.Phrase)
		}
	}
}

func TestURLAuditOnBlockedPlatform(t *testing.T) {
	p, srv := pipelineWithFetch(t, true, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "login required", http.StatusForbidden)
	})
	defer srv.Close()
	res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: "https://www.instagram.com/p/abc/"})
	if res.Metadata == nil || res.Metadata.Status != models.MetadataBlocked || res.Metadata.Platform != "instagram" {
		t.Fatalf("metadata = %+v", res.Metadata)
	}
}

func TestShortLinkIsScoredByItsDestination(t *testing.T) {
	p, srv := pipelineWithFetch(t, true, func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "bit.ly" || strings.HasPrefix(r.URL.Path, "/abc") {
			w.Header().Set("Location", "https://bri-verifikasi.xyz/login")
			w.WriteHeader(http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()
	res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: "https://bit.ly/abc"})
	link := res.Pillars.Link
	if link.FinalURL != "https://bri-verifikasi.xyz/login" || len(link.RedirectChain) != 2 {
		t.Fatalf("link = %+v", link)
	}
	codes := ""
	for _, r := range link.Reasons {
		codes += r.Code + " "
	}
	if !strings.Contains(codes, linkcheck.ReasonShortenerHides) || !strings.Contains(codes, linkcheck.ReasonLookalikeBrand) {
		t.Fatalf("reasons = %s", codes)
	}
}

func TestSafeBrowsingHitSetsMalicious(t *testing.T) {
	sb := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"matches":[{}]}`))
	}))
	defer sb.Close()
	p, srv := pipelineWithFetch(t, true, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	defer srv.Close()
	p.SafeBrowsing = linkcheck.NewSafeBrowsing("k")
	p.SafeBrowsing.Endpoint = sb.URL

	res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: "https://example.org/"})
	if !res.LinkMalicious || res.Pillars.Link.Score != 100 {
		t.Fatalf("malicious=%v score=%d", res.LinkMalicious, res.Pillars.Link.Score)
	}
	if v := Verdict(res, false); v.Status != models.StatusTampered {
		t.Fatalf("verdict = %+v", v)
	}
}

// A screenshot whose OCR text holds a URL gets the link pillar, and the image
// itself is what the model receives.
func TestScreenshotOCRLinkIsChecked(t *testing.T) {
	var sawImage bool
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []struct {
				Content any `json:"content"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if parts, ok := body.Messages[1].Content.([]any); ok && len(parts) == 2 {
			sawImage = true
		}
		inner := `{"claim":"","search_query":"","ai_score":30,"ai_indicators":[],"text_anomalies":[],"ocr_text":"Klaim di https://dana-bonus.xyz/klaim sekarang","tactics":[],"injection_suspected":false}`
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"message": map[string]string{"content": inner}}}})
	}))
	defer gateway.Close()

	p, srv := pipelineWithFetch(t, false, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	defer srv.Close()
	p.LLM = llm.New(gateway.URL, "k", "m", false)

	res := p.Run(context.Background(), Request{InputType: models.InputFile, Image: tinyPNG(t), ImageFormat: "png"})
	if !sawImage {
		t.Fatal("the model never received the image")
	}
	if !res.Pillars.Link.Available || !strings.Contains(res.Pillars.Link.FinalURL, "dana-bonus.xyz") {
		t.Fatalf("link pillar = %+v", res.Pillars.Link)
	}
}

func rewrite(srv *httptest.Server) http.RoundTripper {
	return roundTripper(func(req *http.Request) (*http.Response, error) {
		r2 := req.Clone(req.Context())
		r2.URL.Scheme = "http"
		r2.URL.Host = strings.TrimPrefix(srv.URL, "http://")
		r2.Host = req.URL.Host
		return http.DefaultTransport.RoundTrip(r2)
	})
}

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func tinyPNG(t *testing.T) []byte {
	t.Helper()
	// 2x2 PNG, generated once with image/png.
	return []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02, 0x08, 0x02, 0x00, 0x00, 0x00, 0xfd, 0xd4, 0x9a, 0x73, 0x00, 0x00, 0x00, 0x16, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x62, 0xf8, 0xcf, 0xc0, 0xc0, 0xf0, 0x9f, 0x81, 0x81, 0x01, 0x10, 0x00, 0x00, 0xff, 0xff, 0x1d, 0xec, 0x03, 0xfd, 0x7e, 0x3f, 0x8b, 0x28, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82}
}

// Phishing on a hacked or free-hosting site has no tell-tale in its address;
// the page itself (a bank's name over a password field) gives it away.
func TestPhishingPageOnAnOrdinaryHost(t *testing.T) {
	p, srv := pipelineWithFetch(t, true, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if r.Method == http.MethodHead {
			return
		}
		_, _ = w.Write([]byte(`<html><head><title>KlikBCA Individual</title></head><body><form><input name="user_id"><input type="password" name="pin"></form></body></html>`))
	})
	defer srv.Close()

	res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: "https://wisata-kita.example/wp-content/klik/"})
	link := res.Pillars.Link
	found := false
	for _, r := range link.Reasons {
		if r.Code == linkcheck.ReasonPageImpersonation && r.Text != "" {
			found = true
		}
	}
	if !found || link.Score < 45 {
		t.Fatalf("link = %+v", link)
	}
	if v := Verdict(res, false); v.Status == models.StatusSafe || v.Status == models.StatusInsufficient {
		t.Fatalf("verdict = %+v", v)
	}
}
