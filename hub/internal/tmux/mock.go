package tmux

import "fmt"

// Mock is an in-memory Client for testing.
type Mock struct {
	sessions    map[string]bool
	dieOnSpawn  bool
	LastCmd     string // shell command passed to the most recent NewSession call
}

func NewMock() *Mock { return &Mock{sessions: make(map[string]bool)} }

// NewMockDying returns a Mock where spawned sessions exit immediately,
// simulating a command not found or instant crash.
func NewMockDying() *Mock { return &Mock{sessions: make(map[string]bool), dieOnSpawn: true} }

func (m *Mock) NewSession(name, cwd, shellCmd string) error {
	m.LastCmd = shellCmd
	if m.sessions[name] {
		return fmt.Errorf("session %q already exists", name)
	}
	if !m.dieOnSpawn {
		m.sessions[name] = true
	}
	return nil
}

func (m *Mock) KillSession(name string) error {
	if !m.sessions[name] {
		return fmt.Errorf("session %q not found", name)
	}
	delete(m.sessions, name)
	return nil
}

func (m *Mock) HasSession(name string) bool { return m.sessions[name] }

func (m *Mock) ListSessions() ([]string, error) {
	names := make([]string, 0, len(m.sessions))
	for n := range m.sessions {
		names = append(names, n)
	}
	return names, nil
}

func (m *Mock) SendKeys(name string) error         { return nil }
func (m *Mock) CapturePane(name string) (string, error) { return "", nil }
