package tui

import (
	"fmt"
	"os"
	"preq/internal/cli/paramutils"
	"preq/internal/clientutils"
	"preq/internal/pkg/client"

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

func loadConfig() (client.Client, *client.Repository, error) {
	params := &paramutils.RepositoryParams{}
	paramutils.FillDefaultRepositoryParams(params)

	c, err := clientutils.ClientFactory{}.DefaultClient(params.Provider)
	if err != nil {
		return nil, nil, err
	}

	r, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
		Provider:           params.Provider,
		FullRepositoryName: params.Name,
	})
	if err != nil {
		return nil, nil, err
	}

	return c, r, nil
}

func loadPRs(app *tview.Application, c client.Client, repo *client.Repository, table *tview.Table) {
	app.QueueUpdateDraw(func() {
		table.SetCell(0, 0, tview.NewTableCell("Loading...").SetAlign(tview.AlignCenter))
	})

	nextURL := ""
	for {
		prs, err := c.GetPullRequests(&client.GetPullRequestsOptions{
			Repository: repo,
			State:      client.PullRequestState_OPEN,
			Next:       nextURL,
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(134)
		}

		nextURL = prs.NextURL

		app.QueueUpdateDraw(func() {
			table.Clear()
			table.SetCell(0, 0, tview.NewTableCell("#"))
			table.SetCell(0, 1, tview.NewTableCell("Title"))
			table.SetCell(0, 2, tview.NewTableCell("Source -> Destination"))
			for i, v := range prs.Values {
				table.SetCell(i+1, 0, tview.NewTableCell(v.ID))
				table.SetCell(i+1, 1, tview.NewTableCell(v.Title))
				table.SetCell(i+1, 2, tview.NewTableCell(fmt.Sprintf("%s -> %s", v.Source, v.Destination)))
			}
		})

		if nextURL == "" {
			break
		}

		app.QueueUpdateDraw(func() {
			table.SetCell(len(prs.Values), 0, tview.NewTableCell("Loading..."))
		})
	}
}

func Run() {
	c, repo, err := loadConfig()
	if err != nil {
		os.Exit(123)
	}

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

	app.SetFocus(table.View)
	go loadPRs(app, c, repo, table.View)

	if err := app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
