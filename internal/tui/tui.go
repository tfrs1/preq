package tui

import (
	"os"
	"preq/internal/cli/paramutils"
	"preq/internal/clientutils"
	"preq/internal/pkg/client"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func loadConfig(params *paramutils.RepositoryParams) (client.Client, *client.Repository, error) {
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

func loadPRs(app *tview.Application, c client.Client, repo *client.Repository, table *pullRequestTable) {
	app.QueueUpdateDraw(func() {
		table.View.SetCell(0, 0, tview.NewTableCell("Loading...").SetAlign(tview.AlignCenter))
	})

	nextURL := ""
	for {
		prs, err := c.GetPullRequests(&client.GetPullRequestsOptions{
			Repository: repo,
			State:      client.PullRequestState_OPEN,
			Next:       nextURL,
		})
		if err != nil {
			app.QueueUpdateDraw(func() {
				table.View.SetCell(0, 0,
					tview.
						NewTableCell(err.Error()).
						SetAlign(tview.AlignCenter),
				)
			})
			return
		}

		nextURL = prs.NextURL

		app.QueueUpdateDraw(func() {
			table.Init(prs.Values)
		})

		if nextURL == "" {
			break
		}

		app.QueueUpdateDraw(func() {
			table.View.SetCell(len(prs.Values), 0, tview.NewTableCell("Loading..."))
		})
	}
}

func Run(params *paramutils.RepositoryParams) {
	c, repo, err := loadConfig(params)
	if err != nil {
		os.Exit(123)
	}

	app := tview.NewApplication()
	// app.SetScreen(tcell.NewSimulationScreen("sim"))

	// newPrimitive := func(text string) tview.Primitive {
	// 	return tview.NewTextView().
	// 		SetTextAlign(tview.AlignCenter).
	// 		SetText(text)
	// }
	// menu := newPrimitive("Menu")
	// main := newPrimitive("Main content")
	// sideBar := newPrimitive("Side Bar")

	table := newPullRequestTable()

	searchInput := tview.NewInputField()
	searchInput.
		SetPlaceholder("Filter pull requests").
		SetChangedFunc(func(text string) {
			table.Filter(text)
		}).
		SetBorder(true).
		SetTitle("Filter")

	searchInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			if searchInput.GetText() != "" {
				searchInput.SetText("")
			} else {
				app.SetFocus(table.View)
			}
		}

		return event
	})

	table.View.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			app.Stop()
			return nil
		case tcell.KeyCtrlD:
			// Decline pull requests
			return nil
			// case tcell.KeyUp:
			// 	table.moveSelectionUp()
			// case tcell.KeyDown:
			// 	table.moveSelectionDown()
		}

		switch event.Rune() {
		// case 'j':
		// 	table.moveSelectionDown()
		// case 'k':
		// 	table.moveSelectionUp()
		case 'q':
			app.Stop()
			return nil
		case '/':
			app.SetFocus(searchInput)
			return nil
		case ' ':
			table.selectCurrentRow()
			return nil
		}

		return event
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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

	go loadPRs(app, c, repo, table)
	app.SetRoot(grid, true).EnableMouse(true)
	app.SetFocus(table.View)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
