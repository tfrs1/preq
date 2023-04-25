package tui

import (
	"preq/internal/cli/utils"
	"preq/internal/pkg/client"

	"github.com/rivo/tview"
)

var mergeConfirmationModal = tview.NewModal().
	SetText("Are you sure you want to merge %d pull requests?").
	AddButtons([]string{"Merge", "Cancel"}).
	SetDoneFunc(mergeConfirmationCallback())

func mergeConfirmationCallback() func(int, string) {
	return func(buttonIndex int, buttonLabel string) {
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
				row.PullRequest.State = client.PullRequestState_MERGING
				row.Selected = false
			}

			redraw()

			go processPullRequestMap(
				selectedPRs,
				mergePR,
				func(msg *utils.ProcessPullRequestResponse) {
					if msg.Status == "Done" {
						v := table.GetRowByGlobalID(msg.GlobalID)
						// TODO: return an error instead?
						if v != nil {
							v.PullRequest.State = client.PullRequestState_MERGED
						}
					}

					app.QueueUpdateDraw(redraw)
				},
			)
		}

		eventBus.Publish("mergeModal:closed", nil)
	}
}

func mergePR(
	cl client.Client,
	r *client.Repository,
	id string,
	globalId string,
	ch chan *utils.ProcessPullRequestResponse,
) {
	_, err := cl.Merge(&client.MergeOptions{
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
