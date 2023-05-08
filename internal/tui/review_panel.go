package tui

import (
	"fmt"
	"preq/internal/pkg/client"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sourcegraph/go-diff/diff"
)

type diffFile struct {
	DiffId string
	Type   DiffFileType
	Title  string
	Hunks  []*diff.Hunk
}

type lineCommentListMap map[string][]*client.PullRequestComment

func lineCommentListMapId(before, after int) string {
	if before != 0 {
		after = 0
	} else {
		before = 0
	}

	return fmt.Sprintf("%d___%d", before, after)
}

type ReviewPanel struct {
	*ScrollablePage
	pullRequest   *PullRequest
	loadingError  error
	diffs         []*diff.FileDiff
	IsLoading     bool
	files         map[string]*diffFile
	currentDiff   *diffFile
	currentDiffId string
	commentMap    map[string]map[string][]*client.PullRequestComment
}

func NewReviewPanel() *ReviewPanel {
	return &ReviewPanel{
		ScrollablePage: NewScrollablePage(),
		IsLoading:      true,
	}
}

func (ct *ReviewPanel) makeId(diff *diff.FileDiff) string {
	id := diff.NewName[2:]
	if diff.NewName == "/dev/null" {
		id = diff.OrigName[2:]
	}

	return id
}

func (ct *ReviewPanel) SetData(pr *PullRequest, changes []byte, commentsMap map[string]map[string][]*client.PullRequestComment) {
	_, _, ct.width, ct.height = ct.GetInnerRect()

	ct.Clear()
	ct.pullRequest = pr
	ct.loadingError = nil
	ct.IsLoading = false
	ct.files = make(map[string]*diffFile, 0)
	ct.commentMap = commentsMap

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
}

func (ct *ReviewPanel) rerenderContent() {
	ct.prerenderContent(ct.currentDiffId)
}

func (ct *ReviewPanel) prerenderContent(diffId string) {
	ct.Clear()
	ct.currentDiffId = diffId
	ct.currentDiff = ct.files[diffId]

	if ct.currentDiff == nil {
		return
	}

	content := make([]*ScrollablePageLine, 0)
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
				&ScrollablePageLineStatement{
					Content: tlb + blbPrev,
					Indent:  indent,
				},
				&ScrollablePageLineStatement{
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
			content = append(content, &ScrollablePageLine{
				Reference: comment,
				Statements: []*ScrollablePageLineStatement{
					{
						Content:   fmt.Sprintf("%s%s%s%s", tlb, blbPrev, padding, trb),
						Alignment: tview.AlignLeft,
						Indent:    indent,
					},
				},
			})
		}

		borderColor := "white"

		statements := []*ScrollablePageLineStatement{}
		if comment.IsBeingStored {
			statements = []*ScrollablePageLineStatement{
				{
					Content: fmt.Sprintf("[%s]%s⏳ %s", borderColor, verticalBorder, "Sending..."),
					Indent:  indent,
				},
				{
					Content:   fmt.Sprintf("[%v]%s", borderColor, verticalBorder),
					Alignment: tview.AlignRight,
				},
			}
		} else if comment.IsBeingDeleted {
			statements = []*ScrollablePageLineStatement{
				{
					Content: fmt.Sprintf("[%s]%s⏳ %s", borderColor, verticalBorder, "Deleting..."),
					Indent:  indent,
				},
				{
					Content:   fmt.Sprintf("[%v]%s", borderColor, verticalBorder),
					Alignment: tview.AlignRight,
				},
			}
		} else {
			statements = []*ScrollablePageLineStatement{
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

		content = append(content, &ScrollablePageLine{
			Reference:  comment,
			Statements: statements,
		})

		words := []string{}
		if comment.Deleted {
			words = []string{"[gray::s]This comment has been deleted.[::]"}
		} else {
			words = strings.Split(comment.Content, " ")
		}
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
			content = append(content, &ScrollablePageLine{
				Reference: comment,
				Statements: []*ScrollablePageLineStatement{
					{
						Content: verticalBorder + line,
						Indent:  indent,
					},
					{
						Content:   verticalBorder,
						Alignment: tview.AlignRight,
					},
				},
			})
		}

		padding := ""
		if commentBoxWidth > 2 {
			padding = strings.Repeat(horizontalBorder, commentBoxWidth-2)
		}
		content = append(content, &ScrollablePageLine{
			Reference: comment,
			Statements: []*ScrollablePageLineStatement{
				{
					Content: fmt.Sprintf("%s%s%s", bottomLeftBorder, padding, bottomRightBorder),
					Indent:  indent,
				},
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

	d := ct.currentDiff
	comments := ct.commentMap[ct.currentDiffId]

	content = append(content, &ScrollablePageLine{
		Statements: []*ScrollablePageLineStatement{{Content: d.Title}},
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

			content = append(content, &ScrollablePageLine{
				Reference: &diffLine{
					FilePath:   d.DiffId,
					LineNumber: int(lineNumber),
					Type:       diffLineType,
				},
				Statements: []*ScrollablePageLineStatement{
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
				id := lineCommentListMapId(int(origIdx), int(newIdx))
				if c, ok := comments[id]; ok {
					// FIXME: Sort comments chronologically
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
			content = append(content, &ScrollablePageLine{
				Statements: []*ScrollablePageLineStatement{{Content: ""}},
			})
		}
	}

	ct.content = content
}

func (ct *ReviewPanel) Draw(screen tcell.Screen) {
	ct.DrawForSubclass(screen, ct.ScrollablePage)

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
			tview.Styles.PrimaryTextColor,
		)
		return
	}

	if ct.IsLoading {
		tview.Print(screen, "Loading...", x, y, ct.width, tview.AlignLeft, tcell.ColorWhite)
		return
	}

	ct.ScrollablePage.Draw(screen)

	return

	topLevelComments := []*client.PullRequestComment{}
	for _, prc := range ct.pullRequest.PullRequest.Comments {
		if prc.ParentID == "" {
			topLevelComments = append(topLevelComments, prc)
		}
	}
}
