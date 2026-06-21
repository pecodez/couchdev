package session

import (
	"fmt"

	"github.com/pecodez/couchdev/internal/project"
	"github.com/pecodez/couchdev/internal/tmux"
)

type Service struct {
	projects *project.Store
	sessions *Store
	tmux     tmux.Client
}

func NewService(projects *project.Store, sessions *Store, t tmux.Client) *Service {
	return &Service{projects: projects, sessions: sessions, tmux: t}
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

	if cwd == "" {
		cwd = proj.RepoPath
	}
	tmuxName := tmux.SessionName(projectName, sessionName)
	shellCmd := fmt.Sprintf(`claude --rc "%s"`, canonical)

	if err := svc.tmux.NewSession(tmuxName, cwd, shellCmd); err != nil {
		return nil, fmt.Errorf("spawn session: %w", err)
	}

	sess, err := svc.sessions.Create(Session{
		ProjectID:     proj.ID,
		Session:       sessionName,
		CanonicalName: canonical,
		PassedName:    canonical,
		CWD:           cwd,
		TmuxName:      tmuxName,
	})
	if err != nil {
		svc.tmux.KillSession(tmuxName) // best-effort rollback
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
	return svc.sessions.MarkKilled(canonical)
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
