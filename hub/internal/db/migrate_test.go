package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrate_AddsColumns(t *testing.T) {
	conn, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()

	cols := schemaColumns(t, conn, "projects")
	for _, want := range []string{"source_type", "repo_url", "registry"} {
		if !cols[want] {
			t.Errorf("column %q missing from projects after migration", want)
		}
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	conn, err := Open(path)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	conn.Close()

	conn2, err := Open(path)
	if err != nil {
		t.Fatalf("second open (should be idempotent): %v", err)
	}
	conn2.Close()
}

func TestMigrate_ExistingRowsGetDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "v0.db")

	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open raw: %v", err)
	}
	if _, err := raw.Exec(schema); err != nil {
		t.Fatalf("apply base schema: %v", err)
	}
	if _, err := raw.Exec(
		`INSERT INTO projects (name, repo_path) VALUES ('myproject', '/tmp/myproject')`,
	); err != nil {
		t.Fatalf("insert v0 row: %v", err)
	}
	raw.Close()

	conn, err := Open(path)
	if err != nil {
		t.Fatalf("open with migration: %v", err)
	}
	defer conn.Close()

	var sourceType, repoURL, registry string
	err = conn.QueryRow(
		`SELECT source_type, repo_url, registry FROM projects WHERE name='myproject'`,
	).Scan(&sourceType, &repoURL, &registry)
	if err != nil {
		t.Fatalf("query migrated row: %v", err)
	}
	if sourceType != "clone" {
		t.Errorf("source_type = %q, want 'clone'", sourceType)
	}
	if repoURL != "" {
		t.Errorf("repo_url = %q, want ''", repoURL)
	}
	if registry != "" {
		t.Errorf("registry = %q, want ''", registry)
	}
}

func schemaColumns(t *testing.T, db *sql.DB, table string) map[string]bool {
	t.Helper()
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		t.Fatalf("PRAGMA table_info: %v", err)
	}
	defer rows.Close()
	cols := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan column info: %v", err)
		}
		cols[name] = true
	}
	return cols
}
