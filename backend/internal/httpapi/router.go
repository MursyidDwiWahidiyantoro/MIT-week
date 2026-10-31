package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/amba/amba/backend/internal/config"
	"github.com/amba/amba/backend/internal/security"
)

// NewRouter builds the API router. Routes are added per phase; the envelope,
// headers and recovery apply to everything mounted here.
func NewRouter(cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(security.APISecurityHeaders)
	r.Use(recoverer)

	limiter := security.NewRateLimiter(security.RateLimitConfig{
		PerMinute:      15,
		Burst:          5,
		MaxInFlight:    32,
		IdleTTL:        10 * time.Minute,
		SweepInterval:  5 * time.Minute,
		TrustedProxies: security.ParseTrustedProxies(cfg.TrustedProxy),
	})
	cors := security.NewCORSConfig(cfg.PublicAppURL, !cfg.IsProduction())

	r.Use(cors.Middleware)
	r.Use(limiter.Middleware(func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, errRateLimited)
	}))

	r.NotFound(notFound)
	r.MethodNotAllowed(methodNotAllowed)

	r.Get("/healthz", health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthz", health)
		r.Get("/demo/{id}", getDemo)
	})

	return r
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
