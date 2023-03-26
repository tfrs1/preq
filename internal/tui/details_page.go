package tui

import (
	"fmt"
	"math"
	"preq/internal/pkg/client"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/go-diff/diff"
)

var (
	bottomLeftBorder         = "┗"
	bottomRightBorder        = "┛"
	topLeftBorder            = "┏"
	topLeftReplyBorder       = "┳"
	topRightBorder           = "┓"
	topRightReplyBorder      = "┫"
	topLeftSameLevelBorder   = "┣"
	horizontalBorder         = "━"
	verticalBorder           = "┃"
	bottomLeftPreviousBorder = "┻"
)

type CommentsTable struct {
	*tview.Box
	pullRequest       *PullRequest
	pageOffset        int
	disableScrollDown bool
	loadingError      error
	diffs             []*diff.FileDiff
	content           [][]*contentLineStatement
	IsLoading         bool
	width             int
	height            int
	files             map[string]*diffFile
}

const (
	DiffFileTypeAdded = iota
	DiffFileTypeRemoved
	DiffFileTypeRenamed
	DiffFileTypeUpdated
)

type diffFile struct {
	DiffId string
	Type   int
	Title  string
	Hunks  []*diff.Hunk
}

func NewCommentsTable() *CommentsTable {
	return &CommentsTable{
		Box:        tview.NewBox(),
		pageOffset: 0,
		IsLoading:  true,
	}
}

func (ct *CommentsTable) scrollDown(size int) {
	ct.pageOffset += size

	end := len(ct.content) - ct.height - 1
	if end < 0 {
		end = 0
	}

	// Should not scroll past the end of the content
	if ct.pageOffset > end {
		ct.pageOffset = end
	}
}

func (ct *CommentsTable) scrollUp(size int) {
	ct.pageOffset -= size

	// Should not scroll above the top line
	if ct.pageOffset < 0 {
		ct.pageOffset = 0
	}
}

func (ct *CommentsTable) ScrollDown() {
	ct.scrollDown(1)
}

func (ct *CommentsTable) ScrollHalfPageDown() {
	ct.scrollDown(ct.height / 2)
}

func (ct *CommentsTable) ScrollUp() {
	ct.scrollUp(1)
}

func (ct *CommentsTable) ScrollHalfPageUp() {
	ct.scrollUp(ct.height / 2)
}

func (ct *CommentsTable) makeId(diff *diff.FileDiff) string {
	id := diff.NewName[2:]
	if diff.NewName == "/dev/null" {
		id = diff.OrigName[2:]
	}

	return id
}

func (ct *CommentsTable) SetData(pr *PullRequest) {
	_, _, ct.width, ct.height = ct.GetInnerRect()

	ct.pullRequest = pr
	ct.loadingError = nil
	ct.IsLoading = true
	ct.content = make([][]*contentLineStatement, 0)
	ct.pageOffset = 0
	ct.files = make(map[string]*diffFile, 0)

	changes, err := ct.pullRequest.GitUtil.GetDiffPatch(
		ct.pullRequest.PullRequest.Destination.Hash,
		ct.pullRequest.PullRequest.Source.Hash,
	)

	if err != nil {
		ct.loadingError = err
		return
	}

	diffs, err := diff.ParseMultiFileDiff(changes)
	if err != nil {
		ct.loadingError = err
		return
	}
	ct.diffs = diffs

	for _, d := range diffs {
		newName := d.NewName[2:]
		oldName := d.OrigName[2:]

		id := ct.makeId(d)
		if d.OrigName == "/dev/null" {
			ct.files[id] = &diffFile{
				DiffId: id,
				Title:  newName,
				Type:   DiffFileTypeAdded,
				Hunks:  d.Hunks,
			}
		} else if d.NewName == "/dev/null" {
			ct.files[id] = &diffFile{
				DiffId: id,
				Title:  oldName,
				Type:   DiffFileTypeRemoved,
				Hunks:  d.Hunks,
			}
		} else if oldName != newName {
			ct.files[id] = &diffFile{
				DiffId: id,
				Title:  fmt.Sprintf("%s -> %s", oldName, newName),
				Type:   DiffFileTypeRenamed,
				Hunks:  d.Hunks,
			}
		} else {
			ct.files[id] = &diffFile{
				DiffId: id,
				Title:  newName,
				Type:   DiffFileTypeUpdated,
				Hunks:  d.Hunks,
			}
		}
	}

	ct.pullRequest.IsCommentsLoading = true
	go (func() {
		list, err := ct.pullRequest.Client.GetComments(&client.GetCommentsOptions{
			Repository: ct.pullRequest.Repository,
			ID:         ct.pullRequest.PullRequest.ID,
		})

		if err != nil {
			return
		}

		ct.pullRequest.PullRequest.Comments = list
		ct.pullRequest.IsCommentsLoading = false
		ct.IsLoading = false
		app.QueueUpdateDraw(func() {
			eventBus.Publish("DetailsPage:LoadingFinished", nil)
		})
	})()
}

