package httpapi

import (
	"context"
	"net/http"
	"time"
)

// opsStatus is the honest integration map. Every external dependency is
// either live, explicitly disabled, or a placeholder — never silently absent.
// It carries no key, URL, host or model output.
type opsStatus struct {
	Status       string            `json:"status"`
	Version      string            `json:"version"`
	UptimeS      int64             `json:"uptime_seconds"`
	Requests     int64             `json:"requests"`
	LLM          llmStatus         `json:"llm"`
	Retrieval    integrationStatus `json:"retrieval"`
	SafeBrowsing integrationStatus `json:"safe_browsing"`
	Pillars      map[string]string `json:"pillars"`
	Stats        map[string]any    `json:"stats"`
}

type llmStatus struct {
	Mode        string `json:"mode"`
	Model       string `json:"model"`
	Found       bool   `json:"model_found"`
	ProbeFailed bool   `json:"probe_failed"`
	Vision      bool   `json:"vision"`
	Reasoning   string `json:"reasoning_effort"`
	Placeholder bool   `json:"placeholder"`
}

type integrationStatus struct {
	Enabled     bool   `json:"enabled"`
	Configured  bool   `json:"configured"`
	Placeholder bool   `json:"placeholder"`
	Note        string `json:"note,omitempty"`
}

// getOpsStatus serves GET /api/v1/ops/status.
//
// The gateway's /models call is best-effort: a cold or briefly busy 9router
// would otherwise make a healthy server look misconfigured. The live audit
// path is the real health signal, so a failed probe is reported as "unknown"
// rather than "not found".
func (h *handlers) getOpsStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 6*time.Second)
	defer cancel()
	info := h.llm.HealthCheck(ctx)
	probeFailed := !h.llm.Mock && !info.Found && info.Count == 0

	llmMode := "live"
	placeholder := false
	if h.llm.Mock {
		llmMode = "mock"
		placeholder = true
	}
	vision := false
	for _, c := range info.Capabilities {
		if c == "vision" {
			vision = true
		}
	}

	retrieval := integrationStatus{
		Enabled:    h.cfg.Retrieval && !h.cfg.LLMMock,
		Configured: h.cfg.RetrievalURL != "",
		Note:       "Pencarian berita lewat endpoint tetap, lalu disaring daftar sumber tepercaya.",
	}
	if !retrieval.Enabled {
		retrieval.Placeholder = true
		retrieval.Note = "Nonaktif: pemeriksaan fakta tidak berjalan."
	}

	safe := integrationStatus{
		Enabled:    h.cfg.SafeBrowsing != "",
		Configured: h.cfg.SafeBrowsing != "",
		Note:       "Google Safe Browsing v4.",
	}
	if !safe.Enabled {
		safe.Placeholder = true
		safe.Note = "Placeholder: GOOGLE_SAFE_BROWSING_KEY belum diisi, jadi reputasi tautan tidak diperiksa."
	}

	pillars := map[string]string{
		"ela":  "live",
		"ai":   llmMode,
		"fact": "live",
		"link": "live",
	}
	if !retrieval.Enabled {
		pillars["fact"] = "placeholder"
	}
	if llmMode == "mock" {
		pillars["ai"] = "placeholder"
	}
	if safe.Placeholder {
		pillars["link"] = "live+placeholder"
	}

	writeJSON(w, http.StatusOK, opsStatus{
		Status:   "ok",
		Version:  "1.0.0",
		UptimeS:  int64(time.Since(h.startedAt).Seconds()),
		Requests: h.requests.Load(),
		LLM: llmStatus{
			Mode:        llmMode,
			Model:       h.cfg.RouterModel,
			Found:       info.Found,
			ProbeFailed: probeFailed,
			Vision:      vision,
			Reasoning:   h.cfg.Reasoning,
			Placeholder: placeholder,
		},
		Retrieval:    retrieval,
		SafeBrowsing: safe,
		Pillars:      pillars,
		Stats: map[string]any{
			"pipeline": h.pipeline.Stats.Snapshot(),
			"llm":      h.llm.Stats.Snapshot(),
		},
	})
}

// getMetrics is mounted on the optional internal listener only.
func (h *handlers) getMetrics(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"requests": h.requests.Load(),
		"uptime_s": int64(time.Since(h.startedAt).Seconds()),
		"pipeline": h.pipeline.Stats.Snapshot(),
		"llm":      h.llm.Stats.Snapshot(),
	})
}
