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
	app       = tview.NewApplication()
	flex      = tview.NewFlex()
	details   = newDetailsPage()
	table     = newPullRequestTable()
	eventBus  = NewEventBus()
	prClient  client.Client
	prRepo    *client.Repository
	tableData []*tableRepoData
)

const (
	PAGE_APPROVE_CONFIRMATION_MODAL   = "page_approve_confirmation_modal"
	PAGE_UNAPPROVE_CONFIRMATION_MODAL = "aage_unapprove_confirmation_modal"
	PAGE_MERGE_CONFIRMATION_MODAL     = "page_merge_confirmation_modal"
	PAGE_DECLINE_CONFIRMATION_MODAL   = "confirmation_modal"
)

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
		Provider: client.RepositoryProvider(repoInfo.Provider),
		Name:     repoInfo.Name,
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

func Run(
	params *paramutils.RepositoryParams,
	repos []*persistance.PersistanceRepoInfo,
) {
	// app.SetScreen(tcell.NewSimulationScreen("sim"))

	pages := tview.NewPages()

	errorModal := tview.NewModal().
		SetText("Unknown error").
		AddButtons([]string{"Quit"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				app.Stop()
			}
		})

	eventBus.Subscribe("detailsPage:close", func(_ interface{}) {
		flex.RemoveItem(details.View)
		app.SetFocus(table.View)
	})

	eventBus.Subscribe("detailsPage:open", func(_ interface{}) {
		flex.AddItem(details.View, 0, 1, false)
		app.SetFocus(details.View)
	})

	eventBus.Subscribe("errorModal:open", func(err interface{}) {
		errorModal.SetText(err.(error).Error())
		pages.ShowPage("error_modal")
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

	approveConfirmationModal := tview.NewModal().
		SetText("Are you sure you want to approve %d pull requests?").
		AddButtons([]string{"Approve", "Cancel"}).
		SetDoneFunc(approveConfirmationCallback(pages))

	unapproveConfirmationModal := tview.NewModal().
		SetText("Are you sure you want to unapprove %d pull requests?").
		AddButtons([]string{"Unapprove", "Cancel"}).
		SetDoneFunc(unapproveConfirmationCallback(pages))

	mergeConfirmationModal := tview.NewModal().
		SetText("Are you sure you want to merge %d pull requests?").
		AddButtons([]string{"Merge", "Cancel"}).
		SetDoneFunc(mergeConfirmationCallback(pages))

	declineConfirmationModal := tview.NewModal().
		SetText("Are you sure you want to decline %d pull requests?").
		AddButtons([]string{"Decline", "Cancel"}).
		SetDoneFunc(declineConfirmationCallback(pages))

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
			count := len(table.GetSelectedRows())
			declineConfirmationModal.
				SetText(
					fmt.Sprintf(
						"Are you sure you want to decline %v pull requests?",
						count,
					),
				)
			pages.ShowPage(PAGE_DECLINE_CONFIRMATION_MODAL)
			return event
		case tcell.KeyCtrlA:
			count := len(table.GetSelectedRows())

			approveConfirmationModal.
				SetText(
					fmt.Sprintf(
						"Are you sure you want to approve %v pull requests?",
						count,
					),
				)
			pages.ShowPage(PAGE_APPROVE_CONFIRMATION_MODAL)
			return event
		case tcell.KeyCtrlU:
			count := len(table.GetSelectedRows())

			unapproveConfirmationModal.
				SetText(
					fmt.Sprintf(
						"Are you sure you want to unapprove %v pull requests?",
						count,
					),
				)
			pages.ShowPage(PAGE_UNAPPROVE_CONFIRMATION_MODAL)
			return event
		case tcell.KeyCtrlM:
			count := len(table.GetSelectedRows())
			mergeConfirmationModal.
				SetText(
					fmt.Sprintf(
						"Are you sure you want to merge %v pull requests?",
						count,
					),
				)
			pages.ShowPage(PAGE_MERGE_CONFIRMATION_MODAL)
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
				eventBus.Publish("BrowserUrlOpen", r.PullRequest.URL)
			}
		}

		switch event.Rune() {
		// TODO: Add details feature
		// case 'o':
		// 	eventBus.Publish("detailsPage:open", nil)
		case 'q':
			app.Stop()
			return nil
		case '/':
			app.SetFocus(searchInput)
			return nil
		case ' ':
			table.SelectCurrentRow()
			return nil
		}

		return event
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})

	pages.AddPage("main", flex, true, true)

	pages.AddPage(
		PAGE_DECLINE_CONFIRMATION_MODAL,
		declineConfirmationModal,
		false,
		false,
	)
	pages.AddPage(
		PAGE_UNAPPROVE_CONFIRMATION_MODAL,
		unapproveConfirmationModal,
		false,
		false,
	)
	pages.AddPage(
		PAGE_APPROVE_CONFIRMATION_MODAL,
		approveConfirmationModal,
		false,
		false,
	)
	pages.AddPage(
		PAGE_MERGE_CONFIRMATION_MODAL,
		mergeConfirmationModal,
		false,
		false,
	)
	pages.AddPage("HelpPage", helpPage, true, false)
	pages.AddPage("error_modal", errorModal, false, false)

	tableData = make([]*tableRepoData, 0)
	for _, v := range repos {
		c, repo, err := loadConfig(&persistance.PersistanceRepoInfo{
			Name:     v.Name,
			Provider: string(v.Provider),
			Path:     v.Path,
		})

		if err != nil {
			log.Error().Msg(err.Error())
			os.Exit(123)
		}

		tableData = append(tableData, &tableRepoData{
			Repository: repo,
			Client:     c,
			Path:       v.Path,
		})
	}

	go func() {
		table.Init(tableData)
		app.QueueUpdateDraw(func() {
			redraw()
		})
	}()

	app.SetRoot(pages, true) //.EnableMouse(true)
	app.SetFocus(table.View)

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func redraw() {
	table.redraw()
}

