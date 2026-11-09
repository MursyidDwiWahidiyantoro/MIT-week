// Package pipeline orchestrates one audit: it runs the pillars concurrently
// under a shared deadline, reports each pillar the moment it finishes,
// degrades gracefully when one fails, and caches the final result for a
// short TTL.
package pipeline

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/amba/amba/backend/internal/forensics"
	"github.com/amba/amba/backend/internal/grounding"
	"github.com/amba/amba/backend/internal/linkcheck"
	"github.com/amba/amba/backend/internal/llm"
	"github.com/amba/amba/backend/internal/metadata"
	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/provenance"
	"github.com/amba/amba/backend/internal/scoring"
	"github.com/amba/amba/backend/internal/security"
	"github.com/amba/amba/backend/internal/tactics"
	"github.com/amba/amba/backend/internal/video"
)

// Request is one audit input: a link (URL) or an uploaded image (Image).
type Request struct {
	InputType models.InputType
	URL       string
	Image     []byte
	// ImageFormat is the validated decoder format ("jpeg", "png", "webp").
	ImageFormat string
	// OCRText is an optional client-side hint for FILE input.
	OCRText string
}

// Result is what the pipeline produces before signing.
type Result struct {
	// InputType is the request's: the AI floors apply to uploads only.
	InputType models.InputType
	Pillars   models.Pillars
	Tactics   []models.Tactic
	Metadata  *models.SourceMetadata
	// OCRText is the model's reading of an uploaded image (FILE only).
	OCRText string
	// Video is the downloaded social video's analysis (video links only).
	Video    *models.VideoAnalysis
	Degraded bool
	// InjectionSuspected is set when the model output or the input was
	// flagged, so the scorer raises INJECTION_GUARD (TECH_SPEC §3.3 floor 5).
	InjectionSuspected bool
	// LinkMalicious is set on a Safe Browsing hit (floor 2).
	LinkMalicious bool
	// ELAHotspots and AIAnomalies are the evidence the ELA_CORROBORATED floor
	// compares; they are not serialised to the client.
	ELAHotspots []scoring.Box
	AIAnomalies []scoring.Box
}

// Event is one pillar progress report, emitted as soon as it happens.
type Event struct {
	Pillar string // ela | ai | fact | link
	Status string // running | done | skipped | failed
	// Data is the pillar payload for a "done" event.
	Data any
}

// Pillar names as they appear in SSE events.
const (
	PillarELA  = "ela"
	PillarAI   = "ai"
	PillarFact = "fact"
	PillarLink = "link"
)

// Pipeline runs audits. All fields are safe for concurrent use once built.
type Pipeline struct {
	LLM          *llm.Client
	Grounder     *grounding.Grounder
	Searcher     grounding.Searcher
	SafeBrowsing *linkcheck.SafeBrowsing
	// RDAP looks up a link's domain age (nil: off).
	RDAP *linkcheck.RDAP
	// Detector is an optional dedicated AI-image detector (nil: none).
	Detector AIDetector
	// NewFetchClient returns the SSRF-safe client used for every request to a
	// user-supplied URL. Tests replace it to reach httptest servers.
	NewFetchClient func(timeout time.Duration) *http.Client

	CacheTTL time.Duration
	CacheMax int

	// Per-pillar budgets (TECH_SPEC §4). Each is further bounded by whatever
	// remains of the request's hard deadline.
	ELADeadline      time.Duration
	LinkDeadline     time.Duration
	MetaDeadline     time.Duration
	LLMDeadline      time.Duration
	SearchDeadline   time.Duration
	ClassifyDeadline time.Duration
	// DetectDeadline bounds the AI detector; PreviewDeadline bounds the
	// download of a link's preview image; TextDeadline bounds the pasted-text
	// check after the model answers.
	DetectDeadline  time.Duration
	PreviewDeadline time.Duration
	TextDeadline    time.Duration
	// Video downloads social video links (nil: video analysis off);
	// VideoFetchDeadline bounds one download with its frame extraction.
	Video              VideoSource
	VideoFetchDeadline time.Duration
	videoSlots         chan struct{}

	mu    sync.Mutex
	cache map[string]*list.Element
	order *list.List
}

type cached struct {
	key     string
	result  Result
	expires time.Time
}

