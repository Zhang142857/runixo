package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	GitHub    GitHubConfig    `yaml:"github"`
	CDN       CDNConfig       `yaml:"cdn"`
}

type ServerConfig struct {
	Port            int    `yaml:"port"`
	TrustedProxies  []string `yaml:"trusted_proxies"`
	ReadTimeout     int    `yaml:"read_timeout_seconds"`
	WriteTimeout    int    `yaml:"write_timeout_seconds"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	BurstSize         int `yaml:"burst_size"`
	BanThreshold      int `yaml:"ban_threshold"`
	BanDurationMinutes int `yaml:"ban_duration_minutes"`
}

type GitHubConfig struct {
	Owner string `yaml:"owner"`
	Repo  string `yaml:"repo"`
}

type CDNConfig struct {
	BaseURL string `yaml:"base_url"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	cfg.setDefaults()
	return cfg, nil
}

func (c *Config) setDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 10
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 10
	}
	if c.Database.Path == "" {
		c.Database.Path = "data/runixo.db"
	}
	if c.RateLimit.RequestsPerMinute == 0 {
		c.RateLimit.RequestsPerMinute = 60
	}
	if c.RateLimit.BurstSize == 0 {
		c.RateLimit.BurstSize = 10
	}
	if c.RateLimit.BanThreshold == 0 {
		c.RateLimit.BanThreshold = 300
	}
	if c.RateLimit.BanDurationMinutes == 0 {
		c.RateLimit.BanDurationMinutes = 30
	}
	if c.GitHub.Owner == "" {
		c.GitHub.Owner = "runixo"
	}
	if c.GitHub.Repo == "" {
		c.GitHub.Repo = "runixo"
	}
}
