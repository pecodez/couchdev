package git

import "os"

type Mock struct {
	CloneErr          error
	InitErr           error
	FetchErr          error
	NoRemote          bool // when true, HasRemote reports no remote (simulates a local-only repo)
	HasRemoteErr      error
	AddRemoteErr      error
	PushErr           error
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
	FetchCalls            int    // number of times Fetch was called
	AddedRemote           string // last remote name passed to AddRemote
	AddedRemoteURL        string // last url passed to AddRemote
	PushedRemote          string // last remote passed to Push
	PushedBranch          string // last branch passed to Push
	WorktreeAdded         string // last worktreePath passed to WorktreeAdd
	WorktreeAddStartPoint string // last startPoint passed to WorktreeAdd
	WorktreeRemoved       string // last worktreePath passed to WorktreeRemove
	MergedDefaultBranch   string // last defaultBranch (ref) arg passed to IsFullyMerged
	AheadBase             string // last base ref arg passed to CommitsAhead
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
	m.FetchCalls++
	return m.FetchErr
}

func (m *Mock) HasRemote(_, _ string) (bool, error) {
	return !m.NoRemote, m.HasRemoteErr
}

func (m *Mock) AddRemote(_, remote, url string) error {
	m.AddedRemote = remote
	m.AddedRemoteURL = url
	return m.AddRemoteErr
}

func (m *Mock) Push(_, remote, branch string) error {
	m.PushedRemote = remote
	m.PushedBranch = branch
	return m.PushErr
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

func (m *Mock) CommitsAhead(_, base string) (int, error) {
	m.AheadBase = base
	return m.AheadCount, m.AheadErr
}

func (m *Mock) ChangedFiles(_ string) ([]string, error) {
	return m.Files, m.FilesErr
}

func (m *Mock) IsClean(_ string) (bool, error) {
	return m.CleanResult, m.CleanErr
}

func (m *Mock) IsFullyMerged(_, defaultBranch, _ string) (bool, error) {
	m.MergedDefaultBranch = defaultBranch
	return m.MergedResult, m.MergedErr
}
