package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/amba/amba/backend/internal/demo"
)

// getDemo serves a bundled example dossier. Demo data is never produced for a
// live audit (CLAUDE.md rule 4); this route is the only way it reaches a client.
func getDemo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := demo.Get(id)
	if err != nil {
		if errors.Is(err, demo.ErrNotFound) {
			writeError(w, newAPIError(http.StatusNotFound, "INVALID_URL", "Contoh tidak ditemukan.", false))
			return
		}
		writeError(w, errInternal)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(c.Dossier)
}
