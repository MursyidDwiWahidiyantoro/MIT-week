// Package pipeline orchestrates one audit: it runs the pillars concurrently
// under a shared deadline, degrades gracefully when one fails, and caches the
// final verdict for a short TTL.
package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/amba/amba/backend/internal/forensics"
	"github.com/amba/amba/backend/internal/grounding"
	"github.com/amba/amba/backend/internal/linkcheck"
	"github.com/amba/amba/backend/internal/llm"
	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/scoring"
	"github.com/amba/amba/backend/internal/tactics"
)

// Request is one audit input. Exactly one of URL, Text or Image is set.
type Request struct {
	InputType models.InputType
	URL       string
	Text      string
	Image     []byte
	OCRText   string
}

// Result is what the pipeline produces before signing.
type Result struct {
	Pillars  models.Pillars
	Tactics  []models.Tactic
	Degraded bool
}

// PillarResult carries the outcome of one pillar, including whether it ran.
type pillarOutcome struct {
	available bool
	signal    scoring.Signal
	failure   string
}

// Pipeline runs audits. All fields are safe for concurrent use.
type Pipeline struct {
	LLM      *llm.Client
	Grounder *grounding.Grounder
	CacheTTL time.Duration
	// SubDeadline bounds each pillar so one slow dependency cannot consume the
	// whole request budget.
	SubDeadline time.Duration

	mu    sync.Mutex
	cache map[string]cached
}

type cached struct {
	result  Result
	expires time.Time
}

// New builds a pipeline.
func New(client *llm.Client, g *grounding.Grounder, cacheTTL time.Duration) *Pipeline {
	if cacheTTL <= 0 {
		cacheTTL = 10 * time.Minute
	}
	return &Pipeline{
		LLM:         client,
		Grounder:    g,
		CacheTTL:    cacheTTL,
		SubDeadline: 4 * time.Second,
		cache:       make(map[string]cached),
	}
}

// CacheKey hashes the input so no raw content is ever used as a map key that
// could be logged or dumped.
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

// Run executes every pillar concurrently. It never returns an error for a
// pillar failure: a dead dependency degrades the result instead.
func (p *Pipeline) Run(ctx context.Context, req Request) Result {
	key := CacheKey(req)
	if res, ok := p.fromCache(key); ok {
		return res
	}

	var (
		elaOut  = pillarOutcome{}
		aiOut   = pillarOutcome{}
		factOut = pillarOutcome{}
		linkOut = pillarOutcome{}

		elaPillar    models.ELAPillar
		aiPillar     models.AIPillar
		factPillar   models.FactPillar
		linkPillar   models.LinkPillar
		foundTactics []models.Tactic
		findings     llm.Findings
	)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if req.InputType != models.InputFile || len(req.Image) == 0 {
			return nil
		}
		pctx, cancel := context.WithTimeout(gctx, p.SubDeadline)
		defer cancel()
		done := make(chan struct{})
		var res forensics.Result
		var err error
		go func() { res, err = forensics.Analyze(req.Image); close(done) }()
		select {
		case <-done:
		case <-pctx.Done():
			elaOut.failure = "timeout"
			return nil
		}
		if err != nil {
			elaOut.failure = "decode"
			return nil
		}
		elaPillar = models.ELAPillar{
			Available:   true,
			Score:       res.Score,
			Reliability: res.Reliability,
			Anomalies:   []models.Anomaly{},
		}
		if res.Note != "" {
			note := res.Note
			elaPillar.ReliabilityNote = &note
		}
		elaOut.available = true
		elaOut.signal = scoring.Signal{Available: true, Score: res.Score, Reliability: res.Reliability}
		return nil
	})

	g.Go(func() error {
		pctx, cancel := context.WithTimeout(gctx, p.SubDeadline)
		defer cancel()

		content := req.Text
		if req.InputType == models.InputURL {
			content = req.URL
		}
		f, err := p.LLM.Analyze(pctx, string(req.InputType), content)
		if err != nil {
			aiOut.failure = "unavailable"
			factOut.failure = "unavailable"
			return nil
		}
		findings = f

		aiPillar = models.AIPillar{
			Available:  true,
			Score:      f.AIScore,
			Indicators: f.AIIndicators,
			Caveat:     "Indikatif, bukan bukti pasti",
		}
		aiOut.available = true
		aiOut.signal = scoring.Signal{Available: true, Score: f.AIScore, Reliability: aiReliability(req.InputType)}

		// Fact pillar: only available when the model produced a checkable claim.
		//
		// The status is derived from the VERIFIED sources, never taken on the
		// model's word. A source that fails grounding is downgraded, and if no
		// verified source supports the claim's status the whole pillar falls
		// back to UNVERIFIED. Otherwise an invented citation could carry a
		// CONTRADICTS status into the score with nothing behind it.
		if f.Claim != "" {
			sources := make([]models.Source, 0, len(f.Sources))
			for _, s := range f.Sources {
				match := models.FactStatus(s.Match)
				if !p.Grounder.Verified(s.URL, nil) {
					match = models.FactUnverified
				}
				sources = append(sources, models.Source{
					Title: s.Title, Publisher: s.Publisher, URL: s.URL, Match: match,
				})
			}
			status := statusFromVerifiedSources(models.FactStatus(f.FactStatus), sources)
			factPillar = models.FactPillar{
				Available: true, Claim: f.Claim, Status: status, Sources: sources,
			}
			factOut.available = true
			factOut.signal = scoring.Signal{
				Available: true, Score: factScore(status), Reliability: 1.0,
			}
		}

		candidates := make([]models.Tactic, 0, len(f.Tactics))
		for _, t := range f.Tactics {
			candidates = append(candidates, models.Tactic{
				Code: t.Code, Phrase: t.Phrase, Source: "text",
			})
		}
		source := content
		if req.InputType == models.InputFile && f.OCRText != "" {
			source = f.OCRText
			for i := range candidates {
				candidates[i].Source = "ocr"
			}
		}
		foundTactics = tactics.Verify(candidates, source)
		return nil
	})

	g.Go(func() error {
		if req.InputType != models.InputURL {
			return nil
		}
		pctx, cancel := context.WithTimeout(gctx, p.SubDeadline)
		defer cancel()
		_ = pctx

		analysis := linkcheck.Analyze(req.URL)
		reasons := make([]models.LinkReason, 0, len(analysis.Findings))
		for _, f := range analysis.Findings {
			reasons = append(reasons, models.LinkReason{Code: f.Code, Text: reasonText(f.Code)})
		}
		linkPillar = models.LinkPillar{
			Available:     true,
			Score:         analysis.Score,
			FinalURL:      analysis.FinalURL,
			RedirectChain: []string{analysis.FinalURL},
			Reasons:       reasons,
		}
		linkOut.available = true
		linkOut.signal = scoring.Signal{Available: true, Score: analysis.Score, Reliability: 1.0}
		return nil
	})

	_ = g.Wait()

	res := Result{
		Pillars: models.Pillars{ELA: elaPillar, AI: aiPillar, Fact: factPillar, Link: linkPillar},
		Tactics: foundTactics,
	}
	res.Degraded = !elaOut.available && req.InputType == models.InputFile ||
		aiOut.failure != "" || factOut.failure != "" ||
		(req.InputType == models.InputURL && !linkOut.available)

	_ = findings
	p.toCache(key, res)
	return res
}

