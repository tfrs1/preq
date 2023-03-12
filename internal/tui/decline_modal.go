package tui

import (
	"preq/internal/cli/utils"
	"preq/internal/pkg/client"

	"github.com/rivo/tview"
)

var (
	declineConfirmationModal = tview.NewModal().
		SetText("Are you sure you want to decline %d pull requests?").
		AddButtons([]string{"Decline", "Cancel"}).
		SetDoneFunc(declineConfirmationCallback)
)

func declineConfirmationCallback(buttonIndex int, buttonLabel string) {
	if buttonIndex == 0 {
		selectedPRs := make(map[string]*promptPullRequest)

		for _, row := range table.GetSelectedRows() {
			selectedPRs[row.PullRequest.URL] = &promptPullRequest{
				ID:         row.PullRequest.ID,
				GlobalID:   row.PullRequest.URL,
				Title:      row.PullRequest.Title,
				Repository: row.Repository,
				Client:     row.Client,
			}

			row.PullRequest.State = client.PullRequestState_DECLINING
			row.Selected = false
		}

		redraw()

		go processPullRequestMap(
			selectedPRs,
			declinePR,
			func(msg *utils.ProcessPullRequestResponse) {
				if msg.Status == "Done" {
					v := table.GetRowByGlobalID(msg.GlobalID)
					v.PullRequest.State = client.PullRequestState_DECLINED
				}

				app.QueueUpdateDraw(redraw)
			},
		)
	}

	eventBus.Publish("mergeModal:closed", nil)
}

func declinePR(
	cl client.Client,
	r *client.Repository,
	id string,
	globalId string,
	ch chan *utils.ProcessPullRequestResponse,
) {
	_, err := cl.DeclinePullRequest(&client.DeclinePullRequestOptions{
		Repository: r,
		ID:         id,
	})

	res := &utils.ProcessPullRequestResponse{
		ID:       id,
		GlobalID: globalId,
		Status:   "Done",
	}
	if err != nil {
		res.Status = "Error"
		res.Error = err
	}

	ch <- res
}
