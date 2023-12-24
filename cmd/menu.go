/*
Copyright © 2023 Daniel Gerard Ramirez

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"math/bits"
	"math/rand"
	"os"
	"regexp"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/notnil/chess"
	"github.com/spf13/cobra"
)

type GameMsg int
type programMode int
type direction uint8
type bitboard uint64
type errMsg error
type MenuItem struct {
	title  string
	action tea.Cmd
}
type creditVisual func() string
type Model struct {
	mode programMode

	menuItems  []MenuItem
	menuCursor int

	credits       []creditVisual
	creditsCursor int

	pastMovesView   viewport.Model
	nextMoveField   textinput.Model
	game            chess.Game
	boardDirection  direction
	highlightsBoard bitboard
	guessList       []chess.Move
	guessMenu       string
	guessCursor     int
	err             error
}

// 88888bo 888 888 88888bo 88888bo 888    d88888
// 888 888 888v888 888 888 888 888 888    8888<
// 88888<  8888888 88888<  88888<  888.d8 888888
// 888 888 8888888 888 888 888 888 888888 8888<
// 88888P" ?88888P 88888P" 88888P" 888888 ?88888

const chessString string = ` ,d8888 d88 88b d88888  d88888  d88888
d888888 888v888 8888<  88888<  88888<
88888<  8888888 888888 8888888 8888888
?888888 888^888 8888<   >88888  >88888
 "Y8888 ?88 88? ?88888 88888P  88888P`

const golangString string = `                   d88888b       d88888b
                d88P   ?88b   d88P   ?88b
              d8P       P"  d8P       ?8b
    .oooooo d88           d88         88b
.ooooooooo ?88    od88888888         88P
     .ooo ?8b       >8P ?8b       d8P
          ?8bo   od8P   ?8bo   od8P
           ?88888P       ?88888P`

const sidebar = `bubble-chess 0.0.1`

var chessTitle string = ""

const (
	width       = 64
	columnWidth = 20
	margin      = 1
)

const (
	GameCPUTurn = iota
	GameOver
	GameStart
	GameExit
	GameViewCredits
)

const (
	NO_HIGHLIGHT = iota
	HIGHLIGHT
)

const (
	NO_GUESS = -1 << iota
)

const (
	WhiteDirection = iota
	BlackDirection
)

const (
	MainMenuMode = iota
	GameMode
	CreditsMode
)

var rootCmd = &cobra.Command{
	Use:   "bubble-chess",
	Short: "A simple chess tui",
	Long: `Bubble Chess is a terminal UI chess game.
It is build with Bubbletea and is intended
to be feature complete by December 31 2023.`,

	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(New(""))

		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	},
}

var (
	white       = lipgloss.CompleteColor{TrueColor: "#FFFFFF", ANSI256: "15", ANSI: "15"}
	black       = lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"}
	magenta     = lipgloss.CompleteColor{TrueColor: "#AF48B6", ANSI256: "13", ANSI: "5"}
	cyan        = lipgloss.CompleteColor{TrueColor: "#4DA5C9", ANSI256: "14", ANSI: "6"}
	green       = lipgloss.CompleteColor{TrueColor: "#0dbc79", ANSI256: "2", ANSI: "2"}
	brightgreen = lipgloss.CompleteColor{TrueColor: "#23d18b", ANSI256: "10", ANSI: "10"}
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
	strToFileMap = map[string]chess.File{
		"a": chess.FileA,
		"b": chess.FileB,
		"c": chess.FileC,
		"d": chess.FileD,
		"e": chess.FileE,
		"f": chess.FileF,
		"g": chess.FileG,
		"h": chess.FileH,
	}
)
var (
	pieceNameRegex  = regexp.MustCompile("[KQRBN]")
	fileNameRegex   = regexp.MustCompile("[abcdefgh]")
	squareNameRegex = regexp.MustCompile("[abcdefgh][12345678]")
	moveListRegex   = regexp.MustCompile(` (\d+\.)`)
)

var columnStyle = lipgloss.NewStyle().
	Align(lipgloss.Left).
	Margin(0, margin).
	Width(columnWidth)

var golangStyle = lipgloss.NewStyle().
	Foreground(white).
	Background(cyan)

var chessStyle = lipgloss.NewStyle().
	Background(cyan).
	Foreground(black)

var bannerStyle = lipgloss.NewStyle().
	Align(lipgloss.Left).
	Margin(1)

var creditsStyle = lipgloss.NewStyle().
	Align(lipgloss.Left).
	Margin(2).
	Width(50)

var selectedMenuItemStyle = lipgloss.NewStyle().
	Background(magenta).
	Foreground(white)

var menuListStyle = lipgloss.NewStyle().
	MarginRight(4)

func contains(squares []chess.Square, s chess.Square) bool {
	for i := range squares {
		if s == squares[i] {
			return true
		}
	}
	return false
}

func (gm GameMsg) Msg() {}

func exitGame() tea.Msg {
	return GameMsg(GameExit)
}

func golang() string {
	var styledText string
	var text string
	for _, ru := range golangString {
		if str := string(ru); str != " " && str != "\n" {
			text += str
		} else {
			styledText += golangStyle.Render(text)
			text = ""
			styledText += str
		}
	}
	styledText += golangStyle.Render(text)

	return creditsStyle.Render(styledText)
}

func bubbletea() string {
	return creditsStyle.Render("[insert ascii art of bubbletea logo]")
}

func notnil() string {
	return creditsStyle.Render("[insert ascii art of notnil/chess pkg]")
}

func renderTitle() (title string) {
	if chessTitle == "" {
		var text string
		for _, ru := range chessString {
			if str := string(ru); str != " " && str != "\n" {
				text += str
			} else {
				chessTitle += chessStyle.Render(text)
				text = ""
				chessTitle += str
			}
		}
		chessTitle += chessStyle.Render(text)
	}
	title = bannerStyle.Render(chessTitle)
	return
}

func (m *Model) renderMenuItems() string {
	var menu = ""
	for idx, itm := range m.menuItems {
		baseItm := fmt.Sprintf(" %s ", itm.title)
		if idx == m.menuCursor {
			menu += selectedMenuItemStyle.Render(baseItm)
		} else {
			menu += baseItm
		}
		menu += "\n"
	}
	return menuListStyle.Render(menu)
}

func (m *Model) gameNextStep() tea.Msg {
	if m.game.Outcome() == chess.NoOutcome {
		if m.game.Position().Turn() == chess.Black {
			return GameMsg(GameCPUTurn)
		}
	}

	return nil
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

func (m *Model) namedPieceHighlightUpdate(piece chess.Piece) bitboard {
	var origins []chess.Square
	for _, mov := range m.game.ValidMoves() {
		if m.game.Position().Board().Piece(mov.S1()) == piece {
			origins = append(origins, mov.S1())
		}
	}

	return toBitboard(origins)
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

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func substr(str string, first int, last int) string {
	actualLast := min(len(str), last)
	return str[first:actualLast]
}

func (m *Model) generateGuessList(input string) []chess.Move {
	var moveList []chess.Move = []chess.Move{}
	if len(input) < 1 {
		return moveList
	}

	var piece chess.Piece = chess.NoPiece
	if oneRune := input[0:1]; pieceNameRegex.MatchString(oneRune) {
		pieceType := toPieceType(oneRune)
		if pieceType != chess.NoPieceType {
			input = input[1:]
			piece = toPiece(pieceType, chess.White)
		}
	}

	twoRunes := substr(input, 0, 2)

	var haveSquare1 bool = len(twoRunes) == 2
	var s1 chess.Square = chess.NoSquare
	var haveFile1 bool = len(twoRunes) == 1
	var file1 chess.File = chess.FileA

	if haveSquare1 && squareNameRegex.MatchString(twoRunes) {
		s1 = toSquare(twoRunes)
	} else {
		haveFile1, file1 = toFile(input[0:])
	}

	board := m.game.Position().Board()
	for _, mov := range m.game.ValidMoves() {
		movS1 := mov.S1()
		if piece != chess.NoPiece && board.Piece(movS1) != piece {
			continue
		}

		if haveSquare1 && movS1 != s1 {
			continue
		}

		if haveFile1 && movS1.File() != file1 {
			continue
		}

		if piece == chess.Piece(chess.Pawn) && !haveSquare1 && !haveFile1 {
			continue
		}

		moveList = append(moveList, *mov)
	}

	return moveList
}

func (m *Model) singleRuneHighlightUpdate(piece chess.Piece, input string) bitboard {
	ok, file := toFile(input)
	var origins []chess.Square

	for _, mov := range m.game.ValidMoves() {
		s1 := mov.S1()
		if ok && s1.File() == file &&
			m.game.Position().Board().Piece(s1) == piece {
			origins = append(origins, s1)
		}
	}

	return toBitboard(origins)
}

func toFile(input string) (bool, chess.File) {
	if fileNameRegex.MatchString(input) {
		return true, strToFileMap[input]
	} else {
		return false, chess.FileA
	}
}

func toSquare(input string) chess.Square {
	return strToSquareMap[input]
}

func (m *Model) doubleRuneHighlightUpdate(input string) bitboard {
	sq := toSquare(input)

	var destinations []chess.Square

	for _, mov := range m.game.ValidMoves() {
		if mov.S1() == sq {
			destinations = append(destinations, mov.S2())
		}
	}

	return toBitboard(destinations)
}

func (m *Model) tripleRuneHighlightUpdate(input string) bitboard {
	sq := toSquare(input[0:2])
	ok, file := toFile(input[2:3])

	if sq == chess.NoSquare || !ok {
		return 0
	}

	var destinations []chess.Square

	for _, mov := range m.game.ValidMoves() {
		if mov.S1() == sq && mov.S2().File() == file {
			destinations = append(destinations, mov.S2())
		}
	}

	return toBitboard(destinations)
}

func (m *Model) fullMoveHighlightUpdate(input string) bitboard {
	sq1 := toSquare(input[0:2])
	sq2 := toSquare(input[2:4])

	if sq1 == chess.NoSquare || sq2 == chess.NoSquare {
		return 0
	}

	var destinations []chess.Square

	for _, mov := range m.game.ValidMoves() {
		if mov.S1() == sq1 && mov.S2() == sq2 {
			destinations = append(destinations, mov.S2())
		}
	}

	return toBitboard(destinations)
}

func (m *Model) generateHighlights(input string) (newHighlights bitboard) {
	newHighlights = 0
	var piece chess.Piece
	if len(input) >= 1 && pieceNameRegex.MatchString(input[0:1]) {
		pieceType := toPieceType(input[0:1])
		if pieceType == chess.NoPieceType {
			return
		}
		piece = toPiece(pieceType, chess.White)
		input = input[1:]
	} else {
		piece = chess.WhitePawn
	}

	switch len(input) {
	case 0:
		if piece != chess.WhitePawn {
			newHighlights = m.namedPieceHighlightUpdate(piece)
		}
	case 1:
		if fileNameRegex.MatchString(input) {
			newHighlights = m.singleRuneHighlightUpdate(piece, input[0:1])
		}
	case 2:
		newHighlights = m.doubleRuneHighlightUpdate(input)
	case 3:
		newHighlights = m.tripleRuneHighlightUpdate(input)
	case 4:
		newHighlights = m.fullMoveHighlightUpdate(input)
	}
	return
}

// Reverse returns a bitboard where the bit order is reversed.
func (b bitboard) Reverse() bitboard {
	return bitboard(bits.Reverse64(uint64(b)))
}

func (m *Model) highlighted(sq chess.Square) bool {
	var b bitboard
	if m.boardDirection == WhiteDirection {
		b = m.highlightsBoard
	} else {
		b = m.highlightsBoard.Reverse()
	}

	return (bits.RotateLeft64(uint64(b), int(sq)+1) & 1) == 1
}

func (m *Model) RenderBoard() string {
	const numOfSquaresInRow = 8
	var b *chess.Board

	if m.boardDirection == WhiteDirection {
		b = m.game.Position().Board()
	} else {
		b = m.game.Position().Board().Flip(chess.UpDown).Flip(chess.LeftRight)
	}

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

	s := ""

	if m.boardDirection == WhiteDirection {
		s += borderStyle.Render("  A B C D E F G H   ")
	} else {
		s += borderStyle.Render("  H G F E D C B A   ")
	}

	s += "\n"
	for r := 7; r >= 0; r-- {
		var displayRank string
		if m.boardDirection == WhiteDirection {
			displayRank = borderStyle.Render(chess.Rank(r).String() + " ")
		} else {
			displayRank = borderStyle.Render(chess.Rank(7-r).String() + " ")
		}
		s += displayRank

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
				if m.highlighted(square) {
					sqStyle = squareWhiteHighlight
				} else {
					sqStyle = squareWhite
				}
			} else {
				if m.highlighted(square) {
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

		isWhite = !isWhite
		s += displayRank
		s += "\n"
	}
	if m.boardDirection == WhiteDirection {
		s += borderStyle.Render("  A B C D E F G H   ")
	} else {
		s += borderStyle.Render("  H G F E D C B A   ")
	}
	return s
}

func (m *Model) renderMoveList() string {
	return moveListRegex.ReplaceAllString(m.game.String(),
		"\n$1",
	)
}

func toString(pt chess.PieceType) string {
	switch pt {
	case chess.King:
		return "K"
	case chess.Queen:
		return "Q"
	case chess.Rook:
		return "R"
	case chess.Bishop:
		return "B"
	case chess.Knight:
		return "N"
	}
	return ""
}

func (m *Model) renderMove(mov chess.Move) string {
	var board *chess.Board = m.game.Position().Board()
	var str string = ""
	s1 := mov.S1()
	if pieceType := board.Piece(s1).Type(); pieceType != chess.Pawn {
		str += toString(pieceType)
	}
	str += s1.String()
	if mov.HasTag(chess.Capture) {
		str += "x"
	}
	s2 := mov.S2()
	str += s2.String()
	return str
}

func (m *Model) renderGuessList() string {
	var str string = ""
	for idx, mov := range m.guessList {
		movStr := m.renderMove(mov)
		if idx == m.guessCursor {
			movStr = lipgloss.NewStyle().
				Background(white).
				Foreground(black).
				Render(movStr)
		}
		str += fmt.Sprintf("%s ", movStr)
	}
	return fmt.Sprintf("%s ", str)
}

func New(fen string) *Model {
	nmField := textinput.New()
	nmField.Placeholder = "Your move"
	nmField.Focus()
	nmField.CharLimit = columnWidth - len(nmField.Prompt)
	nmField.Width = columnWidth

	pm := viewport.New(columnWidth, 5)

	gameOptions := []func(*chess.Game){chess.UseNotation(chess.LongAlgebraicNotation{})}

	if fen != "" {
		if newOpts, err := chess.FEN(fen); err == nil {
			gameOptions = append(gameOptions, newOpts)
		}
	}

	return &Model{
		mode: MainMenuMode,
		menuItems: []MenuItem{
			{
				title:  "Vs. Computer",
				action: func() tea.Msg { return GameMsg(GameStart) },
			},
			{
				title:  "Vs. Player",
				action: func() tea.Msg { return GameMsg(GameStart) },
			},
			{
				title:  "Credits",
				action: func() tea.Msg { return GameMsg(GameViewCredits) },
			},
		},
		menuCursor: 0,
		credits: []creditVisual{
			golang,
			bubbletea,
			notnil,
		},
		nextMoveField:   nmField,
		pastMovesView:   pm,
		game:            *chess.NewGame(gameOptions...),
		boardDirection:  WhiteDirection,
		highlightsBoard: 0,
		guessList:       []chess.Move{},
		guessMenu:       "",
		guessCursor:     NO_GUESS,
		err:             nil,
	}
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case MainMenuMode:
		return m.mainMenuUpdate(msg)
	case GameMode:
		return m.gameUpdate(msg)
	case CreditsMode:
		return m.creditsUpdate(msg)
	}

	return m, nil
}

func (m *Model) mainMenuUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:

			return m, m.menuItems[m.menuCursor].action
		case tea.KeyDown:
			if m.menuCursor < len(m.menuItems)-1 {
				m.menuCursor += 1
			} else {
				m.menuCursor = 0
			}
		case tea.KeyUp:
			if m.menuCursor > 0 {
				m.menuCursor -= 1
			} else {
				m.menuCursor = len(m.menuItems) - 1
			}
		}
		return m, nil
	case GameMsg:
		switch msg {
		case GameStart:
			m.mode = GameMode
		case GameViewCredits:
			m.mode = CreditsMode
		}
	}

	return m, nil
}

func (m *Model) gameUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.nextMoveField, tiCmd = m.nextMoveField.Update(msg)
	m.pastMovesView, vpCmd = m.pastMovesView.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			return m, exitGame
		case tea.KeyEnter:
			input := m.nextMoveField.Value()

			if err := m.game.MoveStr(input); err != nil {
				// display err
			} else {
				m.nextMoveField.Reset()
				m.pastMovesView.SetContent(m.renderMoveList())
			}

			return m, m.gameNextStep
		case tea.KeyTab:
			if guessLen := len(m.guessList); guessLen > 1 {
				if m.guessCursor < guessLen-1 {
					m.guessCursor += 1
				} else {
					m.guessCursor = 0
				}

				selection := m.renderMove(m.guessList[m.guessCursor])
				m.nextMoveField.SetValue(selection)
				m.highlightsBoard = m.generateHighlights(selection)
				m.guessMenu = m.renderGuessList()
			}

			return m, nil
		case tea.KeyShiftTab:
			if guessLen := len(m.guessList); guessLen > 1 {
				if m.guessCursor > 0 {
					m.guessCursor -= 1
				} else {
					m.guessCursor = guessLen - 1
				}
			}

			selection := m.renderMove(m.guessList[m.guessCursor])
			m.nextMoveField.SetValue(selection)
			m.highlightsBoard = m.generateHighlights(selection)
			m.guessMenu = m.renderGuessList()
		case tea.KeyCtrlF:
			if m.boardDirection == WhiteDirection {
				m.boardDirection = BlackDirection
			} else {
				m.boardDirection = WhiteDirection
			}
			return m, nil
		case tea.KeyCtrlT:
			m.guessMenu = "--------10--------20--------30--------40--------50--------60--------70"
		}
	default:
		var input string = m.nextMoveField.Value()

		m.guessList = m.generateGuessList(input)
		m.guessCursor = NO_GUESS
		m.guessMenu = m.renderGuessList()

		m.highlightsBoard = m.generateHighlights(input)

		return m, nil

	case GameMsg:
		switch msg {
		case GameExit:
			m.mode = MainMenuMode
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

func (m *Model) creditsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			return m, exitGame
		case tea.KeyLeft:
			if m.creditsCursor < len(m.credits)-1 {
				m.creditsCursor += 1
			} else {
				m.creditsCursor = 0
			}
		case tea.KeyRight:
			if m.creditsCursor > 0 {
				m.creditsCursor -= 1
			} else {
				m.creditsCursor = len(m.credits) - 1
			}
		}
		return m, nil
	case GameMsg:
		switch msg {
		case GameStart:
			m.mode = GameMode
		case GameViewCredits:
			m.mode = CreditsMode
		}
	}

	return m, nil
}

func (m *Model) View() string {
	switch m.mode {
	case MainMenuMode:
		return m.mainMenuView()
	case GameMode:
		return m.gameView()
	case CreditsMode:
		return m.creditsView()
	}

	return ""
}

func (m *Model) mainMenuView() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		renderTitle(),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.renderMenuItems(),
			sidebar,
		),
	)
}

func (m *Model) gameView() string {
	column1 := m.RenderBoard()
	column2 := lipgloss.JoinVertical(
		lipgloss.Left,
		m.pastMovesView.View(),
		m.nextMoveField.View(),
	)
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		columnStyle.Copy().Align(lipgloss.Center).Render(column1),
		columnStyle.Copy().Align(lipgloss.Left).Render(column2),
		columnStyle.Copy().MarginRight(0).Render("esc back\n^C quit\ntab toggle\n^F flip"),
	)

	footer := lipgloss.NewStyle().
		Margin(margin).
		Width(width - margin*2).
		Height(2).
		Render(m.guessMenu)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		mainContent, footer,
	)
}

func (m *Model) creditsView() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.credits[m.creditsCursor](),
		columnStyle.Copy().MarginRight(0).Render("esc back\n^C quit\ntab toggle\n^F flip"),
	)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bubble-chess.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().StringVarP(&customStartFEN, "fen", "f", "", "FEN to start from")
}
