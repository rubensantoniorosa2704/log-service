package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeaders applies HTTP security headers globally (ADR-001).
//
// Exceptions:
//   - /swagger/* and /docs/*: CSP omitted so Swagger UI remains functional.
//   - /api/v1/events/*: Cache-Control omitted; the SSE library manages it.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		h := w.Header()

		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		if !strings.HasPrefix(path, "/swagger/") && !strings.HasPrefix(path, "/docs/") {
			h.Set("Content-Security-Policy", "default-src 'none'")
		}

		if !strings.HasPrefix(path, "/api/v1/events/") {
			h.Set("Cache-Control", "no-store")
		}

		next.ServeHTTP(w, r)
	})
}
