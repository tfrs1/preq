package tui

import (
	"errors"
	"fmt"
	"preq/internal/domain"
	"preq/internal/domain/pullrequest"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type pullRequestTableItem struct {
	ID          pullrequest.EntityID
	Title       string
	Source      string
	Destination string
	Selected    bool
}

type pullRequestTable struct {
	View            *tview.Table
	items           []*pullRequestTableItem
	selectedStyle   tcell.Style
	unselectedStyle tcell.Style
	observers       map[EventType][]Observer
	pages           *tview.Pages
}

func newPullRequestTable() *pullRequestTable {
	table := tview.NewTable()

	// // Set box options
	// table.
	// 	SetTitle("preq").
	// 	SetBorder(true)

	// TODO: Make style configurable
	// NewRGBColor NewHexColor
	var styleInstance tcell.Style
	unselectedStyle := styleInstance.
		Background(tcell.ColorDefault).
		Foreground(tcell.ColorWhite)

	selectedStyle := styleInstance.
		Italic(true).
		Bold(true).
		Background(tcell.ColorNames["maroon"]).
		Foreground(tcell.ColorNames["white"])

	items := []*pullRequestTableItem{}
	instance := &pullRequestTable{
		table,
		items,
		selectedStyle,
		unselectedStyle,
		make(map[EventType][]Observer),
		&tview.Pages{},
	}
	// Set table options
	table.
		SetBorders(false).
		Select(0, 0).
		SetFixed(1, 1).
		SetSelectable(true, false).
		// SetSelectedStyle(tableSelectedStyle).
		SetDoneFunc(func(key tcell.Key) {
			// if key == tcell.KeyEscape {
			// 	// table.SetSelectable(true, false)
			// }
			if key == tcell.KeyEnter {
				// table.SetSelectable(true, false)
			}
		}).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case ' ':
				row, err := instance.HighlightedRow()
				if err != nil {
					return nil
				}

				if !instance.RowSelected(row) {
					instance.selectRow(row)
				} else {
					instance.unselectRow(row)
				}

				return nil
			}

			switch event.Key() {
			// TODO: Make keybindings configurable
			case tcell.KeyCtrlD:
				items := instance.SelectedItems()
				if len(items) == 0 {
					// items = append(items, instance.HighlightedItem())
					return nil
				}

				// TODO: Show confirmation modal, continue only if confirmed
				modal := tview.NewModal().
					SetText(fmt.Sprintf("Are you sure you want close (%d) pull requests?", len(items))).
					AddButtons([]string{"Yes", "No"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						if buttonIndex == 1 {
							// HACK: Event handle sends focus back to the table
							// so the hack is to remove zero items
							items = []*pullRequestTableItem{}
						}

						instance.pages.RemovePage("ConfirmDelete")
						instance.Notify(EVENT_CLOSE_PULL_REQUEST, items)
					})
				instance.pages.AddPage("ConfirmDelete", modal, true, true)
				return nil
			}

			return event
		})

	return instance
}

func (prt *pullRequestTable) RowSelected(row int) bool {
	return prt.items[row].Selected
}

func (prt *pullRequestTable) SelectedItems() []*pullRequestTableItem {
	result := []*pullRequestTableItem{}
	for _, v := range prt.items {
		if v.Selected {
			result = append(result, v)
		}
	}
	return result
}

func (prt *pullRequestTable) HighlightedRow() (int, error) {
	row, _ := prt.View.GetSelection()
	// TODO: Check for header row
	if row == 0 {
		return -1, errors.New("no row selected")
	}
	return row - 1, nil
}

func (prt *pullRequestTable) selectRow(row int) {
	// TODO: Rename
	inTableRow := prt.inTableRow(row)
	for i := 0; i < prt.View.GetColumnCount(); i++ {
		prt.View.GetCell(inTableRow, i).SetStyle(prt.selectedStyle)
	}

	prt.items[row].Selected = true
}

func (prt *pullRequestTable) inTableRow(row int) int {
	return row + 1
}

func (prt *pullRequestTable) unselectRow(row int) {
	// TODO: Rename
	inTableRow := prt.inTableRow(row)
	for i := 0; i < prt.View.GetColumnCount(); i++ {
		prt.View.GetCell(inTableRow, i).SetStyle(prt.unselectedStyle)
	}

	prt.items[row].Selected = false
}

func (prt *pullRequestTable) filter(input string) {

}

