package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	color "github.com/fatih/color"
)

const PAWN = 0
const KNIGHT = 1
const BISHOP = 2
const ROOK = 3
const QUEEN = 4
const KING = 5

const WHITE = 0
const BLACK = 1

const EMPTY = 6

var pieceRunes = map[int][]rune{
	WHITE: {'P', 'N', 'B', 'R', 'Q', 'K'},
	BLACK: {'p', 'n', 'b', 'r', 'q', 'k'},
}

type model struct {
	viewport      viewport.Model
	pastMoves     []string
	nextMoveField textarea.Model
	senderStyle   lipgloss.Style
	pieceBoard    [64]int
	colorBoard    [64]int
	err           error
}

type (
	errMsg error
)

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Your move."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cusror line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to Bubble Chess!
Type your move and press Enter to confirm.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		nextMoveField: ta,
		pastMoves:     []string{},
		viewport:      vp,
		senderStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:           nil,
		pieceBoard: [64]int{
			3, 1, 2, 4, 5, 2, 1, 3,
			0, 0, 0, 0, 0, 0, 0, 0,
			6, 6, 6, 6, 6, 6, 6, 6,
			6, 6, 6, 6, 6, 6, 6, 6,
			6, 6, 6, 6, 6, 6, 6, 6,
			6, 6, 6, 6, 6, 6, 6, 6,
			0, 0, 0, 0, 0, 0, 0, 0,
			3, 1, 2, 4, 5, 2, 1, 3,
		},
		colorBoard: [64]int{
			1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1,
			6, 6, 6, 6, 6, 6, 6, 6,
			6, 6, 6, 6, 6, 6, 6, 6,
			6, 6, 6, 6, 6, 6, 6, 6,
			6, 6, 6, 6, 6, 6, 6, 6,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			fmt.Println(m.nextMoveField.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			m.pastMoves = append(m.pastMoves, m.senderStyle.Render("  You: ")+m.nextMoveField.Value())
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

func (m model) viewBoard() string {
	board_string := ""
	black_bg := color.New(color.BgBlack).SprintFunc()
	white_bg := color.New(color.BgWhite).SprintFunc()
	is_white := true
	for index, square := range m.pieceBoard {
		square_string := ". "
		if square != EMPTY {
			color := m.colorBoard[index]
			square_string = fmt.Sprintf("%c ", pieceRunes[color][square])
		}

		if is_white {
			board_string += fmt.Sprintf(white_bg(square_string))
		} else {
			board_string += fmt.Sprintf(black_bg(square_string))
		}
		is_white = !is_white
		if (index+1)%8 == 0 {
			is_white = !is_white
			board_string += "\n"
		}
	}
	return board_string
}

func (m model) View() string {
	leftPane := m.viewBoard()
	rightPane := fmt.Sprintf(
		"%s\n%s",
		m.viewport.View(),
		m.nextMoveField.View(),
	)
	mainContent := lipgloss.JoinHorizontal(lipgloss.Center, leftPane, rightPane)

	return fmt.Sprintf(
		"\n%s\n\n%s\n\n%s",
		"You gais like ganoo chese",
		mainContent,
		"Press esc to quit.\n",
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
