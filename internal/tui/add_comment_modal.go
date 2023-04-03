package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type AddCommentModal struct {
	*tview.Flex
	textArea *tview.TextArea
}

func (m *AddCommentModal) Clear() {
	m.textArea.SetText("", false)
}

func NewAddCommentModal() *AddCommentModal {
	modal := func(p tview.Primitive, width, height int) *tview.Flex {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(
				tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(nil, 0, 1, false).
					AddItem(p, 0, 1, true).
					AddItem(nil, 0, 1, false),
				width, 1, true,
			).
			AddItem(nil, 0, 1, false)
	}

	textArea := tview.NewTextArea().SetPlaceholder("Add text here...")
	cancelButton := tview.NewButton("Cancel")
	confirmButton := tview.NewButton("Send")
	s := tview.NewFlex()

	textArea.
		SetBorder(true).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEsc:
				app.SetFocus(cancelButton)
				return nil
			case tcell.KeyEnter:
				if event.Modifiers()&tcell.ModShift == 0 {
					if textArea.GetText() != "" {
						eventBus.Publish("AddCommentModal:ConfirmRequested", textArea.GetText())
					}
					return nil
				}
			}

			return event
		})

	cancelButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(confirmButton)
			return nil
		case tcell.KeyEsc:
			eventBus.Publish("AddCommentModal:CancelRequested", nil)
			return nil
		}
		return event
	})

	confirmButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(textArea)
			return nil
		case tcell.KeyEsc:
			eventBus.Publish("AddCommentModal:CancelRequested", nil)
			return nil
		case tcell.KeyEnter:
			if textArea.GetText() != "" {
				eventBus.Publish("AddCommentModal:ConfirmRequested", textArea.GetText())
			}
			return nil
		}

		return event
	})

	s.SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("Content"), 1, 0, false).
		AddItem(textArea, 0, 1, true).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(tview.NewBox(), 0, 1, false).
				AddItem(cancelButton, len(cancelButton.GetLabel())+2, 0, false).
				AddItem(tview.NewBox(), 1, 0, false).
				AddItem(confirmButton, len(confirmButton.GetLabel())+2, 0, false),
			1, 0, false,
		)

	s.SetTitle("Add a comment").
		SetBorder(true).
		SetBorderColor(s.GetBackgroundColor())

	return &AddCommentModal{
		Flex:     modal(s, 80, 20),
		textArea: textArea,
	}
}
