package mocks

import (
	"os"
	"time"
)

type FS struct {
	Info FileInfo
	Err  error
}

func (fs FS) Open(name string) (*os.File, error)    { return nil, fs.Err }
func (fs FS) Stat(name string) (os.FileInfo, error) { return fs.Info, fs.Err }
func (fs FS) Getwd() (string, error)                { return "", fs.Err }

type FileInfo struct {
	IsDirValue bool
}

func (m FileInfo) IsDir() bool        { return m.IsDirValue }
func (m FileInfo) ModTime() time.Time { return time.Now() }
func (m FileInfo) Mode() os.FileMode  { return 0 }
func (m FileInfo) Name() string       { return "" }
func (m FileInfo) Size() int64        { return 1 }
func (m FileInfo) Sys() interface{}   { return nil }