func (prt *pullRequestTable) resetFilter() {

}

func (prt *pullRequestTable) Notify(et EventType, data interface{}) {
	for _, f := range prt.observers[et] {
		f(et, data)
	}
}

func (prt *pullRequestTable) Subscribe(et EventType, o Observer) {
	prt.observers[et] = append(prt.observers[et], o)
}

func (prt *pullRequestTable) Clear() {
	prt.View.Clear()
	prt.View.SetCell(0, 0, tview.NewTableCell("#"))
	prt.View.SetCell(0, 1, tview.NewTableCell("Title"))
	prt.View.SetCell(0, 2, tview.NewTableCell("Source -> Destination"))
}

func (prt *pullRequestTable) AddRow(prti *pullRequestTableItem) {
	prt.items = append(prt.items, prti)
	i := len(prt.items)
	prt.View.SetCell(i, 0, tview.NewTableCell(fmt.Sprint(prti.ID)))
	prt.View.SetCell(i, 1, tview.NewTableCell(prti.Title))
	prt.View.SetCell(i, 2, tview.NewTableCell(fmt.Sprintf("%s -> %s", prti.Source, prti.Destination)))
}

func (prt *pullRequestTable) FindIndex(prti *pullRequestTableItem) (error, int) {
	for i, v := range prt.items {
		if v.ID == prti.ID {
			return nil, i
		}
	}

	return errors.New("item not found"), -1
}

func (prt *pullRequestTable) findTableRow(prti *pullRequestTableItem) (error, int) {
	for i, v := range prt.items {
		if v.ID == prti.ID {
			return nil, i + 1
		}
	}

	return errors.New("item not found"), -1
}

func (prt *pullRequestTable) RemoveItem(prti *pullRequestTableItem) {
	err, itemIndex := prt.FindIndex(prti)
	if err != nil {
		return
	}

	copy(prt.items[itemIndex:], prt.items[itemIndex+1:])
	prt.items = prt.items[:len(prt.items)-1]
	prt.View.RemoveRow(prt.inTableRow(itemIndex))
}

func (prt *pullRequestTable) SetItemState(prti *pullRequestTableItem, state string) {
	err, i := prt.findTableRow(prti)
	if err != nil {
		return
	}

	prt.View.SetCellSimple(i, 1, "Closing...")
	prt.View.SetCell(i, 2, nil)
}

type grid struct {
	grid   *tview.Grid
	table  *tview.Table
	search *tview.InputField
}

func (g *grid) showFilter() {
	g.grid.Clear().
		AddItem(g.table, 0, 0, 1, 1, 0, 0, false).
		AddItem(g.search, 1, 0, 1, 1, 0, 0, false)
}

func (g *grid) hideFilter() {
	g.grid.Clear().
		AddItem(g.table, 0, 0, 2, 1, 0, 0, false)
}

func loadPRs(app *tview.Application, prList pullrequest.EntityPageList, table *pullRequestTable) {
	app.QueueUpdateDraw(func() {
		table.View.SetCell(0, 0, tview.NewTableCell("Loading...").SetAlign(tview.AlignCenter))
	})

	app.QueueUpdateDraw(func() {
		table.Clear()
		values, err := prList.GetPage(1)
		if err != nil {
			return
		}

		for _, v := range values {
			table.AddRow(&pullRequestTableItem{
				ID:          v.ID,
				Title:       v.Title,
				Source:      v.Source,
				Destination: v.Destination,
			})
		}
	})

	// app.QueueUpdateDraw(func() {
	// 	table.View.SetCell(len(prList.Values), 0, tview.NewTableCell("Loading..."))
	// })
}

type TuiPresenter struct {
	client pullrequest.Repository
}

func (tp *TuiPresenter) Start() {
	run(tp.client)
}

func (tp *TuiPresenter) Notify(e *domain.Event) {}

func NewTui(c []pullrequest.Repository) *TuiPresenter {
	return &TuiPresenter{
		client: c[0],
	}
}

type app struct {
	modal       *tview.Modal
	tui         *tview.Application
	table       *pullRequestTable
	searchInput *tview.InputField
	grid        *grid
}

func (app *app) Update(prList pullrequest.EntityPageList) {
	loadPRs(app.tui, prList, app.table)
	// fmt.Println(prList)
}

func (app *app) UpdateFailed(error) {}

type Observer func(EventType, interface{})