func (ct *CommentsTable) prerenderContent(d *diffFile) {
	filesMap := make(map[string]*commentMap)
	for _, prc := range ct.pullRequest.PullRequest.Comments {
		if filesMap[prc.FilePath] == nil {
			filesMap[prc.FilePath] = &commentMap{
				RemovedLineComments: make(map[uint]*client.PullRequestComment),
				AddedLineComments:   make(map[uint]*client.PullRequestComment),
			}
		}

		if prc.BeforeLineNumber != 0 && prc.ParentID == "" {
			filesMap[prc.FilePath].RemovedLineComments[prc.BeforeLineNumber] = prc
		} else if prc.AfterLineNumber != 0 && prc.ParentID == "" {
			filesMap[prc.FilePath].AddedLineComments[prc.AfterLineNumber] = prc
		}
	}

	content := make([][]*contentLineStatement, 0)
	prevIndent := 0
	printComment := func(comment *client.PullRequestComment, indent int) error {
		innerWidth := ct.width - indent

		tlb := ""
		blbPrev := horizontalBorder
		if indent == 0 {
			tlb = topLeftBorder
		} else if indent < prevIndent {
			tlb = topLeftBorder
			blbPrev = bottomLeftPreviousBorder
		} else if indent == prevIndent {
			tlb = topLeftSameLevelBorder
		} else if indent > prevIndent {
			tlb = topLeftReplyBorder
		}

		if indent > 0 {
			statements := &content[len(content)-1]
			*statements = append(*statements,
				&contentLineStatement{
					Content: tlb + blbPrev,
					Indent:  indent,
				},
				&contentLineStatement{
					Content:   topRightReplyBorder,
					Alignment: tview.AlignRight,
				},
			)
		} else {
			trb := topRightBorder
			content = append(content, []*contentLineStatement{
				{
					Content: fmt.Sprintf(
						"%s%s%s%s",
						tlb,
						blbPrev,
						strings.Repeat(horizontalBorder, innerWidth-3),
						trb,
					),
					Alignment: tview.AlignLeft,
					Indent:    indent,
				},
			})
		}

		content = append(content, []*contentLineStatement{
			{
				Content: fmt.Sprintf("[white]%s%s", verticalBorder, comment.User),
				Indent:  indent,
			},
			{
				Content: fmt.Sprintf(
					"%s[%v]%s",
					comment.Created.Local().Format("2006-01-02 15:04:05"),
					"yellow",
					verticalBorder,
				),
				Alignment: tview.AlignRight,
				Indent:    0,
			},
		})

		words := strings.Split(comment.Content, " ")
		commentLines := []string{}
		line := []string{}
		for _, word := range words {
			lineLen := 0
			for _, w := range line {
				lineLen += len(w) + 1
			}

			if lineLen+len(word) > innerWidth-2 {
				commentLines = append(commentLines, strings.Join(line, " "))
				line = []string{}
			}

			line = append(line, word)
		}
		commentLines = append(commentLines, strings.Join(line, " "))
		for _, line := range commentLines {
			content = append(content, []*contentLineStatement{
				{
					Content: verticalBorder + line,
					Indent:  indent,
				},
				{
					Content:   verticalBorder,
					Alignment: tview.AlignRight,
				},
			})
		}

		content = append(content, []*contentLineStatement{
			{
				Content: fmt.Sprintf(
					"%s%s%s",
					bottomLeftBorder,
					strings.Repeat(horizontalBorder, innerWidth-2),
					bottomRightBorder,
				),
				Indent: indent,
			},
		})

		prevIndent = indent

		return nil
	}

	var handleComment func(comment *client.PullRequestComment, depth int) (int, error)
	handleComment = func(comment *client.PullRequestComment, depth int) (int, error) {
		err := printComment(comment, depth)
		if err != nil {
			return -1, err
		}

		for _, prc := range ct.pullRequest.PullRequest.Comments {
			if prc.ParentID == comment.ID {
				_, err = handleComment(prc, depth+1)
				if err != nil {
					return -1, err
				}
			}
		}

		return 0, nil
	}

	comments := filesMap[d.DiffId]

	content = append(content, []*contentLineStatement{
		{Content: d.Title},
	})

	for _, h := range d.Hunks {
		origIdx := h.OrigStartLine
		newIdx := h.NewStartLine

		lines := strings.Split(string(h.Body), "\n")

		maxOrigIdx := h.OrigStartLine
		maxNewIdx := h.NewStartLine
		for _, line := range lines {
			isAddedLine := strings.HasPrefix(line, "+")
			isRemoveLine := strings.HasPrefix(line, "-")
			isCommonLine := strings.HasPrefix(line, " ")

			if isAddedLine || isCommonLine {
				maxNewIdx++
			}

			if isRemoveLine || isCommonLine {
				maxOrigIdx++
			}
		}

		origIdxLen := len(fmt.Sprint(maxOrigIdx))
		newIdxLen := len(fmt.Sprint(maxNewIdx))
		for _, line := range lines {
			isAddedLine := strings.HasPrefix(line, "+")
			isRemoveLine := strings.HasPrefix(line, "-")
			isCommonLine := strings.HasPrefix(line, " ")

			color := "white"
			oldLineNumber := fmt.Sprint(origIdx)
			if isAddedLine {
				oldLineNumber = ""
				color = "green"
			}

			newLineNumber := fmt.Sprint(newIdx)
			if isRemoveLine {
				newLineNumber = ""
				color = "red"
			}

			content = append(content, []*contentLineStatement{
				{
					Content: fmt.Sprintf(
						"%*s %*s│ [%s]%s",
						origIdxLen,
						oldLineNumber,
						newIdxLen,
						newLineNumber,
						color,
						line,
					),
				},
			})

			if comments != nil {
				var comment *client.PullRequestComment = nil
				if isAddedLine {
					if n, err := strconv.Atoi(newLineNumber); err == nil {
						if c, ok := comments.AddedLineComments[uint(n)]; ok {
							comment = c
						}
					}
				} else {
					if n, err := strconv.Atoi(oldLineNumber); err == nil {
						if c, ok := comments.RemovedLineComments[uint(n)]; ok {
							comment = c
						}
					}
				}

				if comment != nil {
					handleComment(comment, 0)
				}
			}

			if isAddedLine || isCommonLine {
				newIdx++
			}

			if isRemoveLine || isCommonLine {
				origIdx++
			}
		}

		content = append(content, []*contentLineStatement{{Content: ""}})
	}

	ct.content = content
}

