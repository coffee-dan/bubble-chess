package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	color "github.com/fatih/color"
)

const HISTORY_SIZE = 400

const PAWN = 0
const KNIGHT = 1
const BISHOP = 2
const ROOK = 3
const QUEEN = 4
const KING = 5

const WHITE = 0
const BLACK = 1

const EMPTY = 6

// Either holds an index on pieceBoard or -1 to represent an inaccessible
// square
var boundaryBoard = [120]int{
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, 0, 1, 2, 3, 4, 5, 6, 7, -1,
	-1, 8, 9, 10, 11, 12, 13, 14, 15, -1,
	-1, 16, 17, 18, 19, 20, 21, 22, 23, -1,
	-1, 24, 25, 26, 27, 28, 29, 30, 31, -1,
	-1, 32, 33, 34, 35, 36, 37, 38, 39, -1,
	-1, 40, 41, 42, 43, 44, 45, 46, 47, -1,
	-1, 48, 49, 50, 51, 52, 53, 54, 55, -1,
	-1, 56, 57, 58, 59, 60, 61, 62, 63, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
}

var upscaleBoard = [64]int{
	21, 22, 23, 24, 25, 26, 27, 28,
	31, 32, 33, 34, 35, 36, 37, 38,
	41, 42, 43, 44, 45, 46, 47, 48,
	51, 52, 53, 54, 55, 56, 57, 58,
	61, 62, 63, 64, 65, 66, 67, 68,
	71, 72, 73, 74, 75, 76, 77, 78,
	81, 82, 83, 84, 85, 86, 87, 88,
	91, 92, 93, 94, 95, 96, 97, 98,
}

// Sliding is considered repeated movements in one direction during one turn,
// thus the knight cannot since the knight's L-Shape move is one (1) move.
var canSlide = [6]bool{false, false, true, true, true, false}
var moveLists = map[int][]int{
	PAWN:   {}, // Pawn is generally handled case by case
	KNIGHT: {-21, -19, -12, -8, 8, 12, 19, 21},
	BISHOP: {-11, -9, 9, 11},
	ROOK:   {-10, -1, 1, 10},
	QUEEN:  {-11, -10, -9, -1, 1, 9, 10, 11},
	KING:   {-11, -10, -9, -1, 1, 9, 10, 11},
}

var pieceRunes = map[int][]rune{
	WHITE: {'P', 'N', 'B', 'R', 'Q', 'K'},
	BLACK: {'p', 'n', 'b', 'r', 'q', 'k'},
}

var (
	ErrNotImplemented    = errors.New("feature not implemented (yet)")
	ErrInvalidMove       = errors.New("this doesn't look like a move")
	ErrCannotCheckYrself = errors.New("this move would place you in check")
	ErrMoveOutOfBounds   = errors.New("move is out of bounds")
)

type Chess struct {
	viewport      viewport.Model
	pastMoves     []string
	nextMoveField textarea.Model
	youStyle      lipgloss.Style
	themStyle     lipgloss.Style
	future        []move
	futureIdx     int
	history       [HISTORY_SIZE]snapshot
	historyIdx    int // Number of half-moves (ply) since the start of the game
	pieceBoard    [64]int
	colorBoard    [64]int
	playerSide    int
	side          int
	xside         int
	err           error
}

type move struct {
	from        int
	to          int
	promotion   int
	capturing   bool
	castling    bool
	enPassant   bool
	pawnPushing bool
	pawnMove    bool
	promoting   bool
}

type byTo []move

func (ml byTo) Len() int           { return len(ml) }
func (ml byTo) Less(i, j int) bool { return ml[i].to < ml[j].to }
func (ml byTo) Swap(i, j int)      { ml[i], ml[j] = ml[j], ml[i] }

type snapshot struct {
	lastMove      move
	capturedPiece int // the dearly departed
	// TODO: castling rights
	// TODO: where was the en passant square
	// TODO: fifty-move-draw rule indicator
	// TODO: sha of the position
}

