package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/amba/amba/backend/internal/grounding"
	"github.com/amba/amba/backend/internal/models"
)

func TestCorePhrase(t *testing.T) {
	for _, tc := range []struct{ name, title, desc, want string }{
		{"plain title", "Banjir besar menerjang Jakarta pagi ini", "", "Banjir besar menerjang Jakarta pagi ini"},
		{"hashtags and links removed", "Gibran mundur dari wapres #viral #fyp https://t.co/x", "", "Gibran mundur dari wapres"},
		{"emoji and punctuation", "WOW!!! Pesawat jatuh di Bandung 😱😱", "", "WOW Pesawat jatuh di Bandung"},
		{"mentions removed", "@akun lihat Prabowo naik motor di Solo", "", "lihat Prabowo naik motor di Solo"},
		{"capped at ten words", "satu dua tiga empat lima enam tujuh delapan sembilan sepuluh sebelas dua belas", "", "satu dua tiga empat lima enam tujuh delapan sembilan sepuluh"},
		{"short title falls back to description", "Viral!", "Seekor singa masuk ke kota Surabaya semalam. Warga panik.", "Seekor singa masuk ke kota Surabaya semalam"},
		{"nothing usable", "#fyp #viral", "ok", ""},
		{"empty", "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := corePhrase(tc.title, tc.desc); got != tc.want {
				t.Fatalf("corePhrase = %q, want %q", got, tc.want)
			}
		})
	}
}

// webPipeline runs the mock model with a fake web search.
func webPipeline(t *testing.T, results []grounding.Result, err error) (*Pipeline, *fakeSearcher, func()) {
	t.Helper()
	p, srv := pipelineWithFetch(t, true, ogPage)
	s := &fakeSearcher{results: results, err: err}
	p.WebSearch = s
	p.WebGrounder = grounding.New(grounding.WebAllowlist)
	return p, s, srv.Close
}

func TestWebProvenanceKeepsOnlyJudgedPagesOnTheVideosSubject(t *testing.T) {
	p, s, closeFn := webPipeline(t, []grounding.Result{
		{Title: "Video banjir besar hari ini ternyata buatan AI", Publisher: "kompas.com", URL: "https://www.kompas.com/cekfakta/read/1"},
		{Title: "[HOAKS] Banjir besar hari ini di Jakarta", Publisher: "turnbackhoax.id", URL: "https://turnbackhoax.id/2026/11/10/x", Rating: "Hoaks"},
		{Title: "Banjir besar hari ini, ini penyebabnya", Publisher: "detik.com", URL: "https://news.detik.com/x"},
		{Title: "Harga emas turun", Publisher: "kompas.com", URL: "https://www.kompas.com/y", Rating: "Hoaks"},
		{Title: "Banjir besar hari ini buatan AI", Publisher: "blog", URL: "https://evil.example/z"},
	}, nil)
	defer closeFn()
	p.Video = &fakeVideo{clip: testClip(t)}

	res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: tiktokURL})
	ev := res.Video.WebEvidence
	if len(ev) != 2 || ev[0].Says != models.WebSaysAI || ev[1].Says != models.WebSaysHoax {
		t.Fatalf("evidence = %+v", ev)
	}
	if len(s.queries) != 1 || s.queries[0] != "Banjir besar hari ini video AI hoaks" {
		t.Fatalf("queries = %v", s.queries)
	}
	if res.Degraded {
		t.Fatal("the web search is optional and must not degrade a result")
	}
}

func TestWebAIEvidenceIsProvenanceAndWebHoaxContradictsTheFactPillar(t *testing.T) {
	p, _, closeFn := webPipeline(t, []grounding.Result{
		{Title: "Video banjir besar hari ini ternyata buatan AI", Publisher: "kompas.com", URL: "https://www.kompas.com/cekfakta/read/1"},
		{Title: "[HOAKS] Banjir besar hari ini di Jakarta", Publisher: "turnbackhoax.id", URL: "https://turnbackhoax.id/2026/11/10/x", Rating: "Hoaks"},
	}, nil)
	defer closeFn()
	p.Video = &fakeVideo{clip: testClip(t)}

	res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: tiktokURL})
	if pv := res.Pillars.AI.Provenance; pv.Kind != models.ProvenanceAIGenerated || pv.Source != "web" {
		t.Fatalf("provenance = %+v", pv)
	}
	f := res.Pillars.Fact
	if !f.Available || f.Status != models.FactContradicts {
		t.Fatalf("fact = %+v", f)
	}
	found := false
	for _, s := range f.Sources {
		if s.URL == "https://turnbackhoax.id/2026/11/10/x" && s.Match == models.FactContradicts {
			found = true
		}
	}
	if !found {
		t.Fatalf("sources = %+v", f.Sources)
	}
	if v := Verdict(res, false); v.Status != models.StatusTampered {
		t.Fatalf("verdict = %+v", v)
	}
}

func TestWebHoaxNeedsAFactCheckArticle(t *testing.T) {
	p, _, closeFn := webPipeline(t, []grounding.Result{
		{Title: "Banjir besar hari ini di Jakarta", Publisher: "detik.com", URL: "https://news.detik.com/berita/1", Rating: "Salah"},
	}, nil)
	defer closeFn()
	p.Video = &fakeVideo{clip: testClip(t)}
	res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: tiktokURL})
	if len(res.Video.WebEvidence) != 0 {
		t.Fatalf("evidence = %+v, a plain news page cannot call a hoax", res.Video.WebEvidence)
	}
}

func TestWebSearchFailureIsSilent(t *testing.T) {
	p, _, closeFn := webPipeline(t, nil, grounding.ErrSearchFailed)
	defer closeFn()
	p.Video = &fakeVideo{clip: testClip(t)}
	res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: tiktokURL})
	if res.Video.Status != models.VideoOK || len(res.Video.WebEvidence) != 0 || res.Video.WebEvidence == nil {
		t.Fatalf("video = %+v", res.Video)
	}
}

// blockingSearcher waits until its context ends.
type blockingSearcher struct{}

func (blockingSearcher) Enabled() bool { return true }
func (blockingSearcher) Search(ctx context.Context, _ string) ([]grounding.Result, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestSlowWebSearchCannotHoldTheVerdict(t *testing.T) {
	p, srv := pipelineWithFetch(t, true, ogPage)
	defer srv.Close()
	p.WebSearch = blockingSearcher{}
	p.WebSearchDeadline = 80 * time.Millisecond
	p.Video = &fakeVideo{clip: testClip(t)}
	start := time.Now()
	res := p.Run(context.Background(), Request{InputType: models.InputURL, URL: tiktokURL})
	if time.Since(start) > 3*time.Second || res.Video.Status != models.VideoOK {
		t.Fatalf("took %v, video %+v", time.Since(start), res.Video)
	}
}

func TestNoWebSearchWhenOffOrNothingToSearch(t *testing.T) {
	p, srv := pipelineWithFetch(t, true, ogPage)
	defer srv.Close()
	if ch := p.startWebProvenance(context.Background(), testClip(t)); ch != nil {
		t.Fatal("search started without a searcher")
	}
	s := &fakeSearcher{}
	p.WebSearch = s
	clip := testClip(t)
	clip.Title, clip.Description = "#fyp", ""
	if ch := p.startWebProvenance(context.Background(), clip); ch != nil || len(s.queries) != 0 {
		t.Fatal("search started with no usable words")
	}
}
