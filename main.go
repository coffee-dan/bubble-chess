package main

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/notnil/chess"
)

type Model struct {
	header        viewport.Model
	pastMovesView viewport.Model
	validations   viewport.Model
	nextMoveField textarea.Model
	game          chess.Game
	err           error
}

type GameMsg int

func (gm GameMsg) Msg() {}

const (
	GameCPUTurn = iota
	GameOver
)

var (
	white   lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#FFFFFF", ANSI256: "15", ANSI: "15"}
	black   lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"}
	magenta lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#AF48B6", ANSI256: "13", ANSI: "5"}
	cyan    lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#4DA5C9", ANSI256: "14", ANSI: "6"}
)

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

	hd := viewport.New(30, 2)
	hd.SetContent(`White to move`)

	pm := viewport.New(30, 5)
	vs := viewport.New(30, 1)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return &Model{
		nextMoveField: ta,
		header:        hd,
		pastMovesView: pm,
		validations:   vs,
		game:          *chess.NewGame(chess.UseNotation(chess.LongAlgebraicNotation{})),
		err:           nil,
	}
}

func (m *Model) gameNextStep() tea.Msg {
	if m.game.Outcome() == chess.NoOutcome {
		if m.game.Position().Turn() == chess.Black {
			return GameMsg(GameCPUTurn)
		}
	}

	return nil
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
	m.pastMovesView, vpCmd = m.pastMovesView.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			input := m.nextMoveField.Value()

			if err := m.game.MoveStr(input); err != nil {
				m.validations.SetContent(err.Error())
			} else {
				m.nextMoveField.Reset()
				m.pastMovesView.SetContent(m.renderMoveList())
			}

			return m, m.gameNextStep
		}
	case GameMsg:
		switch msg {
		case GameCPUTurn:
			moves := m.game.ValidMoves()
			move := moves[rand.Intn(len(moves))]
			m.game.Move(move)
			m.pastMovesView.SetContent(m.renderMoveList())

			return m, m.gameNextStep
		case GameOver:
			return m, tea.Quit
		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *Model) RenderBoard() string {
	const numOfSquaresInRow = 8
	b := m.game.Position().Board()

	borderStyle := lipgloss.NewStyle().
		Background(black).
		Foreground(white)

	var pieceString string
	var pieceColorCode lipgloss.CompleteColor
	isWhite := true
	squareBlack := lipgloss.NewStyle().
		Background(cyan)

	squareWhite := lipgloss.NewStyle().
		Background(magenta)

	s := "\n"
	s += borderStyle.Render("  A B C D E F G H   ")
	s += "\n"
	for r := 7; r >= 0; r-- {
		s += borderStyle.Render(chess.Rank(r).String() + " ")
		for f := 0; f < numOfSquaresInRow; f++ {
			p := b.Piece(chess.NewSquare(chess.File(f), chess.Rank(r)))

			if p == chess.NoPiece {
				pieceString = "  "
				pieceColorCode = black
			} else {
				pieceString = p.String() + " "
				if p.Color() == chess.White {
					pieceColorCode = white
				} else {
					pieceColorCode = black
				}
			}

			var sq string

			if isWhite {
				sq = squareWhite.
					Foreground(pieceColorCode).
					Render(pieceString)
			} else {
				sq = squareBlack.
					Foreground(pieceColorCode).
					Render(pieceString)
			}

			s += sq
			isWhite = !isWhite
		}
		s += borderStyle.Render(" " + chess.Rank(r).String())
		isWhite = !isWhite
		s += "\n"
	}
	s += borderStyle.Render("   A B C D E F G H  ")
	return s
}

var moveListRegex *regexp.Regexp = regexp.MustCompile(` (\d+\.)`)

func (m *Model) renderMoveList() string {
	return moveListRegex.ReplaceAllString(m.game.String(),
		"\n$1",
	)
}

func (m *Model) View() string {
	leftPane := m.RenderBoard()
	rightPane := fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		m.header.View(),
		m.pastMovesView.View(),
		m.validations.View(),
		m.nextMoveField.View(),
	)
	mainContent := lipgloss.JoinHorizontal(lipgloss.Center, leftPane, rightPane)

	return fmt.Sprintf(
		"\n\n\n%s\n\n%s",
		mainContent,
		"Press esc to quit.\n",
	)
}

func main() {
	p := tea.NewProgram(New())

	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}

		defer f.Close()
	}

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
