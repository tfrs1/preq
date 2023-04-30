package gitutils

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"preq/internal/pkg/client"
	"preq/internal/pkg/fs"
	"regexp"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
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

// Return the root diretory of a Git repository when the supplied path
// is a subdirectory of a Git repository. Errors if the root cannot be found.
func GetRepoRootDir(path string) (string, error) {
	for {
		_, err := git.PlainOpen(path)
		if err == nil {
			return path, nil
		}

		// Exit if we are at the root
		parent := filepath.Dir(path)
		if path == parent {
			break
		}

		// Move up to the parent directory
		path = parent
	}

	// Reached the root without finding a repo
	return "", fmt.Errorf("cannot find a git repository at %s", path)
}

var getWorkingDir = func(fs fs.Filesystem) (string, error) {
	return fs.Getwd()
}

var getRemoteInfoList = func(git *GoGit, aliases map[client.RepositoryProvider][]string) ([]*client.Repository, error) {
	var repos []*client.Repository

	repoURLs, err := git.Git.GetRemoteURLs()
	if err != nil {
		return nil, err
	}

	for _, url := range repoURLs {
		pRepo, err := parseRepositoryString(url, aliases)
		if err != nil {
			return nil, err
		}

		repos = append(repos, pRepo)
	}

	return repos, nil
}

var extractRepositoryTokens = func(uri string) ([]string, error) {
	r := regexp.MustCompile(`@([^:/]+)[:/](.*)\.git$`)
	m := r.FindStringSubmatch(uri)
	if len(m) != 3 {
		return nil, ErrUnableToParseRemoteRepositoryURI
	}

	return m[1:], nil
}

var parseRepositoryProvider = func(p string, aliases map[client.RepositoryProvider][]string) (client.RepositoryProvider, error) {
	return client.ParseRepositoryProvider(p, aliases)
}

var parseRepositoryString = func(repoString string, aliases map[client.RepositoryProvider][]string) (*client.Repository, error) {
	m, err := extractRepositoryTokens(repoString)
	if err != nil {
		return nil, err
	}

	p, err := parseRepositoryProvider(m[0], aliases)
	if err != nil {
		return nil, err
	}

	return &client.Repository{
		Provider: p,
		Name:     m[1],
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
	GetRemoteInfo(
		aliases map[client.RepositoryProvider][]string,
	) (*client.Repository, error)
	GetCurrentCommitMessage() (string, error)
	GetCurrentBranch() (string, error)
	GetBranchLastCommitMessage(name string) (string, error)
	GetDiffPatch(fromHash string, toHash string) ([]byte, error)
}

type GoGit struct {
	Git   gitRepository
	goGit *git.Repository
}

func new(path string) (GitUtilsClient, error) {
	repo, err := openRepo(path)
	return &GoGit{
		goGit: repo,
		Git: &repository{
			r: repo,
		},
	}, err
}

func (git *GoGit) GetDiffPatch(fromHash string, toHash string) ([]byte, error) {
	cmd := exec.Command("git", "diff", fmt.Sprintf("%s...%s", fromHash, toHash))
	cfg, err := git.goGit.Worktree()
	if err != nil {
		return nil, err
	}

	root := cfg.Filesystem.Root()
	cmd.Dir = root

	// Capture the output of the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return output, nil
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

func (git *GoGit) GetRemoteInfo(
	aliases map[client.RepositoryProvider][]string,
) (*client.Repository, error) {
	repos, err := getRemoteInfoList(git, aliases)
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
