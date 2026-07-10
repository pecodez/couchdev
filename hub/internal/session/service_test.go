package session_test

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/pecodez/couchdev/internal/db"
	"github.com/pecodez/couchdev/internal/project"
	"github.com/pecodez/couchdev/internal/session"
	"github.com/pecodez/couchdev/internal/tmux"
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

func TestGenesis_SessionDiesImmediately(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	svc := session.NewService(ps, ss, tmux.NewMockDying())

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj"}); err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err := svc.Genesis("proj", "s1", "")
	if err == nil {
		t.Fatal("expected error when spawned session exits immediately, got nil")
	}

	sessions, _ := ss.ListAll()
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions in DB after failed spawn, got %d", len(sessions))
	}
}

func TestGenesis_PersistsWhenSessionStaysAlive(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	svc := session.NewService(ps, ss, tmux.NewMock())

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj"}); err != nil {
		t.Fatalf("create project: %v", err)
	}

	got, err := svc.Genesis("proj", "s1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.CanonicalName != "proj/s1" {
		t.Errorf("canonical_name = %q, want proj/s1", got.CanonicalName)
	}

	sessions, _ := ss.ListAll()
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session in DB, got %d", len(sessions))
	}
}

func TestGenesis_ShellCmdContainsClaude(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	mock := tmux.NewMock()
	svc := session.NewService(ps, ss, mock)

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj"}); err != nil {
		t.Fatalf("create project: %v", err)
	}

	if _, err := svc.Genesis("proj", "s1", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(mock.LastCmd, "claude") {
		t.Errorf("shell cmd %q does not invoke claude", mock.LastCmd)
	}
	if !strings.Contains(mock.LastCmd, "proj/s1") {
		t.Errorf("shell cmd %q does not contain canonical name proj/s1", mock.LastCmd)
	}
	if !strings.Contains(mock.LastCmd, "--dangerously-skip-permissions") {
		t.Errorf("shell cmd %q missing --dangerously-skip-permissions (required to bypass directory trust prompt)", mock.LastCmd)
	}
}
