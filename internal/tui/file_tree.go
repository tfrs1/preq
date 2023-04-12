package tui

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rs/zerolog/log"
)

// TODO: Add NerdFont icons

type FileTree struct {
	*ScrollablePage
	fileList []*FileTreeItem
	root     *FileTreeNode
}

func NewFileTree() *FileTree {
	info := &FileTree{
		ScrollablePage: NewScrollablePage(),
		fileList:       []*FileTreeItem{},
	}
	info.SetBorder(true).SetTitle("Files")
	info.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case ' ':
				eventBus.Publish("FileTree:ToggleExpandCollapseRequest", info.selectedIndex)
				return nil
			}
		case tcell.KeyEnter:
			eventBus.Publish(
				"FileTree:FileSelectionRequested",
				info.selectedIndex,
			)
			return nil
		}

		return event
	})

	info.SetSelectionChangedFunc(func(index int) {
		eventBus.Publish(
			"DetailsPage:OnFileChanged",
			info.GetSelectedReference(),
		)
	})

	eventBus.Subscribe("FileTree:ToggleExpandCollapseRequest", func(input interface{}) {
		ref := info.GetSelectedReference()
		if ref == nil {
			return
		}

		node, ok := ref.(*FileTreeStatementReference)
		if !ok {
			log.Error().Msg("cast to FileTreeStatementReference failed")
			return
		}

		node.Node.Collapsed = !node.Node.Collapsed
	})

	return info
}

func (ft FileTree) Draw(screen tcell.Screen) {
	ft.DrawForSubclass(screen, ft.ScrollablePage)

	ft.content = ft.root.rebuildStatements()

	ft.ScrollablePage.Draw(screen)
}

func (ft *FileTree) Clear() {
	ft.ScrollablePage.Clear()
	ft.fileList = []*FileTreeItem{}
	ft.root = nil
}

func (ft *FileTree) AddFile(file *FileTreeItem) *FileTree {
	ft.fileList = append(ft.fileList, file)
	ft.root = FilesToTree(ft.fileList)
	return ft
}

type FileTreeItem struct {
	Filename    string
	Decoration  string
	reference   interface{}
	hasComments bool
}

func NewFileTreeItem(filename string) *FileTreeItem {
	return &FileTreeItem{
		Filename: filename,
	}
}

func (fti *FileTreeItem) SetDecoration(decoration string) *FileTreeItem {
	fti.Decoration = decoration
	return fti
}

func (fti *FileTreeItem) SetHasComments(value bool) *FileTreeItem {
	fti.hasComments = value
	return fti
}

func (fti *FileTreeItem) SetReference(ref interface{}) *FileTreeItem {
	fti.reference = ref
	return fti
}

func (fti *FileTreeItem) GetReference() interface{} {
	return fti.reference
}

type FileTreeNode struct {
	Filename          string
	Children          []*FileTreeNode
	Collapsed         bool
	Decoration        string
	GlobalDecorations []string
	reference         interface{}
}

func FilesToTree(items []*FileTreeItem) *FileTreeNode {
	findNode := func(input []*FileTreeNode, name string) *FileTreeNode {
		for _, ftn := range input {
			if ftn.Filename == name {
				return ftn
			}
		}
		return nil
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Filename < items[j].Filename
	})

	root := &FileTreeNode{Filename: "root", Collapsed: false}

	for _, item := range items {
		currentNode := root
		for _, v := range strings.Split(item.Filename, "/") {
			if v == "" {
				continue
			}

			globalDecorations := []string{}
			if item.hasComments {
				globalDecorations = append(globalDecorations, "[green::b]💬")
			}

			child := findNode(currentNode.Children, v)
			if child == nil {
				child = &FileTreeNode{
					Filename:          v,
					Collapsed:         false,
					GlobalDecorations: globalDecorations,
					reference:         item.reference,
					Decoration:        item.Decoration,
				}

				currentNode.Children = append(currentNode.Children, child)
				currentNode.reference = nil
				currentNode.GlobalDecorations = []string{}
			}

			currentNode = child
		}
	}

	var traverse func(node *FileTreeNode)
	traverse = func(node *FileTreeNode) {
		if len(node.Children) == 1 && len(node.Children[0].Children) > 0 {
			node.Filename = path.Join(node.Filename, node.Children[0].Filename)
			node.Children = node.Children[0].Children
			node.reference = nil

			traverse(node)
		} else {
			for _, child := range node.Children {
				traverse(child)
			}
		}
	}

	traverse(root)

	return root
}

type FileTreeStatementReference struct {
	Node *FileTreeNode
	Diff interface{}
}

func (node *FileTreeNode) dfs(level int, callback func(node *FileTreeNode)) {
	if node == nil {
		return
	}

	callback(node)

	for _, child := range node.Children {
		child.dfs(level+1, callback)
	}
}

func (node *FileTreeNode) rebuildStatements() []*ScrollablePageLine {
	statements := []*ScrollablePageLine{}
	maxDecorations := 0
	node.dfs(0, func(node *FileTreeNode) {
		if l := len(node.GlobalDecorations); l > maxDecorations {
			maxDecorations = l
		}
	})

	var recurse func(node *FileTreeNode, level int)
	recurse = func(node *FileTreeNode, level int) {
		if node == nil {
			return
		}

		indent := strings.Repeat(" ", level+maxDecorations)
		icon := " "
		decoration := ""

		if len(node.Children) > 0 {
			// TODO: Add nerd font icons
			// decoration = "[blue::][white::b]"
			decoration = ""

			if node.Collapsed == true {
				icon = ""
			}
		} else {
			if node.Decoration != "" {
				decoration = strings.Trim(node.Decoration, " ") + " "
			}
		}

		statements = append(statements, &ScrollablePageLine{
			Reference: &FileTreeStatementReference{
				Node: node,
				Diff: node.reference,
			},
			Statements: []*ScrollablePageLineStatement{{
				Content: fmt.Sprintf("%s%s %s[white::-]%s", indent, icon, decoration, node.Filename),
			}},
		})

		if len(node.GlobalDecorations) > 0 {
			statements[len(statements)-1].Statements = append(statements[len(statements)-1].Statements, &ScrollablePageLineStatement{
				Content: fmt.Sprintf("%s", node.GlobalDecorations[0]),
			})
		}

		if node.Collapsed == false {
			for _, child := range node.Children {
				recurse(child, level+1)
			}
		}
	}

	recurse(node, 0)

	return statements
}
