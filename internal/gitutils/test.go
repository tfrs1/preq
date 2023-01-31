package gitutils

import (
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type MockGoGitRepository struct {
	HeadValue    *plumbing.Reference
	Err          error
	RemotesValue []*git.Remote
}

func (r MockGoGitRepository) Head() (*plumbing.Reference, error) {
	return r.HeadValue, r.Err
}

func (r MockGoGitRepository) Remotes() ([]*git.Remote, error) {
	return r.RemotesValue, r.Err
}

func (r MockGoGitRepository) Reference(
	plumbing.ReferenceName,
	bool,
) (*plumbing.Reference, error) {
	return &plumbing.Reference{}, r.Err
}

func (r MockGoGitRepository) CommitObject(
	plumbing.Hash,
) (*object.Commit, error) {
	return nil, nil
}

type MockGitRepository struct {
	ErrorValue         error
	CurrentBranchValue string
	RemoteURLsValue    []string
	BranchCommitValue  *object.Commit
	Commit             *object.Commit
}

func (r *MockGitRepository) GetCheckedOutBranchShortName() (string, error) {
	return r.CurrentBranchValue, r.ErrorValue
}

func (r *MockGitRepository) BranchCommit(string) (*object.Commit, error) {
	return r.BranchCommitValue, r.ErrorValue
}

func (r *MockGitRepository) CurrentCommit() (*object.Commit, error) {
	return r.Commit, r.ErrorValue
}

func (r *MockGitRepository) GetRemoteURLs() ([]string, error) {
	return r.RemoteURLsValue, r.ErrorValue
}
