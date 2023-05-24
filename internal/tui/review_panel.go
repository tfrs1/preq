package tui

import (
	"fmt"
	"preq/internal/pkg/client"
	"strconv"
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
				Type:   DiffFileTypeModified,
				Hunks:  d.Hunks,
			}
		}
	}

	ct.pullRequest.IsCommentsLoading = true
}

func (ct *ReviewPanel) rerenderContent() {
	ct.prerenderContent(ct.currentDiffId)
}

func (ct *ReviewPanel) handleComment(comment *client.PullRequestComment) (int, error) {
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
			line := ct.content[len(ct.content)-1]
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
			ct.content = append(ct.content, &ScrollablePageLine{
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
					Content: fmt.Sprintf("[%s]%s%s %s", borderColor, verticalBorder, IconsMap["Working"], "Sending..."),
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
					Content: fmt.Sprintf("[%s]%s%s %s", borderColor, verticalBorder, IconsMap["Working"], "Deleting..."),
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

		ct.content = append(ct.content, &ScrollablePageLine{
			Reference:  comment,
			Statements: statements,
		})

		words := []string{}
		if comment.Deleted {
			words = []string{"[gray::s]This comment has been deleted.[-:-:-]"}
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
			ct.content = append(ct.content, &ScrollablePageLine{
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
		ct.content = append(ct.content, &ScrollablePageLine{
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

	return handleComment(comment, 0)
}

func (ct *ReviewPanel) renderStatusPage() {
	ct.Clear()

	ct.addLine("[::b]Description[::-]", nil)
	desc := ct.pullRequest.PullRequest.Description
	if len(desc) > 0 {
		for _, line := range strings.Split(desc, "\n") {
			ct.addLine(line, nil)
		}
	} else {
		ct.addLine("[gray::i]no description[-::-]", nil)
	}

	topLevelComments := []*client.PullRequestComment{}
	for _, c := range ct.pullRequest.PullRequest.Comments {
		if c.Type == client.CommentTypeGlobal {
			topLevelComments = append(topLevelComments, c)
		}
	}
	if len(topLevelComments) > 0 {
		ct.addLine("", nil)
		ct.addLine("[::b]Comments[::-]", nil)
		for _, c := range topLevelComments {
			ct.handleComment(c)
		}
	}

	outdatedComments := []*client.PullRequestComment{}
	// TODO: Add outdated member to comment struct
	// for _, c := range ct.pullRequest.PullRequest.Comments {
	// 	if c.Outdated {
	// 		outdateComments = append(outdateComments, c)
	// 	}
	// }
	// FIXME: Add outdated and file level comments together with diff (shortcut to show hide the comments)
	// collapse/expand inline comments in diff?
	if len(outdatedComments) > 0 {
		ct.addLine("", nil)
		ct.addLine("[::b]Outdated comments[::-]", nil)
		for _, c := range outdatedComments {
			ct.handleComment(c)
		}
	}
}

func (ct *ReviewPanel) prerenderContent(diffId string) {
	ct.Clear()
	ct.currentDiffId = diffId
	ct.currentDiff = ct.files[diffId]

	if ct.currentDiff == nil {
		return
	}

	ct.content = make([]*ScrollablePageLine, 0)

	d := ct.currentDiff
	comments := ct.commentMap[ct.currentDiffId]

	ct.addLine(fmt.Sprintf("[::b]File: %s[::-]", d.Title), nil)
	fileComments := []*client.PullRequestComment{}
	outdatedComments := []*client.PullRequestComment{}
	for _, comment := range ct.pullRequest.PullRequest.Comments {
		if comment.ParentID != "" {
			// Not a top level comment, skip
			continue
		}

		if comment.Type == client.CommentTypeFile {
			fileComments = append(fileComments, comment)
		} else if comment.IsOutdated(ct.pullRequest.PullRequest.Source.Hash) {
			outdatedComments = append(outdatedComments, comment)
		}
	}

	if len(fileComments) > 0 {
		ct.addLine("", nil)
		ct.addLine("[::b]File comments[::-]", nil)

		for _, comment := range fileComments {
			ct.handleComment(comment)
		}
	}

	if len(outdatedComments) > 0 {
		ct.addLine("", nil)
		ct.addLine("[::b]Outdated comments[::-]", nil)

		for _, comment := range outdatedComments {
			ct.handleComment(comment)
		}
	}

	ct.addLine("", nil)
	ct.addLine("[::b]Diff[::-]", nil)
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

			ct.content = append(ct.content, &ScrollablePageLine{
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
				b := 0
				n := 0
				b, _ = strconv.Atoi(oldLineNumber)
				n, _ = strconv.Atoi(newLineNumber)
				id := lineCommentListMapId(b, n)
				if c, ok := comments[id]; ok {
					// FIXME: Sort comments chronologically
					for _, comment := range c {
						if !comment.IsOutdated(ct.pullRequest.PullRequest.Source.Hash) {
							ct.handleComment(comment)
						}
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
			ct.content = append(ct.content, &ScrollablePageLine{
				Statements: []*ScrollablePageLineStatement{{Content: ""}},
			})
		}
	}
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
}