type commentMap struct {
	RemovedLineComments map[uint]*client.PullRequestComment
	AddedLineComments   map[uint]*client.PullRequestComment
}

type contentLineStatement struct {
	Indent    int
	Content   string
	Alignment int
}

func (ct *CommentsTable) Draw(screen tcell.Screen) {
	ct.Box.DrawForSubclass(screen, ct)
	x, y, width, height := ct.GetInnerRect()

	if ct.loadingError != nil {
		tview.Print(
			screen,
			"Could not find the commit hash locally. Please pull.",
			x,
			y,
			width,
			tview.AlignLeft,
			tcell.ColorWhite,
		)
		return
	}

	if ct.IsLoading {
		tview.Print(screen, "Loading...", x, y, width, tview.AlignLeft, tcell.ColorWhite)
		return
	}

	if len(ct.content) != 0 {
		ct.width = width
		ct.height = height
	}

	i := ct.pageOffset
	end := int(math.Min(float64(len(ct.content)), float64(i+height)))
	offset := 0
	for ; i < end; i++ {
		cl := ct.content[i]

		for _, s := range cl {
			tview.Print(
				screen,
				s.Content,
				x+s.Indent,
				y+offset,
				width,
				s.Alignment,
				tcell.ColorWhite,
			)
		}

		offset++

		tview.Print(screen, "", x, y+i, width, tview.AlignLeft, tcell.ColorWhite)
	}
	return

	topLevelComments := []*client.PullRequestComment{}
	for _, prc := range ct.pullRequest.PullRequest.Comments {
		if prc.ParentID == "" {
			topLevelComments = append(topLevelComments, prc)
		}
	}
}

