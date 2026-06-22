package git_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pecodez/couchdev/internal/git"
)

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