type (
	errMsg error
)

func New() *Chess {
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

	return &Chess{
		nextMoveField: ta,
		pastMoves:     []string{},
		viewport:      vp,
		youStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		themStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		history:       [HISTORY_SIZE]snapshot{},
		historyIdx:    0,
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
		side:       WHITE,
		xside:      BLACK,
		playerSide: WHITE,
		err:        nil,
	}
}

func (c *Chess) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return textarea.Blink
}

func (c *Chess) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	c.nextMoveField, tiCmd = c.nextMoveField.Update(msg)
	c.viewport, vpCmd = c.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(c.nextMoveField.Value())
			return c, tea.Quit
		case tea.KeyEnter:
			whoami := "You"
			senderStyle := c.youStyle
			if c.side != c.playerSide {
				whoami = "Them"
				senderStyle = c.themStyle
			}
			input := c.nextMoveField.Value()

			var errors []error

			// validate move syntax (proper form)
			nextMove, parseErr := c.parseMove(input)

			var validation string
			if parseErr != nil {
				validation = " (invalid)"
				errors = append(errors, parseErr)
			}
			senderString := fmt.Sprintf("  %s%s: ", whoami, validation)
			c.pastMoves = append(c.pastMoves, senderStyle.Render(senderString)+input)

			if len(errors) == 0 {
				// c.pastMoves = append(c.pastMoves,
				// 	fmt.Sprintf("parseMove: %d -> %d", nextMove.from, nextMove.to),
				// )

				res, moveErr := c.makeMove(nextMove)

				c.pastMoves = append(c.pastMoves, fmt.Sprintf("makeMove: %t", res))
				if moveErr != nil {
					c.pastMoves = append(c.pastMoves, fmt.Sprintf("%e", moveErr))
				}

				// validate semantics (does the player have those pieces)
				// validate move pragmatics (is that move legal?)
				// 		- is _this_ side in check?
				// 		-
				// - do the move
				// - prompt for promotion
				// - send error message
				// switch turns

				c.viewport.SetContent(strings.Join(c.pastMoves, "\n"))
				c.nextMoveField.Reset()
				c.viewport.GotoBottom()
			}
		}
	case errMsg:
		c.err = msg
		return c, nil

	}

	return c, tea.Batch(tiCmd, vpCmd)
}

// Convert board index to a human readable coordinate
// primarily for debugging
func toCoord(index int) (coord string) {
	letter := rune(index%8 + 'a')
	number := 8 - index/8
	return fmt.Sprintf("%c%d", letter, number)
}

// Convert human readable coordiante to board index, very useful for writing
// tests which you want to be legible
func toIndex(coord string) (index int) {
	index = int(coord[0] - 'a')
	index += 8 * (8 - int(coord[1]-'0'))
	return
}

func (c *Chess) parseMove(userInput string) (mov move, err error) {
	var (
		from int
		to   int
		// index int32
	)

	moveRunes := []rune(userInput)

	// syntax check
	if len(moveRunes) < 4 ||
		moveRunes[0] < 'a' || moveRunes[0] > 'h' ||
		moveRunes[1] < '1' || moveRunes[1] > '9' ||
		moveRunes[2] < 'a' || moveRunes[2] > 'h' ||
		moveRunes[3] < '1' || moveRunes[3] > '9' {
		err = ErrInvalidMove
		return
	}

	// convert user input to board indices
	from = int(moveRunes[0] - 'a')
	from += 8 * (8 - int(moveRunes[1]-'0'))
	to = int(moveRunes[2] - 'a')
	to += 8 * (8 - int(moveRunes[3]-'0'))

	// requested move is compared against list of possible moves generated once per turn
	// TODO: generate all possible moves once per turn

	mov = move{
		from: from,
		to:   to,
	}
	return
}

// Convert index to side to side rows, typically denoted by numbers a-h
func toFile(boardIndex int) int {
	return boardIndex & 7
}

