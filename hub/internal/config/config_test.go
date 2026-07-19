package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pecodez/couchdev/internal/config"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestLoad_DefaultsRequireAuthFalse(t *testing.T) {
	path := writeConfig(t, `{"db_path": "db.sqlite"}`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RequireAuth {
		t.Errorf("expected require_auth to default to false, got true")
	}
}

func TestLoad_DefaultsListenAddr(t *testing.T) {
	path := writeConfig(t, `{"db_path": "db.sqlite"}`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected default listen_addr :8080, got %q", cfg.ListenAddr)
	}
}

func TestLoad_TokenHashOptionalWhenAuthDisabled(t *testing.T) {
	path := writeConfig(t, `{"db_path": "db.sqlite", "require_auth": false}`)
	if _, err := config.Load(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoad_TokenHashRequiredWhenAuthEnabled(t *testing.T) {
	path := writeConfig(t, `{"db_path": "db.sqlite", "require_auth": true}`)
	if _, err := config.Load(path); err == nil {
		t.Fatal("expected error when require_auth is true and token_hash is missing, got nil")
	}
}

func TestLoad_TokenHashPresentWhenAuthEnabled(t *testing.T) {
	path := writeConfig(t, `{"db_path": "db.sqlite", "require_auth": true, "token_hash": "abc123"}`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TokenHash != "abc123" {
		t.Errorf("expected token_hash %q, got %q", "abc123", cfg.TokenHash)
	}
}

func TestLoad_MissingDBPath(t *testing.T) {
	path := writeConfig(t, `{}`)
	if _, err := config.Load(path); err == nil {
		t.Fatal("expected error when db_path is missing, got nil")
	}
}
