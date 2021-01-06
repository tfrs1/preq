package close

import (
	"preq/internal/cli/paramutils"
	"preq/internal/config"
	"preq/internal/errcodes"
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

	t.Run("sets the ID as empty string if args are missing", func(t *testing.T) {
		cmdArgs := parseArgs([]string{})
		assert.Equal(t, cmdArgs.ID, "")
	})
}

func Test_fillDefaultCloseCmdParams(t *testing.T) {
	t.Run("another test", func(t *testing.T) {
		old := getRemoteInfo
		defer func() { getRemoteInfo = old }()
		getRemoteInfo = func() (*client.Repository, error) {
			return &client.Repository{
				Provider: client.RepositoryProviderEnum.BITBUCKET,
				Name:     "repo-name",
				Owner:    "owner",
			}, nil
		}

		params := cmdParams{}
		fillDefaultCloseCmdParams(&params)
		assert.Equal(t, params.Repository.Provider, client.RepositoryProviderEnum.BITBUCKET)
		assert.Equal(t, params.Repository.Name, "owner/repo-name")
	})
}

func Test_validateFlagCloseCmdParams(t *testing.T) {
	t.Run("no param", func(t *testing.T) {
		params := &cmdParams{}
		err := validateFlagCloseCmdParams(params)
		assert.Equal(t, nil, err)
	})

	t.Run("only repo", func(t *testing.T) {
		params := &cmdParams{
			Repository: config.RepositoryParams{
				Name: "owner/repo-name",
			},
		}
		err := validateFlagCloseCmdParams(params)
		assert.Equal(t, errcodes.ErrSomeRepoParamsMissing, err)
	})

	t.Run("only provider", func(t *testing.T) {
		params := &cmdParams{
			Repository: config.RepositoryParams{
				Provider: "provider",
			},
		}
		err := validateFlagCloseCmdParams(params)
		assert.Equal(t, errcodes.ErrSomeRepoParamsMissing, err)
	})

	t.Run("wrong repo", func(t *testing.T) {
		params := &cmdParams{
			Repository: config.RepositoryParams{
				Name:     "wrong",
				Provider: "provider",
			},
		}
		err := validateFlagCloseCmdParams(params)
		assert.Equal(t, errcodes.ErrRepositoryMustBeInFormOwnerRepo, err)
	})

	t.Run("wrong provider", func(t *testing.T) {
		params := &cmdParams{
			Repository: config.RepositoryParams{
				Name:     "owner/repo-name",
				Provider: "wrong",
			},
		}
		err := validateFlagCloseCmdParams(params)
		assert.Equal(t, errcodes.ErrorRepositoryProviderUnknown, err)
	})

	t.Run("succeeds with valid repo and provider", func(t *testing.T) {
		params := &cmdParams{
			Repository: config.RepositoryParams{
				Name:     "owner/repo-name",
				Provider: client.RepositoryProviderEnum.BITBUCKET,
			},
		}
		err := validateFlagCloseCmdParams(params)
		assert.Equal(t, nil, err)
	})
}

func Test_fillFlagCloseCmdParams(t *testing.T) {
	t.Run("fills with flag parameters", func(t *testing.T) {
		repo := "owner/repo"
		params := cmdParams{}
		fillFlagCloseCmdParams(
			&paramutils.MockPreqFlagSet{StringMap: map[string]interface{}{
				"repository": repo,
				"provider":   string(client.RepositoryProviderEnum.BITBUCKET),
			}},
			&params,
		)

		assert.Equal(t, params.Repository.Name, repo)
		assert.Equal(t, params.Repository.Provider, client.RepositoryProviderEnum.BITBUCKET)
	})

	t.Run("fills with fallback parameters", func(t *testing.T) {
		params := cmdParams{}
		fillFlagCloseCmdParams(
			&paramutils.MockPreqFlagSet{},
			&params,
		)

		assert.Equal(t, params.Repository.Name, "")
		assert.Equal(t, params.Repository.Provider.IsValid(), false)
	})
}