// Convert index to fore-to-aft columns, typically denoted by letters 1-8
func toRank(boardIndex int) int {
	// Equivalent to dividing by 8 and discarding remainder. But this is cooler B)
	return (63 - boardIndex) >> 3
}

func getMoveDestination(currentIndex int, moveOffset int) (dest int, err error) {
	dest = boundaryBoard[upscaleBoard[currentIndex]+moveOffset]
	if dest == -1 {
		err = ErrMoveOutOfBounds
	}
	return
}

// Returns true if target is under attack by side, false otherwise
func (c *Chess) underAttack(target int, side int) bool {
	for idx, square := range c.pieceBoard {
		if square == target {
			continue
		}
		if c.colorBoard[idx] == side^1 {
			continue
		}
		if square == PAWN {
			if side == WHITE {
				if toFile(idx) != 0 && idx-9 == target {
					return true
				} else if toFile(idx) != 7 && idx-7 == target {
					return true
				}
			} else {
				if toFile(idx) != 0 && idx+7 == target {
					return true
				} else if toFile(idx) != 7 && idx+9 == target {
					return true
				}
			}
		} else {
			for _, moveOffset := range moveLists[square] {
				dest, err := getMoveDestination(idx, moveOffset)
				if err != nil {
					break
				}
				if dest == target {
					return true
				}
				if c.pieceBoard[dest] != EMPTY {
					break
				}
				if !canSlide[square] {
					break
				}
			}
		}
	}
	return false
}

func (c *Chess) inCheck(side int) bool {
	for idx, square := range c.pieceBoard {
		if square == KING && c.colorBoard[idx] == side {
			return c.underAttack(idx, side^1)
		}
	}
	return false // no king ig
}

// Generate list of psuedo-legal moves for the current position
func (c *Chess) generateFuture(side int) (future []move, err error) {
	for idx, square := range c.pieceBoard {
		if c.colorBoard[idx] == side^1 {
			continue
		}

		if square == PAWN {
			if side == WHITE {
				if dest := idx - 8; c.pieceBoard[dest] == EMPTY {
					future = append(future, move{from: idx, to: dest})
				}
				// twice forward move
				if dest := idx - 16; toRank(idx) == 1 && c.pieceBoard[dest] == EMPTY {
					future = append(future, move{from: idx, to: dest})
				}
				// left-capture
				if dest := idx - 9; toFile(idx) != 0 && c.colorBoard[dest] == BLACK {
					future = append(future, move{from: idx, to: dest})
				}
				// right-capture
				if dest := idx - 7; toFile(idx) != 7 && c.colorBoard[dest] == BLACK {
					future = append(future, move{from: idx, to: dest})
				}
			} else {
				if dest := idx + 8; c.pieceBoard[dest] == EMPTY {
					future = append(future, move{from: idx, to: dest})
				}

				if dest := idx + 16; toRank(idx) == 6 && c.pieceBoard[dest] == EMPTY {
					future = append(future, move{from: idx, to: dest})
				}
				// left-capture
				if dest := idx + 9; toFile(idx) != 0 && c.colorBoard[dest] == BLACK {
					future = append(future, move{from: idx, to: dest})
				}
				// right-capture
				if dest := idx + 7; toFile(idx) != 7 && c.colorBoard[dest] == BLACK {
					future = append(future, move{from: idx, to: dest})
				}
			}
		}
	}
	return
}

