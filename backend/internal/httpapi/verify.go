package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/amba/amba/backend/internal/models"
	"github.com/amba/amba/backend/internal/token"
)

// verify handles POST /verify. The token arrives in the body, never the query,
// so it never reaches server logs.
func (h *handlers) verify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<10)).Decode(&body); err != nil || body.Token == "" {
		writeJSON(w, http.StatusOK, map[string]any{"valid": false})
		return
	}

	payload, err := token.Verify(h.signingKey, body.Token)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"valid": false})
		return
	}

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
// accepted here.
func (h *handlers) reply(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
		Tone  string `json:"tone"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<10)).Decode(&body); err != nil {
		writeError(w, errInvalidURL)
		return
	}
	if body.Tone != "hormat" && body.Tone != "santai" && body.Tone != "singkat" {
		writeError(w, errInvalidURL)
		return
	}

	payload, err := token.Verify(h.signingKey, body.Token)
	if err != nil {
		writeError(w, newAPIError(http.StatusBadRequest, models.CodeInvalidURL, "Laporan tidak sah.", false))
		return
	}

	draft := h.replies.Draft(r.Context(), payload, body.Tone, body.Token)
	writeJSON(w, http.StatusOK, draft)
}
