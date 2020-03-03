package gitutil

import (
	"preq/internal/fs"
	"preq/mocks"
	client "preq/pkg/bitbucket"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func Test_openLocalRepo(t *testing.T) {
	oldGetWorkingDir := getWorkingDir
	oldOpenRepo := openRepo

	t.Run("fails if cannot get working dir", func(t *testing.T) {
		vErr := errors.New("wd err")
		getWorkingDir = func(fs.Filesystem) (string, error) { return "", vErr }

		_, err := openLocalRepo()
		assert.Error(t, err)
	})

	t.Run("fails if cannot open repo", func(t *testing.T) {
		vErr := errors.New("repo err")
		getWorkingDir = func(fs.Filesystem) (string, error) { return "", nil }
		openRepo = func(string) (*git.Repository, error) { return nil, vErr }

		_, err := openLocalRepo()
		assert.Error(t, err)
	})

	t.Run("succeeds if no error", func(t *testing.T) {
		getWorkingDir = func(fs.Filesystem) (string, error) { return "", nil }
		openRepo = func(string) (*git.Repository, error) { return &git.Repository{}, nil }

		_, err := openLocalRepo()
		assert.NoError(t, err)
	})

	getWorkingDir = oldGetWorkingDir
	openRepo = oldOpenRepo
}

func TestGetCurrentBranch(t *testing.T) {
	oldOpenLocalRepo := openLocalRepo

	t.Run("fails when cannot get repo", func(t *testing.T) {
		vErr := errors.New("repo err")
		openLocalRepo = func() (gitRepository, error) { return nil, vErr }

		_, err := GetCurrentBranch()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("fails when cannot get branch", func(t *testing.T) {
		vErr := errors.New("branch err")
		openLocalRepo = func() (gitRepository, error) {
			return &mocks.GitRepository{ErrorValue: vErr}, nil
		}

		_, err := GetCurrentBranch()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("succeeds otherwise", func(t *testing.T) {
		v := "branch-name"
		openLocalRepo = func() (gitRepository, error) {
			return &mocks.GitRepository{CurrentBranchValue: v}, nil
		}

		r, err := GetCurrentBranch()
		assert.Equal(t, v, r)
		assert.NoError(t, err)
	})

	openLocalRepo = oldOpenLocalRepo
}

func Test_getRemoteInfoList(t *testing.T) {
	oldopenLocalRepo := openLocalRepo
	oldParseRepositoryString := parseRepositoryString

	t.Run("fails when cannot get repo", func(t *testing.T) {
		vErr := errors.New("repo err")
		openLocalRepo = func() (gitRepository, error) { return nil, vErr }

		_, err := getRemoteInfoList()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("fails when cannot get remote", func(t *testing.T) {
		vErr := errors.New("remote err")
		openLocalRepo = func() (gitRepository, error) {
			return &mocks.GitRepository{
				ErrorValue: vErr,
			}, nil
		}

		_, err := getRemoteInfoList()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("fails when cannot parse", func(t *testing.T) {
		vErr := errors.New("parse err")
		openLocalRepo = func() (gitRepository, error) {
			return &mocks.GitRepository{
				RemoteURLsValue: []string{"url"},
			}, nil
		}
		parseRepositoryString = func(repoString string) (*client.Repository, error) { return nil, vErr }

		_, err := getRemoteInfoList()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("succeeds otherwise", func(t *testing.T) {
		openLocalRepo = func() (gitRepository, error) {
			return &mocks.GitRepository{
				RemoteURLsValue: []string{"url"},
			}, nil
		}
		parseRepositoryString = func(repoString string) (*client.Repository, error) {
			return &client.Repository{}, nil
		}

		repos, err := getRemoteInfoList()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(repos))
	})

	openLocalRepo = oldopenLocalRepo
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
			return client.RepositoryProvider_BITBUCKET_CLOUD, vErr
		}

		_, err := parseRepositoryString("")
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("fails when cannot parse provider", func(t *testing.T) {
		vErr := errors.New("provider err")
		extractRepositoryTokens = func(uri string) ([]string, error) { return []string{""}, nil }
		parseRepositoryProvider = func(p string) (client.RepositoryProvider, error) {
			return client.RepositoryProvider_BITBUCKET_CLOUD, vErr
		}

		_, err := parseRepositoryString("")
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("succeeds otherwise", func(t *testing.T) {
		extractRepositoryTokens = func(uri string) ([]string, error) { return []string{"", "owner", "repo"}, nil }
		parseRepositoryProvider = func(p string) (client.RepositoryProvider, error) {
			return client.RepositoryProvider_BITBUCKET_CLOUD, nil
		}

		v, err := parseRepositoryString("")
		assert.NoError(t, err)
		assert.Equal(t, client.RepositoryProvider_BITBUCKET_CLOUD, v.Provider)
		assert.Equal(t, "owner", v.Owner)
		assert.Equal(t, "repo", v.Name)
	})

	extractRepositoryTokens = oldExtractRepositoryTokens
	parseRepositoryProvider = oldParseRepositoryProvider
}

func TestGetRemoteInfo(t *testing.T) {
	oldGetRepos := getRemoteInfoList

	t.Run("", func(t *testing.T) {
		vErr := errors.New("repos err")
		getRemoteInfoList = func() ([]*client.Repository, error) { return nil, vErr }
		_, err := GetRemoteInfo()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("", func(t *testing.T) {
		getRemoteInfoList = func() ([]*client.Repository, error) {
			return []*client.Repository{&client.Repository{}}, nil
		}

		repos, err := GetRemoteInfo()
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
		r := &mocks.GitRepository{ErrorValue: vErr}

		_, err := getBranchCommits(r, []string{""})
		assert.EqualError(t, err, ErrCannotFindAnyBranchReference.Error())
	})

	t.Run("fail when cannot find any branch commits", func(t *testing.T) {
		r := &mocks.GitRepository{BranchCommitValue: &object.Commit{}}

		res, err := getBranchCommits(r, []string{""})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(res))
	})
}

func TestGetClosestBranch(t *testing.T) {
	oldOpenLocalRepo := openLocalRepo
	oldGetBranchCommits := getBranchCommits

	t.Run("", func(t *testing.T) {
		vErr := errors.New("repo err")
		openLocalRepo = func() (gitRepository, error) { return nil, vErr }

		_, err := GetClosestBranch([]string{})
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("", func(t *testing.T) {
		vErr := errors.New("head err")
		openLocalRepo = func() (gitRepository, error) {
			return &mocks.GitRepository{ErrorValue: vErr}, nil
		}

		_, err := GetClosestBranch([]string{})
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("", func(t *testing.T) {
		vErr := errors.New("branch commits err")
		openLocalRepo = func() (gitRepository, error) {
			return &mocks.GitRepository{}, nil
		}
		getBranchCommits = func(r gitRepository, branches []string) (branchCommitMap, error) { return nil, vErr }

		_, err := GetClosestBranch([]string{})
		assert.EqualError(t, err, vErr.Error())
	})

	openLocalRepo = oldOpenLocalRepo
	getBranchCommits = oldGetBranchCommits
}

func TestGetCurrentCommitMessage(t *testing.T) {
	oldOpenLocalRepo := openLocalRepo

	t.Run("fails when cannot open repo", func(t *testing.T) {
		vErr := errors.New("repo err")
		openLocalRepo = func() (gitRepository, error) { return nil, vErr }
		_, err := GetCurrentCommitMessage()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("fails when cannot get commit", func(t *testing.T) {
		vErr := errors.New("commit err")
		openLocalRepo = func() (gitRepository, error) {
			return &mocks.GitRepository{
				ErrorValue: vErr,
			}, nil
		}
		_, err := GetCurrentCommitMessage()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("succeeds otherwise", func(t *testing.T) {
		msg := "message"
		openLocalRepo = func() (gitRepository, error) {
			return &mocks.GitRepository{Commit: &object.Commit{Message: msg}}, nil
		}
		val, err := GetCurrentCommitMessage()
		assert.NoError(t, err)
		assert.Equal(t, msg, val)
	})

	openLocalRepo = oldOpenLocalRepo
}
