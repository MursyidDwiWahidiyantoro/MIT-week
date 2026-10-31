package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/amba/amba/backend/internal/config"
)

// NewRouter builds the API router. Routes are added per phase; the envelope,
// headers and recovery apply to everything mounted here.
func NewRouter(cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(securityHeaders)
	r.Use(recoverer)

	r.NotFound(notFound)
	r.MethodNotAllowed(methodNotAllowed)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})
	})

	return r
}
