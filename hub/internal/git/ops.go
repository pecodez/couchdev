package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type Client interface {
	Clone(repoURL, destPath string) error
	Init(path string) error
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
