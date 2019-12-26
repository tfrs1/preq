package gitutil

import (
	"errors"
	"os"

	"gopkg.in/src-d/go-git.v4"
)

var (
	ErrCannotGetLocalRepository = errors.New("cannot get local repository")
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

	// for _, re := range remotes {
	// 	repos = append(repos, re.Config().URLs...)
	// }

	// // return repos, nil
}

func GetRepos() ([]string, error) {
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

func GetRepo() (string, error) {
	repos, err := GetRepos()
	if err != nil {
		return "", err
	}

	return repos[0], nil
}
