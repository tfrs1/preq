package tui

import (
	"fmt"
	"os"
	"os/exec"
	"preq/internal/cli/paramutils"
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

	r, err := client.NewRepositoryFromOptions(
		&client.RepositoryOptions{
			Provider: client.RepositoryProvider(repoInfo.Provider),
			Name:     repoInfo.Name,
		},
	)
	if err != nil {
		return nil, nil, err
	}

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

	eventBus.Subscribe("mergeModal:closed", func(_ interface{}) {
		pages.SwitchToPage("main")
		app.SetFocus(table.View)
	})

	eventBus.Subscribe("approveModal:closed", func(_ interface{}) {
		pages.SwitchToPage("main")
		app.SetFocus(table.View)
	})

	eventBus.Subscribe("unapproveModal:closed", func(_ interface{}) {
		pages.SwitchToPage("main")
		app.SetFocus(table.View)
	})

	eventBus.Subscribe("declineModal:closed", func(_ interface{}) {
		pages.SwitchToPage("main")
		app.SetFocus(table.View)
	})

	eventBus.Subscribe("HelpPage:Open", func(_ interface{}) {
		pages.ShowPage("HelpPage")
	})

	eventBus.Subscribe("HelpPage:Closed", func(_ interface{}) {
		pages.HidePage("HelpPage")
		pages.SwitchToPage("main")
		app.SetFocus(table.View)
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
	flex.AddItem(grid, 0, 1, false)

	helpPage := tview.NewBox().
		SetTitle("Help")

	helpPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			eventBus.Publish("HelpPage:Closed", nil)
			return nil
		}

		switch event.Rune() {
		case 'h':
		case 'q':
			eventBus.Publish("HelpPage:Closed", nil)
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
		case 'o':
			row, _ := table.View.GetSelection()
			pr, err := table.GetPullRequest(row)
			if err == nil {
				eventBus.Publish("detailsPage:open", pr)
			}
			return nil
		case 'h':
			eventBus.Publish("HelpPage:Open", nil)
			return nil
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

	pages.AddPage("AddCommentModal", NewAddCommentModal(), true, false)

	eventBus.Subscribe("DetailsPage:NewCommentRequested", func(ref interface{}) {
		pages.ShowPage("AddCommentModal")
	})

	eventBus.Subscribe("AddCommentModal:CancelRequested", func(_ interface{}) {
		pages.HidePage("AddCommentModal")
		eventBus.Publish("AddCommentModal:Closed", nil)
	})

	eventBus.Subscribe("AddCommentModal:CloseRequested", func(_ interface{}) {
		pages.HidePage("AddCommentModal")
		eventBus.Publish("AddCommentModal:Closed", nil)
	})

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
		app.QueueUpdateDraw(redraw)
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
