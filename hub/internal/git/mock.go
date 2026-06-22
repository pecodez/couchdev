package git

import "os"

type Mock struct {
	CloneErr error
	InitErr  error
}

func (m *Mock) Clone(_, dest string) error {
	os.MkdirAll(dest, 0755)
	return m.CloneErr
}

func (m *Mock) Init(path string) error {
	os.MkdirAll(path, 0755)
	return m.InitErr
}
