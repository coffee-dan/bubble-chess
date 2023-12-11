package main

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"unicode"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/notnil/chess"
)

type Model struct {
	header          viewport.Model
	pastMovesView   viewport.Model
	validations     viewport.Model
	nextMoveField   textarea.Model
	game            chess.Game
	highlightsBoard bitboard
	err             error
}

type bitboard uint64

type GameMsg int

func contains(squares []chess.Square, s chess.Square) bool {
	for i := range squares {
		if s == squares[i] {
			return true
		}
	}
	return false
}

func (gm GameMsg) Msg() {}

const (
	GameCPUTurn = iota
	GameOver
)

const (
	NO_HIGHLIGHT = iota
	HIGHLIGHT
)

var (
	white       lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#FFFFFF", ANSI256: "15", ANSI: "15"}
	black       lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"}
	magenta     lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#AF48B6", ANSI256: "13", ANSI: "5"}
	cyan        lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#4DA5C9", ANSI256: "14", ANSI: "6"}
	green       lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#0dbc79", ANSI256: "2", ANSI: "2"}
	brightgreen lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#23d18b", ANSI256: "10", ANSI: "10"}
)

var (
	strToSquareMap = map[string]chess.Square{
		"a1": chess.A1, "a2": chess.A2, "a3": chess.A3, "a4": chess.A4, "a5": chess.A5, "a6": chess.A6, "a7": chess.A7, "a8": chess.A8,
		"b1": chess.B1, "b2": chess.B2, "b3": chess.B3, "b4": chess.B4, "b5": chess.B5, "b6": chess.B6, "b7": chess.B7, "b8": chess.B8,
		"c1": chess.C1, "c2": chess.C2, "c3": chess.C3, "c4": chess.C4, "c5": chess.C5, "c6": chess.C6, "c7": chess.C7, "c8": chess.C8,
		"d1": chess.D1, "d2": chess.D2, "d3": chess.D3, "d4": chess.D4, "d5": chess.D5, "d6": chess.D6, "d7": chess.D7, "d8": chess.D8,
		"e1": chess.E1, "e2": chess.E2, "e3": chess.E3, "e4": chess.E4, "e5": chess.E5, "e6": chess.E6, "e7": chess.E7, "e8": chess.E8,
		"f1": chess.F1, "f2": chess.F2, "f3": chess.F3, "f4": chess.F4, "f5": chess.F5, "f6": chess.F6, "f7": chess.F7, "f8": chess.F8,
		"g1": chess.G1, "g2": chess.G2, "g3": chess.G3, "g4": chess.G4, "g5": chess.G5, "g6": chess.G6, "g7": chess.G7, "g8": chess.G8,
		"h1": chess.H1, "h2": chess.H2, "h3": chess.H3, "h4": chess.H4, "h5": chess.H5, "h6": chess.H6, "h7": chess.H7, "h8": chess.H8,
	}
)

