package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/runixo/runixo/server/internal/store"
)

type PluginHandler struct {
	store *store.Store
}

func NewPluginHandler(s *store.Store) *PluginHandler {
	return &PluginHandler{store: s}
}

type PluginListResponse struct {
	Plugins []store.Plugin `json:"plugins"`
	Total   int            `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
}

// GET /api/v1/plugins?search=xxx&limit=20&offset=0
func (h *PluginHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := q.Get("search")
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	plugins, total, err := h.store.ListPlugins(search, limit, offset)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if plugins == nil {
		plugins = []store.Plugin{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PluginListResponse{
		Plugins: plugins, Total: total, Limit: limit, Offset: offset,
	})
}

// GET /api/v1/plugins/{id}
func (h *PluginHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id := extractPathParam(r.URL.Path, "/api/v1/plugins/")
	if id == "" || strings.Contains(id, "/") {
		jsonError(w, "invalid plugin id", http.StatusBadRequest)
		return
	}

	plugin, err := h.store.GetPlugin(id)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if plugin == nil {
		jsonError(w, "plugin not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plugin)
}

// GET /api/v1/plugins/{id}/download -> 302 redirect to GitHub/CDN
func (h *PluginHandler) Download(w http.ResponseWriter, r *http.Request) {
	// Extract id from /api/v1/plugins/{id}/download
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/plugins/")
	id := strings.TrimSuffix(path, "/download")
	if id == "" || strings.Contains(id, "/") {
		jsonError(w, "invalid plugin id", http.StatusBadRequest)
		return
	}

	plugin, err := h.store.GetPlugin(id)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if plugin == nil {
		jsonError(w, "plugin not found", http.StatusNotFound)
		return
	}

	// Increment download counter (fire and forget)
	go h.store.IncrementDownloads(id)

	// 302 redirect - our server doesn't serve the file, GitHub/CDN does
	http.Redirect(w, r, plugin.DownloadURL, http.StatusFound)
}

// --- helpers ---

func extractPathParam(path, prefix string) string {
	s := strings.TrimPrefix(path, prefix)
	if i := strings.Index(s, "/"); i >= 0 {
		return s[:i]
	}
	return s
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
