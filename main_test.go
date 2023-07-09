package main

import (
	"fmt"
	"sort"
	"testing"
)

func TestParseMove(t *testing.T) {
	chess := New()
	result, err := chess.parseMove("e2e4")
	if err != nil {
		t.Errorf("Unexpected error %e", err)
	}
	want := move{
		from: 52,
		to:   36,
	}
	if result != want {
		t.Errorf("Error, got: %+v, want: %+v", result, want)
	}

}

func TestGetMoveDestination(t *testing.T) {
	// This tests a knight's move set if the knight were at D5
	currentIndex := toIndex("d5")
	var examples = []struct {
		moveOffset int
		destCoord  string
	}{
		{-21, "c7"},
		{-19, "e7"},
		{-12, "b6"},
		{-8, "f6"},
		{8, "b4"},
		{12, "f4"},
		{19, "c3"},
		{21, "e3"},
	}

	for _, ex := range examples {
		tName := fmt.Sprintf("From d5 Knight can move to %s", ex.destCoord)
		t.Run(tName, func(t *testing.T) {
			dest, err := getMoveDestination(currentIndex, ex.moveOffset)
			if err != nil {
				t.Errorf("Unexpected error %e", err)
			}
			if dest != toIndex(ex.destCoord) {
				t.Error("huh?")
			}
		})
	}
}
func TestUnderAttack(t *testing.T) {
	chess := New()
	chess.pieceBoard = [64]int{
		3, 1, 2, 4, 5, 2, 1, 3,
		0, 0, 0, 0, 0, 0, 0, 0,
		6, 6, 6, 6, 6, 6, 6, 6,
		6, 6, 6, 6, 6, 6, 6, 6,
		6, 6, 6, 6, 0, 6, 6, 6,
		6, 6, 6, 6, 6, 6, 6, 6,
		0, 0, 0, 0, 6, 0, 0, 0,
		3, 1, 2, 4, 5, 2, 1, 3,
	}
	chess.colorBoard = [64]int{
		1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1,
		6, 6, 6, 6, 6, 6, 6, 6,
		6, 6, 6, 6, 6, 6, 6, 6,
		6, 6, 6, 6, 0, 6, 6, 6,
		6, 6, 6, 6, 6, 6, 6, 6,
		0, 0, 0, 0, 6, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
	}
	whiteKing := 60
	result := chess.underAttack(whiteKing, BLACK)
	if result {
		t.Errorf("Error, got: %t, want: %t", result, false)
	}
}

func TestMakeMove(t *testing.T) {
	chess := New()
	success, _ := chess.makeMove(move{
		from: toIndex("e1"),
		to:   toIndex("e6"),
	})

	if success {
		t.Errorf("Error, got: %t, want: %t", success, false)
	}

	pieceAtE1 := chess.pieceBoard[toIndex("e1")]

	if pieceAtE1 == EMPTY {
		t.Errorf("Error, king has gone away")
	}
}

func (c *Chess) setPiece(piece int, color int, coord string) {
	idx := toIndex(coord)
	c.pieceBoard[idx] = piece
	c.colorBoard[idx] = color
}

func (c *Chess) setPlayer(side int) {
	c.side = side
	c.xside = side ^ 1
	c.playerSide = side
}

func equal(slice1 []move, slice2 []move) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	sort.Sort(byTo(slice1))
	sort.Sort(byTo(slice2))

	for i, mov := range slice1 {
		if mov != slice2[i] {
			return false
		}
	}
	return true
}

var blankBoard = [64]int{
	6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6,
}

func NewBlank() *Chess {
	c := New()
	c.pieceBoard = blankBoard
	c.colorBoard = blankBoard
	return c
}

var sideNames = []string{
	"White", "Black",
}

var pieceNames = []string{
	"Pawn", "Knight", "Bishop", "Rook", "Queen", "King",
}

func TestGenerateFutureSinglePieces(t *testing.T) {
	var examples = []struct {
		piece        int
		side         int
		start        string
		destinations []string
	}{
		{PAWN, WHITE, "e2", []string{"e3", "e4"}},
		{PAWN, BLACK, "d7", []string{"d5", "d6"}},
	}

	for _, ex := range examples {
		tName := fmt.Sprintf("%s %s at %s %v",
			sideNames[ex.side], pieceNames[ex.piece], ex.start, ex.destinations,
		)
		t.Run(tName, func(t *testing.T) {
			chess := NewBlank()
			chess.setPiece(ex.piece, ex.side, ex.start)
			moves, _ := chess.generateFuture(ex.side)

			startIdx := toIndex(ex.start)
			var expectedMoves []move
			for _, dest := range ex.destinations {
				expectedMoves = append(expectedMoves, move{
					from: startIdx,
					to:   toIndex(dest),
				})
			}

			if !equal(moves, expectedMoves) {
				fmt.Printf("%+v\n\n", move{})
				t.Errorf("Error, got: %v, want: %v", moves, expectedMoves)
			}
		})
	}
}
