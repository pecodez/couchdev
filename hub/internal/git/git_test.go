package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pecodez/couchdev/internal/git"
)

func runGit(t *testing.T, repoDir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repoDir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

func writeFile(t *testing.T, repoDir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(repoDir, name), []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func setupRepo(t *testing.T) string {
	t.Helper()
	repoDir := t.TempDir()
	runGit(t, repoDir, "init", "-b", "main")
	writeFile(t, repoDir, "a.txt", "a")
	runGit(t, repoDir, "add", "a.txt")
	runGit(t, repoDir, "commit", "-m", "init")
	return repoDir
}

func TestReal_IsClean(t *testing.T) {
	repoDir := setupRepo(t)
	g := git.Real{}

	clean, err := g.IsClean(repoDir)
	if err != nil {
		t.Fatalf("IsClean: %v", err)
	}
	if !clean {
		t.Error("expected clean worktree right after commit")
	}

	writeFile(t, repoDir, "untracked.txt", "oops")
	clean, err = g.IsClean(repoDir)
	if err != nil {
		t.Fatalf("IsClean: %v", err)
	}
	if clean {
		t.Error("expected dirty worktree with an untracked file")
	}
}

func TestReal_IsFullyMerged_NotMerged(t *testing.T) {
	repoDir := setupRepo(t)
	runGit(t, repoDir, "checkout", "-b", "feat")
	writeFile(t, repoDir, "feat.txt", "feat work")
	runGit(t, repoDir, "add", "feat.txt")
	runGit(t, repoDir, "commit", "-m", "feat work")

	g := git.Real{}
	merged, err := g.IsFullyMerged(repoDir, "main", "feat")
	if err != nil {
		t.Fatalf("IsFullyMerged: %v", err)
	}
	if merged {
		t.Error("expected feat to not be merged into main")
	}
}

func TestReal_IsFullyMerged_MergeCommit(t *testing.T) {
	repoDir := setupRepo(t)
	runGit(t, repoDir, "checkout", "-b", "feat")
	writeFile(t, repoDir, "feat.txt", "feat work")
	runGit(t, repoDir, "add", "feat.txt")
	runGit(t, repoDir, "commit", "-m", "feat work")
	runGit(t, repoDir, "checkout", "main")
	runGit(t, repoDir, "merge", "--no-ff", "feat", "-m", "merge feat")

	g := git.Real{}
	merged, err := g.IsFullyMerged(repoDir, "main", "feat")
	if err != nil {
		t.Fatalf("IsFullyMerged: %v", err)
	}
	if !merged {
		t.Error("expected feat to be merged into main via merge commit")
	}
}

func TestReal_IsFullyMerged_SquashMerge(t *testing.T) {
	repoDir := setupRepo(t)
	runGit(t, repoDir, "checkout", "-b", "feat")
	writeFile(t, repoDir, "feat.txt", "feat work")
	runGit(t, repoDir, "add", "feat.txt")
	runGit(t, repoDir, "commit", "-m", "feat work")
	runGit(t, repoDir, "checkout", "main")
	runGit(t, repoDir, "merge", "--squash", "feat")
	runGit(t, repoDir, "commit", "-m", "squash merge feat")

	g := git.Real{}
	merged, err := g.IsFullyMerged(repoDir, "main", "feat")
	if err != nil {
		t.Fatalf("IsFullyMerged: %v", err)
	}
	if !merged {
		t.Error("expected feat to be detected as merged via squash-merge content match")
	}
}

func TestReal_Init_CreatesGitDir(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "repo")
	g := git.Real{}
	if err := g.Init(dest); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, ".git")); err != nil {
		t.Errorf(".git directory not created: %v", err)
	}
}

func TestReal_Init_Idempotent(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "repo")
	g := git.Real{}
	if err := g.Init(dest); err != nil {
		t.Fatalf("first Init: %v", err)
	}
	if err := g.Init(dest); err != nil {
		t.Fatalf("second Init (idempotent): %v", err)
	}
}

func TestReal_Clone_InvalidURL(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "dest")
	g := git.Real{}
	err := g.Clone("not-a-url", dest)
	if err == nil {
		t.Fatal("expected error for invalid clone URL, got nil")
	}
	if !strings.Contains(err.Error(), "git clone") {
		t.Errorf("error %q does not mention 'git clone'", err.Error())
	}
}
