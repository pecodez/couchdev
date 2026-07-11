package tmux

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// loginWrap wraps cmd in a bash login shell so ~/.profile is sourced and
// user-installed tools (e.g. claude via npm/nvm) are on PATH.
func loginWrap(cmd string) string {
	escaped := strings.ReplaceAll(cmd, "'", `'\''`)
	return "bash --login -c '" + escaped + "'"
}

type Client interface {
	NewSession(name, cwd, shellCmd string) error
	KillSession(name string) error
	HasSession(name string) bool
	ListSessions() ([]string, error)
	SendKeys(name string) error
	CapturePane(name string) (string, error)
}

// SessionName returns the tmux session name for a project/session pair.
// Underscores are used as separators because tmux silently converts dots to
// underscores in session names, which would cause HasSession to always miss.
func SessionName(project, session string) string {
	return "cdv_" + project + "_" + session
}

// Exec is the production Client that shells out to tmux.
type Exec struct {
	log *zap.Logger
}

func NewExec(log *zap.Logger) Exec {
	return Exec{log: log}
}

func (e Exec) NewSession(name, cwd, shellCmd string) error {
	wrapped := loginWrap(shellCmd)
	e.log.Debug("tmux new-session",
		zap.String("name", name),
		zap.String("cwd", cwd),
		zap.String("cmd", shellCmd),
		zap.String("wrapped", wrapped),
	)
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", cwd, wrapped)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		e.log.Warn("tmux new-session failed", zap.String("name", name), zap.Error(err), zap.String("stderr", stderr.String()))
		return fmt.Errorf("tmux new-session: %w: %s", err, stderr.String())
	}
	e.log.Debug("tmux new-session ok", zap.String("name", name))
	return nil
}

func (e Exec) KillSession(name string) error {
	e.log.Debug("tmux kill-session", zap.String("name", name))
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		e.log.Warn("tmux kill-session failed", zap.String("name", name), zap.Error(err), zap.String("stderr", stderr.String()))
		return fmt.Errorf("tmux kill-session: %w: %s", err, stderr.String())
	}
	return nil
}

func (e Exec) HasSession(name string) bool {
	ok := exec.Command("tmux", "has-session", "-t", name).Run() == nil
	e.log.Debug("tmux has-session", zap.String("name", name), zap.Bool("exists", ok))
	return ok
}

func (e Exec) SendKeys(name string) error {
	e.log.Debug("tmux send-keys Enter", zap.String("name", name))
	cmd := exec.Command("tmux", "send-keys", "-t", name, "Enter")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		e.log.Warn("tmux send-keys failed", zap.String("name", name), zap.Error(err))
		return fmt.Errorf("tmux send-keys: %w: %s", err, stderr.String())
	}
	return nil
}

func (e Exec) CapturePane(name string) (string, error) {
	out, err := exec.Command("tmux", "capture-pane", "-t", name, "-p").Output()
	if err != nil {
		e.log.Warn("tmux capture-pane failed", zap.String("name", name), zap.Error(err))
		return "", fmt.Errorf("tmux capture-pane: %w", err)
	}
	pane := string(out)
	e.log.Debug("tmux capture-pane", zap.String("name", name), zap.Int("bytes", len(pane)), zap.String("content", pane))
	return pane, nil
}

func (e Exec) ListSessions() ([]string, error) {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
	if err != nil {
		if ex, ok := err.(*exec.ExitError); ok && ex.ExitCode() == 1 {
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
