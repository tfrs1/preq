package gitutils

import (
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test_repository_GetCheckedOutBranchShortName(t *testing.T) {
	vErr := errors.New("branch err")
	r := &repository{
		r: &MockGoGitRepository{
			Err: vErr,
		},
	}

	_, err := r.GetCheckedOutBranchShortName()
	assert.EqualError(t, err, vErr.Error())
}

func Test_repository_GetRemoteURLs(t *testing.T) {
	t.Run("fails when cannot get remotes", func(t *testing.T) {
		vErr := errors.New("remotes err")
		r := &repository{
			r: &MockGoGitRepository{
				Err: vErr,
			},
		}

		_, err := r.GetRemoteURLs()
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("succeeds otherwise", func(t *testing.T) {
		r := &repository{
			r: &MockGoGitRepository{
				RemotesValue: []*git.Remote{
					git.NewRemote(nil, &config.RemoteConfig{
						URLs: []string{"url"},
					}),
				},
			},
		}

		urls, err := r.GetRemoteURLs()
		assert.Equal(t, 1, len(urls))
		assert.NoError(t, err)
		assert.Contains(t, urls, "url")
	})
}

func Test_repository_CurrentCommit(t *testing.T) {
	t.Run("", func(t *testing.T) {
		vErr := errors.New("commit err")
		r := &repository{
			r: &MockGoGitRepository{
				Err: vErr,
			},
		}

		_, err := r.CurrentCommit()
		assert.EqualError(t, err, vErr.Error())
	})
}

func Test_repository_BranchCommit(t *testing.T) {
	t.Run("", func(t *testing.T) {
		vErr := errors.New("branch err")
		r := &repository{
			r: &MockGoGitRepository{
				Err: vErr,
			},
		}

		_, err := r.BranchCommit("")
		assert.EqualError(t, err, vErr.Error())
	})
}
