package tui

import (
	"fmt"
	"os"
	"preq/internal/domain"
	"preq/internal/pkg/client"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type pullRequestTableItem struct {
	ID          string
	Title       string
	Source      string
	Destination string
}

type pullRequestTable struct {
	View            *tview.Table
	items           []*pullRequestTableItem
	selectedStyle   tcell.Style
	unselectedStyle tcell.Style
}

func newPullRequestTable() *pullRequestTable {
	table := tview.NewTable()

	// // Set box options
	// table.
	// 	SetTitle("preq").
	// 	SetBorder(true)

	var styleInstance tcell.Style
	unselectedStyle := styleInstance.
		Background(tcell.ColorDefault).
		Foreground(tcell.ColorWhite)

	selectedStyle := styleInstance.
		Italic(true).
		Bold(true).
		Background(tcell.ColorDarkRed)

	selected := false

	items := []*pullRequestTableItem{}
	instance := &pullRequestTable{table, items, selectedStyle, unselectedStyle}
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
				row, _ := table.GetSelection()
				// Disable selecting the header row
				if row == 0 {
					return nil
				}

				if selected {
					selected = false
					instance.selectRow(row)
				} else {
					selected = true
					instance.unselectRow(row)
				}

				return nil
			}

			return event
		})

	return instance
}

func (prt *pullRequestTable) selectRow(row int) {
	for i := 0; i < prt.View.GetColumnCount(); i++ {
		prt.View.GetCell(row, i).SetStyle(prt.selectedStyle)
	}
}

func (prt *pullRequestTable) unselectRow(row int) {
	for i := 0; i < prt.View.GetColumnCount(); i++ {
		prt.View.GetCell(row, i).SetStyle(prt.unselectedStyle)
	}
}

func (prt *pullRequestTable) filter(input string) {

}

func (prt *pullRequestTable) resetFilter() {

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
	prt.View.SetCell(i, 0, tview.NewTableCell(prti.ID))
	prt.View.SetCell(i, 1, tview.NewTableCell(prti.Title))
	prt.View.SetCell(i, 2, tview.NewTableCell(fmt.Sprintf("%s -> %s", prti.Source, prti.Destination)))
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
func loadPRs(app *tview.Application, c domain.Client, repo *client.Repository, table *pullRequestTable) {
	app.QueueUpdateDraw(func() {
		table.View.SetCell(0, 0, tview.NewTableCell("Loading...").SetAlign(tview.AlignCenter))
	})

	nextURL := ""
	for {
		prs, err := c.GetPullRequests(&domain.GetPullRequestOptions{
			// Repository: repo,
			State: client.PullRequestState_OPEN,
			Next:  nextURL,
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(134)
		}

		nextURL = prs.NextURL

		app.QueueUpdateDraw(func() {
			table.Clear()
			for _, v := range prs.Values {
				table.AddRow(&pullRequestTableItem{
					ID:          v.ID,
					Title:       v.Title,
					Source:      v.Source,
					Destination: v.Destination,
				})
			}
		})

		if nextURL == "" {
			break
		}

		app.QueueUpdateDraw(func() {
			table.View.SetCell(len(prs.Values), 0, tview.NewTableCell("Loading..."))
		})
	}
}

type TuiPresenter struct {
	client domain.Client
}

func (tp *TuiPresenter) Start() {
	run(tp.client)
}

func (tp *TuiPresenter) Notify(e *domain.Event) {}

func NewTui(c domain.Client) *TuiPresenter {
	return &TuiPresenter{
		client: c,
	}
}

type app struct {
	tui         *tview.Application
	table       *pullRequestTable
	searchInput *tview.InputField
	grid        *grid
}

func (app *app) Update() {}

func newApp() *app {
	tui := tview.NewApplication()
	table := newPullRequestTable()
	searchInput := tview.NewInputField()
	grid := &grid{tview.NewGrid(), table.View, searchInput}

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

	tui.SetFocus(table.View)
	tui = tui.SetRoot(grid.grid, true).EnableMouse(true)

	return &app{
		tui:         tui,
		table:       table,
		grid:        grid,
		searchInput: searchInput,
	}
}

func run(c domain.Client) {
	app := newApp()
	go domain.LoadPullRequests(c, app)

	if err := app.tui.Run(); err != nil {
		panic(err)
	}
}
