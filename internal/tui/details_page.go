package tui

import (
	"fmt"
	"math"
	"preq/internal/pkg/client"
	"strings"
	"time"

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

// TODO: Change this code so it's one map with unique string ID (ol--nl, eg 1--21)
type commentMap struct {
	RemovedLineComments map[uint][]*client.PullRequestComment
	AddedLineComments   map[uint][]*client.PullRequestComment
}

type contentLine struct {
	Statements []*contentLineStatement
	Reference  interface{}
}

type contentLineStatement struct {
	Indent    int
	Content   string
	Alignment int
}

type DiffLineType int

const (
	DiffLineTypeAdded DiffLineType = iota
	DiffLineTypeRemoved
)

type diffLine struct {
	FilePath   string
	LineNumber int
	Type       DiffLineType
}

type CommentsTable struct {
	*tview.Box
	pullRequest       *PullRequest
	pageOffset        int
	selectedIndex     int
	disableScrollDown bool
	loadingError      error
	diffs             []*diff.FileDiff
	content           []*contentLine
	IsLoading         bool
	width             int
	height            int
	files             map[string]*diffFile
	currentDiff       *diffFile
}

type DiffFileType int

const (
	DiffFileTypeAdded DiffFileType = iota
	DiffFileTypeRemoved
	DiffFileTypeRenamed
	DiffFileTypeUpdated
)

type diffFile struct {
	DiffId string
	Type   DiffFileType
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

func (ct *CommentsTable) moveSelected(size int) {
	ct.selectedIndex += size

	// Should not scroll past the end of the content
	end := len(ct.content) - 1
	if ct.selectedIndex > end {
		ct.selectedIndex = end
	}

	if ct.selectedIndex < 0 {
		ct.selectedIndex = 0
	}
}

func (ct *CommentsTable) scroll(size int) {
	ct.pageOffset += size

	end := len(ct.content) - ct.height
	if end < 0 {
		end = 0
	}

	// Should not scroll past the end of the content
	if ct.pageOffset > end {
		ct.pageOffset = end
	}

	if ct.pageOffset < 0 {
		ct.pageOffset = 0
	}
}

func (ct *CommentsTable) ScrollDown() {
	if (ct.pageOffset+ct.height)-ct.selectedIndex <= 4 {
		ct.scroll(1)
	}

	ct.moveSelected(1)
}

func (ct *CommentsTable) ScrollHalfPageDown() {
	ct.scroll(ct.height / 2)
	ct.moveSelected(ct.height / 2)
}

func (ct *CommentsTable) ScrollUp() {
	if ct.selectedIndex-ct.pageOffset <= 4 {
		ct.scroll(-1)
	}

	ct.moveSelected(-1)
}

func (ct *CommentsTable) ScrollHalfPageUp() {
	ct.scroll(-ct.height / 2)
	ct.moveSelected(-ct.height / 2)
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
	ct.content = make([]*contentLine, 0)
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
func (ct *CommentsTable) rerenderContent() {
	ct.prerenderContent(ct.currentDiff)
}

func (ct *CommentsTable) prerenderContent(d *diffFile) {
	ct.currentDiff = d
	filesMap := make(map[string]*commentMap)
	for _, prc := range ct.pullRequest.PullRequest.Comments {
		if filesMap[prc.FilePath] == nil {
			filesMap[prc.FilePath] = &commentMap{
				RemovedLineComments: make(map[uint][]*client.PullRequestComment),
				AddedLineComments:   make(map[uint][]*client.PullRequestComment),
			}
		}

		if prc.BeforeLineNumber != 0 && prc.ParentID == "" {
			_, ok := filesMap[prc.FilePath].RemovedLineComments[prc.BeforeLineNumber]
			if !ok {
				filesMap[prc.FilePath].RemovedLineComments[prc.BeforeLineNumber] = make(
					[]*client.PullRequestComment,
					0,
				)
			}

			filesMap[prc.FilePath].RemovedLineComments[prc.BeforeLineNumber] = append(
				filesMap[prc.FilePath].RemovedLineComments[prc.BeforeLineNumber],
				prc,
			)

		} else if prc.AfterLineNumber != 0 && prc.ParentID == "" {
			_, ok := filesMap[prc.FilePath].AddedLineComments[prc.AfterLineNumber]
			if !ok {
				filesMap[prc.FilePath].AddedLineComments[prc.AfterLineNumber] = make([]*client.PullRequestComment, 0)
			}

			filesMap[prc.FilePath].AddedLineComments[prc.AfterLineNumber] = append(
				filesMap[prc.FilePath].AddedLineComments[prc.AfterLineNumber],
				prc,
			)
		}
	}

	content := make([]*contentLine, 0)
	prevIndent := 0
	printComment := func(comment *client.PullRequestComment, indent int) error {
		commentBoxWidth := ct.width - indent

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
			line := content[len(content)-1]
			line.Reference = nil
			line.Statements = append(line.Statements,
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
			padding := ""
			if commentBoxWidth > 3 {
				padding = strings.Repeat(horizontalBorder, commentBoxWidth-3)
			}
			content = append(content, &contentLine{
				Reference: comment,
				Statements: []*contentLineStatement{
					{
						Content:   fmt.Sprintf("%s%s%s%s", tlb, blbPrev, padding, trb),
						Alignment: tview.AlignLeft,
						Indent:    indent,
					},
				},
			})
		}

		borderColor := "white"

		statements := []*contentLineStatement{
			{
				Content: fmt.Sprintf("[%s]%s⏳ %s", borderColor, verticalBorder, "Sending..."),
				Indent:  indent,
			},
			{
				Content:   fmt.Sprintf("[%v]%s", borderColor, verticalBorder),
				Alignment: tview.AlignRight,
			},
		}

		if !comment.IsSending {
			statements = []*contentLineStatement{
				{
					Content: fmt.Sprintf("[%s]%s%s", borderColor, verticalBorder, comment.User),
					Indent:  indent,
				},
				{
					Content: fmt.Sprintf(
						"%s[%s]%s",
						comment.Created.Local().Format("2006-01-02 15:04:05"),
						borderColor,
						verticalBorder,
					),
					Alignment: tview.AlignRight,
					Indent:    0,
				},
			}
		}

		content = append(content, &contentLine{
			Reference:  comment,
			Statements: statements,
		})

		words := strings.Split(comment.Content, " ")
		commentLines := []string{}
		line := []string{}
		for _, word := range words {
			lineLen := 0
			for _, w := range line {
				lineLen += len(w) + 1
			}

			if lineLen+len(word) > commentBoxWidth-2 {
				commentLines = append(commentLines, strings.Join(line, " "))
				line = []string{}
			}

			line = append(line, word)
		}
		commentLines = append(commentLines, strings.Join(line, " "))
		for _, line := range commentLines {
			content = append(content, &contentLine{
				Reference: comment,
				Statements: []*contentLineStatement{
					{
						Content: verticalBorder + line,
						Indent:  indent,
					},
					{
						Content:   verticalBorder,
						Alignment: tview.AlignRight,
					},
				}})
		}

		padding := ""
		if commentBoxWidth > 2 {
			padding = strings.Repeat(horizontalBorder, commentBoxWidth-2)
		}
		content = append(content, &contentLine{
			Reference: comment,
			Statements: []*contentLineStatement{
				{
					Content: fmt.Sprintf("%s%s%s", bottomLeftBorder, padding, bottomRightBorder),
					Indent:  indent,
				}},
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

	content = append(content, &contentLine{
		Statements: []*contentLineStatement{{Content: d.Title}},
	})

	for i, h := range d.Hunks {
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
			diffLineType := DiffLineTypeAdded
			if isAddedLine {
				oldLineNumber = ""
				color = "green"
			}

			newLineNumber := fmt.Sprint(newIdx)
			if isRemoveLine {
				diffLineType = DiffLineTypeRemoved
				newLineNumber = ""
				color = "red"
			}

			lineNumber := origIdx
			if isAddedLine {
				lineNumber = newIdx
			}

			content = append(content, &contentLine{
				Reference: &diffLine{
					FilePath:   d.DiffId,
					LineNumber: int(lineNumber),
					Type:       diffLineType,
				},
				Statements: []*contentLineStatement{
					{Content: fmt.Sprintf(
						"%*s %*s│ [%s]%s",
						origIdxLen,
						oldLineNumber,
						newIdxLen,
						newLineNumber,
						color,
						line,
					)},
				},
			})

			if comments != nil {
				cm := comments.RemovedLineComments
				if isAddedLine {
					cm = comments.AddedLineComments
				}

				if c, ok := cm[uint(lineNumber)]; ok {
					for _, prc := range c {
						handleComment(prc, 0)
					}
				}

			}

			if isAddedLine || isCommonLine {
				newIdx++
			}

			if isRemoveLine || isCommonLine {
				origIdx++
			}
		}

		if i < len(d.Hunks)-1 {
			content = append(content, &contentLine{
				Statements: []*contentLineStatement{{Content: ""}},
			})
		}
	}

	ct.content = content
}

func (ct *CommentsTable) GetSelectedReference() interface{} {
	if ct.selectedIndex > len(ct.content) {
		return nil
	}

	v := ct.content[ct.selectedIndex]
	if v == nil {
		return nil
	}

	return v.Reference
}

func (ct *CommentsTable) Draw(screen tcell.Screen) {
	ct.Box.DrawForSubclass(screen, ct)
	x, y, width, height := ct.GetInnerRect()
	ct.width = width
	ct.height = height

	if ct.loadingError != nil {
		tview.Print(
			screen,
			"Could not find the commit hash locally. Please pull.",
			x,
			y,
			ct.width,
			tview.AlignLeft,
			tcell.ColorWhite,
		)
		return
	}

	if ct.IsLoading {
		tview.Print(screen, "Loading...", x, y, ct.width, tview.AlignLeft, tcell.ColorWhite)
		return
	}

	i := ct.pageOffset
	end := int(math.Min(float64(len(ct.content)), float64(i+ct.height)))
	offset := 0
	for ; i < end; i++ {
		cl := ct.content[i]
		colorPrefix := ""
		if i == ct.selectedIndex {
			colorPrefix = "[:gray]"
		}

		tview.Print(
			screen,
			colorPrefix+strings.Repeat(" ", ct.width),
			x,
			y+offset,
			ct.width,
			tview.AlignRight,
			tcell.ColorWhite,
		)

		for _, s := range cl.Statements {
			tview.Print(
				screen,
				colorPrefix+s.Content,
				x+s.Indent,
				y+offset,
				ct.width,
				s.Alignment,
				tcell.ColorWhite,
			)
		}

		offset++

		tview.Print(screen, "", x, y+i, ct.width, tview.AlignLeft, tcell.ColorWhite)
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

func CommentLineNumberTypeToDiffLineType(d DiffLineType) client.CommentLineNumberType {
	var t client.CommentLineNumberType = client.OriginalLineNumber
	if d == DiffLineTypeAdded {
		t = client.NewLineNumber
	}

	return t
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
			case 'c':
				ref := table.GetSelectedReference()
				if ref != nil {
					eventBus.Publish(
						"DetailsPage:NewCommentRequested",
						table.GetSelectedReference(),
					)
				}
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

	// FIXME: Multiple comments on the same code line are not visible

	eventBus.Subscribe("AddCommentModal:ConfirmRequested", func(input interface{}) {
		// FIXME: Sent comment 7 times because I was spamming enter over the send button?!?
		content, ok := input.(string)
		if !ok {
			log.Error().Msg("cast failed when confirming the comment")
			return
		}

		ref := table.GetSelectedReference()
		var options *client.CreateCommentOptions = nil
		switch ref.(type) {
		case *diffLine:
			if d, ok := ref.(*diffLine); ok && d != nil {
				options = &client.CreateCommentOptions{
					Repository: table.pullRequest.Repository,
					ID:         table.pullRequest.PullRequest.ID,
					Content:    content,
					FilePath:   d.FilePath,
					LineRef: &client.CreateCommentOptionsLineRef{
						LineNumber: d.LineNumber,
						Type:       CommentLineNumberTypeToDiffLineType(d.Type),
					},
				}
			}
		case *client.PullRequestComment:
			if c, ok := ref.(*client.PullRequestComment); ok && c != nil {
				options = &client.CreateCommentOptions{
					Repository: table.pullRequest.Repository,
					ID:         table.pullRequest.PullRequest.ID,
					Content:    content,
					FilePath:   c.FilePath,
					ParentRef: &client.CreateCommentOptionsParentRef{
						ID: c.ID,
					},
				}
			}
		}

		parentId := ""
		if options.ParentRef != nil {
			parentId = options.ParentRef.ID
		}

		beforeLineNumber := 0
		afterLineNumber := 0

		if options.LineRef != nil {
			if options.LineRef.Type == client.OriginalLineNumber {
				beforeLineNumber = options.LineRef.LineNumber
			} else {
				afterLineNumber = options.LineRef.LineNumber
			}
		}

		tempComment := &client.PullRequestComment{
			ID:               fmt.Sprintf("%d", time.Now().UnixNano()),
			IsSending:        true,
			Created:          time.Now(),
			Updated:          time.Now(),
			Deleted:          false,
			User:             "",
			Content:          content,
			ParentID:         parentId,
			BeforeLineNumber: uint(beforeLineNumber),
			AfterLineNumber:  uint(afterLineNumber),
			FilePath:         options.FilePath,
		}

		table.pullRequest.PullRequest.Comments = append(
			table.pullRequest.PullRequest.Comments,
			tempComment,
		)

		go func() {
			comment, err := table.pullRequest.Client.CreateComment(options)
			if err != nil {
				log.Error().Err(err).Msg("failed to create comment")
				return
			}

			tempComment.ID = comment.ID
			tempComment.Created = comment.Created
			tempComment.Updated = comment.Updated
			tempComment.Deleted = comment.Deleted
			tempComment.User = comment.User
			tempComment.Content = comment.Content
			tempComment.ParentID = comment.ParentID
			tempComment.BeforeLineNumber = comment.BeforeLineNumber
			tempComment.AfterLineNumber = comment.AfterLineNumber
			tempComment.FilePath = comment.FilePath
			tempComment.IsSending = comment.IsSending

			app.QueueUpdateDraw(table.rerenderContent)
		}()

		table.rerenderContent()
		eventBus.Publish("AddCommentModal:CloseRequested", nil)
	})

	eventBus.Subscribe("AddCommentModal:Closed", func(_ interface{}) {
		app.SetFocus(table)
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
		table.selectedIndex = 0
		diff := table.files[fileDiff.DiffId]
		table.prerenderContent(diff)
	})

	return &detailsPage{
		View: grid,
	}
}
