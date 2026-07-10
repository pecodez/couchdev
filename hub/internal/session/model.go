package session

import "time"

type State string

const (
	StateStarting  State = "starting"
	StateLive      State = "live"
	StateResumable State = "resumable"
	StateDead      State = "dead"
)

type Session struct {
	ID            int64     `json:"id"`
	ProjectID     int64     `json:"-"`
	Session       string    `json:"session"`
	CanonicalName string    `json:"canonical_name"`
	PassedName    string    `json:"passed_name"`
	CWD           string    `json:"cwd"`
	Branch        string    `json:"branch"`
	WorktreePath  string    `json:"worktree_path"`
	TmuxName      string    `json:"tmux_name"`
	PID           *int      `json:"pid,omitempty"`
	Killed        bool      `json:"-"`
	StartedAt     time.Time `json:"started_at"`
	State         State     `json:"state"`
}
