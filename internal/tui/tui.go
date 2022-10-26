package tui

import (
	"fmt"
	"os"
	"os/exec"
	"preq/internal/cli/paramutils"
	"preq/internal/clientutils"
	"preq/internal/pkg/client"
	"runtime"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

var (
	app      = tview.NewApplication()
	flex     = tview.NewFlex()
	details  = newDetailsPage()
	table    = newPullRequestTable()
	prClient client.Client
	prRepo   *client.Repository
)

// var prClient client.Client

var (
	eventBus = NewEventBus()
)

type EventBus struct {
	subscribers map[string][]EventBusEventCallback
}

type EventBusEventCallback func(data interface{})

func (bus *EventBus) Publish(name string, data interface{}) {
	for _, v := range bus.subscribers[name] {
		v(data)
	}
}

func (bus *EventBus) Subscribe(name string, callback EventBusEventCallback) {
	if eventBus.subscribers[name] == nil {
		eventBus.subscribers[name] = make([]EventBusEventCallback, 0)
	}

	eventBus.subscribers[name] = append(eventBus.subscribers[name], callback)
}

func NewEventBus() *EventBus {
	subscribers := make(map[string][]EventBusEventCallback)
	return &EventBus{
		subscribers: subscribers,
	}
}

func loadConfig(
	params *paramutils.RepositoryParams,
) (client.Client, *client.Repository, error) {
	c, err := clientutils.ClientFactory{}.DefaultClient(params.Provider)
	if err != nil {
		return nil, nil, err
	}

	prClient = c

	r, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
		Provider:           params.Provider,
		FullRepositoryName: params.Name,
	})
	if err != nil {
		return nil, nil, err
	}
	// prRepo = r

	// prClient.GetPullRequestInfo(&client.ApprovePullRequestOptions{
	// 	Repository: prRepo,
	// 	ID:         "77",
	// })

	return c, r, nil
}

func loadPRs(
	app *tview.Application,
	c client.Client,
	repo *client.Repository,
	table *pullRequestTable,
) {
	app.QueueUpdateDraw(func() {
		table.View.SetCell(
			0,
			0,
			tview.NewTableCell("Loading...").
				SetAlign(tview.AlignCenter),
		)
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
			table.View.SetCell(
				len(prs.Values),
				0,
				tview.NewTableCell("Loading..."),
			)
		})
	}
}

func Run(params *paramutils.RepositoryParams) {
	c, repo, err := loadConfig(params)
	if err != nil {
		os.Exit(123)
	}

	if repo != nil {
		persistanceRepo.AddVisited(
			fmt.Sprintf("%s/%s", repo.Owner, repo.Name),
			string(repo.Provider),
		)
	}

	// app.SetScreen(tcell.NewSimulationScreen("sim"))

	eventBus.Subscribe("detailsPage:close", func(_ interface{}) {
		flex.RemoveItem(details.View)
		app.SetFocus(table.View)
	})

	eventBus.Subscribe("detailsPage:open", func(_ interface{}) {
		flex.AddItem(details.View, 0, 1, false)
		app.SetFocus(details.View)
	})

	eventBus.Subscribe("BrowserUrlOpen", func(data interface{}) {
		url := data.(string)
		var err error

		switch runtime.GOOS {
		case "linux":
			err = exec.Command("xdg-open", url).Start()
		case "windows":
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).
				Start()
		case "darwin":
			err = exec.Command("open", url).Start()
		default:
			err = fmt.Errorf("unsupported platform")
		}
		if err != nil {
			log.Fatal().Msg("Unknown system for url open")
		}
	})

	searchInput := tview.NewInputField().
		SetLabel(" Filter ").
		SetLabelColor(tcell.ColorRed)
	searchInput.
		SetPlaceholder(" Filter pull requests").
		SetChangedFunc(func(text string) {
			table.Filter(text)
		})

	searchInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			if searchInput.GetText() != "" {
				searchInput.SetText("")
			} else {
				app.SetFocus(table.View)
			}
		case tcell.KeyEnter:
			app.SetFocus(table.View)
		}

		return event
	})

	grid := tview.NewGrid().
		SetRows(0, 1).
		AddItem(table.View, 0, 0, 1, 1, 0, 0, false).
		AddItem(searchInput, 1, 0, 1, 1, 0, 0, false)

	grid.
		SetBorders(false).
		SetBorder(false)
	pages := tview.NewPages()
	flex.AddItem(grid, 0, 1, false)
	helpPage := tview.NewBox().
		SetTitle("Help")

	helpPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			pages.HidePage("HelpPage")
			return nil
		}

		switch event.Rune() {
		case 'h':
			pages.HidePage("HelpPage")
			return nil
		}

		return event
	})

	table.View.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlD:
			pages.ShowPage("confirmation_modal")
			return event
		case tcell.KeyCtrlM:
			pages.ShowPage("confirmation_modal")
			return event
		case tcell.KeyCtrlH:
			pages.ShowPage("HelpPage")
			return nil
		case tcell.KeyCtrlO:
			rowId, _ := table.View.GetSelection()
			r := table.rows[rowId-1]
			eventBus.Publish("BrowserUrlOpen", r.pullRequest.URL)
		}

		switch event.Rune() {
		case 'o':
			eventBus.Publish("detailsPage:open", nil)
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

	pages.AddPage("main", flex, true, true)
	pages.AddPage("confirmation_modal", tview.NewModal().
		SetText("Are you sure you want to decline %d pull requests?").
		AddButtons([]string{"Decline", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				selectedPRs := make(map[string]*promptPullRequest)
				for index, row := range table.rows {
					if row.selected && row.visible {
						selectedPRs[row.pullRequest.ID] = &promptPullRequest{
							ID:    row.pullRequest.ID,
							Title: row.pullRequest.Title,
						}

						table.View.GetCell(index+1, 4).
							SetText(pad("Declining..."))

						for i := 0; i < table.View.GetColumnCount(); i++ {
							table.View.GetCell(index+1, i).SetSelectable(false)
						}

					}
				}

				go execute(c, repo, selectedPRs,
					func(msg interface{}) string {
						m := msg.(declineResponse)
						if m.Status == "Done" {
							for _, v := range table.rows {
								if v.pullRequest.ID == m.ID {
									v.pullRequest.State = client.PullRequestState_DECLINED
								}
							}
						}
						app.QueueUpdateDraw(func() {
							table.redraw()
						})

						return ""
					},
				)
			}

			pages.SwitchToPage("main")
			app.SetFocus(table.View)
		}),
		false,
		false,
	)
	pages.AddPage("HelpPage", helpPage, true, false)
	pages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'h':
			return tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)
		case 'l':
			return tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)
		}
		return event
	})

	go loadPRs(app, c, repo, table)
	app.SetRoot(pages, true).EnableMouse(true)
	app.SetFocus(table.View)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
