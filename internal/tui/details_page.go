package tui

import (
	"fmt"
	"preq/internal/pkg/client"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
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

type DiffFileType int

const (
	DiffFileTypeAdded DiffFileType = iota
	DiffFileTypeRemoved
	DiffFileTypeRenamed
	DiffFileTypeModified
)

type detailsPage struct {
	*tview.Grid
	fileTree    *FileTree
	reviewPanel *ReviewPanel
	changes     []byte
	commentsMap map[string]map[string][]*client.PullRequestComment
}

func CommentLineNumberTypeToDiffLineType(d DiffLineType) client.CommentLineNumberType {
	var t client.CommentLineNumberType = client.OriginalLineNumber
	if d == DiffLineTypeAdded {
		t = client.NewLineNumber
	}

	return t
}

func newDetailsPage() *detailsPage {
	grid := tview.NewGrid().SetRows(0, 0).SetColumns(-2, -5)
	fileTree := NewFileTree()
	reviewPanel := NewReviewPanel()

	eventBus.Subscribe("DetailsPage:LoadingFinished", func(data interface{}) {
		n, err := fileTree.GetSelectedNode()
		if err == nil && n != nil {
			eventBus.Publish("DetailsPage:OnFileChanged", &FileTreeStatementReference{
				Node: n,
			})
		}
	})

	reviewPanel.
		SetBorder(true).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'c':
				ref := reviewPanel.GetSelectedReference()
				if ref != nil {
					eventBus.Publish(
						"DetailsPage:NewCommentRequested",
						reviewPanel.GetSelectedReference(),
					)
				}
				return nil
			case 'd':
				ref := reviewPanel.GetSelectedReference()
				if ref == nil {
					return nil
				}

				if comment, ok := ref.(*client.PullRequestComment); ok {
					eventBus.Publish(
						"DetailsPage:DeleteCommentRequested",
						comment,
					)
				} else {
					log.Debug().Msg("[DetailsPage] Delete request on line without a comment reference")
				}
			}

			switch event.Key() {
			case tcell.KeyEsc:
				app.SetFocus(fileTree)
				return nil
			}

			return event
		})

	grid.AddItem(fileTree, 0, 0, 3, 1, 0, 0, true)
	grid.AddItem(reviewPanel, 0, 1, 3, 1, 0, 0, true)
	grid.
		SetTitle("Review").
		SetBorder(true).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEscape:
				if !fileTree.HasFocus() {
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

	eventBus.Subscribe("FileTree:FileSelectionRequested", func(input interface{}) {
		app.SetFocus(reviewPanel)
	})

	eventBus.Subscribe("DeleteCommendModal:DeleteCancelled", func(_ interface{}) {
		app.SetFocus(reviewPanel)
	})

	eventBus.Subscribe("DeleteCommendModal:DeleteConfirmed", func(input interface{}) {
		comment, ok := input.(*client.PullRequestComment)
		if !ok {
			log.Debug().Msg("cast failed when confirming comment deletion")
			return
		}

		/**
		 * TODO: Update the state of the comment as deleteing and update the table
		 */
		comment.IsBeingDeleted = true

		go (func() {
			err := reviewPanel.pullRequest.Client.DeleteComment(&client.DeleteCommentOptions{
				Repository: reviewPanel.pullRequest.Repository,
				ID:         reviewPanel.pullRequest.PullRequest.ID,
				CommentID:  comment.ID,
			})
			if err != nil {
				log.Error().Err(err).Msgf("failed to delete comment %s", comment.ID)
			}

			comment.Deleted = true
			comment.IsBeingDeleted = false

			app.QueueUpdateDraw(reviewPanel.rerenderContent)
		})()

		reviewPanel.rerenderContent()
		app.SetFocus(reviewPanel)
	})

	// FIXME: Multiple comments on the same code line are not visible
	eventBus.Subscribe("AddCommentModal:ConfirmRequested", func(input interface{}) {
		// FIXME: Sent comment 7 times because I was spamming enter over the send button?!?
		content, ok := input.(string)
		if !ok {
			log.Error().Msg("cast failed when confirming the comment")
			return
		}

		ref := reviewPanel.GetSelectedReference()
		var options *client.CreateCommentOptions = nil
		switch ref.(type) {
		case *diffLine:
			if d, ok := ref.(*diffLine); ok && d != nil {
				options = &client.CreateCommentOptions{
					Repository: reviewPanel.pullRequest.Repository,
					ID:         reviewPanel.pullRequest.PullRequest.ID,
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
					Repository: reviewPanel.pullRequest.Repository,
					ID:         reviewPanel.pullRequest.PullRequest.ID,
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
			IsBeingStored:    true,
			IsBeingDeleted:   false,
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

		reviewPanel.pullRequest.PullRequest.Comments = append(
			reviewPanel.pullRequest.PullRequest.Comments,
			tempComment,
		)

		go func() {
			comment, err := reviewPanel.pullRequest.Client.CreateComment(options)
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
			tempComment.IsBeingStored = comment.IsBeingStored

			app.QueueUpdateDraw(reviewPanel.rerenderContent)
		}()

		reviewPanel.rerenderContent()
		eventBus.Publish("AddCommentModal:CloseRequested", nil)
	})

	eventBus.Subscribe("AddCommentModal:Closed", func(_ interface{}) {
		app.SetFocus(reviewPanel)
	})

	eventBus.Subscribe("DetailsPage:OnFileChanged", func(input interface{}) {
		reviewPanel.Clear()

		if input == nil {
			return
		}

		rr, ok := input.(*FileTreeStatementReference)
		if !ok {
			log.Error().Msg("cast failed to FileTreeStatementReference when opening the details page")
			return
		}

		if rr.Node.IsRoot() {
			reviewPanel.renderStatusPage()
			return
		}

		if rr.Diff == nil {
			return
		}

		fileDiff, ok := rr.Diff.(*diffFile)
		if !ok {
			log.Error().Msg("cast failed to diffFile when opening the details page")
			return
		}

		reviewPanel.prerenderContent(fileDiff.DiffId)
	})

	return &detailsPage{
		Grid:        grid,
		fileTree:    fileTree,
		reviewPanel: reviewPanel,
	}
}

func (dp *detailsPage) SetData(pr *PullRequest) error {
	dp.fileTree.Clear()
	dp.reviewPanel.Clear()

	changes, err := pr.GitUtil.GetDiffPatch(
		pr.PullRequest.Destination.Hash,
		pr.PullRequest.Source.Hash,
	)
	if err != nil {
		return err
	}
	dp.changes = changes
	dp.commentsMap = make(map[string]map[string][]*client.PullRequestComment)

	dp.reviewPanel.SetData(pr, dp.changes, nil)
	dp.reviewPanel.rerenderContent()

	go func() {
		list, err := pr.Client.GetComments(&client.GetCommentsOptions{
			Repository: pr.Repository,
			ID:         pr.PullRequest.ID,
		})
		if err != nil {
			return
		}

		pr.PullRequest.Comments = list
		pr.IsCommentsLoading = false

		dp.commentsMap = make(map[string]map[string][]*client.PullRequestComment)
		for _, prc := range pr.PullRequest.Comments {
			if dp.commentsMap[prc.FilePath] == nil {
				dp.commentsMap[prc.FilePath] = make(map[string][]*client.PullRequestComment)
			}

			if prc.ParentID == "" {
				id := lineCommentListMapId(int(prc.BeforeLineNumber), int(prc.AfterLineNumber))
				dp.commentsMap[prc.FilePath][id] = append(dp.commentsMap[prc.FilePath][id], prc)
			}
		}

		row := 0
		for f, v := range dp.reviewPanel.files {
			item := NewFileTreeItem(v.Title).SetReference(v)

			switch v.Type {
			case DiffFileTypeAdded:
				item.SetDecoration(fmt.Sprintf("[green::]%s", IconsMap["GitAdded"]))
			case DiffFileTypeRenamed:
				item.SetDecoration(fmt.Sprintf("[white::]%s", IconsMap["GitRenamed"]))
			case DiffFileTypeRemoved:
				item.SetDecoration(fmt.Sprintf("[red::]%s", IconsMap["GitRemoved"]))
			case DiffFileTypeModified:
				item.SetDecoration(fmt.Sprintf("[orange::]%s", IconsMap["GitModified"]))
			}

			if dp.commentsMap[f] != nil {
				item.SetHasComments(true)
			}

			dp.fileTree.AddFile(item)
			row++
		}

		app.QueueUpdateDraw(func() {
			dp.reviewPanel.SetData(pr, dp.changes, dp.commentsMap)
			dp.fileTree.Rerender()
			eventBus.Publish("DetailsPage:LoadingFinished", nil)
		})
	}()

	return nil
}
