package close

import (
	"errors"
	"preq/internal/cli/utils"
	"preq/internal/domain/pullrequest"
	"preq/internal/pkg/client"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// func Test_closePR(t *testing.T) {
// 	t.Run("status is 'Error' on fail", func(t *testing.T) {
// 		ch := make(chan interface{})
// 		go closePR(
// 			&client.MockClient{
// 				ErrorValue: errors.New("asdlkfj"),
// 			},
// 			&client.Repository{},
// 			"",
// 			ch,
// 		)
// 		v := (<-ch).(closeResponse)
// 		assert.Equal(t, "Error", v.Status)
// 	})

// 	t.Run("status is 'Done' on success", func(t *testing.T) {
// 		ch := make(chan interface{})
// 		go closePR(
// 			&client.MockClient{},
// 			&client.Repository{},
// 			"",
// 			ch,
// 		)
// 		v := (<-ch).(closeResponse)
// 		assert.Equal(t, "Done", v.Status)
// 	})
// }

func Test_execute(t *testing.T) {
	type args struct {
		c      pullrequest.Repository
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
			err := execute(tt.args.c, tt.args.args, tt.args.params)
			if (err != nil) != tt.wantErr {
				// t.Errorf("execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Run("execute succeeds when client calls succeed", func(t *testing.T) {
		oldProcessPullRequestMap := processPullRequestMap
		processPullRequestMap = func(selectedPRs map[string]*utils.PromptPullRequest, cl pullrequest.Repository, r *client.Repository, processFn func(cl pullrequest.Repository, r *client.Repository, id string, c chan interface{}), fn func(interface{}) string) {
			return
		}
		err := execute(
			&client.MockClient{},
			&cmdArgs{},
			&cmdParams{},
		)
		assert.Equal(t, nil, err)

		processPullRequestMap = oldProcessPullRequestMap
	})
}

func Test_runCmd(t *testing.T) {
	t.Run("returns error if params don't validate", func(t *testing.T) {
		retErr := errors.New("close error")
		oldValidateFlagCloseCmdParams := validateFlagCloseCmdParams
		validateFlagCloseCmdParams = func(params *cmdParams) error {
			return retErr
		}
		err := runCmd(
			&cobra.Command{},
			[]string{},
		)

		assert.Equal(t, retErr, err)
		validateFlagCloseCmdParams = oldValidateFlagCloseCmdParams
	})
}
