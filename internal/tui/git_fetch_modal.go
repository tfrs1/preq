package tui

import (
	"context"
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

	cleanup := func() {
		// FIXME: Handle errors
		err := ptmx.Close()
		if err != nil {
			log.Error().Err(err).Msg("Error while closing pty")
		}

		err = pw.Close()
		if err != nil {
			log.Error().Err(err).Msg("Error while closing fetch pipe writer")
		}

		err = pr.Close()
		if err != nil {
			log.Error().Err(err).Msg("Error while closing fetch pipe reader")
		}
	}

	defer cleanup()

	// TODO: Is this how you use context?
	ctx, cancelCtx := context.WithCancel(context.TODO())
	m.input.
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				ptmx.WriteString(m.input.GetText() + "\n")
				m.input.SetText("")
			} else if key == tcell.KeyEscape {
				cancelCtx()
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

	go func() {
		_, err = io.Copy(pw, ptmx)
		if err != nil {
			log.Error().Err(err).Msg("Error while coping from pty")
		}

		cancelCtx()
	}()

	select {
	case <-ctx.Done():
		// wait
	}

	eventBus.Publish("GitFetchModal:RequestClose", nil)
}

func NewGitFetchModal() *GitFetchModal {
	outputView := tview.NewTextView().
		SetWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	inputField := tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetMaskCharacter('*')
	inputField.SetBorder(true)

	f := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(outputView, 0, 1, false).
		AddItem(inputField, 3, 0, true)
	f.SetTitle(" Git fetch ").SetBorder(true)

	grid := tview.NewGrid().
		SetColumns(0, 80, 0).
		SetRows(0, 15, 0).
		AddItem(f, 1, 1, 1, 1, 0, 0, true)

	return &GitFetchModal{
		Grid:   grid,
		input:  inputField,
		output: outputView,
	}
}

var gitFetchModal = NewGitFetchModal()
