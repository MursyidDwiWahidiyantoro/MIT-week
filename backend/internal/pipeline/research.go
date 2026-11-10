package pipeline

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/amba/amba/backend/internal/grounding"
	"github.com/amba/amba/backend/internal/llm"
	"github.com/amba/amba/backend/internal/models"
)

// Research (Sir, 10 Nov 2026): the conclusion about a video must rest on data
// from trusted places. What the video itself shows (its fact-check keywords,
// or else its 5W1H) becomes one search over the trusted newsrooms and
// fact-checkers, and the pages it finds are returned as the trail behind the
// verdict. There is no model call: the trail is data the family can open and
// read, not a second opinion. It starts as soon as the video is understood,
// runs beside the fact check, and never degrades or delays a result past its
// own budget.
const (
	researchBudget = 8 * time.Second
	researchMax    = 5
)

// researchTerms is the model's fact-check keywords, or with none, what the
// video shows, where and when, cut to a searchable phrase ("" when too thin).
func researchTerms(vf llm.VideoFindings) string {
	if q := strings.TrimSpace(vf.SearchQuery); q != "" {
		return q
	}
	w := vf.W5H1
	return corePhrase(strings.Join([]string{w.What, w.Where, w.When}, " "), w.Who)
}

// startResearch starts the search and returns a channel that receives
// exactly one result, or nil when research is off or there is nothing to
// search. ctx must outlive the audit's errgroup (it is awaited after it).
func (p *Pipeline) startResearch(ctx context.Context, vf llm.VideoFindings) <-chan models.Research {
	if p.Research == nil || !p.Research.Enabled() {
		return nil
	}
	q := researchTerms(vf)
	if q == "" {
		return nil
	}
	ch := make(chan models.Research, 1)
	go func() {
		defer func() {
			if recover() != nil {
				ch <- models.Research{Query: q, Status: models.ResearchUnavailable, Sources: []models.ResearchSource{}}
			}
		}()
		ch <- p.research(ctx, q)
	}()
	return ch
}

// research runs one search and keeps the trusted, on-topic pages.
func (p *Pipeline) research(ctx context.Context, q string) models.Research {
	out := models.Research{Query: q, Status: models.ResearchNone, Sources: []models.ResearchSource{}}
	sctx, cancel := withBudget(ctx, researchBudget)
	results, err := p.Research.Search(sctx, q)
	cancel()
	if err != nil {
		out.Status = models.ResearchUnavailable
		return out
	}
	allowed := p.researchGrounder()
	seen := map[string]bool{}
	for _, r := range results {
		if len(out.Sources) >= researchMax {
			break
		}
		// No allowlist, no trail: a page is shown only from a trusted host.
		if seen[r.URL] || allowed == nil || !allowed.HostAllowed(r.URL) || !grounding.SameTopic(q, r.Title+" "+r.Snippet) {
			continue
		}
		seen[r.URL] = true
		out.Sources = append(out.Sources, models.ResearchSource{
			Title:   truncateRunes(strings.TrimSpace(r.Title), 200),
			URL:     r.URL,
			Domain:  domainOf(r.URL, r.Publisher),
			Snippet: truncateRunes(strings.TrimSpace(r.Snippet), 280),
		})
	}
	if len(out.Sources) > 0 {
		out.Status = models.ResearchFound
	}
	return out
}

func (p *Pipeline) researchGrounder() *grounding.Grounder {
	if p.ResearchGrounder != nil {
		return p.ResearchGrounder
	}
	return p.webGrounder()
}

// domainOf is the page's host without "www.", or the searcher's publisher.
func domainOf(rawURL, publisher string) string {
	if u, err := url.Parse(rawURL); err == nil && u.Hostname() != "" {
		return strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	}
	return publisher
}

// awaitResearch waits for the search, at most until ctx ends.
func awaitResearch(ctx context.Context, ch <-chan models.Research, q string) models.Research {
	select {
	case r := <-ch:
		return r
	case <-ctx.Done():
		return models.Research{Query: q, Status: models.ResearchUnavailable, Sources: []models.ResearchSource{}}
	}
}
