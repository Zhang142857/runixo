package middleware

import (
	"net/http"
	"strings"
)

// Security adds security headers and basic request validation.
func Security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'none'")
		w.Header().Set("Referrer-Policy", "no-referrer")

		// Only allow GET for API endpoints
		if strings.HasPrefix(r.URL.Path, "/api/") && r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		// Reject oversized URLs (path traversal / fuzzing)
		if len(r.URL.Path) > 512 || len(r.URL.RawQuery) > 1024 {
			http.Error(w, `{"error":"request too large"}`, http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
