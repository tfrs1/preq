package tui

import (
	"errors"
	"fmt"
	"preq/internal/pkg/client"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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
	ct.disableScrollDown = false
}

func (ct *CommentsTable) SetData(pr *PullRequest) {
	ct.pullRequest = pr
}

func (ct *CommentsTable) Draw(screen tcell.Screen) {
	ct.Box.DrawForSubclass(screen, ct)
	x, y, width, height := ct.GetInnerRect()
	if ct.pageOffset < 0 {
		ct.pageOffset = 0
	}

	topLevelComments := []*client.PullRequestComment{}
	for _, prc := range ct.pullRequest.PullRequest.Comments {
		if prc.ParentID == "" {
			topLevelComments = append(topLevelComments, prc)
		}
	}

	index := -ct.pageOffset
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

		trb := topRightBorder
		if indent > 0 {
			trb = topRightReplyBorder
		}

		if index >= 0 {
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
				y+index,
				innerWidth,
				tview.AlignLeft,
				tcell.ColorYellow,
			)
		}

		index++
		if index >= height {
			return errors.New("height reached")
		}

		if index >= 0 {
			tview.Print(
				screen,
				fmt.Sprintf("%s%s", verticalBorder, comment.User),
				x+indent,
				y+index,
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
				y+index,
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
			index++
			if index >= height {
				return errors.New("height reached")
			}

			if index >= 0 {
				tview.Print(
					screen,
					fmt.Sprintf("%s%s", verticalBorder, line),
					x+indent,
					y+index,
					innerWidth,
					tview.AlignLeft,
					tcell.ColorYellow,
				)

				tview.Print(
					screen,
					verticalBorder,
					x+indent,
					y+index,
					innerWidth,
					tview.AlignRight,
					tcell.ColorYellow,
				)
			}
		}

		index++
		if index >= height {
			return errors.New("height reached")
		}

		if index >= 0 {
			tview.Print(
				screen,
				fmt.Sprintf(
					"%s%s%s",
					bottomLeftBorder,
					strings.Repeat(horizontalBorder, innerWidth-2),
					bottomRightBorder,
				),
				x+indent,
				y+index,
				innerWidth,
				tview.AlignLeft,
				tcell.ColorYellow,
			)
		}

		prevIndent = indent

		return nil
	}

	var handleComment func(comment *client.PullRequestComment, depth int) error
	handleComment = func(comment *client.PullRequestComment, depth int) error {
		err := printComment(comment, depth)
		if err != nil {
			return err
		}

		for _, prc := range ct.pullRequest.PullRequest.Comments {
			if prc.ParentID == comment.ID {
				err := handleComment(prc, depth+1)
				if err != nil {
					return err
				}
				continue
			}
		}

		return nil
	}

	reachedEnd := true
	for _, comment := range topLevelComments {
		if index >= height {
			reachedEnd = false
			break
		}

		prevIndent = 0
		err := handleComment(comment, 0)
		if err != nil {
			reachedEnd = false
			break
		}

		index++
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