type detailsPage struct {
	View tview.Primitive
}

func newDetailsPage() *detailsPage {
	grid := tview.NewGrid().SetRows(5, 0).SetColumns(0)
	info := tview.NewFlex()
	table := NewCommentsTable()

	filesTable := tview.NewTable()

	eventBus.Subscribe("DetailsPage:LoadingFinished", func(data interface{}) {
		row, _ := filesTable.GetSelection()
		if row >= filesTable.GetRowCount() {
			row = 0
			filesTable.Select(row, 0)
		}

		tableRow := filesTable.GetCell(row, 0)
		var ref *diffFile
		ref, ok := tableRow.GetReference().(*diffFile)
		if !ok {
			log.Error().Msg("changes table row reference")
			return
		}

		eventBus.Publish("DetailsPage:OnFileChanged", ref)
	})

	filesTable.SetSelectionChangedFunc(func(row, column int) {
		tableRow := filesTable.GetCell(row, 0)
		var ref *diffFile
		ref, ok := tableRow.GetReference().(*diffFile)
		if !ok {
			log.Error().Msg("changes table row reference")
			return
		}

		eventBus.Publish("DetailsPage:OnFileChanged", ref)
	})

	filesTable.SetSelectable(true, false).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				app.SetFocus(table)
				return nil
			}

			return event
		})

	info.AddItem(filesTable, 0, 1, true)
	info.SetTitle("Info").SetBorder(true)

	table.
		SetBorder(true).
		SetTitle("Comments").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'j':
				table.ScrollDown()
				return nil
			case 'k':
				table.ScrollUp()
				return nil
			}

			switch event.Key() {
			case tcell.KeyCtrlD:
				table.ScrollHalfPageDown()
				return nil
			case tcell.KeyCtrlU:
				table.ScrollHalfPageUp()
				return nil
			case tcell.KeyEsc:
				app.SetFocus(info)
				return nil
			}

			return event
		})

	grid.AddItem(info, 0, 0, 1, 1, 1, 1, true)
	grid.AddItem(table, 1, 0, 1, 1, 1, 1, true)
	grid.
		SetBorder(true).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEscape:
				if !info.HasFocus() {
					return event
				}

				eventBus.Publish("detailsPage:close", nil)
				return nil
			}

			switch event.Rune() {
			case 'o':
			case 'q':
				eventBus.Publish("detailsPage:close", nil)
				return nil
			}

			return event
		})

	eventBus.Subscribe("detailsPage:open", func(input interface{}) {
		pr, ok := input.(*PullRequest)
		if !ok {
			log.Error().Msg("cast failed when opening the details page")
		}

		table.SetData(pr)
		filesTable.Clear()

		typeText := ""
		row := 0
		for _, v := range table.files {
			switch v.Type {
			case DiffFileTypeAdded:
				typeText = "A"
			case DiffFileTypeRenamed:
				typeText = "R"
			case DiffFileTypeRemoved:
				typeText = "D"
			case DiffFileTypeUpdated:
				typeText = "U"
			}

			filesTable.SetCell(row, 0, tview.NewTableCell(typeText).SetReference(v))
			filesTable.SetCell(row, 1, tview.NewTableCell(v.Title))

			row++
		}

		app.SetFocus(table)
	})

	eventBus.Subscribe("DetailsPage:OnFileChanged", func(input interface{}) {
		fileDiff, ok := input.(*diffFile)

		if !ok {
			log.Error().Msg("cast failed when opening the details page")
			return
		}

		table.pageOffset = 0
		diff := table.files[fileDiff.DiffId]
		table.prerenderContent(diff)
	})

	return &detailsPage{
		View: grid,
	}
}