// Verdict computes the score from a Result. Kept separate so handlers can
// decide when to compute it.
func Verdict(res Result, injectionSuspected bool) models.Verdict {
	in := scoring.Input{
		ELA:                scoring.Signal{Available: res.Pillars.ELA.Available, Score: res.Pillars.ELA.Score, Reliability: res.Pillars.ELA.Reliability},
		AI:                 scoring.Signal{Available: res.Pillars.AI.Available, Score: res.Pillars.AI.Score, Reliability: aiReliabilityFromPillar(res)},
		Fact:               scoring.Signal{Available: res.Pillars.Fact.Available, Score: factScore(res.Pillars.Fact.Status), Reliability: 1.0},
		Link:               scoring.Signal{Available: res.Pillars.Link.Available, Score: res.Pillars.Link.Score, Reliability: 1.0},
		InjectionSuspected: injectionSuspected,
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

// statusFromVerifiedSources keeps the model's claim status only when at least
// one VERIFIED source backs it. Otherwise the pillar is UNVERIFIED: AMBA does
// not report a contradiction it cannot cite.
func statusFromVerifiedSources(claimed models.FactStatus, sources []models.Source) models.FactStatus {
	hasContradict := false
	hasSupport := false
	for _, s := range sources {
		switch s.Match {
		case models.FactContradicts:
			hasContradict = true
		case models.FactSupports:
			hasSupport = true
		}
	}

	switch claimed {
	case models.FactContradicts:
		if hasContradict {
			return models.FactContradicts
		}
		return models.FactUnverified
	case models.FactSupports:
		if hasSupport {
			return models.FactSupports
		}
		return models.FactUnverified
	default:
		return models.FactUnverified
	}
}

func aiReliability(t models.InputType) float64 {
	if t == models.InputFile {
		return 0.7
	}
	return 0.5
}

func aiReliabilityFromPillar(res Result) float64 {
	// The pillar does not carry its own reliability; the input type does.
	// Handlers pass a file-typed result with an ELA pillar, which is the only
	// case that distinguishes the two values.
	if res.Pillars.ELA.Available {
		return 0.7
	}
	return 0.5
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
	default:
		return "Tanda mencurigakan pada tautan"
	}
}

func (p *Pipeline) fromCache(key string) (Result, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	c, ok := p.cache[key]
	if !ok || time.Now().After(c.expires) {
		delete(p.cache, key)
		return Result{}, false
	}
	return c.result, true
}

func (p *Pipeline) toCache(key string, res Result) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cache[key] = cached{result: res, expires: time.Now().Add(p.CacheTTL)}
}
