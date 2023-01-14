package tui

import (
	"fmt"
	"preq/internal/pkg/client"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type tableRepoData struct {
	Repository *client.Repository
	Client     client.Client
	Path       string
	Values     []*pullRequestTableRow
}

type pullRequestTableRow struct {
	pullRequest *client.PullRequest
	client      client.Client
	repository  *client.Repository
	selected    bool
	visible     bool
	tableRowId  int
}

type pullRequestTable struct {
	View          *tview.Table
	totalRowCount int
	tableData     []*tableRepoData
}

func pad(input string) string {
	return fmt.Sprintf(" %s", input)
}

var headers = []string{
	TableHeaderId, TableHeaderTitle, "SOURCE", "DESTINATION", "STATUS", "COMMENTS",
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
		tableData:     make([]*tableRepoData, 0),
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

func (prt *pullRequestTable) Init(data []*tableRepoData) {
	prt.tableData = data
	table.loadPRs(app)
}

func (prt *pullRequestTable) GetPullRequest(
	rowId int,
) (*pullRequestTableRow, error) {
	for _, trd := range prt.tableData {
		for _, v := range trd.Values {
			if v.tableRowId == rowId {
				return v, nil
			}
		}
	}

	return nil, fmt.Errorf("pull request not found row id %v", rowId)
}

func (prt *pullRequestTable) GetSelectedCount() int {
	count := 0
	for _, trd := range table.tableData {
		for _, row := range trd.Values {
			if row.selected && row.visible {
				count++
			}
		}
	}
	return count
}

func (prt *pullRequestTable) GetSelectedRows() []*pullRequestTableRow {
	rows := make([]*pullRequestTableRow, 0)
	for _, trd := range table.tableData {
		for _, row := range trd.Values {
			if row.selected && row.visible {
				rows = append(rows, row)
			}
		}
	}

	return rows
}

func (prt *pullRequestTable) GetRowByGlobalID(id string) *pullRequestTableRow {
	for _, trd := range prt.tableData {
		for _, v := range trd.Values {
			if v.pullRequest.URL == id {
				return v
			}
		}
	}

	return nil
}

func (prt *pullRequestTable) redraw() {
	prt.View.Clear()

	offset := 0

	for _, trd := range prt.tableData {
		size := 0
		for _, v := range trd.Values {
			if v.visible {
				size++
			}
		}

		if size == 0 {
			continue
		}

		offset = prt.drawTable(offset, trd)
		// Fill the cell with empty string to make tview skip it
		// when moving up and down through selection
		for i := 0; i < len(headers); i++ {
			prt.View.SetCell(
				offset,
				i,
				tview.NewTableCell("").
					SetSelectable(false),
			)
		}
		offset += 1
	}
}
func (prt *pullRequestTable) loadPRs(app *tview.Application) {
	for i := range prt.tableData {
		go prt.loadPR(app, i)
	}
}

func (prt *pullRequestTable) loadPR(
	app *tview.Application,
	rowId int,
) {
	app.QueueUpdateDraw(func() {
		prt.View.SetCell(
			0,
			0,
			tview.NewTableCell("Loading...").
				SetAlign(tview.AlignCenter),
		)
	})

	d := prt.tableData[rowId]

	nextURL := ""
	for {
		prs, err := d.Client.GetPullRequests(&client.GetPullRequestsOptions{
			Repository: d.Repository,
			State:      client.PullRequestState_OPEN,
			Next:       nextURL,
		})
		if err != nil {
			app.QueueUpdateDraw(func() {
				prt.View.SetCell(0, 0,
					tview.
						NewTableCell(err.Error()).
						SetAlign(tview.AlignCenter),
				)
			})
			return
		}

		nextURL = prs.NextURL

		for _, v := range prs.Values {
			d.Values = append(d.Values, &pullRequestTableRow{
				pullRequest: v,
				selected:    false,
				visible:     true,
				client:      d.Client,
				repository:  d.Repository,
			})
		}

		app.QueueUpdateDraw(func() {
			prt.redraw()
		})

		if nextURL == "" {
			break
		}

		// Write loading if we're expecting more data
		app.QueueUpdateDraw(func() {
			prt.View.SetCell(
				len(d.Values),
				0,
				tview.NewTableCell("Loading..."),
			)
		})
	}
}

func (prt *pullRequestTable) drawTable(offset int, trd *tableRepoData) int {
	headerStyle := tcell.StyleDefault.Bold(true)

	// Fill the cell with empty string to make tview skip it
	// when moving up and down through selection
	for i := 0; i < len(headers); i++ {
		prt.View.SetCell(
			offset,
			i,
			tview.NewTableCell("").
				SetSelectable(false).
				SetStyle(headerStyle),
		)
	}

	prt.View.GetCell(offset, 0).SetText("REPO")
	prt.View.GetCell(offset, 1).SetText(trd.Repository.Name)

	offset += 1

	for i := 0; i < len(headers); i++ {
		prt.View.SetCell(
			offset,
			i,
			tview.NewTableCell(pad(headers[i])).
				SetSelectable(false).
				SetStyle(headerStyle),
		)
	}

	offset += 1

	i := 0
	for _, v := range trd.Values {
		v.tableRowId = offset + i
		if v.visible {

			prt.addRow(v.pullRequest, i, offset)
			if v.pullRequest.State == client.PullRequestState_DECLINED {
				prt.View.GetCell(i+offset, 4).SetText(pad("Declined"))
				prt.setRowSelectable(i+offset, false)
				prt.colorRow(i+offset, DeclinedColor)
			} else if v.pullRequest.State == client.PullRequestState_MERGED {
				prt.View.GetCell(i+offset, 4).SetText(pad("Merged"))
				prt.setRowSelectable(i+offset, false)
				prt.colorRow(i+offset, MergedColor)
			} else if v.pullRequest.State == client.PullRequestState_DECLINING {
				prt.View.GetCell(i+offset, 4).SetText(pad("Declining..."))
				prt.setRowSelectable(i+offset, false)
				prt.colorRow(i+offset, tcell.ColorDarkRed)
			} else if v.pullRequest.State == client.PullRequestState_MERGING {
				prt.View.GetCell(i+offset, 4).SetText(pad("Merging..."))
				prt.setRowSelectable(i+offset, false)
				prt.colorRow(i+offset, tcell.ColorDarkOliveGreen)
			} else if v.selected {
				prt.colorRow(i+offset, SelectedColor)
			} else {
				prt.colorRow(i+offset, NormalColor)
			}
			i++
		}
	}

	return offset + i
}

func (prt *pullRequestTable) addRow(
	v *client.PullRequest,
	rowId int,
	offset int,
) {
	commentCount := ""
	if v.CommentCount > 0 {
		commentCount = strconv.FormatUint(uint64(v.CommentCount), 10)
	}

	// Escape the title string
	v.Title = strings.ReplaceAll(v.Title, "]", "[]")

	values := []string{
		v.ID,
		v.Title,
		v.Source,
		v.Destination,
		"Open",
		commentCount,
	}

	for i := 0; i < len(values); i++ {
		prt.View.SetCell(
			rowId+offset,
			i,
			tview.NewTableCell(pad(values[i])),
		)
	}
}

func (prt *pullRequestTable) setRowSelectable(rowId int, selectable bool) {
	for i := 0; i < prt.View.GetColumnCount(); i++ {
		prt.View.GetCell(rowId, i).SetSelectable(selectable)
	}
}

func (prt *pullRequestTable) colorRow(rowId int, color tcell.Color) {
	for i := 0; i < prt.View.GetColumnCount(); i++ {
		prt.View.GetCell(rowId, i).SetTextColor(color)
	}
}

func (prt *pullRequestTable) SelectCurrentRow() {
	row, _ := prt.View.GetSelection()

	pr, err := prt.GetPullRequest(row)
	if err != nil {
		// TODO: Change logging so it's not JSON?
		// TODO: Log error?
	}
	pr.selected = !pr.selected

	prt.redraw()
}

func (prt *pullRequestTable) Filter(input string) {
	for _, trd := range prt.tableData {
		for _, v := range trd.Values {
			v.visible = strings.Contains(
				strings.ToLower(v.pullRequest.Title),
				strings.ToLower(input),
			)
		}
	}

	prt.redraw()
}

func (prt *pullRequestTable) resetFilter() {

}
