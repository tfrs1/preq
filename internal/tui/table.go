package tui

import (
	"fmt"
	"preq/internal/pkg/client"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type pullRequestTableRow struct {
	pullRequest *client.PullRequest
	selected    bool
	visible     bool
}

type pullRequestTable struct {
	View          *tview.Table
	totalRowCount int
	rows          []*pullRequestTableRow
}

func newPullRequestTable() *pullRequestTable {
	table := tview.NewTable()
	table.
		SetFixed(2, 0).
		SetBorder(true).
		SetTitle("Pull requests")
	// // Set box options
	// table.
	// 	SetTitle("preq").
	// 	SetBorder(true)
	prt := &pullRequestTable{
		View:          table,
		totalRowCount: 0,
		rows:          make([]*pullRequestTableRow, 0),
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
	prt.rows = make([]*pullRequestTableRow, 0)
	for _, v := range prList {
		prt.rows = append(prt.rows, &pullRequestTableRow{
			pullRequest: v,
			selected:    false,
			visible:     true,
		})
	}

	prt.redraw()
}

func (prt *pullRequestTable) redraw() {
	prt.View.Clear()
	prt.View.SetCell(0, 0, tview.NewTableCell("#").SetSelectable(false))
	prt.View.SetCell(0, 1, tview.NewTableCell("Title").SetSelectable(false))
	prt.View.SetCell(0, 2, tview.NewTableCell("Source -> Destination").SetSelectable(false))
	prt.View.SetCell(0, 3, tview.NewTableCell("STATUS").SetSelectable(false))

	i := 0
	for _, v := range prt.rows {
		if v.visible {
			prt.addRow(v.pullRequest, i)
			prt.colorRow(i, v.selected)
			i++
		}
	}
}

func (prt *pullRequestTable) addRow(v *client.PullRequest, i int) {
	prt.View.SetCell(i+1, 0, tview.NewTableCell(v.ID))
	prt.View.SetCell(i+1, 1, tview.NewTableCell(v.Title))
	prt.View.SetCell(i+1, 2, tview.NewTableCell(fmt.Sprintf("%s -> %s", v.Source, v.Destination)))
	prt.View.SetCell(i+1, 3, tview.NewTableCell("Open"))
}

func (prt *pullRequestTable) colorRow(rowId int, selected bool) {
	color := tcell.ColorWhite
	if selected {
		color = tcell.ColorRed
	}

	for i := 0; i < prt.View.GetColumnCount(); i++ {
		prt.View.GetCell(rowId+1, i).SetTextColor(color)
	}
}

func (prt *pullRequestTable) selectCurrentRow() {
	row, _ := prt.View.GetSelection()
	selectedRow := row - 1

	rowId := 0
	for _, v := range prt.rows {
		if v.visible {
			if rowId == selectedRow {
				v.selected = !v.selected
				// TODO: Instead of redrawing just color the row? possible dangerous
				// prt.colorRow(rowId, v.selected)
				break
			}

			rowId++
		}
	}

	prt.redraw()
}

func (prt *pullRequestTable) Filter(input string) {
	for _, v := range prt.rows {
		v.visible = strings.Contains(
			strings.ToLower(v.pullRequest.Title),
			strings.ToLower(input),
		)
	}

	prt.redraw()
}

func (prt *pullRequestTable) resetFilter() {

}
