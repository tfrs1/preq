package decline

import (
	"errors"
	"preq/cmd/utils"
	"preq/mocks"
	"preq/pkg/client"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func Test_declinePR(t *testing.T) {
	t.Run("status is 'Error' on fail", func(t *testing.T) {
		ch := make(chan interface{})
		go declinePR(
			&mocks.Client{
				ErrorValue: errors.New("asdlkfj"),
			},
			&client.Repository{},
			"",
			ch,
		)
		v := (<-ch).(declineResponse)
		assert.Equal(t, "Error", v.Status)
	})

	t.Run("status is 'Done' on success", func(t *testing.T) {
		ch := make(chan interface{})
		go declinePR(
			&mocks.Client{},
			&client.Repository{},
			"",
			ch,
		)
		v := (<-ch).(declineResponse)
		assert.Equal(t, "Done", v.Status)
	})
}

func Test_execute(t *testing.T) {
	type args struct {
		c      client.Client
		args   *cmdArgs
		params *cmdParams
		repo   *client.Repository
	}
	tests := []struct {
		name    string
		args    *args
		wantErr bool
	}{
		{
			"execute fails when client calls fail",
			&args{
				&mocks.Client{
					ErrorValue: errors.New("execute error"),
				},
				&cmdArgs{},
				&cmdParams{},
				&client.Repository{},
			},
			true,
		},
		{
			"execute fails for one pull request when client calls fail",
			&args{
				&mocks.Client{
					ErrorValue: errors.New("execute error"),
				},
				&cmdArgs{
					ID: "id",
				},
				&cmdParams{},
				&client.Repository{},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := execute(tt.args.c, tt.args.args, tt.args.params, tt.args.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Run("execute succeeds when client calls succeed", func(t *testing.T) {
		oldPromptPullRequestMultiSelect := promptPullRequestMultiSelect
		oldProcessPullRequestMap := processPullRequestMap
		processPullRequestMap = func(selectedPRs map[string]*utils.PromptPullRequest, cl client.Client, r *client.Repository, processFn func(cl client.Client, r *client.Repository, id string, c chan interface{}), fn func(interface{}) string) {
			return
		}
		promptPullRequestMultiSelect = func(prList *client.PullRequestList) map[string]*utils.PromptPullRequest {
			return map[string]*utils.PromptPullRequest{}
		}
		err := execute(
			&mocks.Client{},
			&cmdArgs{},
			&cmdParams{},
			&client.Repository{},
		)
		assert.Equal(t, nil, err)

		promptPullRequestMultiSelect = oldPromptPullRequestMultiSelect
		processPullRequestMap = oldProcessPullRequestMap
	})
}

func Test_runCmd(t *testing.T) {
	t.Run("returns error if params don't validate", func(t *testing.T) {
		retErr := errors.New("decline error")
		oldValidateFlagDeclineCmdParams := validateFlagDeclineCmdParams
		validateFlagDeclineCmdParams = func(params *cmdParams) error {
			return retErr
		}
		err := runCmd(
			&cobra.Command{},
			[]string{},
		)

		assert.Equal(t, retErr, err)
		validateFlagDeclineCmdParams = oldValidateFlagDeclineCmdParams
	})
}
