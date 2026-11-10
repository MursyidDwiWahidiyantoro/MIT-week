package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/amba/amba/backend/internal/config"
	"github.com/amba/amba/backend/internal/security"
)

// NewRouter builds the API router.
func NewRouter(cfg *config.Config) http.Handler {
	handler, _ := NewRouterWithMetrics(cfg)
	return handler
}

// NewRouterWithMetrics also returns the internal metrics handler, which the
// entrypoint may serve on a separate listener.
func NewRouterWithMetrics(cfg *config.Config) (http.Handler, http.Handler) {
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
	h := newHandlers(cfg, limiter)

	r.Use(cors.Middleware)
	r.Use(limiter.Middleware(func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, errRateLimited)
	}))

	r.NotFound(notFound)
	r.MethodNotAllowed(methodNotAllowed)

	r.Get("/healthz", health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthz", health)
		r.Get("/ops/status", h.getOpsStatus)
		r.Get("/demo/{id}", getDemo)
		r.Post("/audit/url", h.auditURL)
		r.Post("/audit/text", h.auditText)
		r.Post("/audit/file", h.auditFile)
		r.Post("/verify", h.verify)
		r.Post("/reply", h.reply)
	})

	metrics := chi.NewRouter()
	metrics.Get("/metrics", h.getMetrics)

	return r, metrics
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
