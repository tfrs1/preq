package gitutils

import (
	"preq/internal/pkg/client"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func TestGetCurrentBranch(t *testing.T) {
	t.Run("fails when cannot get branch", func(t *testing.T) {
		vErr := errors.New("branch err")
		r := &GoGit{
			Git: &MockGitRepository{ErrorValue: vErr},
		}

		_, err := r.GetCurrentBranch()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("succeeds otherwise", func(t *testing.T) {
		v := "branch-name"
		git := &GoGit{
			Git: &MockGitRepository{CurrentBranchValue: v},
		}

		r, err := git.GetCurrentBranch()
		assert.Equal(t, v, r)
		assert.NoError(t, err)
	})
}

func Test_getRemoteInfoList(t *testing.T) {
	oldParseRepositoryString := parseRepositoryString

	t.Run("fails when cannot get repo", func(t *testing.T) {
		vErr := errors.New("repo err")
		_, err := getRemoteInfoList(&GoGit{
			Git: &MockGitRepository{
				ErrorValue: vErr,
			},
		})
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("fails when cannot get remote", func(t *testing.T) {
		vErr := errors.New("remote err")

		_, err := getRemoteInfoList(&GoGit{
			Git: &MockGitRepository{
				ErrorValue: vErr,
			},
		})
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("fails when cannot parse", func(t *testing.T) {
		vErr := errors.New("parse err")
		parseRepositoryString = func(repoString string) (*client.Repository, error) { return nil, vErr }

		_, err := getRemoteInfoList(&GoGit{
			Git: &MockGitRepository{
				RemoteURLsValue: []string{"url"},
			},
		})
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("succeeds otherwise", func(t *testing.T) {
		parseRepositoryString = func(repoString string) (*client.Repository, error) {
			return &client.Repository{}, nil
		}

		repos, err := getRemoteInfoList(&GoGit{
			Git: &MockGitRepository{
				RemoteURLsValue: []string{"url"},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(repos))
	})

	parseRepositoryString = oldParseRepositoryString
}

func Test_parseRemoteRepositoryURI(t *testing.T) {
	t.Run("fails on empty string", func(t *testing.T) {
		_, err := extractRepositoryTokens("")
		assert.EqualError(t, err, ErrUnableToParseRemoteRepositoryURI.Error())
	})

	t.Run("succeeds on Bitbucket cloud SSH URI", func(t *testing.T) {
		v, err := extractRepositoryTokens("git@provider:owner/repo.git")
		assert.NoError(t, err)
		assert.Equal(t, 3, len(v))
		assert.Equal(t, "provider", v[0])
		assert.Equal(t, "owner", v[1])
		assert.Equal(t, "repo", v[2])
	})
}

func Test_parseRepositoryString(t *testing.T) {
	oldExtractRepositoryTokens := extractRepositoryTokens
	oldParseRepositoryProvider := parseRepositoryProvider

	t.Run("fails when cannot parse remote", func(t *testing.T) {
		vErr := errors.New("remote err")
		extractRepositoryTokens = func(uri string) ([]string, error) { return nil, vErr }
		_, err := parseRepositoryString("")
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("fails when cannot parse remote", func(t *testing.T) {
		vErr := errors.New("provider err")
		extractRepositoryTokens = func(uri string) ([]string, error) { return []string{""}, nil }
		parseRepositoryProvider = func(p string) (client.RepositoryProvider, error) {
			return client.RepositoryProviderEnum.BITBUCKET, vErr
		}

		_, err := parseRepositoryString("")
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("fails when cannot parse provider", func(t *testing.T) {
		vErr := errors.New("provider err")
		extractRepositoryTokens = func(uri string) ([]string, error) { return []string{""}, nil }
		parseRepositoryProvider = func(p string) (client.RepositoryProvider, error) {
			return client.RepositoryProviderEnum.BITBUCKET, vErr
		}

		_, err := parseRepositoryString("")
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("succeeds otherwise", func(t *testing.T) {
		extractRepositoryTokens = func(uri string) ([]string, error) { return []string{"", "owner", "repo"}, nil }
		parseRepositoryProvider = func(p string) (client.RepositoryProvider, error) {
			return client.RepositoryProviderEnum.BITBUCKET, nil
		}

		v, err := parseRepositoryString("")
		assert.NoError(t, err)
		assert.Equal(t, client.RepositoryProviderEnum.BITBUCKET, v.Provider)
		assert.Equal(t, "owner", v.Owner)
		assert.Equal(t, "repo", v.Name)
	})

	extractRepositoryTokens = oldExtractRepositoryTokens
	parseRepositoryProvider = oldParseRepositoryProvider
}

func TestGetRemoteInfo(t *testing.T) {
	oldGetRepos := getRemoteInfoList

	t.Run("fails when getRemoteInfoList fails", func(t *testing.T) {
		vErr := errors.New("repos err")
		getRemoteInfoList = func(git *GoGit) ([]*client.Repository, error) {
			return nil, vErr
		}
		r := &GoGit{}
		_, err := r.GetRemoteInfo()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("", func(t *testing.T) {
		getRemoteInfoList = func(git *GoGit) ([]*client.Repository, error) {
			return []*client.Repository{&client.Repository{}}, nil
		}
		r := &GoGit{}

		repos, err := r.GetRemoteInfo()
		assert.NoError(t, err)
		assert.NotNil(t, repos)
	})

	getRemoteInfoList = oldGetRepos
}

func Test_getBranchCommits(t *testing.T) {
	t.Run("fail when no goals", func(t *testing.T) {
		_, err := getBranchCommits(nil, []string{})
		assert.EqualError(t, err, ErrCannotFindAnyBranchReference.Error())
	})

	t.Run("fail when cannot find any branch commits", func(t *testing.T) {
		vErr := errors.New("branch err")
		r := &MockGitRepository{ErrorValue: vErr}

		_, err := getBranchCommits(r, []string{""})
		assert.EqualError(t, err, ErrCannotFindAnyBranchReference.Error())
	})

	t.Run("fail when cannot find any branch commits", func(t *testing.T) {
		r := &MockGitRepository{BranchCommitValue: &object.Commit{}}

		res, err := getBranchCommits(r, []string{""})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(res))
	})
}

func TestGetClosestBranch(t *testing.T) {
	oldGetBranchCommits := getBranchCommits

	t.Run("fails when repository fails", func(t *testing.T) {
		vErr := errors.New("head err")
		r := &GoGit{
			Git: &MockGitRepository{ErrorValue: vErr},
		}

		_, err := r.GetClosestBranch([]string{})
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("fails when gitBranchCommits fails", func(t *testing.T) {
		vErr := errors.New("branch commits err")
		r := &GoGit{
			Git: &MockGitRepository{},
		}
		getBranchCommits = func(
			r gitRepository, branches []string,
		) (branchCommitMap, error) {
			return nil, vErr
		}

		_, err := r.GetClosestBranch([]string{})
		assert.EqualError(t, err, vErr.Error())
	})

	getBranchCommits = oldGetBranchCommits
}

func TestGetCurrentCommitMessage(t *testing.T) {
	t.Run("fails when cannot get commit", func(t *testing.T) {
		vErr := errors.New("commit err")
		r := &GoGit{
			Git: &MockGitRepository{ErrorValue: vErr},
		}
		_, err := r.GetCurrentCommitMessage()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("succeeds otherwise", func(t *testing.T) {
		msg := "message"
		r := &GoGit{
			Git: &MockGitRepository{Commit: &object.Commit{Message: msg}},
		}
		val, err := r.GetCurrentCommitMessage()
		assert.NoError(t, err)
		assert.Equal(t, msg, val)
	})
}
