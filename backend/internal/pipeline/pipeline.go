// Package pipeline orchestrates one audit. Pillars run concurrently under a
// shared deadline, publish their own progress, and degrade honestly when a
// dependency fails. No pillar failure ever becomes an HTTP error.
package pipeline

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/amba/amba/backend/internal/forensics"
	"github.com/amba/amba/backend/internal/grounding"
	"github.com/amba/amba/backend/internal/imaging"
	"github.com/amba/amba/backend/internal/linkcheck"
	"github.com/amba/amba/backend/internal/llm"
	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/pagefetch"
	"github.com/amba/amba/backend/internal/retrieval"
	"github.com/amba/amba/backend/internal/scoring"
	"github.com/amba/amba/backend/internal/security"
	"github.com/amba/amba/backend/internal/tactics"
)

// Request is one audit input. Exactly one of URL, Text or Image is set.
type Request struct {
	InputType models.InputType
	URL       string
	Text      string
	Image     []byte
	ImageMIME string
	OCRText   string
}

// Event is a real progress update. It is only emitted when work actually
// starts or finishes; nothing here is a timer pretending to be progress.
type Event struct {
	Kind   string // "stage" | "partial"
	Pillar string // ela | ai | fact | link
	Status string // running | done | skipped | failed
	Data   any
}

// Emitter receives events. Implementations must be safe for concurrent use:
// pillars publish from their own goroutines.
type Emitter func(Event)

// Result is what the pipeline produces before signing.
type Result struct {
	Pillars  models.Pillars
	Tactics  []models.Tactic
	Degraded bool
	// AIAnomalies are the model's region boxes, used only for the
	// ELA_CORROBORATED floor. ELAHotspots are the server's own hot cells; the
	// two lists are kept separate so a hotspot can never corroborate itself.
	AIAnomalies []scoring.Box
	ELAHotspots []scoring.Box
	// AIReliability is the weight the AI pillar carries (0.7 image, 0.5 text).
	AIReliability float64
	// InjectionSuspected is set by the input scanner or the model, and raises
	// the INJECTION_GUARD floor.
	InjectionSuspected bool
	// Cached marks a result served from the short-lived verdict cache.
	Cached bool
}

// Stats counts audits without recording any content.
type Stats struct {
	mu       sync.Mutex
	Audits   int64
	Degraded int64
	CacheHit int64
	Pillars  map[string]int64
}

func (s *Stats) inc(field *int64) {
	s.mu.Lock()
	*field++
	s.mu.Unlock()
}

func (s *Stats) pillar(name string) {
	s.mu.Lock()
	if s.Pillars == nil {
		s.Pillars = make(map[string]int64)
	}
	s.Pillars[name]++
	s.mu.Unlock()
}

// Snapshot returns a copy safe to serialise.
func (s *Stats) Snapshot() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	pillars := make(map[string]int64, len(s.Pillars))
	for k, v := range s.Pillars {
		pillars[k] = v
	}
	return map[string]any{
		"audits": s.Audits, "degraded": s.Degraded,
		"cache_hits": s.CacheHit, "pillars": pillars,
	}
}

// FactSearcher is the retrieval seam. It lets tests provide a deterministic
// feed without a network, and keeps the pipeline independent of the provider.
type FactSearcher interface {
	Search(ctx context.Context, query string) ([]retrieval.Item, error)
}

// ReputationChecker is the optional Safe Browsing seam.
type ReputationChecker interface {
	Check(ctx context.Context, target string) (bool, error)
}

// Pipeline runs audits. All fields are safe for concurrent use.
type Pipeline struct {
	LLM        *llm.Client
	Grounder   *grounding.Grounder
	Searcher   FactSearcher
	Reputation ReputationChecker
	CacheTTL   time.Duration
	// Deadline bounds the whole audit. Sub-deadlines are derived from it so a
	// slow model cannot eat the retrieval budget.
	Deadline time.Duration
	Stats    Stats

	mu    sync.Mutex
	cache map[string]*list.Element
	order *list.List
}

type cacheEntry struct {
	key    string
	result Result
	expiry time.Time
}

// MaxCacheEntries bounds memory: a live audit stores only a verdict summary.
const MaxCacheEntries = 500

