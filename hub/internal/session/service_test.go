package session_test

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/pecodez/couchdev/internal/db"
	"github.com/pecodez/couchdev/internal/git"
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
	svc := session.NewService(ps, ss, tmux.NewMockDying(), &git.Mock{}, zap.NewNop())

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
	svc := session.NewService(ps, ss, tmux.NewMock(), &git.Mock{}, zap.NewNop())

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
	svc := session.NewService(ps, ss, mock, &git.Mock{}, zap.NewNop())

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
	if !strings.Contains(mock.LastCmd, "--rc") {
		t.Errorf("shell cmd %q missing --rc flag", mock.LastCmd)
	}
}

func TestGenesis_CreatesWorktreeAndStoresFields(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	gm := &git.Mock{}
	svc := session.NewService(ps, ss, tmux.NewMock(), gm, zap.NewNop())

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj/source"}); err != nil {
		t.Fatalf("create project: %v", err)
	}

	got, err := svc.Genesis("proj", "feat/my-feature", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Branch != "feat/my-feature" {
		t.Errorf("branch = %q, want feat/my-feature", got.Branch)
	}
	if got.WorktreePath != "/tmp/proj/worktrees/feat/my-feature" {
		t.Errorf("worktree_path = %q, want /tmp/proj/worktrees/feat/my-feature", got.WorktreePath)
	}
	if got.CWD != got.WorktreePath {
		t.Errorf("cwd %q should equal worktree_path %q", got.CWD, got.WorktreePath)
	}
	if gm.WorktreeAdded != got.WorktreePath {
		t.Errorf("worktree added at %q, want %q", gm.WorktreeAdded, got.WorktreePath)
	}
	if gm.FetchedRemote != "origin" || gm.FetchedBranch != "main" {
		t.Errorf("fetched (%q, %q), want (origin, main)", gm.FetchedRemote, gm.FetchedBranch)
	}
	if gm.WorktreeAddStartPoint != "origin/main" {
		t.Errorf("worktree add start point = %q, want origin/main", gm.WorktreeAddStartPoint)
	}
}

func TestGenesis_FetchFailureAbortsBeforeWorktreeAdd(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	gm := &git.Mock{FetchErr: fmt.Errorf("could not resolve host")}
	svc := session.NewService(ps, ss, tmux.NewMock(), gm, zap.NewNop())

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj/source"}); err != nil {
		t.Fatalf("create project: %v", err)
	}

	if _, err := svc.Genesis("proj", "s1", ""); err == nil {
		t.Fatal("expected error when fetch fails, got nil")
	}
	if gm.WorktreeAdded != "" {
		t.Errorf("expected WorktreeAdd not to be called after fetch failure, but worktree added at %q", gm.WorktreeAdded)
	}
	sessions, _ := ss.ListAll()
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions in DB after failed fetch, got %d", len(sessions))
	}
}

func TestGenesis_WorktreeFailureRollsBack(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	gm := &git.Mock{WorktreeAddErr: fmt.Errorf("branch already exists")}
	svc := session.NewService(ps, ss, tmux.NewMock(), gm, zap.NewNop())

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj/source"}); err != nil {
		t.Fatalf("create project: %v", err)
	}

	if _, err := svc.Genesis("proj", "s1", ""); err == nil {
		t.Fatal("expected error when worktree add fails, got nil")
	}
	sessions, _ := ss.ListAll()
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions in DB after failed worktree add, got %d", len(sessions))
	}
}

func TestChanges_ReturnsBranchAheadAndFiles(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	gm := &git.Mock{AheadCount: 2, Files: []string{"a.go", "b.go"}}
	svc := session.NewService(ps, ss, tmux.NewMock(), gm, zap.NewNop())

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj/source", DefaultBranch: "main"}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := svc.Genesis("proj", "fix-x", ""); err != nil {
		t.Fatalf("genesis: %v", err)
	}

	ahead, files, err := svc.Changes("proj", "fix-x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ahead != 2 {
		t.Errorf("ahead = %d, want 2", ahead)
	}
	if len(files) != 2 {
		t.Errorf("files = %v, want [a.go b.go]", files)
	}
}