// New builds a pipeline with the TECH_SPEC defaults.
func New(client *llm.Client, g *grounding.Grounder, cacheTTL time.Duration) *Pipeline {
	if cacheTTL <= 0 {
		cacheTTL = 10 * time.Minute
	}
	return &Pipeline{
		LLM:            client,
		Grounder:       g,
		NewFetchClient: func(timeout time.Duration) *http.Client { return security.NewSafeClient(timeout, 3) },
		CacheTTL:       cacheTTL,
		CacheMax:       500,
		ELADeadline:    4 * time.Second,
		LinkDeadline:   2500 * time.Millisecond,
		MetaDeadline:   2500 * time.Millisecond,
		LLMDeadline:    9 * time.Second,
		// Measured 9 Nov: a text-only call takes 4-9 s on our gateways and a
		// Tavily search 1-3 s; with 3 s the classifier timed out every time.
		SearchDeadline:     6 * time.Second,
		ClassifyDeadline:   10 * time.Second,
		DetectDeadline:     4 * time.Second,
		PreviewDeadline:    1500 * time.Millisecond,
		TextDeadline:       1500 * time.Millisecond,
		VideoFetchDeadline: 25 * time.Second,
		videoSlots:         make(chan struct{}, maxVideoDownloads),
		cache:              make(map[string]*list.Element),
		order:              list.New(),
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
	case models.InputFile:
		h.Write(req.Image)
		h.Write([]byte{0})
		h.Write([]byte(req.OCRText))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Run executes every pillar concurrently and returns the final result.
func (p *Pipeline) Run(ctx context.Context, req Request) Result {
	return p.RunStream(ctx, req, nil)
}

// RunStream is Run with progress: emit (which may be nil) receives one
// "running" event when a pillar starts and one terminal event when it ends,
// from whichever goroutine finished the work. It never returns an error for
// a pillar failure: a dead dependency degrades the result instead.
func (p *Pipeline) RunStream(ctx context.Context, req Request, emit func(Event)) Result {
	tracker := newTracker(emit)
	defer tracker.finish()

	key := CacheKey(req)
	if res, ok := p.fromCache(key); ok {
		tracker.replay(res)
		return res
	}

	var (
		mu  sync.Mutex // guards the fields below across goroutines
		res Result

		elaFailed, llmFailed, linkFailed, searchFailed, videoFailed bool
		linkRan                                                     bool
		modelBoxes                                                  []llm.Anomaly
		// textEvidence is TextEvidence for each model box (-1: unmeasured).
		textEvidence []int
	)
	// The ELA step hands its decoded image to the text check, which needs
	// the model's boxes and so runs after the model answers.
	decoded := make(chan image.Image, 1)
	elaRan := false

	res.InputType = req.InputType
	g, gctx := errgroup.WithContext(ctx)

	// ELA: JPEG and PNG uploads. Error level analysis re-compresses the image
	// as JPEG, which says nothing about a WebP file's history.
	if req.InputType == models.InputFile && len(req.Image) > 0 && req.ImageFormat != "webp" {
		elaRan = true
		tracker.start(PillarELA)
		g.Go(func() error {
			pillar, hotspots, src, ok := p.runELA(gctx, req.Image)
			decoded <- src
			mu.Lock()
			res.Pillars.ELA = pillar
			res.ELAHotspots = hotspots
			elaFailed = !ok
			mu.Unlock()
			if ok {
				tracker.done(PillarELA, pillar)
			} else {
				tracker.fail(PillarELA)
			}
			return nil
		})
	}

	// Link: the submitted URL, or the first link in the client's OCR hint
	// for an image. Links that only appear in the model's OCR text are
	// checked after the model answers (below).
	linkTarget := ""
	switch req.InputType {
	case models.InputURL:
		linkTarget = req.URL
	case models.InputFile:
		linkTarget = FirstURL(req.OCRText)
	}
	if linkTarget != "" {
		linkRan = true
		tracker.start(PillarLink)
		g.Go(func() error {
			pillar, malicious, ok := p.runLink(gctx, linkTarget)
			mu.Lock()
			res.Pillars.Link = pillar
			res.LinkMalicious = malicious
			linkFailed = !ok
			mu.Unlock()
			tracker.done(PillarLink, pillar)
			return nil
		})
	}

	// Model: metadata and preview image (URL) -> findings -> AI pillar ->
	// search -> classify. The AI detector, when configured, runs beside the
	// model on the same image.
	tracker.start(PillarAI)
	tracker.start(PillarFact)
	g.Go(func() error {
		tacticSource := ""
		in := llm.Input{Type: string(req.InputType)}
		var (
			judged   bool // an image the AI pillar can judge
			prov     provenance.Result
			detected <-chan detection
			clip     *video.Clip // a downloaded social video, judged instead of the preview image
		)

		switch req.InputType {
		case models.InputURL:
			md := p.resolveMetadata(gctx, req.URL)
			mu.Lock()
			res.Metadata = md.model
			mu.Unlock()
			in.Content = md.forModel
			tacticSource = md.caption
			if tacticSource == "" {
				tacticSource = req.URL
			}
			// A social video link: download the video and judge its keyframes
			// and soundtrack. When that fails, the preview image is judged.
			c, va, transient := p.fetchVideo(gctx, req.URL)
			if va != nil {
				mu.Lock()
				res.Video, videoFailed = va, transient
				mu.Unlock()
			}
			if c != nil {
				clip, judged = c, true
				if c.AILabel {
					// The platform's own AI label, visible on the post.
					prov = provenance.Result{Kind: provenance.AIGenerated, Source: "watermark"}
				}
			} else if md.image != "" {
				// The link's preview image is what people see when it is shared.
				if img, format, ok := p.fetchPreview(gctx, md.image); ok {
					judged, prov = true, provenance.Scan(img)
					detected = p.startDetector(gctx, img, format)
					in.Image, in.ImageMIME = forensics.ModelImage(img, format, 1024)
				}
			}
		case models.InputFile:
			in.Content = req.OCRText
			tacticSource = req.OCRText
			if len(req.Image) > 0 {
				// Metadata blocks only: microseconds, no decoding.
				judged, prov = true, provenance.Scan(req.Image)
				detected = p.startDetector(gctx, req.Image, req.ImageFormat)
				in.Image, in.ImageMIME = forensics.ModelImage(req.Image, req.ImageFormat, 1600)
			}
		}
		label := "text"
		if req.InputType == models.InputFile {
			label = "ocr"
		}

		var (
			f   llm.Findings
			err error
		)
		if clip != nil {
			vctx, cancel := withBudget(gctx, p.videoModelBudget(gctx))
			vf, verr := p.LLM.AnalyzeVideo(vctx, videoInput(clip))
			cancel()
			err = verr
			mu.Lock()
			if verr == nil {
				fillVideo(res.Video, clip, &vf)
			} else {
				fillVideo(res.Video, clip, nil)
			}
			mu.Unlock()
			if verr == nil {
				f = videoFindings(vf)
				tacticSource = videoTacticSource(clip, vf, tacticSource)
			}
		} else {
			lctx, cancel := withBudget(gctx, p.LLMDeadline)
			f, err = p.LLM.Analyze(lctx, in)
			cancel()
		}
		modelOK := err == nil

		if judged {
			ocr := f.OCRText
			if ocr == "" {
				ocr = req.OCRText
			}
			ai, ok := buildAIPillar(aiEvidence{prov: prov, modelOK: modelOK, findings: f, ocr: ocr, det: awaitDetection(gctx, detected)})
			mu.Lock()
			res.Pillars.AI = ai
			mu.Unlock()
			if ok {
				tracker.done(PillarAI, ai)
			} else {
				tracker.fail(PillarAI)
			}
		} else {
			// A link without a preview image has no image to judge.
			tracker.skip(PillarAI)
		}

		if !modelOK {
			mu.Lock()
			llmFailed = true
			mu.Unlock()
			tracker.fail(PillarFact)
			return nil
		}

		anomalies := make([]scoring.Box, 0, len(f.TextAnomalies))
		for _, a := range f.TextAnomalies {
			anomalies = append(anomalies, scoring.Box(a.BoundingBox))
		}
		// Timpa Teks: does the text the model marked react to compression
		// unlike the image's other text?
		if elaRan && len(f.TextAnomalies) > 0 {
			if src := awaitImage(gctx, decoded); src != nil {
				ev := p.measureText(gctx, src, f.TextAnomalies)
				mu.Lock()
				textEvidence = ev
				mu.Unlock()
			}
		}

		// Tactics are verified against the text the user can see: the link's
		// caption, or the model's own OCR of the image.
		source := tacticSource
		if req.InputType == models.InputFile && f.OCRText != "" {
			source = f.OCRText
		}
		candidates := make([]models.Tactic, 0, len(f.Tactics))
		for _, t := range f.Tactics {
			candidates = append(candidates, models.Tactic{Code: t.Code, Phrase: t.Phrase, Source: label})
		}
		found := tactics.Verify(candidates, source)

		mu.Lock()
		res.AIAnomalies = anomalies
		modelBoxes = f.TextAnomalies
		res.Tactics = found
		res.InjectionSuspected = f.InjectionSuspected
		if req.InputType == models.InputFile {
			res.OCRText = f.OCRText
		}
		mu.Unlock()

		// A URL the model read inside the screenshot gets the link pillar
		// too (PRD F4), unless the client already supplied one.
		if req.InputType == models.InputFile && !linkRan {
			if target := FirstURL(f.OCRText); target != "" {
				tracker.start(PillarLink)
				pillar, malicious, ok := p.runLink(gctx, target)
				mu.Lock()
				res.Pillars.Link = pillar
				res.LinkMalicious = malicious
				linkFailed = !ok
				mu.Unlock()
				tracker.done(PillarLink, pillar)
			}
		}

		if f.Claim == "" {
			// No checkable claim: the fact pillar is unavailable, never SAFE
			// by omission (PRD F3 AC).
			tracker.skip(PillarFact)
			return nil
		}
		fact, failed := p.checkClaim(gctx, f.Claim, f.SearchQuery)
		mu.Lock()
		res.Pillars.Fact = fact
		searchFailed = failed
		mu.Unlock()
		tracker.done(PillarFact, fact)
		return nil
	})

	_ = g.Wait()

	mu.Lock()
	defer mu.Unlock()
	// The model's region boxes are shown on the ELA viewer. They are attached
	// only after every goroutine finished, so no two writers share a struct.
	// A box whose text reacts to compression unlike the other text raises the
	// ELA score to its evidence and, from 70, is a hotspot ELA_CORROBORATED
	// can match: the model and the measurement point at the same text.
	if res.Pillars.ELA.Available {
		res.Pillars.ELA.Anomalies = make([]models.Anomaly, 0, len(modelBoxes))
		for i, a := range modelBoxes {
			anomaly := models.Anomaly{BoundingBox: a.BoundingBox, ExtractedText: a.ExtractedText, Reason: a.Reason}
			if i < len(textEvidence) && textEvidence[i] >= 0 {
				ev := textEvidence[i]
				anomaly.Evidence = &ev
				res.Pillars.ELA.Score = max(res.Pillars.ELA.Score, ev)
				if ev >= TextHotspotMin {
					res.ELAHotspots = append(res.ELAHotspots, scoring.Box(a.BoundingBox))
				}
			}
			res.Pillars.ELA.Anomalies = append(res.Pillars.ELA.Anomalies, anomaly)
		}
	}
	res.Degraded = elaFailed || llmFailed || linkFailed || searchFailed || videoFailed
	// A degraded result is not cached: the model or a lookup failing once
	// (a free model's rate limit, a timeout) must not answer the user's
	// retry with the same incomplete result for the next ten minutes.
	if !res.Degraded {
		p.toCache(key, res)
	}
	return res
}

// runELA returns the pillar, its grid hotspots and the decoded image (nil
// when decoding failed or ran out of time).
func (p *Pipeline) runELA(ctx context.Context, encoded []byte) (models.ELAPillar, []scoring.Box, image.Image, bool) {
	ectx, cancel := withBudget(ctx, p.ELADeadline)
	defer cancel()
	done := make(chan struct{})
	var res forensics.Result
	var src image.Image
	var err error
	go func() {
		var format string
		if src, format, err = forensics.Decode(encoded); err == nil {
			res, err = forensics.AnalyzeImage(src, format)
		}
		close(done)
	}()
	select {
	case <-done:
	case <-ectx.Done():
		return models.ELAPillar{Anomalies: []models.Anomaly{}}, nil, nil, false
	}
	if err != nil {
		return models.ELAPillar{Anomalies: []models.Anomaly{}}, nil, nil, false
	}
	// The 16x16 grid cells above the mean are the physical hotspots the
	// ELA_CORROBORATED floor needs; without them the floor can never fire.
	var boxes []scoring.Box
	for _, h := range res.Hotspots(1.5) {
		boxes = append(boxes, scoring.Box(h))
	}
	pillar := models.ELAPillar{
		Available:   true,
		Score:       res.Score,
		Reliability: res.Reliability,
		Anomalies:   []models.Anomaly{},
	}
	if res.Note != "" {
		note := res.Note
		pillar.ReliabilityNote = &note
	}
	return pillar, boxes, src, true
}

// TextHotspotMin is the text evidence from which a model box counts as an
// ELA hotspot for ELA_CORROBORATED. No original line reached it on the
// benchmark (internal/forensics/text.go).
const TextHotspotMin = 70

// maxMeasuredBoxes bounds the text check: each box costs two JPEG
// re-encodes of a band of the image.
const maxMeasuredBoxes = 4

// measureText runs TextEvidence on the model's boxes within the text
// budget; late or skipped boxes are -1 (unmeasured).
func (p *Pipeline) measureText(ctx context.Context, src image.Image, boxes []llm.Anomaly) []int {
	out := make([]int, len(boxes))
	for i := range out {
		out[i] = -1
	}
	n := min(len(boxes), maxMeasuredBoxes)
	norm := make([][4]int, n)
	for i := 0; i < n; i++ {
		norm[i] = boxes[i].BoundingBox
	}
	budget := p.TextDeadline
	if budget <= 0 {
		budget = 1500 * time.Millisecond
	}
	tctx, cancel := withBudget(ctx, budget)
	defer cancel()
	done := make(chan []int, 1)
	go func() { done <- forensics.TextEvidence(src, norm) }()
	select {
	case ev := <-done:
		copy(out, ev)
	case <-tctx.Done():
	}
	return out
}

func awaitImage(ctx context.Context, ch <-chan image.Image) image.Image {
	select {
	case src := <-ch:
		return src
	case <-ctx.Done():
		return nil
	}
}

// runLink scores a URL, follows where it really leads, and asks Safe
// Browsing about every hop. Static analysis cannot fail, so the pillar is
// always available; ok=false only reports that the reputation lookup failed.
func (p *Pipeline) runLink(ctx context.Context, target string) (models.LinkPillar, bool, bool) {
	lctx, cancel := withBudget(ctx, p.LinkDeadline)
	defer cancel()

	analysis := linkcheck.Analyze(target)
	// Domain age runs beside the redirect check: the first host now, the
	// final host once the redirects are known.
	var ages []<-chan time.Time
	asked := map[string]bool{}
	lookupAge := func(host string) {
		reg := linkcheck.Registrable(host)
		if !p.RDAP.Enabled() || host == "" || asked[reg] || linkcheck.IsOfficialHost(host) || linkcheck.IsShortener(host) {
			return
		}
		asked[reg] = true
		ch := make(chan time.Time, 1)
		go func() {
			t, _ := p.RDAP.Registered(lctx, host)
			ch <- t
		}()
		ages = append(ages, ch)
	}
	lookupAge(analysis.Host)

	chain := []string{target}
	client := p.NewFetchClient(p.LinkDeadline)
	reached := false
	if hops, final, err := linkcheck.Unroll(lctx, client, target, 3); len(hops) > 0 {
		chain = hops
		reached = err == nil
		if err == nil && final != target {
			analysis = linkcheck.Combine(analysis, linkcheck.Analyze(final))
			lookupAge(analysis.Host)
		}
	}
	// What the destination page shows: a form asking for a password, PIN,
	// OTP or card number, perhaps under a bank's name. Official sites are
	// not fetched; the page read stays inside the link budget.
	if reached && !linkcheck.IsOfficialHost(analysis.Host) {
		if page, err := metadata.New(client).Inspect(lctx, chain[len(chain)-1]); err == nil {
			analysis = linkcheck.WithPage(analysis, page.Title, page.SiteName, page.AsksSecret)
		}
	}

	ok := true
	malicious := false
	if p.SafeBrowsing.Enabled() {
		hit, err := p.SafeBrowsing.Check(lctx, chain)
		switch {
		case err != nil:
			ok = false
		case hit:
			malicious = true
			analysis = linkcheck.WithSafeBrowsingHit(analysis)
		}
	}

	// The youngest domain on the way counts: an old shortener or an old
	// redirect host does not make a new destination trustworthy.
	var youngest time.Time
	for _, ch := range ages {
		select {
		case t := <-ch:
			if !t.IsZero() && t.After(youngest) {
				youngest = t
			}
		case <-lctx.Done():
		}
	}
	analysis = linkcheck.WithDomainAge(analysis, youngest, time.Now())

	reasons := make([]models.LinkReason, 0, len(analysis.Findings))
	for _, f := range analysis.Findings {
		reasons = append(reasons, models.LinkReason{Code: f.Code, Text: LinkReasonText(f.Code)})
	}
	final := analysis.FinalURL
	if len(chain) > 0 {
		final = chain[len(chain)-1]
	}
	return models.LinkPillar{
		Available:     true,
		Score:         analysis.Score,
		FinalURL:      final,
		RedirectChain: chain,
		Reasons:       reasons,
	}, malicious, ok
}

type resolvedMetadata struct {
	model    *models.SourceMetadata
	forModel string
	caption  string
	// image is the link's preview image URL, "" when it has none.
	image string
}

func (p *Pipeline) resolveMetadata(ctx context.Context, rawURL string) resolvedMetadata {
	mctx, cancel := withBudget(ctx, p.MetaDeadline)
	defer cancel()
	m := metadata.New(p.NewFetchClient(p.MetaDeadline)).Resolve(mctx, rawURL)
	return resolvedMetadata{
		model: &models.SourceMetadata{
			Status:      models.MetadataStatus(m.Status),
			Platform:    m.Platform,
			Via:         m.Via,
			Title:       m.Title,
			Author:      m.Author,
			Description: m.Description,
			SiteName:    m.SiteName,
		},
		forModel: m.ForModel(rawURL),
		caption:  m.Caption(),
		image:    m.Image,
	}
}

// checkClaim searches AMBA's own fact-check index and lets the model judge
// each result against the claim. Only URLs that came back from that search
// can become sources (TECH_SPEC §6 (a)), so an invented citation can never
// carry CONTRADICTS into the score. failed reports a search or classifier
// outage (the pillar then stays UNVERIFIED and the result is degraded).
func (p *Pipeline) checkClaim(ctx context.Context, claim, query string) (models.FactPillar, bool) {
	fact := models.FactPillar{Available: true, Claim: claim, Status: models.FactUnverified, Sources: []models.Source{}}
	if p.Searcher == nil || !p.Searcher.Enabled() {
		return fact, false
	}
	if query == "" {
		query = claim
	}
	sctx, cancel := withBudget(ctx, p.SearchDeadline)
	results, err := p.Searcher.Search(sctx, query)
	cancel()
	if err != nil {
		return fact, true
	}
	if len(results) == 0 {
		return fact, false
	}

	evidence := make([]llm.Evidence, len(results))
	for i, r := range results {
		evidence[i] = llm.Evidence{Title: r.Title, Publisher: r.Publisher, Snippet: r.Snippet, Rating: r.Rating}
	}
	cctx, cancel := withBudget(ctx, p.ClassifyDeadline)
	matches, err := p.LLM.Classify(cctx, claim, evidence)
	cancel()

	toolURLs := grounding.URLSet(results)
	for i, r := range results {
		match := models.FactUnverified
		if i < len(matches) && p.Grounder.Verified(r.URL, toolURLs) {
			match = models.FactStatus(matches[i])
		}
		fact.Sources = append(fact.Sources, models.Source{
			Title: r.Title, Publisher: r.Publisher, URL: r.URL, Match: match,
		})
	}
	fact.Status = statusFromSources(fact.Sources)
	return fact, err != nil
}

// statusFromSources: a verified contradiction with no verified support is
// CONTRADICTS, support with no contradiction is SUPPORTS, anything else
// (including disagreement between fact-checkers) is UNVERIFIED.
func statusFromSources(sources []models.Source) models.FactStatus {
	contradicts, supports := 0, 0
	for _, s := range sources {
		switch s.Match {
		case models.FactContradicts:
			contradicts++
		case models.FactSupports:
			supports++
		}
	}
	switch {
	case contradicts > 0 && supports == 0:
		return models.FactContradicts
	case supports > 0 && contradicts == 0:
		return models.FactSupports
	default:
		return models.FactUnverified
	}
}

// Verdict computes the score from a Result.
func Verdict(res Result, injectionSuspected bool) models.Verdict {
	// The pipeline's own finding wins; the argument exists for tests that
	// build a Result by hand.
	if res.InjectionSuspected {
		injectionSuspected = true
	}
	// A claim no fact-checker has ruled on (UNVERIFIED) is shown, but it is not
	// evidence either way (team decision, 9 Nov 2026): counted as risk it made
	// every real news video and every untouched price graphic SUSPICIOUS.
	factCounts := res.Pillars.Fact.Available && res.Pillars.Fact.Status != models.FactUnverified
	in := scoring.Input{
		ELA:                scoring.Signal{Available: res.Pillars.ELA.Available, Score: res.Pillars.ELA.Score, Reliability: res.Pillars.ELA.Reliability},
		AI:                 scoring.Signal{Available: res.Pillars.AI.Available, Score: res.Pillars.AI.Score, Reliability: res.Pillars.AI.Reliability},
		Fact:               scoring.Signal{Available: factCounts, Score: factScore(res.Pillars.Fact.Status), Reliability: 1.0},
		Link:               scoring.Signal{Available: res.Pillars.Link.Available, Score: res.Pillars.Link.Score, Reliability: 1.0},
		LinkMalicious:      res.LinkMalicious,
		InjectionSuspected: injectionSuspected,
		ELAHotspots:        res.ELAHotspots,
		AIAnomalies:        res.AIAnomalies,
		UploadedImage:      res.InputType == models.InputFile,
	}
	if v := res.Video; v != nil && v.Status == models.VideoOK {
		in.Video = true
		in.VideoManipulation, in.VideoContext = v.Manipulation.Score, v.Context.Score
	}
	switch res.Pillars.AI.Provenance.Kind {
	case models.ProvenanceAIGenerated:
		in.AIDeclared = scoring.DeclaredAIGenerated
	case models.ProvenanceAIEdited:
		in.AIDeclared = scoring.DeclaredAIEdited
	}
	if res.Pillars.Fact.Available && res.Pillars.Fact.Status == models.FactContradicts {
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

// LinkReasonText is the static Indonesian copy for each link reason code.
func LinkReasonText(code string) string {
	switch code {
	case linkcheck.ReasonLookalikeBrand:
		return "Domain meniru nama layanan populer"
	case linkcheck.ReasonSuspiciousTLD:
		return "Akhiran domain sering dipakai penipuan"
	case linkcheck.ReasonIPHost:
		return "Alamat berupa nomor IP, bukan nama situs"
	case linkcheck.ReasonPunycode:
		return "Nama domain memakai karakter yang mirip huruf biasa"
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
	case linkcheck.ReasonScamKeywords:
		return "Nama situs memuat kata umpan seperti hadiah atau bansos"
	case linkcheck.ReasonShortenerHides:
		return "Tautan pendek mengarah ke situs yang mencurigakan"
	case linkcheck.ReasonSafeBrowsing:
		return "Tercatat sebagai situs berbahaya oleh Google Safe Browsing"
	case linkcheck.ReasonUnparseable:
		return "Alamat tautan tidak dapat dibaca"
	case linkcheck.ReasonPageAsksSecret:
		return "Halaman meminta kata sandi, PIN, OTP, atau nomor kartu"
	case linkcheck.ReasonPageImpersonation:
		return "Halaman memakai nama layanan populer dan meminta kata sandi atau kode"
	case linkcheck.ReasonNewDomain:
		return "Domain baru didaftarkan kurang dari sebulan lalu"
	case linkcheck.ReasonYoungDomain:
		return "Domain berumur kurang dari enam bulan"
	default:
		return "Tanda mencurigakan pada tautan"
	}
}

// withBudget bounds a step by its own budget and by what remains of the
// parent deadline, keeping a small reserve so the verdict can still be
// computed and streamed.
func withBudget(parent context.Context, budget time.Duration) (context.Context, context.CancelFunc) {
	if deadline, ok := parent.Deadline(); ok {
		if remaining := time.Until(deadline) - 300*time.Millisecond; remaining < budget {
			budget = max(remaining, 50*time.Millisecond)
		}
	}
	return context.WithTimeout(parent, budget)
}

func nonNil(v []string) []string {
	if v == nil {
		return []string{}
	}
	return v
}

func (p *Pipeline) fromCache(key string) (Result, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	el, ok := p.cache[key]
	if !ok {
		return Result{}, false
	}
	c := el.Value.(*cached)
	if time.Now().After(c.expires) {
		p.order.Remove(el)
		delete(p.cache, key)
		return Result{}, false
	}
	p.order.MoveToFront(el)
	return c.result, true
}

// toCache stores a result in a bounded LRU (TECH_SPEC §4: 500 entries, TTL).
// Expired entries are evicted on the way in, so an idle cache cannot grow.
func (p *Pipeline) toCache(key string, res Result) {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	if el, ok := p.cache[key]; ok {
		el.Value.(*cached).result = res
		el.Value.(*cached).expires = now.Add(p.CacheTTL)
		p.order.MoveToFront(el)
		return
	}
	p.cache[key] = p.order.PushFront(&cached{key: key, result: res, expires: now.Add(p.CacheTTL)})
	for el := p.order.Back(); el != nil; {
		prev := el.Prev()
		c := el.Value.(*cached)
		if p.order.Len() > p.cacheMax() || now.After(c.expires) {
			p.order.Remove(el)
			delete(p.cache, c.key)
		} else {
			break
		}
		el = prev
	}
}

func (p *Pipeline) cacheMax() int {
	if p.CacheMax <= 0 {
		return 500
	}
	return p.CacheMax
}

// CacheLen reports the number of cached results (for tests and metrics).
func (p *Pipeline) CacheLen() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.order.Len()
}

// FirstURL returns the first link in a text, so a link inside a screenshot
// or a message shared to the Telegram bot gets the link pillar (PRD F4).
// Links in screenshots are often written without a scheme ("klik
// bri-verifikasi.my.id", "bit.ly/abc"); those are read as https, so the
// missing scheme is not counted against the link. An e-mail address is not
// a link.
func FirstURL(text string) string {
	for _, field := range strings.Fields(text) {
		field = strings.Trim(field, "\"'()[]<>,.;!?")
		lower := strings.ToLower(field)
		if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
			return field
		}
	}
	for _, loc := range bareLink.FindAllStringIndex(text, 8) {
		if loc[0] > 0 && (text[loc[0]-1] == '@' || text[loc[0]-1] == '.' || text[loc[0]-1] == '/') {
			continue
		}
		return "https://" + strings.TrimRight(text[loc[0]:loc[1]], ".,;:!?")
	}
	return ""
}

// bareLink is a host name with a common or abused top-level domain, with an
// optional path, written without a scheme.
var bareLink = regexp.MustCompile(`(?i)\b(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+(?:com|net|org|id|co|io|me|ly|xyz|top|site|online|link|info|shop|store|click|icu|live|cc|app|vip|biz|website|space|fun|tk|ml|ga|cf|gq|pw|cfd|sbs|cyou|lol|bond|buzz|rest|cam|monster|club|pro|asia|win|work|life|tech)\b(?:/[^\s"'<>()\[\]]*)?`)

// tracker turns pillar work into ordered progress events and guarantees that
// every pillar ends with exactly one terminal event.
type tracker struct {
	mu      sync.Mutex
	emit    func(Event)
	started map[string]bool
	ended   map[string]bool
}

func newTracker(emit func(Event)) *tracker {
	return &tracker{emit: emit, started: map[string]bool{}, ended: map[string]bool{}}
}

func (t *tracker) send(e Event) {
	if t.emit != nil {
		t.emit(e)
	}
}

func (t *tracker) start(pillar string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.started[pillar] {
		return
	}
	t.started[pillar] = true
	t.send(Event{Pillar: pillar, Status: "running"})
}

func (t *tracker) end(pillar, status string, data any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ended[pillar] {
		return
	}
	t.ended[pillar] = true
	t.send(Event{Pillar: pillar, Status: status, Data: data})
}

func (t *tracker) done(pillar string, data any) { t.end(pillar, "done", data) }
func (t *tracker) fail(pillar string)           { t.end(pillar, "failed", nil) }
func (t *tracker) skip(pillar string)           { t.end(pillar, "skipped", nil) }

// finish closes every pillar that never reached a terminal state, so the
// client's progress list always settles.
func (t *tracker) finish() {
	for _, p := range []string{PillarELA, PillarLink, PillarAI, PillarFact} {
		t.mu.Lock()
		started, ended := t.started[p], t.ended[p]
		t.mu.Unlock()
		switch {
		case ended:
		case started:
			t.fail(p)
		default:
			t.skip(p)
		}
	}
}

// replay reports a cached result as if each pillar had just finished.
func (t *tracker) replay(res Result) {
	pillars := []struct {
		name      string
		available bool
		data      any
	}{
		{PillarELA, res.Pillars.ELA.Available, res.Pillars.ELA},
		{PillarLink, res.Pillars.Link.Available, res.Pillars.Link},
		{PillarAI, res.Pillars.AI.Available, res.Pillars.AI},
		{PillarFact, res.Pillars.Fact.Available, res.Pillars.Fact},
	}
	for _, p := range pillars {
		if p.available {
			t.start(p.name)
			t.done(p.name, p.data)
		}
	}
}