func declineConfirmationCallback(pages *tview.Pages) func(int, string) {
	return func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			selectedPRs := make(map[string]*promptPullRequest)

			for _, row := range table.GetSelectedRows() {
				selectedPRs[row.PullRequest.URL] = &promptPullRequest{
					ID:         row.PullRequest.ID,
					GlobalID:   row.PullRequest.URL,
					Title:      row.PullRequest.Title,
					Repository: row.Repository,
					Client:     row.Client,
				}

				row.PullRequest.State = client.PullRequestState_DECLINING
				row.Selected = false
			}

			redraw()

			go processPullRequestMap(
				selectedPRs,
				declinePR,
				func(msg utils.ProcessPullRequestResponse) string {
					if msg.Status == "Done" {
						v := table.GetRowByGlobalID(msg.GlobalID)
						v.PullRequest.State = client.PullRequestState_DECLINED
					}

					app.QueueUpdateDraw(table.redraw)

					return ""
				},
			)
		}

		pages.SwitchToPage("main")
		app.SetFocus(table.View)
	}
}

func approveConfirmationCallback(pages *tview.Pages) func(int, string) {
	return func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			selectedPRs := make(map[string]*promptPullRequest)

			for _, row := range table.GetSelectedRows() {
				// approved := false
				// for i, pra := range row.PullRequest.Approvals {
				//   if pra.User == row.Client.User {
				//     approved = true
				//     break
				//   }
				// }

				// if (approved) {
				//   continue
				// }

				selectedPRs[row.PullRequest.URL] = &promptPullRequest{
					ID:         row.PullRequest.ID,
					GlobalID:   row.PullRequest.URL,
					Title:      row.PullRequest.Title,
					Client:     row.Client,
					Repository: row.Repository,
				}

				// TODO: This should probably be a method in table instead
				row.Selected = false
				row.IsApprovalsLoading = true
			}

			table.redraw()

			go processPullRequestMap(
				selectedPRs,
				approvePR,
				func(msg utils.ProcessPullRequestResponse) string {
					v := table.GetRowByGlobalID(msg.GlobalID)
					if msg.Error != nil {
						v.IsApprovalsLoading = false
					}

					if msg.Status == "Done" {
						// TODO: return an error instead?
						if v != nil {
							go func(v *PullRequest) {
								err := v.Client.FillMiscInfoAsync(
									v.Repository,
									v.PullRequest,
								)

								if err != nil {
									// TODO: Handle error
									return
								}

								v.IsApprovalsLoading = false

								app.QueueUpdateDraw(table.redraw)
							}(v)
						}
					}

					app.QueueUpdateDraw(table.redraw)

					return ""
				},
			)
		}

		pages.SwitchToPage("main")
		app.SetFocus(table.View)
	}
}

func unapproveConfirmationCallback(pages *tview.Pages) func(int, string) {
	return func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			selectedPRs := make(map[string]*promptPullRequest)

			for _, row := range table.GetSelectedRows() {
				selectedPRs[row.PullRequest.URL] = &promptPullRequest{
					ID:         row.PullRequest.ID,
					GlobalID:   row.PullRequest.URL,
					Title:      row.PullRequest.Title,
					Client:     row.Client,
					Repository: row.Repository,
				}

				// TODO: This should probably be a method in table instead
				row.Selected = false
				row.IsApprovalsLoading = true
			}

			table.redraw()

			go processPullRequestMap(
				selectedPRs,
				unapprovePR,
				func(msg utils.ProcessPullRequestResponse) string {
					v := table.GetRowByGlobalID(msg.GlobalID)
					v.IsApprovalsLoading = false

					if msg.Status == "Done" {
						// TODO: return an error instead?
						if v != nil {
							go func(v *PullRequest) {
								err := v.Client.FillMiscInfoAsync(
									v.Repository,
									v.PullRequest,
								)

								if err != nil {
									return
								}

								v.IsApprovalsLoading = false

								app.QueueUpdateDraw(table.redraw)
							}(v)
						}
					}

					app.QueueUpdateDraw(table.redraw)

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

			for _, row := range table.GetSelectedRows() {
				selectedPRs[row.PullRequest.URL] = &promptPullRequest{
					ID:         row.PullRequest.ID,
					GlobalID:   row.PullRequest.URL,
					Title:      row.PullRequest.Title,
					Client:     row.Client,
					Repository: row.Repository,
				}

				// TODO: This should probably be a method in table instead
				row.PullRequest.State = client.PullRequestState_MERGING
				row.Selected = false
			}

			table.redraw()

			go processPullRequestMap(
				selectedPRs,
				mergePR,
				func(msg utils.ProcessPullRequestResponse) string {
					if msg.Status == "Done" {
						v := table.GetRowByGlobalID(msg.GlobalID)
						// TODO: return an error instead?
						if v != nil {
							v.PullRequest.State = client.PullRequestState_MERGED
						}
					}

					app.QueueUpdateDraw(table.redraw)

					return ""
				},
			)
		}

		pages.SwitchToPage("main")
		app.SetFocus(table.View)
	}
}