var (
	pieceNameRegex = regexp.MustCompile("[KQRBN]")
	fileNameRegex  = regexp.MustCompile("[abcdefgh]")
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
		nextMoveField:   ta,
		header:          hd,
		pastMovesView:   pm,
		validations:     vs,
		game:            *chess.NewGame(chess.UseNotation(chess.LongAlgebraicNotation{})),
		highlightsBoard: 0,
		err:             nil,
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
	default:
		input := m.nextMoveField.Value()

		// K for king, Q for queen, R for rook, B for bishop, and N for knight

		var piece chess.Piece
		if len(input) >= 1 && pieceNameRegex.MatchString(input[0:1]) {
			pieceType := toPieceType(input[0:1])
			if pieceType == chess.NoPieceType {
				m.clearHighlights()
				return m, nil
			}
			piece = toPiece(pieceType, chess.White)
			input = input[1:]
		} else {
			piece = chess.WhitePawn
		}

		switch len(input) {
		case 0:
			if piece == chess.WhitePawn {
				m.clearHighlights()
			} else {
				m.namedPieceHighlightUpdate(piece)
			}
		case 1:
			if fileNameRegex.MatchString(input) {
				m.singleRuneHighlightUpdate(piece, rune(input[0]))
			} else {
				m.clearHighlights()
			}
		case 2:
			m.doubleRuneHighlightUpdate(input)
		case 3:
			m.tripleRuneHighlightUpdate(input)
		case 4:
			m.fullMoveHighlightUpdate(input)
		default:
			m.clearHighlights()
		}

		return m, nil

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

func toPieceType(s string) chess.PieceType {
	switch s {
	case "K":
		return chess.King
	case "Q":
		return chess.Queen
	case "R":
		return chess.Rook
	case "B":
		return chess.Bishop
	case "N":
		return chess.Knight
	case "P":
		return chess.Pawn
	}
	return chess.NoPieceType
}

func toPiece(typ chess.PieceType, col chess.Color) chess.Piece {
	switch col {
	case chess.White:
		switch typ {
		case chess.King:
			return chess.WhiteKing
		case chess.Queen:
			return chess.WhiteQueen
		case chess.Bishop:
			return chess.WhiteBishop
		case chess.Knight:
			return chess.WhiteKnight
		case chess.Pawn:
			return chess.WhitePawn
		}
	case chess.Black:
		switch typ {
		case chess.King:
			return chess.BlackKing
		case chess.Queen:
			return chess.BlackQueen
		case chess.Bishop:
			return chess.BlackBishop
		case chess.Knight:
			return chess.BlackKnight
		case chess.Pawn:
			return chess.BlackPawn
		}
	}
	return chess.NoPiece
}

func (m *Model) namedPieceHighlightUpdate(piece chess.Piece) {
	var origins []chess.Square
	for _, mov := range m.game.ValidMoves() {
		if m.game.Position().Board().Piece(mov.S1()) == piece {
			origins = append(origins, mov.S1())
		}
	}

	m.highlightsBoard = toBitboard(origins)
}

func toBitboard(squares []chess.Square) bitboard {
	if len(squares) == 0 {
		return 0
	}

	var str string
	for i := 0; i < 64; i++ {
		if contains(squares, chess.Square(i)) {
			str += "1"
		} else {
			str += "0"
		}
	}
	bb, err := strconv.ParseUint(str, 2, 64)
	if err != nil {
		panic(err)
	}
	return bitboard(bb)
}

func toColIndex(input rune) int {
	lower := unicode.ToLower(input)
	return int(lower - 'a')
}

func (m *Model) singleRuneHighlightUpdate(piece chess.Piece, input rune) {
	idx := toColIndex(input)

	file := chess.File(idx)
	var origins []chess.Square

	for _, mov := range m.game.ValidMoves() {
		s1 := mov.S1()
		if s1.File() == file &&
			m.game.Position().Board().Piece(s1) == piece {
			origins = append(origins, s1)
		}
	}

	m.highlightsBoard = toBitboard(origins)
}

func toSquare(input string) chess.Square {
	return strToSquareMap[input]
}

func (m *Model) doubleRuneHighlightUpdate(input string) {
	sq := toSquare(input)

	var destinations []chess.Square

	for _, mov := range m.game.ValidMoves() {
		if mov.S1() == sq {
			destinations = append(destinations, mov.S2())
		}
	}

	m.highlightsBoard = toBitboard(destinations)
}

func (m *Model) tripleRuneHighlightUpdate(input string) {
	sq := toSquare(input[0:2])
	idx := toColIndex(rune(input[2]))

	if sq == chess.NoSquare || idx > 8 || idx < 0 {
		m.clearHighlights()
		return
	}

	file := chess.File(idx)

	var destinations []chess.Square

	for _, mov := range m.game.ValidMoves() {
		if mov.S1() == sq && mov.S2().File() == file {
			destinations = append(destinations, mov.S2())
		}
	}

	m.highlightsBoard = toBitboard(destinations)
}

func (m *Model) fullMoveHighlightUpdate(input string) {
	sq1 := toSquare(input[0:2])
	sq2 := toSquare(input[2:4])

	if sq1 == chess.NoSquare || sq2 == chess.NoSquare {
		m.clearHighlights()
		return
	}

	var destinations []chess.Square

	for _, mov := range m.game.ValidMoves() {
		if mov.S1() == sq1 && mov.S2() == sq2 {
			destinations = append(destinations, mov.S2())
		}
	}

	m.highlightsBoard = toBitboard(destinations)
}

func (m *Model) clearHighlights() {
	m.highlightsBoard = 0
}

func (b bitboard) highlighted(sq chess.Square) bool {
	return (uint64(b) >> uint64(63-sq) & 1) == 1
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

	squareBlackHighlight := lipgloss.NewStyle().
		Background(brightgreen)

	squareWhiteHighlight := lipgloss.NewStyle().
		Background(green)

	s := "\n"
	s += borderStyle.Render("  A B C D E F G H   ")
	s += "\n"
	for r := 7; r >= 0; r-- {
		s += borderStyle.Render(chess.Rank(r).String() + " ")
		for f := 0; f < numOfSquaresInRow; f++ {

			square := chess.NewSquare(chess.File(f), chess.Rank(r))
			p := b.Piece(square)

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

			var sqStyle lipgloss.Style

			if isWhite {
				if m.highlightsBoard.highlighted(square) {
					sqStyle = squareWhiteHighlight
				} else {
					sqStyle = squareWhite
				}
			} else {
				if m.highlightsBoard.highlighted(square) {
					sqStyle = squareBlackHighlight
				} else {
					sqStyle = squareBlack
				}
			}
			sq := sqStyle.
				Foreground(pieceColorCode).
				Render(pieceString)

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