// New builds a pipeline.
func New(client *llm.Client, g *grounding.Grounder, searcher FactSearcher, cacheTTL time.Duration) *Pipeline {
	if cacheTTL <= 0 {
		cacheTTL = 10 * time.Minute
	}
	return &Pipeline{
		LLM:      client,
		Grounder: g,
		Searcher: searcher,
		CacheTTL: cacheTTL,
		Deadline: 12 * time.Second,
		cache:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// CacheKey hashes the input so no raw content is ever used as a map key.
func CacheKey(req Request) string {
	h := sha256.New()
	h.Write([]byte(req.InputType))
	h.Write([]byte{0})
	switch req.InputType {
	case models.InputURL:
		h.Write([]byte(req.URL))
	case models.InputText:
		h.Write([]byte(req.Text))
	case models.InputFile:
		h.Write(req.Image)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Run executes every pillar and reports progress through emit. It never
// returns an error for a pillar failure.
func (p *Pipeline) Run(ctx context.Context, req Request, emit Emitter) Result {
	if emit == nil {
		emit = func(Event) {}
	}
	p.Stats.inc(&p.Stats.Audits)

	key := CacheKey(req)
	if res, ok := p.fromCache(key); ok {
		p.Stats.inc(&p.Stats.CacheHit)
		res.Cached = true
		replay(emit, req, res)
		return res
	}

	ctx, cancel := context.WithTimeout(ctx, p.Deadline)
	defer cancel()

	var (
		mu           sync.Mutex
		elaPillar    models.ELAPillar
		elaOK        bool
		elaFailed    bool
		aiPillar     models.AIPillar
		aiOK         bool
		aiFailed     bool
		factPillar   models.FactPillar
		factOK       bool
		factFailed   bool
		linkPillar   models.LinkPillar
		linkOK       bool
		linkFailed   bool
		foundTactics []models.Tactic
		aiBoxes      []scoring.Box
		elaHotspots  []scoring.Box
		content      = req.Text
		ocrText      = req.OCRText
		findings     llm.Findings
		page         pagefetch.Page
		pageErr      error
		aiRel        = 0.5
	)

	// The page fetch is a prerequisite for a live URL audit: the model needs
	// the actual text, and the link pillar wants the final URL. It is not a
	// pillar of its own, so it publishes under "link" while it runs. Mock mode
	// skips the network so previews and e2e stay offline and deterministic.
	fetchCtx, fetchCancel := context.WithTimeout(ctx, minDuration(p.Deadline/4, 4000*time.Millisecond))
	fetchDone := make(chan struct{})
	go func() {
		defer close(fetchDone)
		if req.InputType != models.InputURL {
			return
		}
		emit(Event{Kind: "stage", Pillar: "link", Status: "running"})
		if p.LLM.Mock {
			return
		}
		page, pageErr = pagefetch.Fetch(fetchCtx, security.NewSafeClient(4000*time.Millisecond, 3), req.URL, 4000*time.Millisecond)
	}()
	defer fetchCancel()

	// aiReady closes when the model goroutine has stored its findings or
	// failed. The fact and link pillars wait on it instead of sleeping.
	aiReady := make(chan struct{})

	g, gctx := errgroup.WithContext(ctx)

	// ELA: only for images.
	g.Go(func() error {
		if req.InputType != models.InputFile || len(req.Image) == 0 {
			return nil
		}
		emit(Event{Kind: "stage", Pillar: "ela", Status: "running"})
		p.Stats.pillar("ela")
		pctx, cancel := context.WithTimeout(gctx, minDuration(p.Deadline/6, 900*time.Millisecond))
		defer cancel()

		type elaResult struct {
			res forensics.Result
			err error
		}
		done := make(chan elaResult, 1)
		go func() {
			res, err := forensics.Analyze(req.Image)
			done <- elaResult{res, err}
		}()

		var out elaResult
		select {
		case out = <-done:
		case <-pctx.Done():
			mu.Lock()
			elaFailed = true
			mu.Unlock()
			emit(Event{Kind: "stage", Pillar: "ela", Status: "failed"})
			return nil
		}
		if out.err != nil {
			mu.Lock()
			elaFailed = true
			mu.Unlock()
			emit(Event{Kind: "stage", Pillar: "ela", Status: "failed"})
			return nil
		}

		// Hotspots become visible boxes even when the model never answers:
		// they are evidence from our own analysis, labelled as such. The same
		// boxes are kept separately for the ELA_CORROBORATED floor.
		hotspots := out.res.Hotspots(1.5)
		anomalies := make([]models.Anomaly, 0, 5)
		boxes := make([]scoring.Box, 0, len(hotspots))
		for i, box := range hotspots {
			boxes = append(boxes, scoring.Box(box))
			if i < 5 {
				anomalies = append(anomalies, models.Anomaly{
					BoundingBox: box,
					Reason:      "Area dengan perbedaan kompresi tinggi",
				})
			}
		}
		pillar := models.ELAPillar{
			Available:   true,
			Score:       out.res.Score,
			Reliability: out.res.Reliability,
			Anomalies:   anomalies,
		}
		if out.res.Note != "" {
			note := out.res.Note
			pillar.ReliabilityNote = &note
		}
		mu.Lock()
		elaPillar = pillar
		elaHotspots = boxes
		elaOK = true
		mu.Unlock()
		emit(Event{Kind: "partial", Pillar: "ela", Status: "done", Data: pillar})
		emit(Event{Kind: "stage", Pillar: "ela", Status: "done"})
		return nil
	})

	// LLM: produces the AI pillar, the claim, OCR text, tactics and the
	// search query. It waits for the page fetch on URL audits.
	g.Go(func() error {
		defer close(aiReady)
		emit(Event{Kind: "stage", Pillar: "ai", Status: "running"})
		p.Stats.pillar("ai")

		if req.InputType == models.InputURL {
			select {
			case <-fetchDone:
			case <-gctx.Done():
				mu.Lock()
				aiFailed = true
				mu.Unlock()
				emit(Event{Kind: "stage", Pillar: "ai", Status: "failed"})
				return nil
			}
			// The URL itself is always evidence: lookalike hosts, punycode and
			// scam keywords are visible in the address even when the page
			// refuses to load. Page text is an enrichment, not a prerequisite.
			var b strings.Builder
			b.WriteString("URL: ")
			b.WriteString(req.URL)
			b.WriteString("\n")
			if pageErr == nil {
				if page.Title != "" {
					b.WriteString("Judul halaman: ")
					b.WriteString(page.Title)
					b.WriteString("\n")
				}
				if page.Description != "" {
					b.WriteString("Deskripsi: ")
					b.WriteString(page.Description)
					b.WriteString("\n")
				}
				if page.Text != "" {
					b.WriteString("Isi halaman:\n")
					b.WriteString(page.Text)
				}
			} else {
				b.WriteString("Catatan: isi halaman tidak dapat dibaca otomatis. Nilai hanya dari alamat dan pola URL.")
			}
			content = b.String()
		}

		// Vision is the slowest dependency: give it the larger share of the
		// budget, but leave room for signing and the response.
		llmBudget := minDuration(p.Deadline-2*time.Second, 11*time.Second)
		if llmBudget < 3*time.Second {
			llmBudget = 3 * time.Second
		}
		llmCtx, cancel := context.WithTimeout(gctx, llmBudget)
		defer cancel()

		// ELA already ran on the original bytes. The model gets a downscaled
		// JPEG: vision latency and cost scale with pixel count.
		imageForModel, mimeForModel := req.Image, req.ImageMIME
		if len(req.Image) > 0 {
			if prepared, err := imaging.Prepare(req.Image); err == nil {
				imageForModel, mimeForModel = prepared, "image/jpeg"
			}
		}
		f, err := p.LLM.Analyze(llmCtx, llm.Request{
			InputType: string(req.InputType),
			Content:   content,
			Image:     imageForModel,
			ImageMIME: mimeForModel,
			OCRText:   ocrText,
		})
		if err != nil {
			mu.Lock()
			aiFailed = true
			mu.Unlock()
			emit(Event{Kind: "stage", Pillar: "ai", Status: "failed"})
			return nil
		}

		pillar := models.AIPillar{
			Available:  true,
			Score:      f.AIScore,
			Indicators: f.AIIndicators,
			Caveat:     "Indikatif, bukan bukti pasti",
		}
		rel := 0.5
		if req.InputType == models.InputFile {
			rel = 0.7
		}

		mu.Lock()
		findings = f
		aiPillar = pillar
		aiRel = rel
		aiOK = true
		ocrText = f.OCRText
		for _, a := range f.TextAnomalies {
			aiBoxes = append(aiBoxes, scoring.Box(a.BoundingBox))
		}
		mu.Unlock()

		emit(Event{Kind: "partial", Pillar: "ai", Status: "done", Data: pillar})
		emit(Event{Kind: "stage", Pillar: "ai", Status: "done"})

		// Tactics are informational and verified against the user's own text.
		source := content
		sourceKind := "text"
		if req.InputType == models.InputFile && f.OCRText != "" {
			source = f.OCRText
			sourceKind = "ocr"
		}
		candidates := make([]models.Tactic, 0, len(f.Tactics))
		for _, t := range f.Tactics {
			candidates = append(candidates, models.Tactic{Code: t.Code, Phrase: t.Phrase, Source: sourceKind})
		}
		verified := tactics.Verify(candidates, source)
		mu.Lock()
		foundTactics = verified
		mu.Unlock()
		return nil
	})

	// Fact: retrieval after the model supplies a query. A heuristic query is
	// the fallback when the model is unavailable, so a dead gateway still
	// produces a real check instead of an empty panel.
	g.Go(func() error {
		emit(Event{Kind: "stage", Pillar: "fact", Status: "running"})
		p.Stats.pillar("fact")

		// Wait for the model's claim/query but not past its own budget. The
		// channel closes on success and on failure, so a dead gateway costs
		// one bounded wait, not a fixed sleep.
		select {
		case <-aiReady:
		case <-time.After(minDuration(p.Deadline/2, 12*time.Second)):
		case <-gctx.Done():
		}

		mu.Lock()
		f := findings
		mu.Unlock()

		query := f.SearchQuery
		if strings.TrimSpace(query) == "" {
			query = retrieval.HeuristicQuery(primaryText(req, content))
		}
		claim := f.Claim
		if claim == "" {
			// The claim shown to the user is the user's own sentence, not the
			// search query (which is truncated and stripped of punctuation).
			claim = truncate(strings.TrimSpace(primaryText(req, content)), 200)
		}

		if p.Searcher == nil {
			// Mock mode has no retrieval by design (previews and e2e must not
			// touch the network). Reporting an available UNVERIFIED claim is
			// the honest answer: the claim exists, no source was checked.
			if p.LLM.Mock {
				if claim == "" {
					claim = truncate(strings.TrimSpace(primaryText(req, content)), 200)
				}
				pillar := models.FactPillar{
					Available: true, Claim: truncate(claim, 200),
					Status: models.FactUnverified, Sources: []models.Source{},
				}
				mu.Lock()
				factPillar = pillar
				factOK = true
				mu.Unlock()
				emit(Event{Kind: "partial", Pillar: "fact", Status: "done", Data: pillar})
				emit(Event{Kind: "stage", Pillar: "fact", Status: "done"})
				return nil
			}
			mu.Lock()
			factPillar = models.FactPillar{Available: false, Status: models.FactUnverified}
			mu.Unlock()
			emit(Event{Kind: "stage", Pillar: "fact", Status: "skipped"})
			return nil
		}
		if strings.TrimSpace(query) == "" {
			mu.Lock()
			factPillar = models.FactPillar{Available: false, Status: models.FactUnverified}
			mu.Unlock()
			emit(Event{Kind: "stage", Pillar: "fact", Status: "skipped"})
			return nil
		}

		searchCtx, cancel := context.WithTimeout(gctx, minDuration(p.Deadline/3, 4000*time.Millisecond))
		defer cancel()
		items, err := p.Searcher.Search(searchCtx, query)
		// When the model's query finds nothing (long paraphrases are a common
		// miss), retry once with the user's own words before reporting an
		// unverified claim.
		if err == nil && len(items) == 0 {
			if fallback := retrieval.HeuristicQuery(primaryText(req, content)); fallback != "" && fallback != query {
				query = fallback
				items, err = p.Searcher.Search(searchCtx, query)
			}
		}
		if err != nil {
			mu.Lock()
			factPillar = models.FactPillar{Available: true, Claim: claim, Status: models.FactUnverified}
			factOK = true
			factFailed = true
			mu.Unlock()
			emit(Event{Kind: "partial", Pillar: "fact", Status: "done", Data: factPillar})
			emit(Event{Kind: "stage", Pillar: "fact", Status: "failed"})
			return nil
		}

		items = retrieval.Annotate(claim, items)
		sources := make([]models.Source, 0, len(items))
		for _, item := range items {
			if !p.Grounder.Verified(item.URL, nil) {
				continue
			}
			sources = append(sources, models.Source{
				Title: item.Title, Publisher: item.Publisher, URL: item.URL, Match: item.Match,
			})
			if len(sources) == 3 {
				break
			}
		}
		pillar := models.FactPillar{
			Available: true,
			Claim:     truncate(claim, 200),
			Status:    retrieval.Aggregate(items),
			Sources:   sources,
		}
		if len(sources) == 0 {
			pillar.Status = models.FactUnverified
		}
		mu.Lock()
		factPillar = pillar
		factOK = true
		mu.Unlock()
		emit(Event{Kind: "partial", Pillar: "fact", Status: "done", Data: pillar})
		emit(Event{Kind: "stage", Pillar: "fact", Status: "done"})
		return nil
	})

	// Link: deterministic heuristics plus redirect unrolling. For FILE input
	// it runs after OCR text arrives, because that is where a URL can be.
	g.Go(func() error {
		if req.InputType == models.InputFile {
			select {
			case <-aiReady:
			case <-time.After(minDuration(p.Deadline/2, 8000*time.Millisecond)):
			case <-gctx.Done():
			}
		}

		target := req.URL
		switch req.InputType {
		case models.InputText:
			target = firstURL(req.Text)
		case models.InputFile:
			mu.Lock()
			target = firstURL(findings.OCRText)
			mu.Unlock()
		}
		if target == "" {
			emit(Event{Kind: "stage", Pillar: "link", Status: "skipped"})
			return nil
		}
		// A URL audit already published "running" before the fetch.
		if req.InputType != models.InputURL {
			emit(Event{Kind: "stage", Pillar: "link", Status: "running"})
		}
		p.Stats.pillar("link")

		select {
		case <-fetchDone:
		default:
			// The fetch is bounded; wait for it so the final URL is known.
			select {
			case <-fetchDone:
			case <-gctx.Done():
			}
		}

		final := target
		chain := []string{target}
		if req.InputType == models.InputURL && pageErr == nil && page.FinalURL != "" {
			final = page.FinalURL
			if final != target {
				chain = []string{target, final}
			}
		} else {
			unrollCtx, cancel := context.WithTimeout(gctx, minDuration(p.Deadline/5, 2500*time.Millisecond))
			defer cancel()
			if hops, last, err := linkcheck.Unroll(unrollCtx, security.NewSafeClient(2500*time.Millisecond, 3), target, 3); err == nil && len(hops) > 1 {
				chain = hops
				final = last
			}
		}

		analysis := linkcheck.Analyze(final)
		if p.Reputation != nil {
			repCtx, repCancel := context.WithTimeout(gctx, 2500*time.Millisecond)
			hit, err := p.Reputation.Check(repCtx, final)
			repCancel()
			if err == nil && hit {
				if analysis.Score < 90 {
					analysis.Score = 90
				}
				analysis.Findings = append(analysis.Findings, linkcheck.Finding{Code: linkcheck.ReasonReputation, Weight: 90})
			}
		}
		reasons := make([]models.LinkReason, 0, len(analysis.Findings))
		for _, f := range analysis.Findings {
			reasons = append(reasons, models.LinkReason{Code: f.Code, Text: reasonText(f.Code)})
		}
		pillar := models.LinkPillar{
			Available:     true,
			Score:         analysis.Score,
			FinalURL:      final,
			RedirectChain: chain,
			Reasons:       reasons,
		}
		mu.Lock()
		linkPillar = pillar
		linkOK = true
		mu.Unlock()
		emit(Event{Kind: "partial", Pillar: "link", Status: "done", Data: pillar})
		emit(Event{Kind: "stage", Pillar: "link", Status: "done"})
		return nil
	})

	_ = g.Wait()
	<-fetchDone

	// The model's own anomaly boxes are the regions the ELA viewer should
	// highlight, alongside the server's hotspot grid.
	for _, a := range findings.TextAnomalies {
		elaPillar.Anomalies = append(elaPillar.Anomalies, models.Anomaly{
			BoundingBox:   a.BoundingBox,
			ExtractedText: a.ExtractedText,
			Reason:        a.Reason,
		})
	}

	mu.Lock()
	injectionText := primaryText(req, content)
	if page.Text != "" {
		injectionText += " " + page.Text
	}
	res := Result{
		Pillars:            models.Pillars{ELA: elaPillar, AI: aiPillar, Fact: factPillar, Link: linkPillar},
		Tactics:            foundTactics,
		AIAnomalies:        aiBoxes,
		ELAHotspots:        elaHotspots,
		AIReliability:      aiRel,
		InjectionSuspected: findings.InjectionSuspected || InjectionDetected(injectionText),
	}
	expected := expectedPillars(req)
	if p.Searcher == nil && !p.LLM.Mock {
		expected["fact"] = false
	}
	res.Degraded = (expected["ela"] && !elaOK) ||
		(expected["ai"] && !aiOK) ||
		(expected["fact"] && !factOK) ||
		(expected["link"] && !linkOK) ||
		elaFailed || aiFailed || linkFailed
	// A retrieval failure still yields an honest UNVERIFIED fact pillar, but
	// the dossier must say that the check did not complete. A page fetch that
	// was refused is the same kind of honest gap for a URL audit.
	if factFailed || (req.InputType == models.InputURL && pageErr != nil) {
		res.Degraded = true
	}
	mu.Unlock()

	if !res.Degraded {
		p.toCache(key, res)
	}
	if res.Degraded {
		p.Stats.inc(&p.Stats.Degraded)
	}
	return res
}

func expectedPillars(req Request) map[string]bool {
	switch req.InputType {
	case models.InputURL:
		return map[string]bool{"ela": false, "ai": true, "fact": true, "link": true}
	case models.InputFile:
		return map[string]bool{"ela": true, "ai": true, "fact": true, "link": false}
	default:
		return map[string]bool{"ela": false, "ai": true, "fact": true, "link": false}
	}
}

// primaryText is what the user actually submitted, without the URL preamble
// the model sees. Injection detection and the fallback query use it.
func primaryText(req Request, content string) string {
	switch req.InputType {
	case models.InputText:
		return req.Text
	case models.InputURL:
		return req.URL
	default:
		if req.OCRText != "" {
			return req.OCRText
		}
		return content
	}
}

var injectionPatterns = regexp.MustCompile(`(?i)(ignore (all |any )?(previous|prior|above) (instructions|prompts)|abaikan (semua )?(instruksi|perintah) (sebelumnya|di atas)|system prompt|set (the )?verdict|tetapkan (vonis|skor)|you are now|kamu sekarang adalah|developer mode|jailbreak)`)

// InjectionDetected scans the user's own text for prompt-injection markers.
// It is a signal, never a verdict: the floor it raises is 40, not TAMPERED.
func InjectionDetected(text string) bool {
	return injectionPatterns.MatchString(text)
}

// Verdict computes the score from a Result.
func Verdict(res Result) models.Verdict {
	in := scoring.Input{
		ELA:                scoring.Signal{Available: res.Pillars.ELA.Available, Score: res.Pillars.ELA.Score, Reliability: res.Pillars.ELA.Reliability},
		AI:                 scoring.Signal{Available: res.Pillars.AI.Available, Score: res.Pillars.AI.Score, Reliability: res.AIReliability},
		Fact:               scoring.Signal{Available: res.Pillars.Fact.Available, Score: factScore(res.Pillars.Fact.Status), Reliability: 1.0},
		Link:               scoring.Signal{Available: res.Pillars.Link.Available, Score: res.Pillars.Link.Score, Reliability: 1.0},
		InjectionSuspected: res.InjectionSuspected,
		ELAHotspots:        res.ELAHotspots,
		AIAnomalies:        res.AIAnomalies,
	}
	if res.Pillars.Fact.Available {
		for _, s := range res.Pillars.Fact.Sources {
			if s.Match == models.FactContradicts {
				in.FactContradicted = true
				break
			}
		}
	}
	return scoring.Compute(in)
}

func factScore(s models.FactStatus) int {
	switch s {
	case models.FactContradicts:
		return 100
	case models.FactUnverified:
		return 40
	default:
		return 0
	}
}

func reasonText(code string) string {
	switch code {
	case linkcheck.ReasonLookalikeBrand:
		return "Domain meniru nama layanan populer"
	case linkcheck.ReasonSuspiciousTLD:
		return "Akhiran domain sering dipakai penipuan"
	case linkcheck.ReasonIPHost:
		return "Alamat berupa nomor IP, bukan nama situs"
	case linkcheck.ReasonPunycode:
		return "Nama domain memakai karakter khusus"
	case linkcheck.ReasonManySubdomains:
		return "Terlalu banyak subdomain"
	case linkcheck.ReasonCredentials:
		return "Tautan memuat nama pengguna dan sandi"
	case linkcheck.ReasonURLShortener:
		return "Tautan pendek menyembunyikan alamat asli"
	case linkcheck.ReasonExcessiveLength:
		return "Tautan sangat panjang"
	case linkcheck.ReasonAtSymbol:
		return "Tautan memuat karakter @ yang menyesatkan"
	case linkcheck.ReasonNoHTTPS:
		return "Tautan tidak memakai koneksi aman"
	case linkcheck.ReasonPortNonStandard:
		return "Tautan memakai port tidak biasa"
	case linkcheck.ReasonSuspiciousPath:
		return "Alamat memuat kata yang sering dipakai penipuan"
	case linkcheck.ReasonReputation:
		return "Alamat ini tercatat sebagai situs berbahaya"
	default:
		return "Tanda mencurigakan pada tautan"
	}
}

// replay publishes a cached result in the same shape as a live run, so the
// SSE contract does not change just because the answer was fast.
func replay(emit Emitter, req Request, res Result) {
	for _, pillar := range orderedPillars(req) {
		if !pillarAvailable(res, pillar) {
			emit(Event{Kind: "stage", Pillar: pillar, Status: "skipped"})
			continue
		}
		emit(Event{Kind: "stage", Pillar: pillar, Status: "running"})
		emit(Event{Kind: "partial", Pillar: pillar, Status: "done", Data: pillarPayload(res, pillar)})
		emit(Event{Kind: "stage", Pillar: pillar, Status: "done"})
	}
}

// OrderedPillars lists the pillars an input type can produce.
func OrderedPillars(req Request) []string { return orderedPillars(req) }

func orderedPillars(req Request) []string {
	switch req.InputType {
	case models.InputURL:
		return []string{"link", "ai", "fact"}
	case models.InputFile:
		return []string{"ela", "ai", "fact", "link"}
	default:
		return []string{"ai", "fact", "link"}
	}
}

func pillarAvailable(res Result, pillar string) bool {
	switch pillar {
	case "ela":
		return res.Pillars.ELA.Available
	case "ai":
		return res.Pillars.AI.Available
	case "fact":
		return res.Pillars.Fact.Available
	case "link":
		return res.Pillars.Link.Available
	}
	return false
}

// PillarPayload exposes a pillar's payload for the SSE partial event.
func PillarPayload(res Result, pillar string) any { return pillarPayload(res, pillar) }

func pillarPayload(res Result, pillar string) any {
	switch pillar {
	case "ela":
		return res.Pillars.ELA
	case "ai":
		return res.Pillars.AI
	case "fact":
		return res.Pillars.Fact
	case "link":
		return res.Pillars.Link
	}
	return nil
}

func (p *Pipeline) fromCache(key string) (Result, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	el, ok := p.cache[key]
	if !ok {
		return Result{}, false
	}
	entry := el.Value.(*cacheEntry)
	if time.Now().After(entry.expiry) {
		p.order.Remove(el)
		delete(p.cache, key)
		return Result{}, false
	}
	p.order.MoveToFront(el)
	return entry.result, true
}

func (p *Pipeline) toCache(key string, res Result) {
	p.mu.Lock()
	defer p.mu.Unlock()
	entry := &cacheEntry{key: key, result: res, expiry: time.Now().Add(p.CacheTTL)}
	if el, ok := p.cache[key]; ok {
		el.Value = entry
		p.order.MoveToFront(el)
		return
	}
	p.cache[key] = p.order.PushFront(entry)
	for p.order.Len() > MaxCacheEntries {
		oldest := p.order.Back()
		if oldest == nil {
			break
		}
		p.order.Remove(oldest)
		delete(p.cache, oldest.Value.(*cacheEntry).key)
	}
}

func firstURL(text string) string {
	for _, field := range strings.Fields(text) {
		field = strings.Trim(field, "\"'()[]<>,.;!?")
		lower := strings.ToLower(field)
		if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
			return field
		}
	}
	return ""
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
