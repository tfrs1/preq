package tui

import (
	"preq/internal/cli/utils"

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

		table.redraw()

		go processPullRequestMap(
			selectedPRs,
			unapprovePR,
			func(msg utils.ProcessPullRequestResponse) string {
				v := table.GetRowByGlobalID(msg.GlobalID)
				v.IsApprovalsLoading = false

				if msg.Status == "Done" {
					// TODO: return an error instead?
					if v != nil {
						go func(v *PullRequest) {
							err := v.Client.FillMiscInfoAsync(
								v.Repository,
								v.PullRequest,
							)

							if err != nil {
								return
							}

							v.IsApprovalsLoading = false

							app.QueueUpdateDraw(table.redraw)
						}(v)
					}
				}

				app.QueueUpdateDraw(table.redraw)

				return ""
			},
		)
	}

	eventBus.Publish("unapproveModal:closed", nil)
}
