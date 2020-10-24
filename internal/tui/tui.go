package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type pullRequestTable struct {
	View *tview.Table
}

func newPullRequestTable() *pullRequestTable {
	table := tview.NewTable()

	// // Set box options
	// table.
	// 	SetTitle("preq").
	// 	SetBorder(true)

	// Set table options
	table.
		SetBorders(false).
		Select(0, 0).
		SetFixed(1, 1).
		SetSelectable(true, false).
		SetDoneFunc(func(key tcell.Key) {
			// if key == tcell.KeyEscape {
			// 	app.Stop()
			// }
			// if key == tcell.KeyEnter {
			// 	table.SetSelectable(true, false)
			// }
		}).
		SetSelectedFunc(func(row int, column int) {
			table.GetCell(row, column).SetTextColor(tcell.ColorRed)
			table.SetSelectable(false, false)
		})

	return &pullRequestTable{table}
}

func (prt *pullRequestTable) filter(input string) {

}

func (prt *pullRequestTable) resetFilter() {

}

func (prt *pullRequestTable) load() {
	lorem := strings.Split("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.", " ")
	cols, rows := 10, 40
	word := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			// color := tcell.ColorWhite
			// if c < 1 || r < 1 {
			// 	color = tcell.ColorYellow
			// }
			prt.View.SetCell(r, c,
				tview.NewTableCell(lorem[word]).
					// SetTextColor(color).
					SetAlign(tview.AlignCenter))
			word = (word + 1) % len(lorem)
		}
	}
}

func Run() {
	app := tview.NewApplication()

	// newPrimitive := func(text string) tview.Primitive {
	// 	return tview.NewTextView().
	// 		SetTextAlign(tview.AlignCenter).
	// 		SetText(text)
	// }
	// menu := newPrimitive("Menu")
	// main := newPrimitive("Main content")
	// sideBar := newPrimitive("Side Bar")

	table := newPullRequestTable()
	table.View.
		SetBorder(true).
		SetTitle("Pull requests")

	searchInput := tview.NewInputField().
		SetPlaceholder("Filter pull requests")
	searchInput.
		SetBorder(true).
		SetTitle("Filter")

	searchInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			searchInput.SetText("")
		}

		return event
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			if app.GetFocus() != searchInput || searchInput.GetText() == "" {
				app.Stop()
			}
		}

		switch event.Rune() {
		case 'q':
			if app.GetFocus() != searchInput {
				app.Stop()
			}
		case '/':
			if app.GetFocus() != searchInput {
				app.SetFocus(searchInput)
				return nil
			}
		}

		return event
	})

	grid := tview.NewGrid().
		// SetRows(3, 0, 3).
		// SetColumns(30, 0, 30).
		SetRows(0, 3).
		SetBorders(true).
		AddItem(table.View, 0, 0, 1, 1, 0, 0, false).
		AddItem(searchInput, 1, 0, 1, 1, 0, 0, false)
	// AddItem(newPrimitive("Header"), 0, 0, 1, 3, 0, 0, false).
	// 	AddItem(newPrimitive("Footer"), 2, 0, 1, 3, 0, 0, false)

	grid.
		// 	SetTitle("preq").
		SetBorders(false).
		SetBorder(false)

	table.load()

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

	if err := app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		app.SetFocus(table.View)
		panic(err)
	}
}
