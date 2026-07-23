package git

import "os"

type Mock struct {
	CloneErr          error
	InitErr           error
	FetchErr          error
	NoRemote          bool // when true, HasRemote reports no remote (simulates a local-only repo)
	HasRemoteErr      error
	WorktreeAddErr    error
	WorktreeRemoveErr error
	AheadCount   int
	AheadErr     error
	Files        []string
	FilesErr     error
	CleanResult  bool
	CleanErr     error
	MergedResult bool
	MergedErr    error

	FetchedRemote         string // last remote passed to Fetch
	FetchedBranch         string // last branch passed to Fetch
	WorktreeAdded         string // last worktreePath passed to WorktreeAdd
	WorktreeAddStartPoint string // last startPoint passed to WorktreeAdd
	WorktreeRemoved       string // last worktreePath passed to WorktreeRemove
}

func (m *Mock) Clone(_, dest string) error {
	os.MkdirAll(dest, 0755)
	return m.CloneErr
}

func (m *Mock) Init(path string) error {
	os.MkdirAll(path, 0755)
	return m.InitErr
}

func (m *Mock) Fetch(_, remote, branch string) error {
	m.FetchedRemote = remote
	m.FetchedBranch = branch
	return m.FetchErr
}

func (m *Mock) HasRemote(_, _ string) (bool, error) {
	return !m.NoRemote, m.HasRemoteErr
}

func (m *Mock) WorktreeAdd(_, worktreePath, _, startPoint string) error {
	if m.WorktreeAddErr != nil {
		return m.WorktreeAddErr
	}
	os.MkdirAll(worktreePath, 0755)
	m.WorktreeAdded = worktreePath
	m.WorktreeAddStartPoint = startPoint
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

func (m *Mock) IsClean(_ string) (bool, error) {
	return m.CleanResult, m.CleanErr
}

func (m *Mock) IsFullyMerged(_, _, _ string) (bool, error) {
	return m.MergedResult, m.MergedErr
}