func TestTeardown_MergedAndClean_RemovesWorktree(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	gm := &git.Mock{CleanResult: true, MergedResult: true}
	svc := session.NewService(ps, ss, tmux.NewMock(), gm, zap.NewNop())

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj/source"}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := svc.Genesis("proj", "s1", ""); err != nil {
		t.Fatalf("genesis: %v", err)
	}

	deleted, reason, err := svc.Teardown("proj", "s1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatalf("expected deleted = true, reason = %q", reason)
	}
	if gm.WorktreeRemoved == "" {
		t.Error("expected worktree to be removed")
	}

	sess, err := svc.Status("proj", "s1")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if sess.State != session.StateDead {
		t.Errorf("state = %q, want dead", sess.State)
	}
}

func TestTeardown_NotMerged_PreservesWorktree(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	gm := &git.Mock{CleanResult: true, MergedResult: false}
	svc := session.NewService(ps, ss, tmux.NewMock(), gm, zap.NewNop())

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj/source"}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := svc.Genesis("proj", "s1", ""); err != nil {
		t.Fatalf("genesis: %v", err)
	}

	deleted, reason, err := svc.Teardown("proj", "s1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted {
		t.Fatal("expected deleted = false when branch is not merged")
	}
	if reason == "" {
		t.Error("expected a non-empty reason")
	}
	if gm.WorktreeRemoved != "" {
		t.Error("expected worktree to be preserved")
	}

	sess, err := svc.Status("proj", "s1")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if sess.State != session.StateResumable {
		t.Errorf("state = %q, want resumable", sess.State)
	}
}

func TestTeardown_Dirty_PreservesWorktree(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	gm := &git.Mock{CleanResult: false, MergedResult: true}
	svc := session.NewService(ps, ss, tmux.NewMock(), gm, zap.NewNop())

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj/source"}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := svc.Genesis("proj", "s1", ""); err != nil {
		t.Fatalf("genesis: %v", err)
	}

	deleted, reason, err := svc.Teardown("proj", "s1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted {
		t.Fatal("expected deleted = false when worktree is dirty")
	}
	if reason == "" {
		t.Error("expected a non-empty reason")
	}
	if gm.WorktreeRemoved != "" {
		t.Error("expected worktree to be preserved")
	}
}

func TestTeardown_Force_RemovesWorktreeEvenWhenNotMerged(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	gm := &git.Mock{CleanResult: false, MergedResult: false}
	svc := session.NewService(ps, ss, tmux.NewMock(), gm, zap.NewNop())

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj/source"}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := svc.Genesis("proj", "s1", ""); err != nil {
		t.Fatalf("genesis: %v", err)
	}

	deleted, _, err := svc.Teardown("proj", "s1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatal("expected deleted = true when force is set")
	}
	if gm.WorktreeRemoved == "" {
		t.Error("expected worktree to be removed when forced")
	}
}

func TestTeardown_SyncsBeforeMergeCheck_MergedOnRemote(t *testing.T) {
	// Simulates a branch merged on GitHub: the local default-branch ref
	// wouldn't show it merged, but a fresh fetch of origin/main would. The
	// mock can't fork behavior by ref, so this asserts the fix's actual
	// mechanism: Teardown fetches before checking, and checks against the
	// "origin/<default-branch>" ref rather than the bare local branch name.
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	gm := &git.Mock{CleanResult: true, MergedResult: true}
	svc := session.NewService(ps, ss, tmux.NewMock(), gm, zap.NewNop())

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj/source", DefaultBranch: "main"}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := svc.Genesis("proj", "s1", ""); err != nil {
		t.Fatalf("genesis: %v", err)
	}

	if _, _, err := svc.Teardown("proj", "s1", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gm.MergedDefaultBranch != "origin/main" {
		t.Errorf("IsFullyMerged called with defaultBranch = %q, want origin/main", gm.MergedDefaultBranch)
	}
	if gm.FetchedRemote != "origin" || gm.FetchedBranch != "main" {
		t.Errorf("expected a fetch of origin/main before the merge check, got (%q, %q)", gm.FetchedRemote, gm.FetchedBranch)
	}
}

