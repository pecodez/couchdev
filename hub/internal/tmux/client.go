package tmux

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Client interface {
	NewSession(name, cwd, shellCmd string) error
	KillSession(name string) error
	HasSession(name string) bool
	ListSessions() ([]string, error)
}

// SessionName returns the tmux session name for a project/session pair.
func SessionName(project, session string) string {
	return "cdv." + project + "." + session
}

// Exec is the production Client that shells out to tmux.
type Exec struct{}

func (Exec) NewSession(name, cwd, shellCmd string) error {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", cwd, shellCmd)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tmux new-session: %w: %s", err, stderr.String())
	}
	return nil
}

func (Exec) KillSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tmux kill-session: %w: %s", err, stderr.String())
	}
	return nil
}

func (Exec) HasSession(name string) bool {
	return exec.Command("tmux", "has-session", "-t", name).Run() == nil
}

func (Exec) ListSessions() ([]string, error) {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok && e.ExitCode() == 1 {
			return nil, nil // no sessions — not an error
		}
		return nil, fmt.Errorf("tmux list-sessions: %w", err)
	}
	var result []string
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if l != "" {
			result = append(result, l)
		}
	}
	return result, nil
}
