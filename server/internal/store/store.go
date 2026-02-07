package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

// --- Models ---

type AgentRelease struct {
	Version     string `json:"version"`
	Channel     string `json:"channel"` // stable, beta, nightly
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	DownloadURL string `json:"download_url"`
	Checksum    string `json:"checksum"`
	Size        int64  `json:"size"`
	ReleaseNote string `json:"release_note"`
	ReleasedAt  string `json:"released_at"`
}

type Plugin struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Version     string `json:"version"`
	Type        string `json:"type"` // client, agent, hybrid
	MinAgent    string `json:"min_agent_version"`
	DownloadURL string `json:"download_url"`
	Checksum    string `json:"checksum"`
	Size        int64  `json:"size"`
	Downloads   int64  `json:"downloads"`
	Icon        string `json:"icon"`
	Tags        string `json:"tags"` // comma-separated
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// --- Init ---

func New(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	return s, s.migrate()
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS agent_releases (
			version TEXT NOT NULL,
			channel TEXT NOT NULL DEFAULT 'stable',
			os TEXT NOT NULL,
			arch TEXT NOT NULL,
			download_url TEXT NOT NULL,
			checksum TEXT NOT NULL,
			size INTEGER DEFAULT 0,
			release_note TEXT DEFAULT '',
			released_at TEXT NOT NULL,
			PRIMARY KEY (version, os, arch)
		);
		CREATE TABLE IF NOT EXISTS plugins (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			author TEXT DEFAULT '',
			version TEXT NOT NULL,
			type TEXT DEFAULT 'hybrid',
			min_agent_version TEXT DEFAULT '',
			download_url TEXT NOT NULL,
			checksum TEXT DEFAULT '',
			size INTEGER DEFAULT 0,
			downloads INTEGER DEFAULT 0,
			icon TEXT DEFAULT '',
			tags TEXT DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_releases_channel ON agent_releases(channel);
		CREATE INDEX IF NOT EXISTS idx_plugins_tags ON plugins(tags);
	`)
	return err
}

// --- Agent Releases ---

func (s *Store) GetLatestRelease(osName, arch, channel string) (*AgentRelease, error) {
	if channel == "" {
		channel = "stable"
	}
	r := &AgentRelease{}
	err := s.db.QueryRow(`
		SELECT version, channel, os, arch, download_url, checksum, size, release_note, released_at
		FROM agent_releases WHERE os = ? AND arch = ? AND channel = ?
		ORDER BY released_at DESC LIMIT 1
	`, osName, arch, channel).Scan(&r.Version, &r.Channel, &r.OS, &r.Arch, &r.DownloadURL, &r.Checksum, &r.Size, &r.ReleaseNote, &r.ReleasedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func (s *Store) UpsertRelease(r *AgentRelease) error {
	_, err := s.db.Exec(`
		INSERT INTO agent_releases (version, channel, os, arch, download_url, checksum, size, release_note, released_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(version, os, arch) DO UPDATE SET
			channel=excluded.channel, download_url=excluded.download_url,
			checksum=excluded.checksum, size=excluded.size,
			release_note=excluded.release_note, released_at=excluded.released_at
	`, r.Version, r.Channel, r.OS, r.Arch, r.DownloadURL, r.Checksum, r.Size, r.ReleaseNote, r.ReleasedAt)
	return err
}

// --- Plugins ---

func (s *Store) ListPlugins(search string, limit, offset int) ([]Plugin, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var total int
	query := "FROM plugins"
	args := []any{}
	if search != "" {
		query += " WHERE name LIKE ? OR description LIKE ? OR tags LIKE ?"
		pat := "%" + search + "%"
		args = append(args, pat, pat, pat)
	}
	if err := s.db.QueryRow("SELECT COUNT(*) "+query, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.Query(
		fmt.Sprintf("SELECT id, name, description, author, version, type, min_agent_version, download_url, checksum, size, downloads, icon, tags, created_at, updated_at %s ORDER BY downloads DESC LIMIT ? OFFSET ?", query),
		append(args, limit, offset)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var plugins []Plugin
	for rows.Next() {
		var p Plugin
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Author, &p.Version, &p.Type, &p.MinAgent, &p.DownloadURL, &p.Checksum, &p.Size, &p.Downloads, &p.Icon, &p.Tags, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		plugins = append(plugins, p)
	}
	return plugins, total, nil
}

func (s *Store) GetPlugin(id string) (*Plugin, error) {
	p := &Plugin{}
	err := s.db.QueryRow(`
		SELECT id, name, description, author, version, type, min_agent_version, download_url, checksum, size, downloads, icon, tags, created_at, updated_at
		FROM plugins WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &p.Description, &p.Author, &p.Version, &p.Type, &p.MinAgent, &p.DownloadURL, &p.Checksum, &p.Size, &p.Downloads, &p.Icon, &p.Tags, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (s *Store) IncrementDownloads(id string) error {
	_, err := s.db.Exec("UPDATE plugins SET downloads = downloads + 1 WHERE id = ?", id)
	return err
}

func (s *Store) UpsertPlugin(p *Plugin) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if p.CreatedAt == "" {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	_, err := s.db.Exec(`
		INSERT INTO plugins (id, name, description, author, version, type, min_agent_version, download_url, checksum, size, downloads, icon, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name, description=excluded.description, author=excluded.author,
			version=excluded.version, type=excluded.type, min_agent_version=excluded.min_agent_version,
			download_url=excluded.download_url, checksum=excluded.checksum, size=excluded.size,
			icon=excluded.icon, tags=excluded.tags, updated_at=excluded.updated_at
	`, p.ID, p.Name, p.Description, p.Author, p.Version, p.Type, p.MinAgent, p.DownloadURL, p.Checksum, p.Size, p.Downloads, p.Icon, p.Tags, p.CreatedAt, p.UpdatedAt)
	return err
}
