package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/reply"
	"github.com/amba/amba/backend/internal/token"
)

// verify handles POST /verify. The token arrives in the body, never the query,
// so it never reaches server logs.
func (h *handlers) verify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<10)).Decode(&body); err != nil || body.Token == "" {
		h.metrics.Inc("amba_verifications_total", "result", "invalid")
		writeJSON(w, http.StatusOK, map[string]any{"valid": false})
		return
	}

	payload, err := token.Verify(h.signingKey, body.Token)
	if err != nil {
		h.metrics.Inc("amba_verifications_total", "result", "invalid")
		writeJSON(w, http.StatusOK, map[string]any{"valid": false})
		return
	}

	h.metrics.Inc("amba_verifications_total", "result", "valid")
	writeJSON(w, http.StatusOK, map[string]any{
		"valid": true,
		"payload": map[string]any{
			"id":     payload.ID,
			"ts":     payload.TS,
			"status": payload.Status,
			"score":  payload.Score,
			"title":  payload.Title,
			"claim":  payload.Claim,
			"srcs":   payload.Srcs,
		},
	})
}

// reply handles POST /reply: token-only input, so no raw user content is ever
// accepted here. A model draft is returned when it passes validation;
// otherwise the deterministic template (TECH_SPEC §6.2).
func (h *handlers) reply(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
		Tone  string `json:"tone"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<10)).Decode(&body); err != nil {
		writeError(w, errInvalidReply)
		return
	}
	if !reply.ValidTone(body.Tone) {
		writeError(w, errInvalidReply)
		return
	}

	payload, err := token.Verify(h.signingKey, body.Token)
	if err != nil {
		writeError(w, newAPIError(http.StatusBadRequest, models.CodeInvalidURL, "Laporan tidak sah.", false))
		return
	}

	// A model call shares the audit concurrency cap; when it is full the
	// template answers immediately instead of queueing.
	text, source := reply.Template(payload.Status, body.Tone), "template"
	if h.limiter.AcquireInFlight() {
		text, source = h.replies.Draft(r.Context(), body.Token, payload, body.Tone)
		h.limiter.ReleaseInFlight()
	}
	h.metrics.Inc("amba_replies_total", "source", source)
	writeJSON(w, http.StatusOK, map[string]string{"text": text, "source": source})
}
