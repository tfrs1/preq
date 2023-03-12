package tui

import (
	"preq/internal/cli/utils"
	"preq/internal/pkg/client"

	"github.com/rivo/tview"
)

var (
	unapproveConfirmationModal = tview.NewModal().
		SetText("Are you sure you want to unapprove %d pull requests?").
		AddButtons([]string{"Unapprove", "Cancel"}).
		SetDoneFunc(unapproveConfirmationCallback)
)

func unapproveConfirmationCallback(buttonIndex int, buttonLabel string) {
	if buttonIndex == 0 {
		selectedPRs := make(map[string]*promptPullRequest)

		for _, row := range table.GetSelectedRows() {
			selectedPRs[row.PullRequest.URL] = &promptPullRequest{
				ID:         row.PullRequest.ID,
				GlobalID:   row.PullRequest.URL,
				Title:      row.PullRequest.Title,
				Client:     row.Client,
				Repository: row.Repository,
			}

			// TODO: This should probably be a method in table instead
			row.Selected = false
			row.IsApprovalsLoading = true
		}

		redraw()

		go processPullRequestMap(
			selectedPRs,
			unapprovePR,
			func(msg *utils.ProcessPullRequestResponse) {
				v := table.GetRowByGlobalID(msg.GlobalID)

				if msg.Error != nil {
					// app.QueueUpdateDraw(table.redraw)
					return
				}

				if msg.Status == "Done" && v != nil {
					go func(v *PullRequest) {
						err := v.Client.FillMiscInfoAsync(v.Repository, v.PullRequest)
						if err != nil {
							return
						}

						v.IsApprovalsLoading = false
						app.QueueUpdateDraw(redraw)
					}(v)
				}
			},
		)
	}

	eventBus.Publish("unapproveModal:closed", nil)
}

func unapprovePR(
	cl client.Client,
	r *client.Repository,
	id string,
	globalId string,
	ch chan *utils.ProcessPullRequestResponse,
) {
	_, err := cl.Unapprove(&client.UnapproveOptions{
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
