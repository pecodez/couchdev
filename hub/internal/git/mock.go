package git

import "os"

type Mock struct {
	CloneErr        error
	InitErr         error
	WorktreeAddErr  error
	WorktreeRemoveErr error
	AheadCount      int
	AheadErr        error
	Files           []string
	FilesErr        error

	WorktreeAdded   string // last worktreePath passed to WorktreeAdd
	WorktreeRemoved string // last worktreePath passed to WorktreeRemove
}

func (m *Mock) Clone(_, dest string) error {
	os.MkdirAll(dest, 0755)
	return m.CloneErr
}

func (m *Mock) Init(path string) error {
	os.MkdirAll(path, 0755)
	return m.InitErr
}

func (m *Mock) WorktreeAdd(_, worktreePath, _ string) error {
	if m.WorktreeAddErr != nil {
		return m.WorktreeAddErr
	}
	os.MkdirAll(worktreePath, 0755)
	m.WorktreeAdded = worktreePath
	return nil
}

func (m *Mock) WorktreeRemove(_, worktreePath string) error {
	m.WorktreeRemoved = worktreePath
	return m.WorktreeRemoveErr
}

func (m *Mock) CommitsAhead(_, _ string) (int, error) {
	return m.AheadCount, m.AheadErr
}

func (m *Mock) ChangedFiles(_ string) ([]string, error) {
	return m.Files, m.FilesErr
}
