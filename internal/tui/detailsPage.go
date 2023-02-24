package tui

import (
	"preq/internal/pkg/client"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type detailsPage struct {
	View tview.Primitive
}

func newDetailsPage() *detailsPage {
	box := tview.NewBox().SetBorder(true).SetTitle("Details")
	box.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			eventBus.Publish("detailsPage:close", nil)
		}

		switch event.Rune() {
		case 'o':
			eventBus.Publish("detailsPage:close", nil)
		}

		return event
	})

	eventBus.Subscribe("detailsPage:open", func(_ interface{}) {
		// TODO: clear page and start loading the PR info
		id := "1"
		prClient.GetPullRequestInfo(&client.ApproveOptions{
			Repository: prRepo,
			ID:         id,
		})
	})

	return &detailsPage{
		View: box,
	}
}
