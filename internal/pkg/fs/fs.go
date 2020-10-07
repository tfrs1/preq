package fs

import "os"

type Filesystem interface {
	Stat(string) (os.FileInfo, error)
	Open(string) (*os.File, error)
	Getwd() (string, error)
}

type OS struct{}

func (OS) Open(name string) (*os.File, error)    { return os.Open(name) }
func (OS) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }
func (OS) Getwd() (string, error)                { return os.Getwd() }
