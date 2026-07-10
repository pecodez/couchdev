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