func (c *Chess) makeMove(mov move) (res bool, err error) {
	res = true

	if mov.castling {
		err = ErrNotImplemented
	}

	// back up move and irreversible board info
	c.history[c.historyIdx] = snapshot{
		lastMove:      mov,
		capturedPiece: c.pieceBoard[mov.to],
	}
	c.historyIdx++

	// TODO: update castling, en passant, and 50-move-draw vars

	// move the piece; the piece leaves before it arrives. this handles a wild
	// edge case that only matters when move gen is not implmented
	piece := c.pieceBoard[mov.from]
	c.colorBoard[mov.from] = EMPTY
	c.pieceBoard[mov.from] = EMPTY

	c.colorBoard[mov.to] = c.side
	if mov.promoting {
		c.pieceBoard[mov.to] = mov.promotion
	} else {
		c.pieceBoard[mov.to] = piece
	}

	// how do we determine that the move was an enPassant?
	if mov.enPassant {
		if c.side == WHITE {
			c.colorBoard[mov.to+8] = EMPTY
			c.pieceBoard[mov.to+8] = EMPTY
		} else {
			c.colorBoard[mov.to-8] = EMPTY
			c.pieceBoard[mov.to-8] = EMPTY
		}
	}

	c.side ^= 1
	c.xside ^= 1
	if c.inCheck(c.xside) {
		c.takeback()
		res = false
	}

	// TODO: generate unique hash of board state for repetition tracking

	return
}

// bits info from tscp182
// 1	capture
// 2	castle
// 4	en passant capture
// 8	pushing a pawn 2 squares
// 16	pawn move
// 32	promote

func (c *Chess) takeback() {

	c.side ^= 1
	c.xside ^= 1
	c.historyIdx--
	mov := c.history[c.historyIdx].lastMove
	// TOOD: reset the various game info flags
	c.colorBoard[mov.from] = c.side
	c.pieceBoard[mov.from] = c.pieceBoard[mov.to]
	// TODO: undo promotion if it happened

	if c.history[c.historyIdx].capturedPiece == EMPTY {
		c.colorBoard[mov.to] = EMPTY
		c.pieceBoard[mov.to] = EMPTY
	} else {
		c.colorBoard[mov.to] = c.xside
		c.pieceBoard[mov.to] = c.history[c.historyIdx].capturedPiece
	}

	// TODO: undo castling if it happened
	// TODO: undo enPassant if it happened

}

func (c *Chess) viewBoard() string {
	var board_string string
	var pieceString string

	border := color.New(color.BgBlack)
	squareBlack := color.BgMagenta
	squareWhite := color.BgCyan
	pieceBlack := color.FgBlack
	pieceWhite := color.FgWhite

	isWhite := true

	var fileLabels string
	switch c.side {
	case WHITE:
		fileLabels = "  A B C D E F G H   "
	case BLACK:
		fileLabels = "  H G F E D C B A   "
	}

	board_string += border.Sprint(fileLabels)
	board_string += "\n"

	var idx int

	for i := range c.pieceBoard {
		switch c.side {
		case WHITE:
			idx = i
		case BLACK:
			idx = len(c.pieceBoard) - 1 - i
		}

		if i%8 == 0 {
			board_string += border.Sprintf("%d ", 8-(idx/8))
		}

		cellStyle := color.New()
		if pType := c.pieceBoard[idx]; pType == EMPTY {
			pieceString = "  "
		} else {
			pColor := c.colorBoard[idx]
			if pColor == WHITE {
				cellStyle.Add(pieceWhite)
			} else {
				cellStyle.Add(pieceBlack)
			}
			pieceString = fmt.Sprintf("%c ", pieceRunes[pColor][pType])
		}

		if isWhite {
			cellStyle.Add(squareWhite)
		} else {
			cellStyle.Add(squareBlack)
		}

		if idx == 0 {
			cellStyle.Add(color.BgRed)
		}

		board_string += cellStyle.Sprintf(pieceString)

		isWhite = !isWhite
		if i%8 == 7 {
			isWhite = !isWhite
			board_string += border.Sprintf("%d ", 8-(idx/8))
			board_string += "\n"
		}
	}
	board_string += border.Sprint(fileLabels)
	board_string += "\n"
	return board_string
}

func (c *Chess) View() string {
	leftPane := c.viewBoard()
	rightPane := fmt.Sprintf(
		"%s\n%s",
		c.viewport.View(),
		c.nextMoveField.View(),
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
	p := tea.NewProgram(New())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
