package gitutils

import (
	"fmt"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type goGitRepository interface {
	Head() (*plumbing.Reference, error)
	Remotes() ([]*git.Remote, error)
	Reference(plumbing.ReferenceName, bool) (*plumbing.Reference, error)
	CommitObject(plumbing.Hash) (*object.Commit, error)
}

type gitRepository interface {
	GetRemoteURLs() ([]string, error)
	GetCheckedOutBranchShortName() (string, error)
	CurrentCommit() (*object.Commit, error)
	BranchCommit(string) (*object.Commit, error)
}

type repository struct {
	r     goGitRepository
	goGit *git.Repository
}

var openRepo = func(path string) (*git.Repository, error) {
	// return git.PlainOpen(path)
	return OpenRepoRecursevely(path)
}

func OpenRepoRecursevely(input string) (*git.Repository, error) {
	dir := input
	for dir != "/" && dir != "." {
		repo, err := git.PlainOpen(dir)
		if err == nil {
			return repo, nil
		}

		dir = path.Dir(dir)
	}

	return nil, fmt.Errorf("Could not recursivelly open a repo at %s", input)
}

func (r *repository) GetRemoteURLs() ([]string, error) {
	var repoURLs []string
	remotes, err := r.r.Remotes()
	if err != nil {
		return nil, err
	}

	for _, re := range remotes {
		repoURLs = append(repoURLs, re.Config().URLs...)
	}

	return repoURLs, nil
}

func (r *repository) GetCheckedOutBranchShortName() (string, error) {
	headRef, err := r.r.Head()
	if err != nil {
		return "", err
	}

	return headRef.Name().Short(), nil
}

func (r *repository) CurrentCommit() (*object.Commit, error) {
	head, err := r.r.Head()
	if err != nil {
		return nil, err
	}

	return r.r.CommitObject(head.Hash())
}

func (r *repository) BranchCommit(b string) (*object.Commit, error) {
	bRef, err := r.r.Reference(plumbing.NewBranchReferenceName(b), false)
	if err != nil {
		return nil, err
	}

	return r.r.CommitObject(bRef.Hash())
}
