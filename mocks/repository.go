package mocks

import (
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type Repository struct {
	HeadValue    *plumbing.Reference
	Err          error
	RemotesValue []*git.Remote
}

func (r Repository) Head() (*plumbing.Reference, error) {
	return r.HeadValue, r.Err
}

func (r Repository) Remotes() ([]*git.Remote, error) {
	return r.RemotesValue, r.Err
}

func (r Repository) Reference(plumbing.ReferenceName, bool) (*plumbing.Reference, error) {
	return nil, nil
}

func (r Repository) CommitObject(plumbing.Hash) (*object.Commit, error) {
	return nil, nil
}
