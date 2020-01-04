package gitutil

import (
	"prctl/internal/fs"
	"prctl/mocks"
	client "prctl/pkg/bitbucket"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
)

func Test_getLocalRepo(t *testing.T) {
	oldGetWorkingDir := getWorkingDir
	oldOpenRepo := openRepo

	t.Run("fails if cannot get working dir", func(t *testing.T) {
		vErr := errors.New("wd err")
		getWorkingDir = func(fs.Filesystem) (string, error) { return "", vErr }

		_, err := getLocalRepo()
		assert.Error(t, err)
	})

	t.Run("fails if cannot open repo", func(t *testing.T) {
		vErr := errors.New("repo err")
		getWorkingDir = func(fs.Filesystem) (string, error) { return "", nil }
		openRepo = func(string) (*git.Repository, error) { return nil, vErr }

		_, err := getLocalRepo()
		assert.Error(t, err)
	})

	t.Run("succeeds if no error", func(t *testing.T) {
		getWorkingDir = func(fs.Filesystem) (string, error) { return "", nil }
		openRepo = func(string) (*git.Repository, error) { return &git.Repository{}, nil }

		_, err := getLocalRepo()
		assert.NoError(t, err)
	})

	getWorkingDir = oldGetWorkingDir
	openRepo = oldOpenRepo
}

func Test_getCheckedOutBranchShortName(t *testing.T) {
	vErr := errors.New("branch err")
	_, err := getCheckedOutBranchShortName(mocks.Repository{
		Err: vErr,
	})
	assert.EqualError(t, err, vErr.Error())
}

func TestGetBranch(t *testing.T) {
	oldGetLocalRepo := getLocalRepo
	oldGetCheckedOutBranchShortName := getCheckedOutBranchShortName

	t.Run("fails when cannot get repo", func(t *testing.T) {
		vErr := errors.New("repo err")
		getLocalRepo = func() (repository, error) { return nil, vErr }

		_, err := GetBranch()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("fails when cannot get branch", func(t *testing.T) {
		vErr := errors.New("branch err")
		getLocalRepo = func() (repository, error) {
			return mocks.Repository{Err: nil}, nil
		}
		getCheckedOutBranchShortName = func(repository) (string, error) {
			return "", vErr
		}

		_, err := GetBranch()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("succeeds otherwise", func(t *testing.T) {
		v := "branch-name"
		getLocalRepo = func() (repository, error) {
			return mocks.Repository{Err: nil}, nil
		}
		getCheckedOutBranchShortName = func(repository) (string, error) {
			return v, nil
		}

		r, err := GetBranch()
		assert.Equal(t, v, r)
		assert.NoError(t, err)
	})

	getLocalRepo = oldGetLocalRepo
	getCheckedOutBranchShortName = oldGetCheckedOutBranchShortName
}

func Test_getRemoteURLs(t *testing.T) {
	t.Run("fails when cannot get remotes", func(t *testing.T) {
		vErr := errors.New("remotes err")
		_, err := getRemoteURLs(mocks.Repository{Err: vErr})
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("succeeds otherwise", func(t *testing.T) {
		urls, err := getRemoteURLs(mocks.Repository{
			RemotesValue: []*git.Remote{
				git.NewRemote(nil, &config.RemoteConfig{
					URLs: []string{"url"},
				}),
			},
		})
		assert.Equal(t, 1, len(urls))
		assert.NoError(t, err)
		assert.Contains(t, urls, "url")
	})
}

func Test_getRepos(t *testing.T) {
	oldGetLocalRepo := getLocalRepo
	oldGetRemoteURLs := getRemoteURLs
	oldParseRepositoryString := parseRepositoryString

	t.Run("fails when cannot get repo", func(t *testing.T) {
		vErr := errors.New("repo err")
		getLocalRepo = func() (repository, error) { return nil, vErr }
		_, err := getRepos()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("fails when cannot get remote", func(t *testing.T) {
		vErr := errors.New("remote err")
		getLocalRepo = func() (repository, error) { return nil, nil }
		getRemoteURLs = func(r repository) ([]string, error) { return nil, vErr }
		_, err := getRepos()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("fails when cannot parse", func(t *testing.T) {
		vErr := errors.New("parse err")
		getLocalRepo = func() (repository, error) { return nil, nil }
		getRemoteURLs = func(r repository) ([]string, error) {
			return []string{"url"}, nil
		}
		parseRepositoryString = func(repoString string) (*client.Repository, error) { return nil, vErr }
		_, err := getRepos()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("succeeds otherwise", func(t *testing.T) {
		getLocalRepo = func() (repository, error) { return nil, nil }
		getRemoteURLs = func(r repository) ([]string, error) {
			return []string{"url"}, nil
		}
		parseRepositoryString = func(repoString string) (*client.Repository, error) {
			return &client.Repository{}, nil
		}

		repos, err := getRepos()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(repos))
	})

	getLocalRepo = oldGetLocalRepo
	getRemoteURLs = oldGetRemoteURLs
	parseRepositoryString = oldParseRepositoryString
}

func Test_parseRemoteRepositoryURI(t *testing.T) {
	t.Run("fails on empty string", func(t *testing.T) {
		_, err := extractRepositoryTokens("")
		assert.EqualError(t, err, ErrUnableToParseRemoteRepositoryURI.Error())
	})

	t.Run("succeds on Bitbucket cloud SSH URI", func(t *testing.T) {
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

func TestGetRepo(t *testing.T) {
	oldGetRepos := getRepos

	t.Run("", func(t *testing.T) {
		vErr := errors.New("repos err")
		getRepos = func() ([]*client.Repository, error) { return nil, vErr }
		_, err := GetRepo()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("", func(t *testing.T) {
		getRepos = func() ([]*client.Repository, error) {
			return []*client.Repository{&client.Repository{}}, nil
		}

		repos, err := GetRepo()
		assert.NoError(t, err)
		assert.NotNil(t, repos)
	})

	getRepos = oldGetRepos
}

func TestGetClosestBranch(t *testing.T) {
	oldGetLocalRepo := getLocalRepo

	t.Run("", func(t *testing.T) {
		vErr := errors.New("repo err")
		getLocalRepo = func() (repository, error) { return nil, vErr }
		_, err := GetClosestBranch([]string{})
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("", func(t *testing.T) {
		vErr := errors.New("head err")
		getLocalRepo = func() (repository, error) {
			return mocks.Repository{
				Err: vErr,
			}, nil
		}
		_, err := GetClosestBranch([]string{})
		assert.EqualError(t, err, vErr.Error())
	})

	getLocalRepo = oldGetLocalRepo
}
