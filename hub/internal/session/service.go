package session

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/pecodez/couchdev/internal/git"
	"github.com/pecodez/couchdev/internal/project"
	"github.com/pecodez/couchdev/internal/tmux"
)

type Service struct {
	projects *project.Store
	sessions *Store
	tmux     tmux.Client
	git      git.Client
}

func NewService(projects *project.Store, sessions *Store, t tmux.Client, g git.Client) *Service {
	return &Service{projects: projects, sessions: sessions, tmux: t, git: g}
}

func (svc *Service) Genesis(projectName, sessionName, cwd string) (*Session, error) {
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
	if err := svc.git.WorktreeAdd(proj.RepoPath, worktreePath, branch); err != nil {
		return nil, fmt.Errorf("create worktree: %w", err)
	}

	if cwd == "" {
		cwd = worktreePath
	}
	tmuxName := tmux.SessionName(projectName, sessionName)
	shellCmd := fmt.Sprintf(`claude --rc "%s"`, canonical)

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

	// tmux new-session -d always exits 0; give the process a moment to fail
	// fast (e.g. claude not in PATH) before we commit the record.
	time.Sleep(300 * time.Millisecond)
	if !svc.tmux.HasSession(tmuxName) {
		return nil, fmt.Errorf("session exited immediately after spawn — is claude installed and in PATH?")
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
		svc.tmux.KillSession(tmuxName)                       // best-effort rollback
		svc.git.WorktreeRemove(proj.RepoPath, worktreePath)  // best-effort rollback
		return nil, fmt.Errorf("persist session: %w", err)
	}
	sess.State = StateStarting
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

func (svc *Service) Teardown(projectName, sessionName string) error {
	canonical := projectName + "/" + sessionName
	sess, err := svc.sessions.GetByCanonical(canonical)
	if err != nil {
		return err
	}
	if svc.tmux.HasSession(sess.TmuxName) {
		if err := svc.tmux.KillSession(sess.TmuxName); err != nil {
			return fmt.Errorf("kill tmux: %w", err)
		}
	}
	if sess.WorktreePath != "" {
		proj, err := svc.projects.GetByID(sess.ProjectID)
		if err == nil {
			_ = svc.git.WorktreeRemove(proj.RepoPath, sess.WorktreePath) // best-effort
		}
	}
	return svc.sessions.MarkKilled(canonical)
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

func (svc *Service) deriveState(sess *Session) State {
	if sess.Killed {
		return StateDead
	}
	if svc.tmux.HasSession(sess.TmuxName) {
		return StateLive
	}
	return StateResumable
}
