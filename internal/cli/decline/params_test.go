package decline

import (
	"preq/internal/cli/paramutils"
	"preq/internal/errcodes"
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
		fillDefaultDeclineCmdParams(&params)
		assert.Equal(
			t,
			params.Provider,
			client.RepositoryProviderEnum.BITBUCKET,
		)
		assert.Equal(t, params.Repository, "owner/repo-name")
	})
}

func Test_validateFlagDeclineCmdParams(t *testing.T) {
	t.Run("no param", func(t *testing.T) {
		params := &cmdParams{}
		err := validateFlagDeclineCmdParams(params)
		assert.Equal(t, nil, err)
	})

	t.Run("only repo", func(t *testing.T) {
		params := &cmdParams{
			Repository: "owner/repo-name",
		}
		err := validateFlagDeclineCmdParams(params)
		assert.Equal(t, errcodes.ErrSomeRepoParamsMissing, err)
	})

	t.Run("only provider", func(t *testing.T) {
		params := &cmdParams{
			Provider: "provider",
		}
		err := validateFlagDeclineCmdParams(params)
		assert.Equal(t, errcodes.ErrSomeRepoParamsMissing, err)
	})

	t.Run("wrong repo", func(t *testing.T) {
		params := &cmdParams{
			Repository: "wrong",
			Provider:   "provider",
		}
		err := validateFlagDeclineCmdParams(params)
		assert.Equal(t, errcodes.ErrRepositoryMustBeInFormOwnerRepo, err)
	})

	t.Run("wrong provider", func(t *testing.T) {
		params := &cmdParams{
			Repository: "owner/repo-name",
			Provider:   "wrong",
		}
		err := validateFlagDeclineCmdParams(params)
		assert.Equal(t, errcodes.ErrorRepositoryProviderUnknown, err)
	})

	t.Run("succeeds with valid repo and provider", func(t *testing.T) {
		params := &cmdParams{
			Repository: "owner/repo-name",
			Provider:   client.RepositoryProviderEnum.BITBUCKET,
		}
		err := validateFlagDeclineCmdParams(params)
		assert.Equal(t, nil, err)
	})
}

func Test_fillFlagDeclineCmdParams(t *testing.T) {
	t.Run("fills with flag parameters", func(t *testing.T) {
		repo := "owner/repo"
		params := cmdParams{}
		fillFlagDeclineCmdParams(
			&paramutils.MockPreqFlagSet{StringMap: map[string]interface{}{
				"repository": repo,
				"provider":   string(client.RepositoryProviderEnum.BITBUCKET),
			}},
			&params,
		)

		assert.Equal(t, params.Repository, repo)
		assert.Equal(
			t,
			params.Provider,
			client.RepositoryProviderEnum.BITBUCKET,
		)
	})

	t.Run("fills with fallback parameters", func(t *testing.T) {
		params := cmdParams{}
		fillFlagDeclineCmdParams(
			&paramutils.MockPreqFlagSet{},
			&params,
		)

		assert.Equal(t, params.Repository, "")
		assert.Equal(t, params.Provider.IsValid(), false)
	})
}
