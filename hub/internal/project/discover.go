package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// FlatRepo is a git repo found directly under projectsDir that has not yet
// been placed under a src/ subdirectory (old layout).
type FlatRepo struct {
	Name string
	Path string // full path: <projectsDir>/<name>
}

// ScanFlat returns repos under projectsDir whose git root is at the top
// level of the project directory (i.e. no src/ subdir yet).
// Returns nil without error if projectsDir does not exist.
func ScanFlat(projectsDir string) ([]FlatRepo, error) {
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan %s: %w", projectsDir, err)
	}
	var found []FlatRepo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(projectsDir, e.Name())
		if isDir(filepath.Join(dir, ".git")) && !isDir(filepath.Join(dir, "src")) {
			found = append(found, FlatRepo{Name: e.Name(), Path: dir})
		}
	}
	return found, nil
}

// MigrateFlat reorganises a flat git repo into the src/ layout:
//
//	<projectsDir>/<name>          (git root)
//	→ <projectsDir>/<name>/src/   (git root under project root)
//
// The operation uses two renames so it is as atomic as the filesystem allows.
func MigrateFlat(projectsDir, name string) error {
	projectDir := filepath.Join(projectsDir, name)
	tmpDir := filepath.Join(projectsDir, name+"_couchdev_tmp")
	srcDir := filepath.Join(projectDir, "src")

	if err := os.Rename(projectDir, tmpDir); err != nil {
		return fmt.Errorf("move %s aside: %w", name, err)
	}
	if err := os.Mkdir(projectDir, 0o755); err != nil {
		os.Rename(tmpDir, projectDir) // best-effort rollback
		return fmt.Errorf("create project root for %s: %w", name, err)
	}
	if err := os.Rename(tmpDir, srcDir); err != nil {
		os.Remove(projectDir)
		os.Rename(tmpDir, projectDir) // best-effort rollback
		return fmt.Errorf("move %s to src/: %w", name, err)
	}
	return nil
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}
