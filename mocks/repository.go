package mocks

import (
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type GoGitRepository struct {
	HeadValue    *plumbing.Reference
	Err          error
	RemotesValue []*git.Remote
}

func (r GoGitRepository) Head() (*plumbing.Reference, error) {
	return r.HeadValue, r.Err
}

func (r GoGitRepository) Remotes() ([]*git.Remote, error) {
	return r.RemotesValue, r.Err
}

func (r GoGitRepository) Reference(plumbing.ReferenceName, bool) (*plumbing.Reference, error) {
	return &plumbing.Reference{}, r.Err
}

func (r GoGitRepository) CommitObject(plumbing.Hash) (*object.Commit, error) {
	return nil, nil
}

type GitRepository struct {
	ErrorValue         error
	CurrentBranchValue string
	RemoteURLsValue    []string
	BranchCommitValue  *object.Commit
	Commit             *object.Commit
}

func (r *GitRepository) GetCheckedOutBranchShortName() (string, error) {
	return r.CurrentBranchValue, r.ErrorValue
}

func (r *GitRepository) BranchCommit(string) (*object.Commit, error) {
	return r.BranchCommitValue, r.ErrorValue
}

func (r *GitRepository) CurrentCommit() (*object.Commit, error) {
	return r.Commit, r.ErrorValue
}

func (r *GitRepository) GetRemoteURLs() ([]string, error) {
	return r.RemoteURLsValue, r.ErrorValue
}
