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
		app.QueueUpdateDraw(func() {})
	})()

	ct.diffs = diffs
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

		if prc.BeforeLineNumber != 0 && prc.ParentID == "" {
			filesMap[prc.FilePath].RemovedLineComments[prc.BeforeLineNumber] = prc
		} else if prc.AfterLineNumber != 0 && prc.ParentID == "" {
			filesMap[prc.FilePath].AddedLineComments[prc.AfterLineNumber] = prc
		}
	}

	// index := -ct.pageOffset

	content := make([][]*contentLineStatement, 0)
	prevIndent := 0
	printComment := func(comment *client.PullRequestComment, indent int) error {
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

	reachedEnd := true

	for _, d := range ct.diffs {
		filename := d.NewName
		if filename == "/dev/null" {
			filename = d.OrigName
		}

		comments := filesMap["Dockerfile"]

		content = append(content, []*contentLineStatement{
			{Content: filename},
		})

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
				content = append(content, []*contentLineStatement{
					{Content: output},
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
		}

		i := 0
		for i, cl := range content {
			if i >= height {
				break
			}

			for _, s := range cl {
				tview.Print(
					screen,
					s.Content,
					x+s.Indent,
					y+i,
					width,
					s.Alignment,
					tcell.ColorWhite,
				)
			}
		}

		tview.Print(screen, "", x, y+i, width, tview.AlignLeft, tcell.ColorWhite)
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
