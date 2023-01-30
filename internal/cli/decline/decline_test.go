package decline

import (
	"errors"
	"preq/internal/cli/utils"
	"preq/internal/pkg/client"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_declinePR(t *testing.T) {
	t.Run("status is 'Error' on fail", func(t *testing.T) {
		ch := make(chan interface{})
		go declinePR(
			&client.MockClient{
				ErrorValue: errors.New("asdlkfj"),
			},
			&client.Repository{},
			"",
			ch,
		)
		v := (<-ch).(ProcessPullRequestResponse)
		assert.Equal(t, "Error", v.Status)
	})

	t.Run("status is 'Done' on success", func(t *testing.T) {
		ch := make(chan interface{})
		go declinePR(
			&client.MockClient{},
			&client.Repository{},
			"",
			ch,
		)
		v := (<-ch).(ProcessPullRequestResponse)
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
				&client.MockClient{
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
				&client.MockClient{
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
			err := execute(
				tt.args.c,
				tt.args.args,
				tt.args.params,
				tt.args.repo,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Run("execute succeeds when client calls succeed", func(t *testing.T) {
		oldPromptPullRequestMultiSelect := promptPullRequestMultiSelect
		oldProcessPullRequestMap := processPullRequestMap
		processPullRequestMap = func(
			selectedPRs map[string]*utils.PromptPullRequest,
			cl client.Client,
			r *client.Repository,
			processFn func(cl client.Client, r *client.Repository, id string, c chan interface{}),
			fn func(interface{}) string,
		) {
		}

		promptPullRequestMultiSelect = func(prList *client.PullRequestList) map[string]*utils.PromptPullRequest {
			return map[string]*utils.PromptPullRequest{}
		}
		err := execute(
			&client.MockClient{},
			&cmdArgs{},
			&cmdParams{},
			&client.Repository{},
		)
		assert.Equal(t, nil, err)

		promptPullRequestMultiSelect = oldPromptPullRequestMultiSelect
		processPullRequestMap = oldProcessPullRequestMap
	})
}
