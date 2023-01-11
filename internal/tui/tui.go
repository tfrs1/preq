package tui

import (
	"fmt"
	"os"
	"os/exec"
	"preq/internal/cli/paramutils"
	"preq/internal/cli/utils"
	"preq/internal/configutils"
	"preq/internal/persistance"
	"preq/internal/pkg/bitbucket"
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
	eventBus = NewEventBus()
	prClient client.Client
	prRepo   *client.Repository
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
	repoInfo *persistance.PersistanceRepoInfo,
) (client.Client, *client.Repository, error) {
	config, err := configutils.DefaultConfig()
	if err != nil {
		// TODO: Do something
	}
	err = configutils.MergeLocalConfig(config, repoInfo.Path)
	if err != nil {
		// TODO: Do something
	}
	c := bitbucket.New(&bitbucket.ClientOptions{
		Username: config.GetString("bitbucket.username"),
		Password: config.GetString("bitbucket.password"),
	})
	if err != nil {
		return nil, nil, err
	}

	prClient = c

	r, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
		Provider:           client.RepositoryProvider(repoInfo.Provider),
		FullRepositoryName: repoInfo.Name,
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
	values *[]*pullRequestTableRow,
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

		for _, v := range prs.Values {
			*values = append(*values, &pullRequestTableRow{
				pullRequest: v,
				selected:    false,
				visible:     true,
				client:      &c,
				repository:  repo,
			})
		}

		app.QueueUpdateDraw(func() {
			table.Init()
		})

		if nextURL == "" {
			break
		}

		// Write loading if we're expecting more data
		app.QueueUpdateDraw(func() {
			table.View.SetCell(
				len(*values),
				0,
				tview.NewTableCell("Loading..."),
			)
		})
	}
}

var tableData = make([]*tableRepoData, 0)

func Run(
	params *paramutils.RepositoryParams,
	repos []*persistance.PersistanceRepoInfo,
) {
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
		SetLabelColor(tcell.ColorRed).
		SetPlaceholderTextColor(tcell.ColorLightGray)
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
			pages.ShowPage("merge_confirmation_modal")
			return event
		case tcell.KeyCtrlH:
			pages.ShowPage("HelpPage")
			return nil
		case tcell.KeyCtrlO:
			rowId, _ := table.View.GetSelection()
			r, err := table.GetPullRequest(rowId)
			if err != nil {
				// TODO: Log error?
			} else {
				eventBus.Publish("BrowserUrlOpen", r.pullRequest.URL)
			}
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
		SetDoneFunc(declineConfirmationCallback(pages)),
		false,
		false,
	)
	pages.AddPage("merge_confirmation_modal", tview.NewModal().
		SetText("Are you sure you want to merge %d pull requests?").
		AddButtons([]string{"Merge", "Cancel"}).
		SetDoneFunc(mergeConfirmationCallback(pages)),
		false,
		false,
	)
	pages.AddPage("HelpPage", helpPage, true, false)

	for _, v := range repos {
		tableData = append(tableData, &tableRepoData{
			Provider: client.RepositoryProvider(v.Provider),
			Name:     v.Name,
			Path:     v.Path,
		})
	}

	for _, v := range tableData {
		// TODO get data for each repo
		// update table after getting repo data
		// Throw event to redraw?
		c, repo, err := loadConfig(&persistance.PersistanceRepoInfo{
			Name:     v.Name,
			Provider: string(v.Provider),
			Path:     v.Path,
		})

		if err != nil {
			log.Error().Msg(err.Error())
			os.Exit(123)
		}

		go loadPRs(app, c, repo, table, &v.Values)
	}

	app.SetRoot(pages, true).EnableMouse(true)
	app.SetFocus(table.View)

	if err := app.Run(); err != nil {
		panic(err)
	}
}

type tableRepoData struct {
	Provider client.RepositoryProvider
	Name     string
	Path     string
	Values   []*pullRequestTableRow
}

func declineConfirmationCallback(pages *tview.Pages) func(int, string) {
	return func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			selectedPRs := make(map[string]*promptPullRequest)
			for _, trd := range tableData {
				for _, row := range trd.Values {
					if row.selected && row.visible {
						selectedPRs[row.pullRequest.URL] = &promptPullRequest{
							ID:         row.pullRequest.ID,
							GlobalID:   row.pullRequest.URL,
							Title:      row.pullRequest.Title,
							Repository: row.repository,
							Client:     row.client,
						}

						row.pullRequest.State = client.PullRequestState_DECLINING
						row.selected = false
					}
				}
			}

			table.redraw()

			go processPullRequestMap(
				selectedPRs,
				declinePR,
				func(msg utils.ProcessPullRequestResponse) string {
					if msg.Status == "Done" {
						for _, trd := range tableData {
							for _, v := range trd.Values {
								if v.pullRequest.URL == msg.GlobalID {
									v.pullRequest.State = client.PullRequestState_DECLINED
								}
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
	}
}

func mergeConfirmationCallback(pages *tview.Pages) func(int, string) {
	return func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			selectedPRs := make(map[string]*promptPullRequest)

			for _, trd := range tableData {
				for _, row := range trd.Values {
					if row.selected && row.visible {
						selectedPRs[row.pullRequest.URL] = &promptPullRequest{
							ID:         row.pullRequest.ID,
							GlobalID:   row.pullRequest.URL,
							Title:      row.pullRequest.Title,
							Client:     row.client,
							Repository: row.repository,
						}

						row.pullRequest.State = client.PullRequestState_MERGING
						row.selected = false
					}
				}
			}

			table.redraw()

			go processPullRequestMap(
				selectedPRs,
				mergePR,
				func(msg utils.ProcessPullRequestResponse) string {
					if msg.Status == "Done" {
						// v, err := table.FindPullRequestFunc((v) {
						// 	return v.pullRequest == m.ID
						// })
						// if err == nil {
						// 	// TODO: Log error
						// }
						// v.pullRequest.State = client.PullRequestState_MERGED

						for _, trd := range tableData {
							for _, v := range trd.Values {
								// TODO: Comparing URL to ID weird, but URL is the only true ID now because of multi repo table
								if v.pullRequest.URL == msg.ID {
									v.pullRequest.State = client.PullRequestState_MERGED
								}
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
	}
}
