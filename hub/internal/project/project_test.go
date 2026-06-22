package project

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/pecodez/couchdev/internal/db"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestDetectRegistry_GitHub(t *testing.T) {
	if got := detectRegistry("git@github.com:org/repo.git"); got != "github" {
		t.Errorf("got %q, want 'github'", got)
	}
}

func TestDetectRegistry_GitLab(t *testing.T) {
	if got := detectRegistry("https://gitlab.com/org/repo.git"); got != "gitlab" {
		t.Errorf("got %q, want 'gitlab'", got)
	}
}

func TestDetectRegistry_Custom(t *testing.T) {
	if got := detectRegistry("https://git.company.io/org/repo.git"); got != "custom" {
		t.Errorf("got %q, want 'custom'", got)
	}
}

func TestProjectStore_CreateAndList(t *testing.T) {
	s := NewStore(openTestDB(t))
	p := Project{Name: "alpha", RepoPath: "/tmp/alpha", SourceType: "greenfield"}

	created, err := s.Create(p)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected non-zero ID")
	}

	list, err := s.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 project, got %d", len(list))
	}
	if list[0].Name != "alpha" {
		t.Errorf("name = %q, want 'alpha'", list[0].Name)
	}
}

func TestProjectStore_DuplicateNameErrors(t *testing.T) {
	s := NewStore(openTestDB(t))
	p := Project{Name: "dup", RepoPath: "/tmp/dup"}

	if _, err := s.Create(p); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if _, err := s.Create(p); err == nil {
		t.Fatal("expected error on duplicate name, got nil")
	}
}

func TestProjectStore_GetByName(t *testing.T) {
	s := NewStore(openTestDB(t))
	p := Project{Name: "beta", RepoPath: "/tmp/beta", SourceType: "clone", RepoURL: "git@github.com:o/r.git", Registry: "github"}

	if _, err := s.Create(p); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := s.GetByName("beta")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.RepoURL != "git@github.com:o/r.git" {
		t.Errorf("repo_url = %q, want git@github.com:o/r.git", got.RepoURL)
	}
	if got.Registry != "github" {
		t.Errorf("registry = %q, want 'github'", got.Registry)
	}
}

func TestProjectStore_GetByNameNotFound(t *testing.T) {
	s := NewStore(openTestDB(t))

	_, err := s.GetByName("ghost")
	if err == nil {
		t.Fatal("expected error for missing project, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error %q does not contain 'not found'", err.Error())
	}
}
