package gitutil

import (
	"errors"
	"os"
	client "prctl/pkg/bitbucket"
	"regexp"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

var (
	ErrCannotGetLocalRepository            = errors.New("cannot get local repository")
	ErrUnableToDetermineRepositoryProvider = errors.New("unable to determine repository provider")
	ErrAncestorCommitNotFound              = errors.New("ancestor commit not found")
	ErrCannotFindAnyBranchReference        = errors.New("cannot find any branch reference")
)

func GetBranch() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", ErrCannotGetLocalRepository
	}

	r, err := git.PlainOpen(wd)
	if err != nil {
		return "", ErrCannotGetLocalRepository
	}

	headRef, err := r.Head()
	if err != nil {
		return "", ErrCannotGetLocalRepository
	}

	return headRef.Name().Short(), nil
}

func getLocalRepo() (*git.Repository, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, ErrCannotGetLocalRepository
	}

	r, err := git.PlainOpen(wd)
	if err != nil {
		return nil, ErrCannotGetLocalRepository
	}

	return r, nil
}

func getRepos() ([]string, error) {
	var repos []string
	r, err := getLocalRepo()
	if err != nil {
		return nil, err
	}

	remotes, err := r.Remotes()
	if err != nil {
		return nil, ErrCannotGetLocalRepository
	}

	for _, re := range remotes {
		repos = append(repos, re.Config().URLs...)
	}

	return repos, nil
}

func parseRepositoryString(repoString string) (*client.Repository, error) {
	r := regexp.MustCompile(`git@(.*):(.*)/(.*)\.git`)
	m := r.FindStringSubmatch(repoString)
	if len(m) != 4 {
		return nil, ErrUnableToDetermineRepositoryProvider
	}

	p, err := client.ParseRepositoryProvider(m[1])
	if err != nil {
		return nil, err
	}

	return &client.Repository{
		Provider: p,
		Owner:    m[2],
		Name:     m[3],
	}, nil
}

func GetRepo() (*client.Repository, error) {
	repos, err := getRepos()
	if err != nil {
		return nil, err
	}

	return parseRepositoryString(repos[0])
}

// TODO: Find a better name
func GetClosestBranch(branches []string) (string, error) {
	// TODO: What is the history branches? Use BFS for looking up history. Perhaps git.GetLog()?
	r, err := getLocalRepo()
	if err != nil {
		return "", err
	}

	head, err := r.Head()
	if err != nil {
		return "", err
	}

	c, err := r.CommitObject(head.Hash())
	if err != nil {
		return "", err
	}

	type branchWrapper struct {
		c *object.Commit
		n string
	}

	cSlice := make([]branchWrapper, 0, len(branches))
	for _, v := range branches {
		bRef, err := r.Reference(plumbing.NewBranchReferenceName(v), false)
		if err != nil {
			continue
		}

		bCommit, err := r.CommitObject(bRef.Hash())
		if err != nil {
			continue
		}

		cSlice = append(cSlice, branchWrapper{bCommit, v})
	}

	if len(cSlice) == 0 {
		return "", ErrCannotFindAnyBranchReference
	}

	// TODO: Implement --log-depth flag
	p := c
	for i := 0; i < 10; i++ {
		p, err = p.Parent(0)
		if err != nil {
			return "", err
		}

		for _, v := range cSlice {
			if v.c.Hash == p.Hash {
				return v.n, nil
			}
		}
	}

	return "", ErrAncestorCommitNotFound
}
