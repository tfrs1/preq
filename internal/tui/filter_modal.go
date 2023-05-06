package tui

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type FilterModalItem struct {
	Line string
	Ref  interface{}
}

type FilterModal struct {
	*tview.Flex
	textArea      *tview.InputField
	filterContent []*ScrollablePageLine
	contentTable  *ScrollablePage
	callback      func(i *FilterModalItem)
}

func (m *FilterModal) Clear() {
	m.textArea.SetText("")
	m.filterContent = []*ScrollablePageLine{}
}

func (m *FilterModal) SetData(data []*FilterModalItem, cb func(item *FilterModalItem)) {
	filterContent := make([]*ScrollablePageLine, 0)
	for _, item := range data {
		filterContent = append(filterContent, &ScrollablePageLine{
			Reference: item,
			Statements: []*ScrollablePageLineStatement{
				{Content: item.Line},
			},
		})
	}

	m.filterContent = filterContent
	m.contentTable.content = filterContent
	m.callback = cb
}

func (m *FilterModal) filter(input string) {
	filtered := make([]*ScrollablePageLine, 0)
	for _, line := range m.filterContent {
		for _, stmnt := range line.Statements {
			lc := strings.ToLowerSpecial(unicode.CaseRanges, stmnt.Content)
			li := strings.ToLowerSpecial(unicode.CaseRanges, input)
			if strings.Contains(lc, li) {
				filtered = append(filtered, line)
			}
		}
	}

	m.contentTable.content = filtered
}

func NewFilterModal() *FilterModal {
	modal := func(p tview.Primitive, width, height int) *tview.Flex {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(
				tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(nil, 0, 1, false).
					AddItem(p, height, 1, true).
					AddItem(nil, 0, 1, false),
				width, 1, true,
			).
			AddItem(nil, 0, 1, false)
	}

	m := &FilterModal{}
	contentTable := NewScrollablePage()
	filterInput := tview.NewInputField()
	filterInput.
		SetBorder(true).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyCtrlJ:
				contentTable.ScrollDown()
				return nil
			case tcell.KeyCtrlK:
				contentTable.ScrollUp()
				return nil
			case tcell.KeyEsc:
				// eventBus.Publish("FilterModal:CloseRequested", nil)
				if m.callback != nil {
					m.callback(nil)
				}
				return nil
			case tcell.KeyEnter:
				if fmi, ok := contentTable.GetSelectedReference().(*FilterModalItem); ok {
					if m.callback != nil {
						m.callback(fmi)
					}
				}
				// if event.Modifiers()&tcell.ModShift == 0 {
				// 	if filterInput.GetText() != "" {
				// 		eventBus.Publish("AddCommentModal:ConfirmRequested", textArea.GetText())
				// 	}
				// 	return nil
				// }
			}

			return event
		})

	contentTable.SetBorder(true)

	s := tview.NewFlex()
	style := tcell.Style{}
	style.Background(s.GetBackgroundColor())
	filterInput.SetFieldStyle(style).SetBorder(true)
	filterInput.SetChangedFunc(m.filter)

	s.SetDirection(tview.FlexRow).
		AddItem(filterInput, 3, 1, true).
		AddItem(contentTable, 0, 1, false)

	s.SetBorder(true).
		SetTitle("Filter").
		SetBorderColor(s.GetBackgroundColor())

	m.Flex = modal(s, 80, 20)
	m.textArea = filterInput
	m.contentTable = contentTable

	return m
}
