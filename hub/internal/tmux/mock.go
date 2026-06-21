package tmux

import "fmt"

// Mock is an in-memory Client for testing.
type Mock struct{ sessions map[string]bool }

func NewMock() *Mock { return &Mock{sessions: make(map[string]bool)} }

func (m *Mock) NewSession(name, cwd, shellCmd string) error {
	if m.sessions[name] {
		return fmt.Errorf("session %q already exists", name)
	}
	m.sessions[name] = true
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
