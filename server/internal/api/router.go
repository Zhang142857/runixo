package api

import (
	"net/http"
	"strings"

	"github.com/runixo/runixo/server/internal/config"
	"github.com/runixo/runixo/server/internal/middleware"
	"github.com/runixo/runixo/server/internal/store"
)

func NewRouter(cfg *config.Config, s *store.Store) http.Handler {
	mux := http.NewServeMux()

	update := NewUpdateHandler(s, cfg)
	plugin := NewPluginHandler(s)

	// Agent update check - matches agent's existing URL: /api/check?version=...
	mux.HandleFunc("/api/check", update.Check)

	// Plugin marketplace
	mux.HandleFunc("/api/v1/plugins", plugin.List)
	mux.HandleFunc("/api/v1/plugins/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/download") {
			plugin.Download(w, r)
		} else {
			plugin.Detail(w, r)
		}
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Apply middleware: security -> rate limit -> routes
	rl := middleware.NewRateLimiter(
		cfg.RateLimit.RequestsPerMinute,
		cfg.RateLimit.BurstSize,
		cfg.RateLimit.BanThreshold,
		cfg.RateLimit.BanDurationMinutes,
	)
	return rl.Middleware(middleware.Security(mux))
}
