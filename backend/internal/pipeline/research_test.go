package pipeline

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/amba/amba/backend/internal/grounding"
	"github.com/amba/amba/backend/internal/llm"
	"github.com/amba/amba/backend/internal/models"
)

// fakeResearch answers every query with the same results (or error) and
// records what it was asked.
type fakeResearch struct {
	results []grounding.Result
	err     error
	asked   []string
}

func (f *fakeResearch) Enabled() bool { return true }
func (f *fakeResearch) Search(_ context.Context, q string) ([]grounding.Result, error) {
	f.asked = append(f.asked, q)
	return f.results, f.err
}

func TestResearchTermsPrefersKeywordsThenWhatTheVideoShows(t *testing.T) {
	vf := llm.VideoFindings{SearchQuery: "banjir jakarta mobil terseret"}
	if got := researchTerms(vf); got != "banjir jakarta mobil terseret" {
		t.Fatalf("keywords first: got %q", got)
	}
	vf = llm.VideoFindings{W5H1: llm.W5H1{What: "Banjir menyeret mobil", Where: "di Jakarta Utara", When: "Senin pagi"}}
	if got := researchTerms(vf); got != "Banjir menyeret mobil di Jakarta Utara Senin pagi" {
		t.Fatalf("5W1H next: got %q", got)
	}
	if got := researchTerms(llm.VideoFindings{W5H1: llm.W5H1{What: "Kucing"}}); got != "" {
		t.Fatalf("too thin to search: got %q", got)
	}
}

func TestResearchKeepsOnlyTrustedOnTopicPages(t *testing.T) {
	fake := &fakeResearch{results: []grounding.Result{
		{Title: "Banjir rendam Jakarta Utara, mobil terseret arus", URL: "https://www.kompas.com/a", Publisher: "kompas.com", Snippet: strings.Repeat("banjir jakarta ", 40)},
		{Title: "Banjir Jakarta mobil terseret", URL: "https://blog-acak.example/banjir", Publisher: "blog-acak.example", Snippet: "banjir jakarta"},
		{Title: "Harga cabai naik di pasar", URL: "https://www.detik.com/b", Publisher: "detik.com", Snippet: "cabai rawit"},
		{Title: "Banjir rendam Jakarta Utara, mobil terseret arus", URL: "https://www.kompas.com/a", Publisher: "kompas.com", Snippet: "ulang"},
		{Title: "Video mobil terseret banjir di Jakarta bukan hoaks", URL: "http://turnbackhoax.id/c", Publisher: "turnbackhoax.id", Snippet: "banjir jakarta"},
	}}
	p := &Pipeline{Research: fake, ResearchGrounder: grounding.New(grounding.WebAllowlist)}
	r := p.research(context.Background(), "banjir jakarta mobil terseret")
	if r.Status != models.ResearchFound || len(r.Sources) != 1 {
		t.Fatalf("want 1 trusted on-topic page, got %s %+v", r.Status, r.Sources)
	}
	s := r.Sources[0]
	if s.Domain != "kompas.com" || s.URL != "https://www.kompas.com/a" {
		t.Fatalf("domain without www and the page's own URL: %+v", s)
	}
	if n := len([]rune(s.Snippet)); n > 280 {
		t.Fatalf("snippet capped at 280 runes, got %d", n)
	}
	if r.Query != "banjir jakarta mobil terseret" || len(fake.asked) != 1 {
		t.Fatalf("one search with the query, got %q %v", r.Query, fake.asked)
	}
}

func TestResearchSaysPlainlyWhenItFoundNothingOrCouldNotRun(t *testing.T) {
	g := grounding.New(grounding.WebAllowlist)
	none := (&Pipeline{Research: &fakeResearch{}, ResearchGrounder: g}).research(context.Background(), "banjir jakarta")
	if none.Status != models.ResearchNone || none.Sources == nil || len(none.Sources) != 0 {
		t.Fatalf("nothing found: %+v", none)
	}
	down := (&Pipeline{Research: &fakeResearch{err: errors.New("432")}, ResearchGrounder: g}).research(context.Background(), "banjir jakarta")
	if down.Status != models.ResearchUnavailable {
		t.Fatalf("a failed search is unavailable, not 'nothing found': %+v", down)
	}
}

func TestResearchIsOffWithoutASearcherOrTerms(t *testing.T) {
	vf := llm.VideoFindings{SearchQuery: "banjir jakarta mobil"}
	if ch := (&Pipeline{}).startResearch(context.Background(), vf); ch != nil {
		t.Fatal("no searcher: no research")
	}
	if ch := (&Pipeline{Research: &fakeResearch{}}).startResearch(context.Background(), llm.VideoFindings{}); ch != nil {
		t.Fatal("no terms: no research")
	}
	ch := (&Pipeline{Research: &fakeResearch{}, ResearchGrounder: grounding.New(grounding.WebAllowlist)}).startResearch(context.Background(), vf)
	if ch == nil {
		t.Fatal("searcher and terms: research runs")
	}
	if r := awaitResearch(context.Background(), ch, "banjir jakarta mobil"); r.Status != models.ResearchNone {
		t.Fatalf("got %+v", r)
	}
}
