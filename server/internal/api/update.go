package api

import (
	"encoding/json"
	"net/http"

	"github.com/runixo/runixo/server/internal/config"
	"github.com/runixo/runixo/server/internal/store"
)

type UpdateHandler struct {
	store *store.Store
	cfg   *config.Config
}

func NewUpdateHandler(s *store.Store, cfg *config.Config) *UpdateHandler {
	return &UpdateHandler{store: s, cfg: cfg}
}

// UpdateResponse matches agent's UpdateInfo struct
type UpdateResponse struct {
	Available    bool   `json:"available"`
	LatestVersion string `json:"latest_version"`
	ReleaseNotes string `json:"release_notes"`
	DownloadURL  string `json:"download_url"`
	Size         int64  `json:"size"`
	Checksum     string `json:"checksum"`
	ReleaseDate  string `json:"release_date"`
	IsCritical   bool   `json:"is_critical"`
}

// GET /api/check?version=x.x.x&channel=stable&os=linux&arch=amd64
// Compatible with agent's existing updater URL pattern
func (h *UpdateHandler) Check(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	currentVer := q.Get("version")
	channel := q.Get("channel")
	osName := q.Get("os")
	arch := q.Get("arch")

	if currentVer == "" || osName == "" || arch == "" {
		jsonError(w, "missing required params: version, os, arch", http.StatusBadRequest)
		return
	}

	release, err := h.store.GetLatestRelease(osName, arch, channel)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if release == nil {
		json.NewEncoder(w).Encode(UpdateResponse{Available: false})
		return
	}

	// Build download URL: prefer CDN, fallback to stored URL
	downloadURL := release.DownloadURL
	if h.cfg.CDN.BaseURL != "" {
		downloadURL = h.cfg.CDN.BaseURL + "/" + release.Version + "/runixo-agent-" + osName + "-" + arch
	}

	resp := UpdateResponse{
		Available:     release.Version != currentVer,
		LatestVersion: release.Version,
		ReleaseNotes:  release.ReleaseNote,
		DownloadURL:   downloadURL,
		Size:          release.Size,
		Checksum:      release.Checksum,
		ReleaseDate:   release.ReleasedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
