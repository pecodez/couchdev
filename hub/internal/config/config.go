package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	ListenAddr  string `json:"listen_addr"`
	TLSCert     string `json:"tls_cert"`
	TLSKey      string `json:"tls_key"`
	RequireAuth bool   `json:"require_auth"`
	TokenHash   string `json:"token_hash"` // SHA-256 of bearer token, hex-encoded (64 chars)
	DBPath      string `json:"db_path"`
	ProjectsDir string `json:"projects_dir"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var c Config
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	if c.RequireAuth && c.TokenHash == "" {
		return nil, fmt.Errorf("config: token_hash required when require_auth is true")
	}
	if c.DBPath == "" {
		return nil, fmt.Errorf("config: db_path required")
	}
	if c.ListenAddr == "" {
		c.ListenAddr = ":8080"
	}
	if c.ProjectsDir == "" {
		c.ProjectsDir = "projects"
	}
	return &c, nil
}
