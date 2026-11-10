package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/pipeline"
	"github.com/amba/amba/backend/internal/security"
)

// auditRequest is the JSON body for the url and text routes.
type auditRequest struct {
	URL  string `json:"url"`
	Text string `json:"text"`
}

// auditURL handles POST /audit/url.
func (h *handlers) auditURL(w http.ResponseWriter, r *http.Request) {
	var body auditRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<10)).Decode(&body); err != nil {
		writeError(w, errInvalidURL)
		return
	}
	raw := strings.TrimSpace(body.URL)
	if raw == "" || len(raw) > 2048 {
		writeError(w, errInvalidURL)
		return
	}
	if _, err := security.ValidateFetchURL(raw); err != nil {
		// Generic message: never reveal which rule refused the target.
		writeError(w, errBlockedTarget)
		return
	}
	h.runAudit(w, r, pipeline.Request{InputType: models.InputURL, URL: raw})
}

// auditText handles POST /audit/text.
func (h *handlers) auditText(w http.ResponseWriter, r *http.Request) {
	var body auditRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 64<<10)).Decode(&body); err != nil {
		writeError(w, errInvalidURL)
		return
	}
	text := strings.TrimSpace(body.Text)
	if text == "" || len(text) > 5000 {
		writeError(w, errInvalidURL)
		return
	}
	h.runAudit(w, r, pipeline.Request{InputType: models.InputText, Text: text})
}

// auditFile handles POST /audit/file.
func (h *handlers) auditFile(w http.ResponseWriter, r *http.Request) {
	maxBytes := h.cfg.MaxUploadMB << 20
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		writeError(w, errPayloadTooLarge)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, errUnsupportedMedia)
		return
	}
	defer func() { _ = file.Close() }()

	data, format, _, err := security.ValidateUpload(file, maxBytes)
	switch {
	case errors.Is(err, security.ErrTooLarge):
		writeError(w, errPayloadTooLarge)
		return
	case errors.Is(err, security.ErrUnsupported), errors.Is(err, security.ErrAnimated):
		writeError(w, errUnsupportedMedia)
		return
	case errors.Is(err, security.ErrTooManyPixels):
		writeError(w, errPayloadTooLarge)
		return
	case err != nil:
		writeError(w, errUnsupportedMedia)
		return
	}

	// The filename is never stored or logged; only the format is used.
	h.runAudit(w, r, pipeline.Request{
		InputType: models.InputFile,
		Image:     data,
		ImageMIME: mimeFor(format),
		OCRText:   strings.TrimSpace(r.FormValue("ocr_text")),
	})
}

func mimeFor(format string) string {
	switch format {
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

// runAudit executes the pipeline and renders either SSE or a single JSON body.
func (h *handlers) runAudit(w http.ResponseWriter, r *http.Request, req pipeline.Request) {
	h.requests.Add(1)
	if !h.limiter.AcquireInFlight() {
		writeError(w, errRateLimited)
		return
	}
	defer h.limiter.ReleaseInFlight()

	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.HardTimeout)
	defer cancel()

	stream := strings.Contains(r.Header.Get("Accept"), "text/event-stream")
	if stream {
		h.streamAudit(ctx, w, req)
		return
	}

	res := h.pipeline.Run(ctx, req, nil)
	dossier := h.buildDossier(res, req)
	writeJSON(w, http.StatusOK, dossier)
}

// streamAudit forwards the pipeline's real events. A pillar announces
// "running" when it actually starts and "done"/"failed"/"skipped" when it
// actually finishes; there is no timer anywhere in this path.
func (h *handlers) streamAudit(ctx context.Context, w http.ResponseWriter, req pipeline.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, errInternal)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	var writeMu chan struct{} = make(chan struct{}, 1)
	writeMu <- struct{}{}

	send := func(event string, payload any) bool {
		select {
		case <-ctx.Done():
			return false
		case <-writeMu:
		}
		defer func() { writeMu <- struct{}{} }()
		data, err := json.Marshal(payload)
		if err != nil {
			return false
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	// A comment heartbeat keeps proxies from closing an idle stream. It is
	// never an event, so the client's stage list cannot mistake it for work.
	done := make(chan struct{})
	defer close(done)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				select {
				case <-writeMu:
					_, _ = fmt.Fprint(w, ": ping\n\n")
					flusher.Flush()
					writeMu <- struct{}{}
				default:
				}
			}
		}
	}()

	res := h.pipeline.Run(ctx, req, func(e pipeline.Event) {
		switch e.Kind {
		case "stage":
			send("stage", map[string]string{"pillar": e.Pillar, "status": e.Status})
		case "partial":
			send("partial", map[string]any{"pillar": e.Pillar, "data": e.Data})
		}
	})

	dossier := h.buildDossier(res, req)
	if !send("verdict", dossier) {
		return
	}
	send("done", struct{}{})
}

