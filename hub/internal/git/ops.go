package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Client interface {
	Clone(repoURL, destPath string) error
	Init(path string) error
	WorktreeAdd(sourceDir, worktreePath, branch string) error
	WorktreeRemove(sourceDir, worktreePath string) error
	CommitsAhead(worktreePath, base string) (int, error)
	ChangedFiles(worktreePath string) ([]string, error)
	IsClean(worktreePath string) (bool, error)
	IsFullyMerged(repoPath, defaultBranch, branch string) (bool, error)
}

type Real struct{}

func (Real) Clone(repoURL, destPath string) error {
	cmd := exec.Command("git", "clone", repoURL, destPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %w: %s", err, stderr.String())
	}
	return nil
}

func (Real) Init(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	cmd := exec.Command("git", "init", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git init: %w: %s", err, stderr.String())
	}
	return nil
}

func (Real) WorktreeAdd(sourceDir, worktreePath, branch string) error {
	cmd := exec.Command("git", "-C", sourceDir, "worktree", "add", worktreePath, "-b", branch)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git worktree add: %w: %s", err, stderr.String())
	}
	return nil
}

func (Real) WorktreeRemove(sourceDir, worktreePath string) error {
	cmd := exec.Command("git", "-C", sourceDir, "worktree", "remove", "--force", worktreePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git worktree remove: %w: %s", err, stderr.String())
	}
	return nil
}

func (Real) CommitsAhead(worktreePath, base string) (int, error) {
	out, err := exec.Command("git", "-C", worktreePath, "rev-list", "--count", base+"..HEAD").Output()
	if err != nil {
		return 0, fmt.Errorf("git rev-list: %w", err)
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("parse commit count: %w", err)
	}
	return n, nil
}

// IsClean reports whether the worktree has no staged, unstaged, or untracked changes.
func (Real) IsClean(worktreePath string) (bool, error) {
	out, err := exec.Command("git", "-C", worktreePath, "status", "--porcelain").Output()
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	return strings.TrimSpace(string(out)) == "", nil
}

// IsFullyMerged reports whether branch's changes are fully present in defaultBranch, such
// that removing branch would not lose any work. It detects plain merge-commit/fast-forward
// merges (via ancestry) and squash merges (via content comparison scoped to the files branch
// touched). Anything it can't positively confirm is reported as not merged.
func (Real) IsFullyMerged(repoPath, defaultBranch, branch string) (bool, error) {
	err := exec.Command("git", "-C", repoPath, "merge-base", "--is-ancestor", branch, defaultBranch).Run()
	if err == nil {
		return true, nil
	}
	if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
		return false, fmt.Errorf("git merge-base --is-ancestor: %w", err)
	}

	baseOut, err := exec.Command("git", "-C", repoPath, "merge-base", defaultBranch, branch).Output()
	if err != nil {
		return false, fmt.Errorf("git merge-base: %w", err)
	}
	mergeBase := strings.TrimSpace(string(baseOut))

	filesOut, err := exec.Command("git", "-C", repoPath, "diff", "--name-only", mergeBase, branch).Output()
	if err != nil {
		return false, fmt.Errorf("git diff --name-only: %w", err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(filesOut)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	if len(files) == 0 {
		return true, nil
	}

	args := append([]string{"-C", repoPath, "diff", defaultBranch, branch, "--"}, files...)
	diffOut, err := exec.Command("git", args...).Output()
	if err != nil {
		return false, fmt.Errorf("git diff: %w", err)
	}
	return strings.TrimSpace(string(diffOut)) == "", nil
}

func (Real) ChangedFiles(worktreePath string) ([]string, error) {
	out, err := exec.Command("git", "-C", worktreePath, "diff", "--name-only", "HEAD").Output()
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}
