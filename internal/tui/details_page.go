package tui

import (
	"errors"
	"fmt"
	"preq/internal/pkg/client"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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
}

func NewCommentsTable() *CommentsTable {
	return &CommentsTable{
		Box:        tview.NewBox(),
		pageOffset: 0,
	}
}

func (ct *CommentsTable) ScrollUp() {
	if !ct.disableScrollDown {
		ct.pageOffset++
	}
}

func (ct *CommentsTable) ScrollDown() {
	ct.pageOffset--

	// Should never scroll above the top line
	if ct.pageOffset < 0 {
		ct.pageOffset = 0
	}

	ct.disableScrollDown = false
}

func (ct *CommentsTable) SetData(pr *PullRequest) {
	ct.pullRequest = pr
	ct.loadingError = nil

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
	})()

	ct.diffs = diffs
}

type commentMap struct {
	RemovedLineComments map[uint]*client.PullRequestComment
	AddedLineComments   map[uint]*client.PullRequestComment
}

func (ct *CommentsTable) Draw(screen tcell.Screen) {
	ct.Box.DrawForSubclass(screen, ct)
	x, y, width, height := ct.GetInnerRect()

	if ct.pullRequest.IsCommentsLoading {
		tview.Print(screen, "Loading comments...", x, y, width, tview.AlignLeft, tcell.ColorWhite)
		return
	}

	if ct.loadingError != nil {
		tview.Print(screen, "Please pull", x, y, width, tview.AlignLeft, tcell.ColorWhite)
		return
	}

	filesMap := make(map[string]*commentMap)
	for _, prc := range ct.pullRequest.PullRequest.Comments {
		if filesMap[prc.FilePath] == nil {
			filesMap[prc.FilePath] = &commentMap{
				RemovedLineComments: make(map[uint]*client.PullRequestComment),
				AddedLineComments:   make(map[uint]*client.PullRequestComment),
			}
		}

		if prc.BeforeLineNumber != 0 {
			filesMap[prc.FilePath].RemovedLineComments[prc.BeforeLineNumber] = prc
		} else if prc.AfterLineNumber != 0 {
			filesMap[prc.FilePath].AddedLineComments[prc.AfterLineNumber] = prc
		}
	}

	index1 := 0

	// index := -ct.pageOffset

	prevIndent := 0
	printComment := func(comment *client.PullRequestComment, yHeight int, indent int) (int, error) {
		localIndex := 0
		innerWidth := width - indent

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

		trb := topRightBorder
		if indent > 0 {
			trb = topRightReplyBorder
		}

		if localIndex+yHeight >= 0 {
			tview.Print(
				screen,
				fmt.Sprintf(
					"%s%s%s%s",
					tlb,
					blbPrev,
					strings.Repeat(horizontalBorder, innerWidth-3),
					trb,
				),
				x+indent,
				yHeight+localIndex,
				innerWidth,
				tview.AlignLeft,
				tcell.ColorYellow,
			)
		}

		localIndex++
		if yHeight+localIndex >= height {
			return -1, errors.New("height reached")
		}

		if localIndex+yHeight >= 0 {
			tview.Print(
				screen,
				fmt.Sprintf("%s%s", verticalBorder, comment.User),
				x+indent,
				yHeight+localIndex,
				innerWidth,
				tview.AlignLeft,
				tcell.ColorYellow,
			)
			tview.Print(
				screen,
				fmt.Sprintf(
					"%s[%v]%s",
					comment.Created.Local().Format("2006-01-02 15:04:05"),
					"yellow",
					verticalBorder,
				),
				x+indent,
				yHeight+localIndex,
				innerWidth,
				tview.AlignRight,
				tcell.ColorWhite,
			)
		}

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
			localIndex++
			if yHeight+localIndex >= height {
				return -1, errors.New("height reached")
			}

			if localIndex+yHeight >= 0 {
				tview.Print(
					screen,
					fmt.Sprintf("%s%s", verticalBorder, line),
					x+indent,
					yHeight+localIndex,
					innerWidth,
					tview.AlignLeft,
					tcell.ColorYellow,
				)

				tview.Print(
					screen,
					verticalBorder,
					x+indent,
					yHeight+localIndex,
					innerWidth,
					tview.AlignRight,
					tcell.ColorYellow,
				)
			}
		}

		localIndex++
		if yHeight+localIndex >= height {
			return -1, errors.New("height reached")
		}

		if localIndex >= 0 {
			tview.Print(
				screen,
				fmt.Sprintf(
					"%s%s%s",
					bottomLeftBorder,
					strings.Repeat(horizontalBorder, innerWidth-2),
					bottomRightBorder,
				),
				x+indent,
				yHeight+localIndex,
				innerWidth,
				tview.AlignLeft,
				tcell.ColorYellow,
			)
		}

		prevIndent = indent

		return yHeight + localIndex, nil
	}

	var handleComment func(comment *client.PullRequestComment, y int, depth int) (int, error)
	handleComment = func(comment *client.PullRequestComment, y int, depth int) (int, error) {
		height, err := printComment(comment, y, depth)
		if err != nil {
			return -1, err
		}

		for _, prc := range ct.pullRequest.PullRequest.Comments {
			if prc.ParentID == comment.ID {
				height, err = handleComment(prc, height, depth+1)
				if err != nil {
					return -1, err
				}
				continue
			}
		}

		return height, nil
	}

	reachedEnd := true

	for _, d := range ct.diffs {
		filename := d.NewName
		if filename == "/dev/null" {
			filename = d.OrigName
		}

		comments := filesMap["Dockerfile"]

		tview.Print(screen, filename, x, y+index1, width, tview.AlignLeft, tcell.ColorWhite)
		index1++

		for _, h := range d.Hunks {
			origIdx := h.OrigStartLine
			newIdx := h.NewStartLine

			lines := strings.Split(string(h.Body), "\n")
			for _, line := range lines {
				isAddedLine := strings.HasPrefix(line, "+")
				isRemoveLine := strings.HasPrefix(line, "-")
				isCommonLine := strings.HasPrefix(line, " ")

				color := "white"
				oldLineNumber := fmt.Sprint(origIdx)
				if isAddedLine {
					oldLineNumber = strings.Repeat(" ", len(oldLineNumber))
					color = "green"
				}

				newLineNumber := fmt.Sprint(newIdx)
				if isRemoveLine {
					newLineNumber = strings.Repeat(" ", len(newLineNumber))
					color = "red"
				}

				output := fmt.Sprintf("%s %s│ [%s]", oldLineNumber, newLineNumber, color) + line
				tview.Print(
					screen,
					output,
					x,
					y+index1,
					width,
					tview.AlignLeft,
					tcell.ColorWhite,
				)

				index1++

				if comments != nil {
					var comment *client.PullRequestComment = nil
					if isAddedLine {
						n, err := strconv.Atoi(newLineNumber)
						if err == nil {
							c, ok := comments.AddedLineComments[uint(n)]
							if ok {
								comment = c
							}
						}
					} else {
						n, err := strconv.Atoi(oldLineNumber)
						if err == nil {
							c, ok := comments.RemovedLineComments[uint(n)]
							if ok {
								comment = c
							}
						}
					}

					if comment != nil {
						height, err := handleComment(comment, y+index1, 0)
						if err == nil {
							index1 = height - y
						}

						index1++
					}
				}

				if comments != nil && isAddedLine {
				}

				if isAddedLine || isCommonLine {
					newIdx++
				}

				if isRemoveLine || isCommonLine {
					origIdx++
				}
			}

			index1++
		}

		index1++
		tview.Print(screen, "", x, y+index1, width, tview.AlignLeft, tcell.ColorWhite)
		index1++
	}
	return

	topLevelComments := []*client.PullRequestComment{}
	for _, prc := range ct.pullRequest.PullRequest.Comments {
		if prc.ParentID == "" {
			topLevelComments = append(topLevelComments, prc)
		}
	}

	// If everything was printed it means there is nothing more to scroll
	// therefore the scroll can be disabled until scrolled in the other direction
	if reachedEnd {
		ct.disableScrollDown = true
	}
}

