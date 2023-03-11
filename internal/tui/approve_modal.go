package tui

import (
	"preq/internal/cli/utils"

	"github.com/rivo/tview"
)

var (
	approveConfirmationModal = tview.NewModal().
		SetText("Are you sure you want to approve %d pull requests?").
		AddButtons([]string{"Approve", "Cancel"}).
		SetDoneFunc(approveConfirmationCallback)
)

func approveConfirmationCallback(buttonIndex int, buttonLabel string) {
	if buttonIndex == 0 {
		selectedPRs := make(map[string]*promptPullRequest)

		for _, row := range table.GetSelectedRows() {
			// approved := false
			// for i, pra := range row.PullRequest.Approvals {
			//   if pra.User == row.Client.User {
			//     approved = true
			//     break
			//   }
			// }

			// if (approved) {
			//   continue
			// }

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
			approvePR,
			func(msg utils.ProcessPullRequestResponse) string {
				v := table.GetRowByGlobalID(msg.GlobalID)
				if msg.Error != nil {
					v.IsApprovalsLoading = false
				}

				if msg.Status == "Done" {
					// TODO: return an error instead?
					if v != nil {
						go func(v *PullRequest) {
							err := v.Client.FillMiscInfoAsync(
								v.Repository,
								v.PullRequest,
							)

							if err != nil {
								// TODO: Handle error
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

	eventBus.Publish("approveModal:closed", nil)
}
