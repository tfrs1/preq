package tui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"preq/internal/cli/paramutils"
	"preq/internal/clientutils"
	"preq/internal/configutils"
	"preq/internal/persistance"
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
		log.Error().Err(err).Msgf("error while loading default confg")
		return nil, nil, err
	}

	err = configutils.MergeLocalConfig(config, repoInfo.Path)
	if err != nil {
		// TODO: Do something
	}

	c, err := clientutils.ClientFactory{}.NewClient(
		client.RepositoryProviderEnum.BITBUCKET,
		config,
	)
	if err != nil {
		return nil, nil, err
	}

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

	addCommentModal := NewAddCommentModal()

	var deletionCommentReference interface{}
	deleteCommentModal := tview.NewModal().
		SetText("Are you sure want to delete this comment?").
		AddButtons([]string{"No", "Yes"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			log.Info().
				Msgf("DeleteCommendModal answered buttonIndex: %d, buttonLabel: %s", buttonIndex, buttonLabel)
			pages.HidePage("DeleteCommentModal")

			if buttonIndex == 0 || buttonIndex < 0 {
				eventBus.Publish(
					"DeleteCommendModal:DeleteCancelled",
					deletionCommentReference,
				)
			} else if buttonIndex == 1 {
				eventBus.Publish("DeleteCommendModal:DeleteConfirmed", deletionCommentReference)
			}

			deletionCommentReference = nil
		})

	errorModal := tview.NewModal().
		SetText("Unknown error").
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				pages.HidePage("ErrorModal")
				app.SetFocus(table.View)
			}
		})

	fatalErrorModal := tview.NewModal().
		SetText("Unknown error").
		AddButtons([]string{"Quit"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				app.Stop()
			}
		})

	eventBus.Subscribe("detailsPage:close", func(_ interface{}) {
		pages.HidePage("details_page")
		app.SetFocus(table.View)
	})

	eventBus.Subscribe("detailsPage:open", func(input interface{}) {
		pr, ok := input.(*PullRequest)
		if !ok {
			err := errors.New("cast failed when opening the details page")
			log.Error().Msg(err.Error())
			return
		}

		err := details.SetData(pr)
		if err != nil {
			eventBus.Publish("ErrorModal:RequestOpen", err)
			log.Error().Msg(err.Error())
			return
		}

		pages.ShowPage("details_page")
		app.SetFocus(details)
	})

	eventBus.Subscribe("ErrorModal:RequestOpen", func(err interface{}) {
		errorModal.SetText(err.(error).Error())
		pages.ShowPage("ErrorModal")
	})

	eventBus.Subscribe("FatalErrorModal:RequestOpen", func(err interface{}) {
		fatalErrorModal.SetText(err.(error).Error())
		pages.ShowPage("FatalErrorModal")
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

	grid := tview.NewGrid().
		SetRows(0, 1).
		AddItem(table.View, 0, 0, 1, 1, 0, 0, false).
		AddItem(tview.NewTextView().SetScrollable(true).SetText("Help: / filter ctrl+u unapprove j/k up/down"), 1, 0, 1, 1, 0, 0, false)

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
			if count := len(table.GetSelectedRows()); count > 0 {
				declineConfirmationModal.
					SetText(
						fmt.Sprintf(
							"Are you sure you want to decline %v pull requests?",
							count,
						),
					)
				pages.ShowPage(PAGE_DECLINE_CONFIRMATION_MODAL)
				return nil
			}
			return event
		case tcell.KeyCtrlA:
			if count := len(table.GetSelectedRows()); count > 0 {
				approveConfirmationModal.
					SetText(
						fmt.Sprintf(
							"Are you sure you want to approve %v pull requests?",
							count,
						),
					)
				pages.ShowPage(PAGE_APPROVE_CONFIRMATION_MODAL)
				return nil
			}
			return event
		case tcell.KeyCtrlU:
			if count := len(table.GetSelectedRows()); count > 0 {
				unapproveConfirmationModal.
					SetText(
						fmt.Sprintf(
							"Are you sure you want to unapprove %v pull requests?",
							count,
						),
					)
				pages.ShowPage(PAGE_UNAPPROVE_CONFIRMATION_MODAL)
				return nil
			}
			return event
		case tcell.KeyCtrlO:
			rowId, _ := table.View.GetSelection()
			r, err := table.GetPullRequest(rowId)
			if err != nil {
				// TODO: Log error?
			} else {
				eventBus.Publish("BrowserUrlOpen", r.PullRequest.URL)
			}
			return nil
		case tcell.KeyEnter:
			row, _ := table.View.GetSelection()
			pr, err := table.GetPullRequest(row)
			if err == nil && pr != nil {
				eventBus.Publish("detailsPage:open", pr)
			}
			return nil
		}

		switch event.Rune() {
		case 'm':
			if count := len(table.GetSelectedRows()); count > 0 {
				mergeConfirmationModal.
					SetText(
						fmt.Sprintf(
							"Are you sure you want to merge %v pull requests?",
							count,
						),
					)
				pages.ShowPage(PAGE_MERGE_CONFIRMATION_MODAL)
				return nil
			}
			return event
		case 'o':
			row, _ := table.View.GetSelection()
			pr, err := table.GetPullRequest(row)
			if err == nil && pr != nil {
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
			filterData := []*FilterModalItem{}
			for _, pr := range table.GetPullRequestList() {
				filterData = append(filterData, &FilterModalItem{
					Line: escapeString(pr.PullRequest.Title),
					Ref:  pr,
				})
			}
			eventBus.Publish("FilterModal:OpenRequested", filterData)
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

	pages.AddPage("details_page", details, true, false)

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
	pages.AddPage("FatalErrorModal", fatalErrorModal, false, false)
	pages.AddPage("ErrorModal", errorModal, false, false)

	pages.AddPage("AddCommentModal", addCommentModal, true, false)
	pages.AddPage("DeleteCommentModal", deleteCommentModal, true, false)

	filterModal := NewFilterModal()
	pages.AddPage("FilterModal", filterModal, true, false)

	eventBus.Subscribe(
		"DetailsPage:NewCommentRequested",
		func(ref interface{}) {
			addCommentModal.Clear()
			pages.ShowPage("AddCommentModal")
		},
	)

	eventBus.Subscribe(
		"DetailsPage:DeleteCommentRequested",
		func(ref interface{}) {
			deletionCommentReference = ref
			pages.ShowPage("DeleteCommentModal")
		},
	)

	eventBus.Subscribe("AddCommentModal:CancelRequested", func(_ interface{}) {
		pages.HidePage("AddCommentModal")
		eventBus.Publish("AddCommentModal:Closed", nil)
	})

	eventBus.Subscribe("AddCommentModal:CloseRequested", func(_ interface{}) {
		pages.HidePage("AddCommentModal")
		eventBus.Publish("AddCommentModal:Closed", nil)
	})

	eventBus.Subscribe("FilterModal:OpenRequested", func(input interface{}) {
		if filterData, ok := input.([]*FilterModalItem); ok {
			filterModal.Clear()
			filterModal.SetData(filterData, func(item *FilterModalItem) {
				eventBus.Publish("FilterModal:CloseRequested", nil)

				if item != nil {
					if pr, ok := item.Ref.(*PullRequest); ok {
						eventBus.Publish("detailsPage:open", pr)
					}
				}
			})

			pages.ShowPage("FilterModal")
			eventBus.Publish("FilterModal:Opened", nil)
		}
	})

	eventBus.Subscribe("FilterModal:CloseRequested", func(_ interface{}) {
		pages.HidePage("FilterModal")
		app.SetFocus(table.View)
		eventBus.Publish("FilterModal:Closed", nil)
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
