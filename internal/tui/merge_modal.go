package tui

import (
	"preq/internal/cli/utils"
	"preq/internal/pkg/client"

	"github.com/rivo/tview"
)

var (
	mergeConfirmationModal = tview.NewModal().
		SetText("Are you sure you want to merge %d pull requests?").
		AddButtons([]string{"Merge", "Cancel"}).
		SetDoneFunc(mergeConfirmationCallback())
)

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

			table.redraw()

			go processPullRequestMap(
				selectedPRs,
				mergePR,
				func(msg utils.ProcessPullRequestResponse) string {
					if msg.Status == "Done" {
						v := table.GetRowByGlobalID(msg.GlobalID)
						// TODO: return an error instead?
						if v != nil {
							v.PullRequest.State = client.PullRequestState_MERGED
						}
					}

					app.QueueUpdateDraw(table.redraw)

					return ""
				},
			)
		}

		eventBus.Publish("mergeModal:closed", nil)
	}
}
