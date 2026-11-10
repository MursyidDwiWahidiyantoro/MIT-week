package pipeline

import (
	"context"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/amba/amba/backend/internal/grounding"
	"github.com/amba/amba/backend/internal/llm"
	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/video"
)

// Web provenance search: the video's own title and description become one
// search over reputable newsrooms and fact-checkers, and each hit is judged
// against the video: AI-made, hoax, real, or about something else. It runs
// beside the frame analysis and never holds the verdict past the video's
// own time budget.
const (
	webSearchBudget   = 6 * time.Second
	webClassifyBudget = 8 * time.Second
	webMaxEvidence    = 5
	// webPhraseWords is how many words of the title form the search phrase.
	webPhraseWords = 10
)

// webOutcome is what the search found. Evidence holds only pages judged to be
// about this video; Phrase is the text that was searched.
type webOutcome struct {
	Evidence []models.WebEvidence
	Phrase   string
	// Judged is the classification of every page that reached the model, by URL,
	// so the claim check does not send the same page again.
	Judged map[string]string
}

// says returns the evidence with the given verdict.
func (w webOutcome) says(v string) []models.WebEvidence {
	var out []models.WebEvidence
	for _, e := range w.Evidence {
		if e.Says == v {
			out = append(out, e)
		}
	}
	return out
}

var (
	urlRe     = regexp.MustCompile(`https?://\S+|www\.\S+`)
	hashRe    = regexp.MustCompile(`[#@][\p{L}\p{N}_]+`)
	spacesRe  = regexp.MustCompile(`\s+`)
	phraseCut = regexp.MustCompile(`[|\n\r]`)
)

// corePhrase is the searchable heart of a post: the title (or, with no
// usable title, the description) without links, hashtags, mentions, emoji
// and punctuation, at most webPhraseWords words. It returns "" when fewer
// than three words remain: a search on that would find unrelated news.
func corePhrase(title, description string) string {
	for _, text := range []string{title, firstSentence(description)} {
		text = urlRe.ReplaceAllString(text, " ")
		text = hashRe.ReplaceAllString(text, " ")
		text = strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '\'' {
				return r
			}
			return ' '
		}, text)
		words := strings.Fields(spacesRe.ReplaceAllString(text, " "))
		if len(words) < 3 {
			continue
		}
		if len(words) > webPhraseWords {
			words = words[:webPhraseWords]
		}
		return strings.Join(words, " ")
	}
	return ""
}

func firstSentence(s string) string {
	s = phraseCut.Split(strings.TrimSpace(s), 2)[0]
	for i, r := range s {
		if (r == '.' || r == '!' || r == '?') && i > 20 {
			return s[:i]
		}
	}
	return s
}

// startWebProvenance searches the web for the clip and returns a channel
// that receives exactly one outcome, or nil when the search is off or the
// clip has no usable words.
func (p *Pipeline) startWebProvenance(ctx context.Context, c *video.Clip) <-chan webOutcome {
	if p.WebSearch == nil || !p.WebSearch.Enabled() {
		return nil
	}
	phrase := corePhrase(c.Title, c.Description)
	if phrase == "" {
		return nil
	}
	ch := make(chan webOutcome, 1)
	go func() {
		ch <- p.webProvenance(ctx, c, phrase)
	}()
	return ch
}

// webProvenance runs one search and one classification. Any failure yields
// an empty outcome: this signal is optional and never degrades a result.
func (p *Pipeline) webProvenance(ctx context.Context, c *video.Clip, phrase string) webOutcome {
	out := webOutcome{Phrase: phrase}
	sb, cb := p.WebSearchDeadline, p.WebClassifyDeadline
	if sb <= 0 {
		sb = webSearchBudget
	}
	if cb <= 0 {
		cb = webClassifyBudget
	}
	sctx, cancel := withBudget(ctx, sb)
	results, err := p.WebSearch.Search(sctx, phrase+" video AI hoaks")
	cancel()
	if err != nil || len(results) == 0 {
		return out
	}
	// Only pages on the video's own subject reach the model.
	var onTopic []grounding.Result
	for _, r := range results {
		if grounding.SameTopic(phrase, r.Title+" "+r.Snippet) {
			onTopic = append(onTopic, r)
		}
	}
	if len(onTopic) == 0 {
		return out
	}
	evidence := make([]llm.Evidence, len(onTopic))
	for i, r := range onTopic {
		evidence[i] = llm.Evidence{Title: r.Title, Publisher: r.Publisher, Snippet: r.Snippet, Rating: r.Rating}
	}
	subject := "Judul: " + c.Title + " | Keterangan: " + truncateRunes(c.Description, 300) + " | Akun: " + c.Uploader
	cctx, cancel := withBudget(ctx, cb)
	says, err := p.LLM.ClassifyWeb(cctx, subject, evidence)
	cancel()
	if err != nil {
		return out
	}
	out.Judged = map[string]string{}
	for i, r := range onTopic {
		if i < len(says) {
			out.Judged[r.URL] = says[i]
		}
	}
	for i, r := range onTopic {
		if len(out.Evidence) >= webMaxEvidence {
			break
		}
		if i >= len(says) || !p.webGrounder().HostAllowed(r.URL) {
			continue
		}
		switch says[i] {
		case llm.WebAI, llm.WebReal:
		case llm.WebHoax:
			// Only a fact-check article may call something a hoax.
			if !grounding.IsFactCheck(r.Title, r.URL) {
				continue
			}
		default:
			continue
		}
		out.Evidence = append(out.Evidence, models.WebEvidence{
			Title: r.Title, URL: r.URL, Publisher: r.Publisher, Says: says[i],
		})
	}
	return out
}

func (p *Pipeline) webGrounder() *grounding.Grounder {
	if p.WebGrounder != nil {
		return p.WebGrounder
	}
	return p.Grounder
}

// awaitWeb waits for the search, at most until ctx ends.
func awaitWeb(ctx context.Context, ch <-chan webOutcome) webOutcome {
	if ch == nil {
		return webOutcome{}
	}
	select {
	case w := <-ch:
		return w
	case <-ctx.Done():
		return webOutcome{}
	}
}

// addWebHoax adds the fact-checks that call the video a hoax to the fact
// pillar as CONTRADICTS, and recomputes its status. claim names the pillar
// when it had none.
func addWebHoax(fact models.FactPillar, w webOutcome, claim string) models.FactPillar {
	hoax := w.says(llm.WebHoax)
	if len(hoax) == 0 {
		return fact
	}
	if !fact.Available {
		fact = models.FactPillar{Available: true, Claim: claim, Status: models.FactUnverified, Sources: []models.Source{}}
	}
	seen := map[string]bool{}
	for _, s := range fact.Sources {
		seen[s.URL] = true
	}
	for _, e := range hoax {
		if seen[e.URL] {
			continue
		}
		fact.Sources = append(fact.Sources, models.Source{
			Title: e.Title, Publisher: e.Publisher, URL: e.URL, Match: models.FactContradicts,
		})
	}
	fact.Status = statusFromSources(fact.Sources)
	return fact
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// onSubject keeps the evidence that is about this video once the model has
// read it: a page calling it a hoax always stays, any other page must share
// the video's subject (claim, title, summary) in its own title. At most
// webMaxEvidence.
func onSubject(ev []models.WebEvidence, topic string) []models.WebEvidence {
	out := make([]models.WebEvidence, 0, len(ev))
	for _, e := range ev {
		if len(out) >= webMaxEvidence {
			break
		}
		if e.Says == llm.WebHoax || grounding.SameTopic(topic, e.Title) {
			out = append(out, e)
		}
	}
	return out
}
