package gitutil

import (
	"prctl/internal/fs"
	client "prctl/pkg/bitbucket"
	"regexp"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

var (
	ErrCannotGetLocalRepository         = errors.New("cannot get local repository")
	ErrUnableToParseRemoteRepositoryURI = errors.New("unabled to parse remote repository URI")
	ErrAncestorCommitNotFound           = errors.New("ancestor commit not found")
	ErrCannotFindAnyBranchReference     = errors.New("cannot find any branch reference")
)

var getWorkingDir = func(fs fs.Filesystem) (string, error) {
	return fs.Getwd()
}

var openRepo = func(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}

type repository interface {
	Head() (*plumbing.Reference, error)
	Remotes() ([]*git.Remote, error)
	Reference(plumbing.ReferenceName, bool) (*plumbing.Reference, error)
	CommitObject(plumbing.Hash) (*object.Commit, error)
}

var getLocalRepo = func() (repository, error) {
	wd, err := getWorkingDir(fs.OS{})
	if err != nil {
		return nil, errors.Wrap(err, ErrCannotGetLocalRepository.Error())
	}

	r, err := openRepo(wd)
	if err != nil {
		return nil, errors.Wrap(err, ErrCannotGetLocalRepository.Error())
	}

	return r, nil
}

var getCheckedOutBranchShortName = func(r repository) (string, error) {
	headRef, err := r.Head()
	if err != nil {
		return "", err
	}

	return headRef.Name().Short(), nil
}

func GetBranch() (string, error) {
	r, err := getLocalRepo()
	if err != nil {
		return "", err
	}

	return getCheckedOutBranchShortName(r)
}

var getRemoteURLs = func(r repository) ([]string, error) {
	var repoURLs []string
	remotes, err := r.Remotes()
	if err != nil {
		return nil, err
	}

	for _, re := range remotes {
		repoURLs = append(repoURLs, re.Config().URLs...)
	}

	return repoURLs, nil
}

var getRepos = func() ([]*client.Repository, error) {
	var repos []*client.Repository
	r, err := getLocalRepo()
	if err != nil {
		return nil, err
	}

	repoURLs, err := getRemoteURLs(r)
	if err != nil {
		return nil, err
	}

	for _, url := range repoURLs {
		pRepo, err := parseRepositoryString(url)
		if err != nil {
			return nil, err
		}

		repos = append(repos, pRepo)
	}

	return repos, nil
}

var extractRepositoryTokens = func(uri string) ([]string, error) {
	r := regexp.MustCompile(`git@(.*):(.*)/(.*)\.git`)
	m := r.FindStringSubmatch(uri)
	if len(m) != 4 {
		return nil, ErrUnableToParseRemoteRepositoryURI
	}

	return m[1:], nil
}

var parseRepositoryProvider = func(p string) (client.RepositoryProvider, error) {
	return client.ParseRepositoryProvider(p)
}

var parseRepositoryString = func(repoString string) (*client.Repository, error) {
	m, err := extractRepositoryTokens(repoString)
	if err != nil {
		return nil, err
	}

	p, err := parseRepositoryProvider(m[0])
	if err != nil {
		return nil, err
	}

	return &client.Repository{
		Provider: p,
		Owner:    m[1],
		Name:     m[2],
	}, nil
}

func GetRepo() (*client.Repository, error) {
	repos, err := getRepos()
	if err != nil {
		return nil, err
	}

	return repos[0], nil
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
