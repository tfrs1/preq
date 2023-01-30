package decline

import (
	"preq/internal/gitutils"
	"preq/internal/pkg/client"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseArgs(t *testing.T) {
	t.Run("sets the ID to the value of first arg", func(t *testing.T) {
		id := "id"
		cmdArgs := parseArgs([]string{id})
		assert.Equal(t, cmdArgs.ID, id)
	})

	t.Run(
		"sets the ID as empty string if args are missing",
		func(t *testing.T) {
			cmdArgs := parseArgs([]string{})
			assert.Equal(t, cmdArgs.ID, "")
		},
	)
}

type MockGoGit struct {
	gitutils.GoGit
}

func (git *MockGoGit) GetRemoteInfo() (*client.Repository, error) {
	return &client.Repository{
		Provider: client.RepositoryProviderEnum.BITBUCKET,
		Name:     "repo-name",
		Owner:    "owner",
	}, nil
}

func Test_fillDefaultDeclineCmdParams(t *testing.T) {
	t.Run("another test", func(t *testing.T) {
		old := getWorkingDirectoryRepo
		defer func() { getWorkingDirectoryRepo = old }()
		getWorkingDirectoryRepo = func() (gitutils.GitUtilsClient, error) {
			return &MockGoGit{}, nil
		}

		params := cmdParams{}
		assert.Equal(
			t,
			params.Repository.Provider,
			client.RepositoryProviderEnum.BITBUCKET,
		)
		assert.Equal(t, params.Repository, "owner/repo-name")
	})
}
