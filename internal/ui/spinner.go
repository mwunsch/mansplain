package ui

import (
	"context"
	"fmt"
	"os"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type actionDoneMsg struct{ err error }

type spinnerModel struct {
	spinner spinner.Model
	start   time.Time
	title   string
	err     error
	done    bool
	action  func(context.Context) error
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.runAction())
}

func (m spinnerModel) runAction() tea.Cmd {
	return func() tea.Msg {
		return actionDoneMsg{err: m.action(context.Background())}
	}
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case actionDoneMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			m.err = fmt.Errorf("interrupted")
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m spinnerModel) View() tea.View {
	if m.done {
		return tea.NewView("")
	}
	elapsed := time.Since(m.start).Truncate(time.Second)
	return tea.NewView(fmt.Sprintf("%s %s  %s\n",
		m.spinner.View(),
		m.title,
		StyleSubtle.Render(elapsed.String()),
	))
}

// RunGeneration runs the action with an animated spinner on TTY,
// or a simple status line on non-TTY.
func RunGeneration(title string, action func(context.Context) error) error {
	if !isTTY() {
		fmt.Fprintf(os.Stderr, "%s\n", title)
		return action(context.Background())
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa"))

	p := tea.NewProgram(spinnerModel{
		spinner: s,
		start:   time.Now(),
		title:   title,
		action:  action,
	}, tea.WithOutput(os.Stderr))

	result, err := p.Run()
	if err != nil {
		return err
	}
	return result.(spinnerModel).err
}

func isTTY() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
