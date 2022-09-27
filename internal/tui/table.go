package tui

import (
	"fmt"
	"preq/internal/pkg/client"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type pullRequestTable struct {
	View                *tview.Table
	totalRowCount       int
	selectedMap         map[int]bool
	pullRequestList     []*client.PullRequest
	filteredRequestList []*client.PullRequest
}

func newPullRequestTable() *pullRequestTable {
	table := tview.NewTable()
	table.
		SetBorder(true).
		SetTitle("Pull requests")

	// // Set box options
	// table.
	// 	SetTitle("preq").
	// 	SetBorder(true)
	prt := &pullRequestTable{
		View:                table,
		totalRowCount:       0,
		selectedMap:         make(map[int]bool),
		pullRequestList:     make([]*client.PullRequest, 0),
		filteredRequestList: make([]*client.PullRequest, 0),
	}

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
			// table.GetCell(row, column).SetTextColor(tcell.ColorRed)
			// table.SetSelectable(false, false)
		})

	return prt
}

func (prt *pullRequestTable) Init(prList []*client.PullRequest) {
	prt.pullRequestList = prList
	prt.filteredRequestList = prList

	prt.redraw()
}

func (prt *pullRequestTable) redraw() {
	prt.View.Clear()
	prt.View.SetCell(0, 0, tview.NewTableCell("#").SetSelectable(false))
	prt.View.SetCell(0, 1, tview.NewTableCell("Title").SetSelectable(false))
	prt.View.SetCell(0, 2, tview.NewTableCell("Source -> Destination").SetSelectable(false))

	for i, v := range prt.filteredRequestList {
		prt.addRow(v, i)
	}
}

func (prt *pullRequestTable) addRow(v *client.PullRequest, i int) {
	prt.View.SetCell(i+1, 0, tview.NewTableCell(v.ID))
	prt.View.SetCell(i+1, 1, tview.NewTableCell(v.Title))
	prt.View.SetCell(i+1, 2, tview.NewTableCell(fmt.Sprintf("%s -> %s", v.Source, v.Destination)))
	prt.selectedMap[i] = false
}

func (prt *pullRequestTable) selectCurrentRow() {
	row, _ := prt.View.GetSelection()
	selectedRow := row - 1

	color := tcell.ColorRed
	if prt.selectedMap[selectedRow] {
		color = tcell.ColorWhite
	}

	for i := 0; i < prt.View.GetColumnCount(); i++ {
		prt.View.GetCell(selectedRow+1, i).SetTextColor(color)
	}
	prt.selectedMap[selectedRow] = !prt.selectedMap[selectedRow]
}

func (prt *pullRequestTable) Filter(input string) {
	prt.filteredRequestList = make([]*client.PullRequest, 0)
	for _, v := range prt.pullRequestList {
		if strings.Contains(strings.ToLower(v.Title), strings.ToLower(input)) {
			prt.filteredRequestList = append(prt.filteredRequestList, v)
		}
	}

	prt.redraw()
}

func (prt *pullRequestTable) resetFilter() {

}