func TestTeardown_FetchFailure_FallsBackToLocalRef(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	gm := &git.Mock{CleanResult: true, MergedResult: true}
	svc := session.NewService(ps, ss, tmux.NewMock(), gm, zap.NewNop())

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj/source", DefaultBranch: "main"}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := svc.Genesis("proj", "s1", ""); err != nil {
		t.Fatalf("genesis: %v", err)
	}
	// Fetch only starts failing after Genesis (which has its own, unrelated
	// fetch call already covered by TestGenesis_FetchFailureAbortsBeforeWorktreeAdd);
	// here we only care that Teardown degrades gracefully when *its* fetch fails.
	gm.FetchErr = fmt.Errorf("could not resolve host")

	deleted, _, err := svc.Teardown("proj", "s1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatal("expected teardown to still succeed by falling back to the local ref")
	}
	if gm.MergedDefaultBranch != "main" {
		t.Errorf("IsFullyMerged called with defaultBranch = %q, want local ref \"main\" as fallback", gm.MergedDefaultBranch)
	}
}

func TestTeardown_LocalOnlyProject_SkipsFetch(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	gm := &git.Mock{CleanResult: true, MergedResult: true, NoRemote: true}
	svc := session.NewService(ps, ss, tmux.NewMock(), gm, zap.NewNop())

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj/source", DefaultBranch: "main"}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := svc.Genesis("proj", "s1", ""); err != nil {
		t.Fatalf("genesis: %v", err)
	}

	if _, _, err := svc.Teardown("proj", "s1", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gm.MergedDefaultBranch != "main" {
		t.Errorf("IsFullyMerged called with defaultBranch = %q, want local ref \"main\" for a remote-less project", gm.MergedDefaultBranch)
	}
	if gm.FetchCalls != 0 {
		t.Errorf("expected no fetch for a local-only project, got %d calls", gm.FetchCalls)
	}
}

func TestChanges_ComparesAgainstOriginRef(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	ss := session.NewStore(conn)
	gm := &git.Mock{AheadCount: 2, Files: []string{"a.go", "b.go"}}
	svc := session.NewService(ps, ss, tmux.NewMock(), gm, zap.NewNop())

	if _, err := ps.Create(project.Project{Name: "proj", RepoPath: "/tmp/proj/source", DefaultBranch: "main"}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := svc.Genesis("proj", "fix-x", ""); err != nil {
		t.Fatalf("genesis: %v", err)
	}

	if _, _, err := svc.Changes("proj", "fix-x"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gm.AheadBase != "origin/main" {
		t.Errorf("CommitsAhead called with base = %q, want origin/main", gm.AheadBase)
	}
}

func TestSyncAllProjects_FetchesEachRemoteProjectAndContinuesPastErrors(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	gm := &git.Mock{FetchErr: fmt.Errorf("could not resolve host")}

	if _, err := ps.Create(project.Project{Name: "a", RepoPath: "/tmp/a", DefaultBranch: "main"}); err != nil {
		t.Fatalf("create project a: %v", err)
	}
	if _, err := ps.Create(project.Project{Name: "b", RepoPath: "/tmp/b", DefaultBranch: "main"}); err != nil {
		t.Fatalf("create project b: %v", err)
	}

	session.SyncAllProjects(ps, gm, zap.NewNop())

	if gm.FetchCalls != 2 {
		t.Errorf("expected a fetch attempt per project despite errors, got %d calls", gm.FetchCalls)
	}
}

func TestSyncAllProjects_SkipsLocalOnlyProjects(t *testing.T) {
	conn := openTestDB(t)
	ps := project.NewStore(conn)
	gm := &git.Mock{NoRemote: true}

	if _, err := ps.Create(project.Project{Name: "a", RepoPath: "/tmp/a", DefaultBranch: "main"}); err != nil {
		t.Fatalf("create project a: %v", err)
	}

	session.SyncAllProjects(ps, gm, zap.NewNop())

	if gm.FetchCalls != 0 {
		t.Errorf("expected no fetch for a local-only project, got %d calls", gm.FetchCalls)
	}
}