type Subject interface {
	Subscribe(EventType, Observer)
	Unsubscribe()
}

type EventType string

const (
	EVENT_CLOSE_PULL_REQUEST = "CLOSE_PULL_REQUEST"
)

func newApp(c pullrequest.Repository) *app {
	tui := tview.NewApplication()
	table := newPullRequestTable()
	searchInput := tview.NewInputField()
	grid := &grid{tview.NewGrid(), table.View, searchInput}

	table.Subscribe(EVENT_CLOSE_PULL_REQUEST, func(et EventType, i interface{}) {
		switch et {
		case EVENT_CLOSE_PULL_REQUEST:
			// TODO: Move to a more appropriate place
			tui.SetFocus(table.View)

			items, ok := i.([]*pullRequestTableItem)
			if !ok {
				// TODO: Notify unexpected data
				return
			}

			service := pullrequest.NewCloseService(c)
			for _, v := range items {
				go func(v *pullRequestTableItem) {
					tui.QueueUpdateDraw(func() {
						table.SetItemState(v, "CLOSING")
					})

					service.Close(&pullrequest.CloseOptions{ID: v.ID})

					tui.QueueUpdateDraw(func() {
						table.RemoveItem(v)
					})
				}(v)
			}
		}
	})

	// newPrimitive := func(text string) tview.Primitive {
	// 	return tview.NewTextView().
	// 		SetTextAlign(tview.AlignCenter).
	// 		SetText(text)
	// }
	// menu := newPrimitive("Menu")
	// main := newPrimitive("Main content")
	// sideBar := newPrimitive("Side Bar")

	table.View.
		SetBorder(true).
		SetTitle("Pull requests")

	searchInput.SetPlaceholder("Filter pull requests")
	searchInput.
		SetBorder(true).
		SetTitle("Filter")

	searchInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			grid.hideFilter()
			searchInput.SetText("")
			tui.SetFocus(table.View)
			// grid.RemoveItem(searchInput)
		}

		return event
	})

	tui.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			// if app.GetFocus() != searchInput || searchInput.GetText() == "" {
			// 	app.Stop()
			// }
		}

		switch event.Rune() {
		case 'q':
			if tui.GetFocus() != searchInput {
				tui.Stop()
			}
		case '/':
			if tui.GetFocus() != searchInput {
				grid.showFilter()
				// grid.AddItem(searchInput, 1, 0, 1, 1, 0, 0, false)
				tui.SetFocus(searchInput)
				return nil
			}
		}

		return event
	})

	grid.grid.
		// SetRows(3, 0, 3).
		// SetColumns(30, 0, 30).
		SetRows(0, 3).
		SetBorders(true).
		AddItem(table.View, 0, 0, 2, 1, 0, 0, false)
	// AddItem(newPrimitive("Header"), 0, 0, 1, 3, 0, 0, false).
	// 	AddItem(newPrimitive("Footer"), 2, 0, 1, 3, 0, 0, false)

	grid.grid.
		// 	SetTitle("preq").
		SetBorders(false).
		SetBorder(false)

	// // Layout for screens narrower than 100 cells (menu and side bar are hidden).
	// grid.AddItem(menu, 0, 0, 0, 0, 0, 0, false).
	// 	AddItem(table, 1, 0, 1, 3, 0, 0, false).
	// 	AddItem(sideBar, 0, 0, 0, 0, 0, 0, false)

	// // Layout for screens wider than 100 cells.
	// grid.AddItem(menu, 1, 0, 1, 1, 0, 100, false).
	// 	AddItem(table, 1, 1, 1, 1, 0, 100, false).
	// 	AddItem(sideBar, 1, 2, 1, 1, 0, 100, false)

	// if err := app.SetRoot(table, true).EnableMouse(true).Run(); err != nil {
	// 	panic(err)
	// }

	pages := tview.NewPages().
		AddPage("main", grid.grid, true, true)

	table.pages = pages

	// TODO: Make enable mouse configurable
	tui = tui.SetRoot(pages, true).EnableMouse(true)

	return &app{
		// modal:       modal,
		tui:         tui,
		table:       table,
		grid:        grid,
		searchInput: searchInput,
	}
}

func (app *app) Run() error {
	app.tui.SetFocus(app.table.View)
	return app.tui.Run()
}

func run(c pullrequest.Repository) {
	app := newApp(c)
	go domain.LoadPullRequests(c, app)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
