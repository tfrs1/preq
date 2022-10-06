package tui

import (
	"fmt"
	"preq/internal/pkg/client"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	statusColumn = 4
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

func pad(input string) string {
	return fmt.Sprintf(" %s", input)
}

func newPullRequestTable() *pullRequestTable {
	table := tview.NewTable()
	table.SetBackgroundColor(tcell.ColorSlateGray)
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
		// SetDoneFunc(func(key tcell.Key) {
		// 	if key == tcell.KeyEscape {
		// 		app.Stop()
		// 	}
		// 	if key == tcell.KeyEnter {
		// 		table.SetSelectable(true, false)
		// 	}
		// }).
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
	headerStyle := tcell.StyleDefault.
		Bold(true)

	prt.View.SetCell(
		0,
		0,
		tview.NewTableCell(pad("#")).
			SetSelectable(false).
			SetStyle(headerStyle),
	)
	prt.View.SetCell(
		0,
		1,
		tview.NewTableCell(pad("TITLE")).
			SetSelectable(false).
			SetStyle(headerStyle),
	)
	prt.View.SetCell(
		0,
		2,
		tview.NewTableCell(pad("SOURCE")).
			SetSelectable(false).
			SetStyle(headerStyle),
	)
	prt.View.SetCell(
		0,
		3,
		tview.NewTableCell(pad("DESTINATION")).
			SetSelectable(false).
			SetStyle(headerStyle),
	)
	prt.View.SetCell(
		0,
		statusColumn,
		tview.NewTableCell(pad("STATUS")).
			SetSelectable(false).
			SetStyle(headerStyle),
	)

	i := 0
	for _, v := range prt.rows {
		if v.visible {
			prt.addRow(v.pullRequest, i)
			// TODO: If merged green
			// TODO: If declined red
			if v.pullRequest.State == client.PullRequestState_DECLINED {
				prt.View.GetCell(i+1, 4).
					SetText(pad("Declined")).
					SetSelectable(false)
				for j := 0; j < prt.View.GetColumnCount(); j++ {
					prt.View.GetCell(i+1, j).SetSelectable(false)
				}
				prt.colorRow(i, DeclinedColor)
			} else if v.selected {
				prt.colorRow(i, SelectedColor)
			} else {
				prt.colorRow(i, NormalColor)
			}
			i++
		}
	}
}

func (prt *pullRequestTable) addRow(v *client.PullRequest, i int) {
	prt.View.SetCell(
		i+1,
		0,
		tview.NewTableCell(pad(v.ID)),
	)
	prt.View.SetCell(i+1, 1, tview.NewTableCell(pad(v.Title)))
	prt.View.SetCell(
		i+1,
		2,
		tview.NewTableCell(pad(v.Source)),
	)
	prt.View.SetCell(
		i+1,
		3,
		tview.NewTableCell(pad(v.Destination)),
	)
	prt.View.SetCell(i+1, statusColumn, tview.NewTableCell(pad("Open")))
}

func (prt *pullRequestTable) colorRow(rowId int, color tcell.Color) {
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
				// TODO: Instead of redrawing just color the row? possibly dangerous
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
