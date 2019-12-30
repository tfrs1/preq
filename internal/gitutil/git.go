package gitutil

import (
	"errors"
	"os"
	client "prctl/pkg/bitbucket"
	"regexp"

	"gopkg.in/src-d/go-git.v4"
)

var (
	ErrCannotGetLocalRepository            = errors.New("cannot get local repository")
	ErrUnableToDetermineRepositoryProvider = errors.New("unable to determine repository provider")
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

func getRepos() ([]string, error) {
	var repos []string
	wd, err := os.Getwd()
	if err != nil {
		return nil, ErrCannotGetLocalRepository
	}

	r, err := git.PlainOpen(wd)
	if err != nil {
		return nil, ErrCannotGetLocalRepository
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
