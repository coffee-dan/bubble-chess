package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	viewport      viewport.Model
	pastMoves     []string
	nextMoveField textarea.Model
	err           error
}

type (
	errMsg error
)

func New() *Model {
	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.Placeholder = "Your move."
	ta.Focus()

	ta.Prompt = "> "
	ta.CharLimit = 10

	ta.SetWidth(30)
	ta.SetHeight(1)

	// Remove cusror line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to Bubble Chess!
Type your move and press Enter to confirm.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return &Model{
		nextMoveField: ta,
		pastMoves:     []string{},
		viewport:      vp,
		err:           nil,
	}
}

func (m *Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.nextMoveField, tiCmd = m.nextMoveField.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			input := m.nextMoveField.Value()
			m.pastMoves = append(m.pastMoves, "  You: "+input)
			m.viewport.SetContent(strings.Join(m.pastMoves, "\n"))
			m.nextMoveField.Reset()
			m.viewport.GotoBottom()
		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *Model) View() string {
	rightPane := fmt.Sprintf(
		"%s\n%s",
		m.viewport.View(),
		m.nextMoveField.View(),
	)
	mainContent := lipgloss.JoinHorizontal(lipgloss.Center, rightPane)

	return fmt.Sprintf(
		"\n\n\n%s\n\n%s",
		mainContent,
		"Press esc to quit.\n",
	)
}

func main() {
	p := tea.NewProgram(New())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
