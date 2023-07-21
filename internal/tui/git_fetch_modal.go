package tui

import (
	"io"
	"os/exec"

	"github.com/creack/pty"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

type GitFetchModal struct {
	*tview.Grid
	input  *tview.InputField
	output *tview.TextView
}

func (m *GitFetchModal) StartGitFetch(path string) {
	m.output.SetText("")
	m.input.SetText("")

	cmd := exec.Command("git", "fetch")
	cmd.Dir = path

	// Create a pipe for capturing the terminal output
	pr, pw := io.Pipe()

	// Start the command
	ptmx, err := pty.Start(cmd)
	if err != nil {
		panic(err)
	}
	defer func() {
		// FIXME: Handle errors
		_ = ptmx.Close()
		_ = pw.Close()
		_ = pr.Close()
	}()

	// Create an InputField for entering commands
	m.input.
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				ptmx.WriteString(m.input.GetText() + "\n")
				m.input.SetText("")
			} else if key == tcell.KeyEscape {
				eventBus.Publish("GitFetchModal:RequestClose", nil)
			}
		})

	go func() {
		output := ""
		bfr := make([]byte, 128)
		for {
			n, err := pr.Read(bfr)
			if err != nil {
				log.Error().Err(err)
				break
			}
			output += string(bfr[:n])
			m.output.SetText(output)
		}
	}()

	_, _ = io.Copy(pw, ptmx)

	eventBus.Publish("GitFetchModal:RequestClose", nil)
}

func NewGitFetchModal() *GitFetchModal {
	// Create a TextView to display the terminal output
	outputView := tview.NewTextView().
		SetWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	inputField := tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite).
		SetMaskCharacter('✱')

	inputField.SetBorder(true)
	outputView.SetBorder(true)

	f := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("Git fetch"), 1, 0, false).
		AddItem(outputView, 0, 1, false).
		AddItem(inputField, 3, 0, true)

	grid := tview.NewGrid().
		SetColumns(0, 50, 0).
		SetRows(0, 20, 0).
		AddItem(f, 1, 1, 1, 1, 0, 0, true)

	return &GitFetchModal{
		Grid:   grid,
		input:  inputField,
		output: outputView,
	}
}

var gitFetchModal = NewGitFetchModal()
