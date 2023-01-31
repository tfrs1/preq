package gitutils

import (
	"preq/internal/pkg/client"
	"preq/internal/pkg/fs"
	"regexp"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

var (
	ErrCannotGetLocalRepository = errors.New(
		"cannot get local repository",
	)
	ErrUnableToParseRemoteRepositoryURI = errors.New(
		"unable to parse remote repository URI",
	)
	ErrAncestorCommitNotFound = errors.New(
		"ancestor commit not found",
	)
	ErrCannotFindAnyBranchReference = errors.New(
		"cannot find any branch reference",
	)
)

func IsDirGitRepo(path string) bool {
	if path == "" {
		return false
	}

	_, err := openRepo(path)
	return err == nil
}

var getWorkingDir = func(fs fs.Filesystem) (string, error) {
	return fs.Getwd()
}

var getRemoteInfoList = func(git *GoGit) ([]*client.Repository, error) {
	var repos []*client.Repository

	repoURLs, err := git.Git.GetRemoteURLs()
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
	r := regexp.MustCompile(`@(.*):(.*)/(.*)\.git$`)
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

type branchCommitMap map[string]*object.Commit

var getBranchCommits = func(r gitRepository, branches []string) (branchCommitMap, error) {
	cSlice := make(branchCommitMap)
	for _, v := range branches {
		bCommit, err := r.BranchCommit(v)
		if err != nil {
			continue
		}

		cSlice[v] = bCommit
	}

	if len(cSlice) == 0 {
		return nil, ErrCannotFindAnyBranchReference
	}

	return cSlice, nil
}

// TODO: Find a more appropriate name
func walkHistory(
	c *object.Commit,
	goalMap branchCommitMap,
	depth int,
) (string, error) {
	p := c
	for i := 0; i < depth; i++ {
		p, err := p.Parent(0)
		if err != nil {
			return "", err
		}

		for b, v := range goalMap {
			if v.Hash == p.Hash {
				return b, nil
			}
		}
	}

	return "", ErrAncestorCommitNotFound
}

type GitUtilsClient interface {
	// GetClosestBranch documentation
	// TODO: Find a better name
	GetClosestBranch(branches []string) (string, error)
	GetRemoteInfo() (*client.Repository, error)
	GetCurrentCommitMessage() (string, error)
	GetCurrentBranch() (string, error)
	GetBranchLastCommitMessage(name string) (string, error)
}

type GoGit struct {
	Git gitRepository
}

func new(path string) (GitUtilsClient, error) {
	repo, err := openRepo(path)
	return &GoGit{Git: &repository{
		r: repo,
	}}, err
}

// TODO: Return a Branch type? Or rename to GetClosesBranchName?
func (git *GoGit) GetClosestBranch(branches []string) (string, error) {
	// TODO: What if the history branches? Use BFS for looking up history. Perhaps git.GetLog()?
	c, err := git.Git.CurrentCommit()
	if err != nil {
		return "", err
	}

	cSlice, err := getBranchCommits(git.Git, branches)
	if err != nil {
		return "", err
	}

	// TODO: Implement --log-depth flag
	return walkHistory(c, cSlice, 10)
}

func (git *GoGit) GetRemoteInfo() (*client.Repository, error) {
	repos, err := getRemoteInfoList(git)
	if err != nil {
		return nil, err
	}

	return repos[0], nil
}

func (git *GoGit) GetBranchLastCommitMessage(name string) (string, error) {
	c, err := git.Git.BranchCommit(name)
	if err != nil {
		return "", err
	}

	return c.Message, nil
}

func (git *GoGit) GetCurrentCommitMessage() (string, error) {
	c, err := git.Git.CurrentCommit()
	if err != nil {
		return "", err
	}

	return c.Message, nil
}

func (git *GoGit) GetCurrentBranch() (string, error) {
	return git.Git.GetCheckedOutBranchShortName()
}

func GetWorkingDirectoryRepo() (GitUtilsClient, error) {
	wd, err := getWorkingDir(fs.OS{})
	if err != nil {
		return nil, err
	}

	return new(wd)
}

func GetRepo(path string) (GitUtilsClient, error) {
	return new(path)
}
