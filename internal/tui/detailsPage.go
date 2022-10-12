package tui

import (
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
			eventBus.Publish("detailsPage:close")
		}

		switch event.Rune() {
		case 'o':
			eventBus.Publish("detailsPage:close")
		}

		return event
	})

	eventBus.Subscribe("detailsPage:open", func() {
		// TODO: clear page and start loading the PR info
	})

	return &detailsPage{
		View: box,
	}
}