// buildDossier assembles the contract object.
func (h *handlers) buildDossier(res pipeline.Result, req pipeline.Request) models.Dossier {
	verdict := pipeline.Verdict(res)
	d := models.Dossier{
		SchemaVersion:   2,
		ID:              newID(),
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		InputType:       req.InputType,
		SourceReference: sourceReference(req),
		Degraded:        res.Degraded,
		IsDemo:          false,
		Verdict:         verdict,
		Pillars:         res.Pillars,
		Tactics:         res.Tactics,
	}
	// The wire contract says these are arrays. Go marshals a nil slice as
	// null, which every strict client rejects; normalise before signing.
	if d.Tactics == nil {
		d.Tactics = []models.Tactic{}
	}
	d.Pillars.ELA.Anomalies = nonNilAnomalies(d.Pillars.ELA.Anomalies)
	d.Pillars.AI.Indicators = nonNilStrings(d.Pillars.AI.Indicators)
	d.Pillars.Fact.Sources = nonNilSources(d.Pillars.Fact.Sources)
	d.Pillars.Link.RedirectChain = nonNilStrings(d.Pillars.Link.RedirectChain)
	d.Pillars.Link.Reasons = nonNilLinkReasons(d.Pillars.Link.Reasons)
	if d.Pillars.Fact.Status == "" {
		d.Pillars.Fact.Status = models.FactUnverified
	}
	d.Verdict.Basis = nonNilStrings(d.Verdict.Basis)
	d.Verdict.NotChecked = nonNilStrings(d.Verdict.NotChecked)
	d.Verdict.FloorsApplied = nonNilStrings(d.Verdict.FloorsApplied)
	d.Verdict.ShortTitle = shortTitle(verdict.Status)
	if d.Verdict.Explanation == "" {
		d.Verdict.Explanation = explanation(verdict)
	}
	if signed, err := h.sign(d); err == nil {
		d.ShareToken = signed
	}
	return d
}

func sourceReference(req pipeline.Request) string {
	var ref string
	switch req.InputType {
	case models.InputURL:
		ref = req.URL
	case models.InputText:
		ref = req.Text
	default:
		ref = "Gambar yang diunggah"
	}
	if len(ref) > 200 {
		ref = ref[:200]
	}
	return ref
}

func shortTitle(s models.Status) string {
	switch s {
	case models.StatusTampered:
		return "Terindikasi Hoax/Manipulasi"
	case models.StatusSuspicious:
		return "Perlu diperiksa lebih lanjut"
	case models.StatusSafe:
		return "Tidak ditemukan tanda bahaya"
	default:
		return "Belum cukup data"
	}
}

func explanation(v models.Verdict) string {
	var parts []string
	if len(v.Basis) > 0 {
		parts = append(parts, fmt.Sprintf("Pemeriksaan yang berjalan: %s.", strings.Join(v.Basis, ", ")))
	}
	if len(v.NotChecked) > 0 {
		parts = append(parts, fmt.Sprintf("Tidak dapat diperiksa: %s.", strings.Join(v.NotChecked, ", ")))
	}
	parts = append(parts, "Hasil ini indikatif, bukan bukti pasti.")
	out := strings.Join(parts, " ")
	if len(out) > 400 {
		out = out[:400]
	}
	return out
}

// The helpers below exist only so JSON never contains null where the contract
// promises an array. They are deliberately tiny: allocating a zero-length
// slice is cheaper than teaching every client to tolerate null.
func nonNilStrings(v []string) []string {
	if v == nil {
		return []string{}
	}
	return v
}

func nonNilAnomalies(v []models.Anomaly) []models.Anomaly {
	if v == nil {
		return []models.Anomaly{}
	}
	return v
}

func nonNilSources(v []models.Source) []models.Source {
	if v == nil {
		return []models.Source{}
	}
	return v
}

func nonNilLinkReasons(v []models.LinkReason) []models.LinkReason {
	if v == nil {
		return []models.LinkReason{}
	}
	return v
}
