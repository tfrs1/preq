package tui

import (
	"fmt"
	"preq/internal/pkg/client"
	"sort"
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

// TODO: Use a string instead, then it can be configurable
var headers = []string{
	TableHeaderId, TableHeaderTitle, "SOURCE", "DESTINATION", "STATUS", "APPROVED", "CHANGES REQUESTED", "COMMENTS",
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
) (*PullRequest, error) {
	for _, data := range state.RepositoryData {
		for _, pr := range data.PullRequests {
			if pr.TableRowId == rowId {
				return pr, nil
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

	prt.drawTable()
}

func (prt *pullRequestTable) loadPRs(app *tview.Application) {
	state.RepositoryData = make(map[string]*RepositoryData)
	for i := range prt.tableData {
		id := repoId(prt.tableData[i].Repository)

		if _, ok := state.RepositoryData[id]; !ok {
			state.RepositoryData[id] = &RepositoryData{
				Name:         prt.tableData[i].Repository.Name,
				IsLoading:    true,
				PullRequests: make(map[string]*PullRequest),
			}
		}

		go prt.loadPR(app, i)
	}
}

func repoId(repo *client.Repository) string {
	return fmt.Sprintf(
		"%v___%v",
		repo.Provider,
		repo.Name,
	)
}

func (prt *pullRequestTable) loadPR(
	app *tview.Application,
	rowId int,
) {
	// TODO: This load should be in table write code
	// TODO here just the state should be updater

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

		id := repoId(d.Repository)
		for _, v := range prs.Values {
			d.Values = append(d.Values, &pullRequestTableRow{
				pullRequest: v,
				selected:    false,
				visible:     true,
				client:      d.Client,
				repository:  d.Repository,
			})

			state.RepositoryData[id].PullRequests[v.ID] = &PullRequest{
				PullRequest:              v,
				Selected:                 false,
				Visible:                  true,
				Client:                   d.Client,
				Repository:               d.Repository,
				IsApprovalsLoading:       true,
				IsCommentsLoading:        true,
				IsChangesRequestsLoading: true,
			}

			go func(v *client.PullRequest) {
				err := d.Client.FillMiscInfoAsync(
					d.Repository,
					v,
				)

				if err != nil {
					return
				}

				id := repoId(d.Repository)
				pr := state.RepositoryData[id].PullRequests[v.ID]
				pr.IsApprovalsLoading = false
				pr.IsCommentsLoading = false
				pr.IsChangesRequestsLoading = false

				app.QueueUpdateDraw(func() {
					prt.redraw()
				})
			}(v)
		}

		state.RepositoryData[id].IsLoading = false
		app.QueueUpdateDraw(func() {
			prt.redraw()
		})

		// TODO: Only the first page of pull requests is fetched
		break
	}
}

func addEmptyRow(table *tview.Table, offset int) {
	for i := 0; i < len(headers); i++ {
		table.SetCell(
			offset,
			i,
			tview.NewTableCell(""),
		)
	}
}
func setRowStyle(table *tview.Table, offset int, style tcell.Style) {
	for i := 0; i < len(headers); i++ {
		table.GetCell(
			offset,
			i,
		).SetStyle(style)
	}
}

func (prt *pullRequestTable) drawTable() {
	headerStyle := tcell.StyleDefault.Bold(true)
	offset := 0

	keys := make([]string, len(state.RepositoryData))
	for k := range state.RepositoryData {
		keys = append(keys, k)
	}

	// TODO: Sort by recency instead from the state file
	sort.Strings(keys)

	for _, k := range keys {
		data, ok := state.RepositoryData[k]
		if !ok {
			continue
		}

		visible := false
		for _, pr := range data.PullRequests {
			if pr.Visible {
				visible = true
				break
			}
		}

		if !data.IsLoading && (len(data.PullRequests) == 0 || !visible) {
			continue
		}

		addEmptyRow(prt.View, offset)
		setRowStyle(prt.View, offset, headerStyle)
		prt.setRowSelectable(offset, false)
		prt.View.GetCell(offset, 0).SetText("REPO")
		prt.View.GetCell(offset, 1).SetText(data.Name)

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

		if data.IsLoading {
			addEmptyRow(prt.View, offset)
			prt.View.SetCell(offset, 0, tview.NewTableCell("Loading..."))
			prt.setRowSelectable(offset, false)
			offset += 1
			continue
		}

		if len(data.PullRequests) > 0 {
			keys := make([]string, len(data.PullRequests))
			for k := range data.PullRequests {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				pr, ok := data.PullRequests[k]
				if !ok || !pr.Visible {
					continue
				}

				pr.TableRowId = offset
				prt.addRow(pr.PullRequest, offset)
				if pr.PullRequest.State == client.PullRequestState_DECLINED {
					prt.updateRowStatus(offset, "Declined", DeclinedColor, false)
				} else if pr.PullRequest.State == client.PullRequestState_DECLINING {
					prt.updateRowStatus(offset, "Declining...", tcell.ColorDarkRed, false)
				} else if pr.PullRequest.State == client.PullRequestState_MERGED {
					prt.updateRowStatus(offset, "Merged", tcell.ColorLightYellow, false)
				} else if pr.PullRequest.State == client.PullRequestState_MERGING {
					prt.updateRowStatus(offset, "Merging...", tcell.ColorYellow, false)
				} else if pr.PullRequest.State == client.PullRequestState_APPROVING {
					prt.updateRowStatus(offset, "Approving...", tcell.ColorDarkOliveGreen, false)
				} else if pr.PullRequest.State == client.PullRequestState_APPROVED {
					prt.updateRowStatus(offset, "Approved", tcell.ColorGreen, true)
				} else if pr.Selected {
					prt.colorRow(offset, tcell.ColorPowderBlue)
				} else {
					prt.colorRow(offset, NormalColor)
				}

				if pr.IsApprovalsLoading {
					prt.View.GetCell(offset, 5).SetText("⏳")
				} else {
					if len(pr.PullRequest.Approvals) > 0 {
						prt.View.GetCell(offset, 5).SetText("✅")
					} else {
						prt.View.GetCell(offset, 5).SetText("")
					}
				}

				if pr.IsChangesRequestsLoading {
					prt.View.GetCell(offset, 6).SetText("⏳")
				} else {
					if len(pr.PullRequest.ChangesRequests) > 0 {
						prt.View.GetCell(offset, 6).SetText("⚠️")
					} else {
						prt.View.GetCell(offset, 6).SetText("")
					}
				}

				if pr.IsCommentsLoading {
					prt.View.GetCell(offset, 7).SetText("⏳")
				} else {
					prt.View.GetCell(offset, 7).SetText(fmt.Sprint(len(pr.PullRequest.Comments)))
				}

				offset++
			}
		}

		addEmptyRow(prt.View, offset)
		prt.setRowSelectable(offset, false)
		offset++
	}
}

func (prt *pullRequestTable) updateRowStatus(
	rowId int,
	text string,
	color tcell.Color,
	selectable bool,
) {
	prt.View.GetCell(rowId, 4).SetText(pad(text))
	prt.setRowSelectable(rowId, selectable)
	prt.colorRow(rowId, color)
}

func (prt *pullRequestTable) addRow(
	v *client.PullRequest,
	rowId int,
) {
	// Escape the title string
	v.Title = strings.ReplaceAll(v.Title, "]", "[]")

	values := []string{
		v.ID,
		v.Title,
		v.Source,
		v.Destination,
		"Open",
		"⏳",
		"⏳",
		"⏳",
	}

	for i := 0; i < len(values); i++ {
		prt.View.SetCell(rowId, i, tview.NewTableCell(pad(values[i])))
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
		// TODO: Log error?
	}
	pr.Selected = !pr.Selected

	prt.redraw()
}

func (prt *pullRequestTable) Filter(input string) {
	for _, data := range state.RepositoryData {
		for _, v := range data.PullRequests {
			v.Visible = strings.Contains(
				strings.ToLower(v.PullRequest.Title),
				strings.ToLower(input),
			)
		}
	}

	prt.redraw()
}

func (prt *pullRequestTable) resetFilter() {

}