type detailsPage struct {
	View tview.Primitive
}

func newDetailsPage() *detailsPage {
	grid := tview.NewGrid().SetRows(5, 0).SetColumns(0)
	info := tview.NewFlex()

	title := tview.NewTextView()
	info.AddItem(title, 0, 1, false)
	info.SetTitle("Info").SetBorder(true)

	table := NewCommentsTable()
	table.
		SetBorder(false).
		SetBorder(true).
		SetTitle("Comments").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'j':
				table.ScrollUp()
				return nil
			case 'k':
				table.ScrollDown()
				return nil
			}

			return event
		})

	grid.AddItem(info, 0, 0, 1, 1, 1, 1, false)
	grid.AddItem(table, 1, 0, 1, 1, 1, 1, true)
	grid.
		SetBorder(true).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEscape:
				eventBus.Publish("detailsPage:close", nil)
			}

			switch event.Rune() {
			case 'o':
			case 'q':
				eventBus.Publish("detailsPage:close", nil)
			}

			return event
		})

	eventBus.Subscribe("detailsPage:open", func(input interface{}) {
		if pr, ok := input.(*PullRequest); ok {
			title.SetText(pr.PullRequest.ID)
			table.SetData(pr)
		} else {
			title.SetText("cast failed")
		}

	})

	return &detailsPage{
		View: grid,
	}
}
