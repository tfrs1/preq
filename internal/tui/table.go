package tui

import (
	"fmt"
	"preq/internal/gitutils"
	"preq/internal/pkg/client"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func escapeString(s string) string {
	return strings.ReplaceAll(s, "]", "[]")
}

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

// TODO: Use a string instead, then it can be configurable
var headers = []string{
	// TableHeaderId, TableHeaderTitle, "SOURCE", "DESTINATION", "STATUS", "APPROVED", "CHANGES REQUESTED", "COMMENTS",
	TableHeaderId, TableHeaderTitle, "🛫", "🛬", "📖", "✅", "✋", "💬",
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

func (prt *pullRequestTable) GetPullRequestList() []*PullRequest {
	list := []*PullRequest{}
	for _, data := range state.RepositoryData {
		for _, pr := range data.PullRequests {
			list = append(list, pr)
		}
	}

	return list
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
	for _, rd := range state.RepositoryData {
		for _, pr := range rd.PullRequests {
			if pr.Selected && pr.Visible {
				count++
			}
		}
	}
	return count
}

func (prt *pullRequestTable) GetSelectedRows() []*PullRequest {
	rows := []*PullRequest{}

	for _, rd := range state.RepositoryData {
		for _, pr := range rd.PullRequests {
			if pr.Selected && pr.Visible {
				rows = append(rows, pr)
			}
		}
	}

	if len(rows) == 0 {
		row, _ := table.View.GetSelection()
		r, err := table.GetPullRequest(row)
		if err == nil {
			rows = append(rows, r)
		}
	}

	return rows
}

func (prt *pullRequestTable) GetRowByGlobalID(id string) *PullRequest {
	for _, rd := range state.RepositoryData {
		for _, pr := range rd.PullRequests {
			if pr.PullRequest.URL == id {
				return pr
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
	for _, data := range prt.tableData {
		id := repoId(data.Repository)

		utilClient, _ := gitutils.GetRepo(data.Path)

		if _, ok := state.RepositoryData[id]; !ok {
			state.RepositoryData[id] = &RepositoryData{
				Name:         data.Repository.Name,
				IsLoading:    true,
				PullRequests: make(map[string]*PullRequest),
				GitUtil:      utilClient,
			}
		}

		go prt.loadPR(app, data)
	}
}

func repoId(repo *client.Repository) string {
	return fmt.Sprintf(
		"%v___%v",
		repo.Provider,
		repo.Name,
	)
}

func (prt *pullRequestTable) loadPR(app *tview.Application, data *tableRepoData) {
	// TODO: This load should be in table write code
	// TODO here just the state should be updater

	nextURL := ""
	for {
		prs, err := data.Client.GetPullRequests(&client.GetPullRequestsOptions{
			Repository: data.Repository,
			State:      client.PullRequestState_OPEN,
			Next:       nextURL,
		})
		if err != nil {
			app.QueueUpdateDraw(func() {
				prt.View.SetCell(0, 0,
					tview.
						NewTableCell(err.Error()).
						SetAlign(tview.AlignLeft),
				)
			})
			return
		}

		id := repoId(data.Repository)
		for _, v := range prs.Values {
			data.Values = append(data.Values, &pullRequestTableRow{
				pullRequest: v,
				selected:    false,
				visible:     true,
				client:      data.Client,
				repository:  data.Repository,
			})

			state.RepositoryData[id].PullRequests[v.ID] = &PullRequest{
				PullRequest:              v,
				Selected:                 false,
				Visible:                  true,
				Client:                   data.Client,
				Repository:               data.Repository,
				IsApprovalsLoading:       true,
				IsCommentsLoading:        true,
				IsChangesRequestsLoading: true,
				GitUtil:                  state.RepositoryData[id].GitUtil,
			}

			go func(v *client.PullRequest) {
				err := data.Client.FillMiscInfoAsync(
					data.Repository,
					v,
				)
				if err != nil {
					return
				}

				id := repoId(data.Repository)
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
		// prt.setRowSelectable(offset, false)
		prt.View.GetCell(offset, 0).SetText("REPO")
		prt.View.GetCell(offset, 1).SetText(data.Name)

		offset += 1

		for i := 0; i < len(headers); i++ {
			prt.View.SetCell(
				offset,
				i,
				tview.NewTableCell(headers[i]).
					// SetSelectable(false).
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
					// } else if pr.PullRequest.State == client.PullRequestState_APPROVING {
					// 	prt.updateRowStatus(offset, "⏳", tcell.ColorDarkOliveGreen, false)
					// } else if pr.PullRequest.State == client.PullRequestState_APPROVED {
					// 	prt.updateRowStatus(offset, "Approved", tcell.ColorGreen, true)
				} else if pr.Selected {
					prt.colorRow(offset, tcell.ColorPowderBlue)
				} else {
					prt.colorRow(offset, NormalColor)
				}

				approvalsText := ""
				if pr.IsApprovalsLoading {
					approvalsText = "⏳"
				} else if len(pr.PullRequest.Approvals) > 0 {
					approvalsText = fmt.Sprintf("[%s::]%d[-::]", "green", len(pr.PullRequest.Approvals))
				}
				prt.View.GetCell(offset, 5).SetText(approvalsText)

				changesRequestText := ""
				if pr.IsChangesRequestsLoading {
					changesRequestText = "⏳"
				} else if len(pr.PullRequest.ChangesRequests) > 0 {
					changesRequestText = fmt.Sprintf("[%s::]%d[-::]", "orange", len(pr.PullRequest.ChangesRequests))
				}
				prt.View.GetCell(offset, 6).SetText(changesRequestText)

				commentsText := ""
				if pr.IsCommentsLoading {
					commentsText = "⏳"
				} else if pr.PullRequest.CommentCount > 0 {
					commentsText = fmt.Sprint(pr.PullRequest.CommentCount)
				}
				prt.View.GetCell(offset, 7).SetText(commentsText)

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
	prt.View.GetCell(rowId, 4).SetText(text)
	prt.setRowSelectable(rowId, selectable)
	prt.colorRow(rowId, color)
}

func cropString(input string, max int) string {
	if max > 3 && len(input) > max {
		return input[:max-3] + "..."
	}

	return input
}

func (prt *pullRequestTable) addRow(
	v *client.PullRequest,
	rowId int,
) {
	maxLen := 40
	source := cropString(v.Source.Name, maxLen)
	destination := cropString(v.Destination.Name, maxLen)
	title := cropString(v.Title, maxLen)

	title = escapeString(title)

	values := []string{
		v.ID,
		title,
		source,
		destination,
		"Open",
		"⏳",
		"⏳",
		"⏳",
	}

	for i := 0; i < len(values); i++ {
		prt.View.SetCell(rowId, i, tview.NewTableCell(values[i]))
	}
}

func (prt *pullRequestTable) setRowSelectable(rowId int, selectable bool) {
	return
	// for i := 0; i < prt.View.GetColumnCount(); i++ {
	// 	prt.View.GetCell(rowId, i).SetSelectable(selectable)
	// }
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

	if pr != nil {
		pr.Selected = !pr.Selected
		prt.redraw()
	}
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
