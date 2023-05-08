package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ScrollablePageLine struct {
	Statements []*ScrollablePageLineStatement
	Reference  interface{}
}

type ScrollablePageLineStatement struct {
	Indent    int
	Content   string
	Alignment int
}

type ScrollablePage struct {
	*tview.Box
	width                    int
	height                   int
	pageOffset               int
	selectedIndex            int
	content                  []*ScrollablePageLine
	selectionChangedCallback func(index int)
}

func NewScrollablePage() *ScrollablePage {
	sp := &ScrollablePage{
		Box: tview.NewBox(),
	}

	return sp
}

func (sp *ScrollablePage) SetSelectionChangedFunc(changed func(index int)) {
	sp.selectionChangedCallback = changed
}

func (sp *ScrollablePage) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return sp.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				sp.ScrollDown()
			case 'k':
				sp.ScrollUp()
			}
		case tcell.KeyUp:
			sp.ScrollUp()
		case tcell.KeyDown:
			sp.ScrollDown()
		case tcell.KeyCtrlD, tcell.KeyPgDn:
			sp.ScrollHalfPageDown()
		case tcell.KeyCtrlU, tcell.KeyPgUp:
			sp.ScrollHalfPageUp()
		}
	})
}

func (sp *ScrollablePage) Clear() *ScrollablePage {
	sp.pageOffset = 0
	sp.selectedIndex = 0
	sp.content = []*ScrollablePageLine{}

	return sp
}

func (sp *ScrollablePage) GetSelectedReference() interface{} {
	if sp.selectedIndex < 0 || sp.selectedIndex >= len(sp.content) {
		return nil
	}

	v := sp.content[sp.selectedIndex]
	if v == nil {
		return nil
	}

	return v.Reference
}

// Moves the highlighted line up or down
func (sp *ScrollablePage) moveSelected(size int) {
	old := sp.selectedIndex

	sp.selectedIndex += size

	// Should not scroll past the end of the content
	end := len(sp.content) - 1
	if sp.selectedIndex > end {
		sp.selectedIndex = end
	}

	if sp.selectedIndex < 0 {
		sp.selectedIndex = 0
	}

	if old != sp.selectedIndex && sp.selectionChangedCallback != nil {
		sp.selectionChangedCallback(sp.selectedIndex)
	}
}

// Scrolls the table up or down
func (sp *ScrollablePage) scroll(size int) {
	sp.pageOffset += size

	end := len(sp.content) - sp.height
	if end < 0 {
		end = 0
	}

	// Should not scroll past the end of the content
	if sp.pageOffset > end {
		sp.pageOffset = end
	}

	if sp.pageOffset < 0 {
		sp.pageOffset = 0
	}
}

func (sp *ScrollablePage) ScrollDown() {
	if (sp.pageOffset+sp.height)-sp.selectedIndex <= 4 {
		sp.scroll(1)
	}

	sp.moveSelected(1)
}

func (sp *ScrollablePage) ScrollHalfPageDown() {
	sp.scroll(sp.height / 2)
	sp.moveSelected(sp.height / 2)
}

func (sp *ScrollablePage) ScrollUp() {
	if sp.selectedIndex-sp.pageOffset <= 4 {
		sp.scroll(-1)
	}

	sp.moveSelected(-1)
}

func (sp *ScrollablePage) ScrollHalfPageUp() {
	sp.scroll(-sp.height / 2)
	sp.moveSelected(-sp.height / 2)
}

func (sp *ScrollablePage) Draw(screen tcell.Screen) {
	sp.Box.DrawForSubclass(screen, sp)
	x, y, width, height := sp.GetInnerRect()
	sp.height = height
	sp.width = width

	i := sp.pageOffset
	end := int(math.Min(float64(len(sp.content)), float64(i+sp.height)))
	offset := 0
	for ; i < end; i++ {
		cl := sp.content[i]

		highlightPrefix := ""
		if i == sp.selectedIndex {
			highlightPrefix = fmt.Sprintf("[:%s]", "gray")
		}

		tview.Print(
			screen,
			highlightPrefix+strings.Repeat(" ", sp.width),
			x,
			y+offset,
			sp.width,
			tview.AlignRight,
			tview.Styles.PrimaryTextColor,
		)

		for _, s := range cl.Statements {
			tview.Print(
				screen,
				highlightPrefix+s.Content,
				x+s.Indent,
				y+offset,
				sp.width,
				s.Alignment,
				tview.Styles.PrimaryTextColor,
			)
		}

		offset++
	}
}
