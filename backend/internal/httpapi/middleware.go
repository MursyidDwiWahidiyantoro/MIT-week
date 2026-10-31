package httpapi

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/amba/amba/backend/internal/models"
)

// securityHeaders sets the API's own headers. The API returns JSON only, so
// the policy can be maximally restrictive.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cache-Control", "no-store")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

// recoverer converts a panic into the generic error envelope. The stack is
// logged server-side only; the response never carries it.
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered", "path", r.URL.Path)
				debug.PrintStack()
				writeError(w, errInternal)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// notFound returns the envelope shape for unknown routes, so clients never
// have to parse two different error formats.
func notFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, newAPIError(http.StatusNotFound, models.CodeInvalidURL, "Alamat tidak ditemukan.", false))
}

// methodNotAllowed mirrors the same shape.
func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	writeError(w, newAPIError(http.StatusMethodNotAllowed, models.CodeInvalidURL, "Metode tidak didukung.", false))
}
