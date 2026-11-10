package httpapi

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/amba/amba/backend/internal/models"
)

// recoverer converts a panic into the generic error envelope. The stack is
// logged server-side only; the response never carries it.
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// Never log the request path: a client can put arbitrary text in
				// it, and rule 2 forbids logging anything derived from user
				// input. The method plus a fixed marker is enough to correlate.
				slog.Error("panic recovered", "method", r.Method)
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
