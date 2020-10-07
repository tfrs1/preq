package fs

import (
	"os"
	"time"
)

type MockFileInfo struct {
	IsDirValue bool
}

func (m MockFileInfo) IsDir() bool        { return m.IsDirValue }
func (m MockFileInfo) ModTime() time.Time { return time.Now() }
func (m MockFileInfo) Mode() os.FileMode  { return 0 }
func (m MockFileInfo) Name() string       { return "" }
func (m MockFileInfo) Size() int64        { return 1 }
func (m MockFileInfo) Sys() interface{}   { return nil }

type MockFS struct {
	Info MockFileInfo
	Err  error
}

func (fs MockFS) Open(name string) (*os.File, error)    { return nil, fs.Err }
func (fs MockFS) Stat(name string) (os.FileInfo, error) { return fs.Info, fs.Err }
func (fs MockFS) Getwd() (string, error)                { return "", fs.Err }
