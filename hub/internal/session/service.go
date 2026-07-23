package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/pecodez/couchdev/internal/git"
	"github.com/pecodez/couchdev/internal/project"
	"github.com/pecodez/couchdev/internal/tmux"
)

type Service struct {
	projects *project.Store
	sessions *Store
	tmux     tmux.Client
	git      git.Client
	log      *zap.Logger
}

func NewService(projects *project.Store, sessions *Store, t tmux.Client, g git.Client, log *zap.Logger) *Service {
	return &Service{projects: projects, sessions: sessions, tmux: t, git: g, log: log}
}

// checkClaudeBridge reads ~/.claude.json and returns an error if Claude Code
// is not authenticated or the remote-control bridge is in a backoff period.
func checkClaudeBridge() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil // can't determine home dir — skip check
	}
	data, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		return nil // file absent — first run, proceed and let Claude report the error
	}
	var cfg struct {
		OauthAccount        *json.RawMessage `json:"oauthAccount"`
		BridgeDeadExpiresAt int64            `json:"bridgeOauthDeadExpiresAt"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil // unreadable — skip check
	}
	if cfg.OauthAccount == nil {
		return errors.New("Claude Code is not authenticated — run 'claude login' first")
	}
	if cfg.BridgeDeadExpiresAt > time.Now().UnixMilli() {
		exp := time.UnixMilli(cfg.BridgeDeadExpiresAt).UTC().Format(time.RFC3339)
		return fmt.Errorf("remote control bridge is in backoff until %s — run 'claude login' to reset it", exp)
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func (svc *Service) Genesis(projectName, sessionName, cwd string) (*Session, error) {
	if err := checkClaudeBridge(); err != nil {
		return nil, err
	}

	proj, err := svc.projects.GetByName(projectName)
	if err != nil {
		return nil, fmt.Errorf("project %q not found: %w", projectName, err)
	}

	canonical := projectName + "/" + sessionName
	if existing, _ := svc.sessions.GetByCanonical(canonical); existing != nil {
		if !existing.Killed && svc.tmux.HasSession(existing.TmuxName) {
			return nil, fmt.Errorf("session %q is already live", canonical)
		}
	}

	branch := sessionName
	worktreePath := filepath.Join(filepath.Dir(proj.RepoPath), "worktrees", sessionName)

	startPoint := proj.DefaultBranch
	hasOrigin, err := svc.git.HasRemote(proj.RepoPath, "origin")
	if err != nil {
		return nil, fmt.Errorf("check remote: %w", err)
	}
	if hasOrigin {
		if err := svc.git.Fetch(proj.RepoPath, "origin", proj.DefaultBranch); err != nil {
			return nil, fmt.Errorf("fetch default branch: %w", err)
		}
		startPoint = "origin/" + proj.DefaultBranch
	}
	if err := svc.git.WorktreeAdd(proj.RepoPath, worktreePath, branch, startPoint); err != nil {
		return nil, fmt.Errorf("create worktree: %w", err)
	}

	if cwd == "" {
		cwd = worktreePath
	}

	// Pre-create the Claude project-state directory so subsequent sessions at
	// the same worktree path have a known entry; Claude populates it after the
	// first trust acceptance.
	if home, err := os.UserHomeDir(); err == nil {
		slug := strings.ReplaceAll(worktreePath, "/", "-")
		trustDir := filepath.Join(home, ".claude", "projects", slug)
		_ = os.MkdirAll(trustDir, 0700)
	}

	tmuxName := tmux.SessionName(projectName, sessionName)
	shellCmd := fmt.Sprintf(`claude --rc "%s"`, canonical)

	svc.log.Info("genesis: spawning session",
		zap.String("canonical", canonical),
		zap.String("tmux_name", tmuxName),
		zap.String("cwd", cwd),
		zap.String("shell_cmd", shellCmd),
		zap.String("worktree", worktreePath),
	)

	// Kill any orphaned tmux session with this name before spawning.  This
	// happens when a previous Genesis attempt created the tmux session but
	// failed (or the DB was reset) before the record was committed, leaving
	// a session that would otherwise block a retry with "already exists".
	if svc.tmux.HasSession(tmuxName) {
		_ = svc.tmux.KillSession(tmuxName)
	}

	if err := svc.tmux.NewSession(tmuxName, cwd, shellCmd); err != nil {
		return nil, fmt.Errorf("spawn session: %w", err)
	}
	svc.log.Info("genesis: tmux session created", zap.String("tmux_name", tmuxName))

	// tmux new-session -d always exits 0; give the process a moment to fail
	// fast (e.g. claude not in PATH) before we commit the record.
	time.Sleep(300 * time.Millisecond)
	if !svc.tmux.HasSession(tmuxName) {
		return nil, fmt.Errorf("session exited immediately after spawn — is claude installed and in PATH?")
	}

	// Poll pane for workspace-trust dialog or Claude startup header.
	const (
		trustPoll     = 500 * time.Millisecond
		trustAttempts = 20 // 10 s max
	)
	trustedOrStarted := false
	for i := 0; i < trustAttempts; i++ {
		time.Sleep(trustPoll)
		if !svc.tmux.HasSession(tmuxName) {
			return nil, fmt.Errorf("session exited while waiting for Claude startup")
		}
		pane, err := svc.tmux.CapturePane(tmuxName)
		if err != nil {
			svc.log.Warn("trust poll: capture-pane error", zap.Int("attempt", i), zap.Error(err))
			continue
		}
		svc.log.Debug("trust poll: pane snapshot",
			zap.Int("attempt", i),
			zap.String("pane", truncate(pane, 500)),
		)
		if strings.Contains(pane, "I trust this folder") {
			svc.log.Info("trust poll: trust dialog detected — sending Enter", zap.Int("attempt", i))
			_ = svc.tmux.SendKeys(tmuxName)
			trustedOrStarted = true
			break
		}
		if strings.Contains(pane, "Claude Code") {
			svc.log.Info("trust poll: Claude started (already trusted)", zap.Int("attempt", i))
			trustedOrStarted = true
			break
		}
	}
	if !trustedOrStarted {
		svc.log.Warn("trust poll: timed out — session may be stuck at trust dialog or startup")
	}

	// Poll pane for remote-control confirmation or errors.
	// Claude Code outputs RC status after startup; we watch for known patterns.
	const (
		rcPoll     = 500 * time.Millisecond
		rcAttempts = 30 // 15 s max
	)
	var warnings []string
	rcConfirmed := false
	for i := 0; i < rcAttempts; i++ {
		time.Sleep(rcPoll)
		if !svc.tmux.HasSession(tmuxName) {
			svc.log.Warn("rc poll: session exited before RC could be confirmed")
			warnings = append(warnings, "session exited before remote control could be confirmed")
			break
		}
		pane, err := svc.tmux.CapturePane(tmuxName)
		if err != nil {
			continue
		}
		lower := strings.ToLower(pane)
		svc.log.Debug("rc poll: pane snapshot",
			zap.Int("attempt", i),
			zap.String("pane", truncate(pane, 500)),
		)
		if strings.Contains(lower, "remote control") || strings.Contains(lower, "remote-control") || strings.Contains(lower, "remotely") {
			svc.log.Info("rc poll: remote control confirmed", zap.Int("attempt", i), zap.String("pane_excerpt", truncate(pane, 200)))
			rcConfirmed = true
			break
		}
		if strings.Contains(lower, "not authenticated") || strings.Contains(lower, "oauth") ||
			strings.Contains(lower, "unauthorized") || strings.Contains(lower, "bridge") {
			svc.log.Warn("rc poll: error pattern detected in pane",
				zap.Int("attempt", i),
				zap.String("pane_excerpt", truncate(pane, 500)),
			)
			warnings = append(warnings, "remote control error detected — check 'claude login' status and couchdev logs")
			break
		}
	}
	if !rcConfirmed && len(warnings) == 0 {
		svc.log.Warn("rc poll: no remote control confirmation after 15s", zap.String("tmux_name", tmuxName))
		warnings = append(warnings, "remote control not confirmed after 15s — session is running but may not appear in Claude mobile; check couchdev --verbose logs")
	}

	sess, err := svc.sessions.Create(Session{
		ProjectID:     proj.ID,
		Session:       sessionName,
		CanonicalName: canonical,
		PassedName:    canonical,
		CWD:           cwd,
		Branch:        branch,
		WorktreePath:  worktreePath,
		TmuxName:      tmuxName,
	})
	if err != nil {
		svc.tmux.KillSession(tmuxName)                      // best-effort rollback
		svc.git.WorktreeRemove(proj.RepoPath, worktreePath) // best-effort rollback
		return nil, fmt.Errorf("persist session: %w", err)
	}
	sess.State = StateStarting
	sess.Warnings = warnings
	svc.log.Info("genesis: session persisted",
		zap.Int64("id", sess.ID),
		zap.String("canonical", canonical),
		zap.Strings("warnings", warnings),
	)
	return sess, nil
}

func (svc *Service) Status(projectName, sessionName string) (*Session, error) {
	canonical := projectName + "/" + sessionName
	sess, err := svc.sessions.GetByCanonical(canonical)
	if err != nil {
		return nil, err
	}
	sess.State = svc.deriveState(sess)
	return sess, nil
}

func (svc *Service) List() ([]Session, error) {
	all, err := svc.sessions.ListAll()
	if err != nil {
		return nil, err
	}
	for i := range all {
		all[i].State = svc.deriveState(&all[i])
	}
	return all, nil
}

// Teardown stops the session's tmux session (if live) and, when the worktree can be proven
// safe to discard (clean working tree and branch fully merged into the project's default
// branch), removes the worktree and marks the session dead. When force is true, or the
// session has no worktree, the safety check is skipped. Otherwise the worktree and session
// record are left untouched (deleted is false) so the caller can inform the user why, and the
// session reappears as resumable since tmux was stopped but Killed stays false.
func (svc *Service) Teardown(projectName, sessionName string, force bool) (deleted bool, reason string, err error) {
	canonical := projectName + "/" + sessionName
	sess, err := svc.sessions.GetByCanonical(canonical)
	if err != nil {
		return false, "", err
	}

	if svc.tmux.HasSession(sess.TmuxName) {
		if err := svc.tmux.KillSession(sess.TmuxName); err != nil {
			return false, "", fmt.Errorf("kill tmux: %w", err)
		}
	}

	if sess.WorktreePath == "" {
		return true, "", svc.sessions.MarkKilled(canonical)
	}

	proj, err := svc.projects.GetByID(sess.ProjectID)
	if err != nil {
		return false, "", err
	}

	if !force {
		clean, err := svc.git.IsClean(sess.WorktreePath)
		if err != nil {
			return false, "", fmt.Errorf("check clean: %w", err)
		}
		if !clean {
			return false, "worktree has uncommitted or untracked changes", nil
		}
		merged, err := svc.git.IsFullyMerged(proj.RepoPath, proj.DefaultBranch, sess.Branch)
		if err != nil {
			return false, "", fmt.Errorf("check merged: %w", err)
		}
		if !merged {
			return false, fmt.Sprintf("branch %q is not merged into %q", sess.Branch, proj.DefaultBranch), nil
		}
	}

	if err := svc.git.WorktreeRemove(proj.RepoPath, sess.WorktreePath); err != nil {
		return false, "", fmt.Errorf("remove worktree: %w", err)
	}
	return true, "", svc.sessions.MarkKilled(canonical)
}

func (svc *Service) Changes(projectName, sessionName string) (int, []string, error) {
	canonical := projectName + "/" + sessionName
	sess, err := svc.sessions.GetByCanonical(canonical)
	if err != nil {
		return 0, nil, err
	}
	if sess.WorktreePath == "" {
		return 0, nil, nil
	}
	proj, err := svc.projects.GetByID(sess.ProjectID)
	if err != nil {
		return 0, nil, err
	}
	ahead, err := svc.git.CommitsAhead(sess.WorktreePath, proj.DefaultBranch)
	if err != nil {
		return 0, nil, fmt.Errorf("commits ahead: %w", err)
	}
	files, err := svc.git.ChangedFiles(sess.WorktreePath)
	if err != nil {
		return 0, nil, fmt.Errorf("changed files: %w", err)
	}
	return ahead, files, nil
}

func (svc *Service) Resume(projectName, sessionName string) (*Session, error) {
	if err := checkClaudeBridge(); err != nil {
		return nil, err
	}

	canonical := projectName + "/" + sessionName
	sess, err := svc.sessions.GetByCanonical(canonical)
	if err != nil {
		return nil, err
	}
	if sess.Killed {
		return nil, fmt.Errorf("session %q is dead", canonical)
	}
	if svc.tmux.HasSession(sess.TmuxName) {
		return nil, fmt.Errorf("session %q is already live", canonical)
	}

	shellCmd := fmt.Sprintf(`claude --rc "%s"`, canonical)
	svc.log.Info("resume: spawning session",
		zap.String("canonical", canonical),
		zap.String("tmux_name", sess.TmuxName),
		zap.String("shell_cmd", shellCmd),
	)

	if err := svc.tmux.NewSession(sess.TmuxName, sess.CWD, shellCmd); err != nil {
		return nil, fmt.Errorf("spawn session: %w", err)
	}
	time.Sleep(300 * time.Millisecond)
	if !svc.tmux.HasSession(sess.TmuxName) {
		return nil, fmt.Errorf("session exited immediately after spawn — is claude installed and in PATH?")
	}
	svc.log.Info("resume: session started", zap.String("canonical", canonical))
	sess.State = StateStarting
	return sess, nil
}

func (svc *Service) deriveState(sess *Session) State {
	if sess.Killed {
		return StateDead
	}
	if svc.tmux.HasSession(sess.TmuxName) {
		return StateLive
	}
	return StateResumable
}
