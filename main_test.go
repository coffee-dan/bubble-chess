package main

import (
	"fmt"
	"sort"
	"testing"
)

var sideNames = []string{
	"White", "Black",
}

var pieceNames = []string{
	"Pawn", "Knight", "Bishop", "Rook", "Queen", "King",
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

func equal(slice1, slice2 []move) bool {
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

// is []move small a subset of []move large
// assumption: no duplicates
func isSubset(small, large []move) bool {
	seen := make(map[move]bool)
	for _, mov := range large {
		seen[mov] = true
	}

	for _, mov := range small {
		if !seen[mov] {
			return false
		}
	}

	return true
}

func NewBlank() *Chess {
	c := New()
	c.pieceBoard = blankBoard
	c.colorBoard = blankBoard
	return c
}

func TestIsSubset(t *testing.T) {
	small := []move{{from: 52, to: 44}}
	large := []move{
		{from: 52, to: 44},
		{from: 52, to: 36},
	}

	if !isSubset(small, large) {
		t.Error("Error, got: false, want: true")
	}
}

func TestToFile(t *testing.T) {
	file := toFile(toIndex("e2"))

	if file != 4 {
		t.Errorf("Error, got: %d, want: 4", file)
	}
}

func TestToRank(t *testing.T) {
	rank := toRank(toIndex("e2"))

	if rank != 1 {
		t.Errorf("Error, got: %d, want: 1", rank)
	}
}

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

func TestGenerateFutureSinglePieces(t *testing.T) {
	var examples = []struct {
		piece        int
		side         int
		start        string
		destinations []string
	}{
		{PAWN, WHITE, "e2", []string{"e3", "e4"}},
		{PAWN, WHITE, "a2", []string{"a3", "a4"}},
		{PAWN, WHITE, "h2", []string{"h3", "h4"}},
		{PAWN, WHITE, "e3", []string{"e4"}},
		{PAWN, BLACK, "d7", []string{"d5", "d6"}},
		{PAWN, BLACK, "a7", []string{"a5", "a6"}},
		{PAWN, BLACK, "h7", []string{"h5", "h6"}},
		{PAWN, BLACK, "d6", []string{"d5"}},
		{KNIGHT, WHITE, "b1", []string{"a3", "c3", "d2"}},
		{KNIGHT, WHITE, "d5", []string{"e3", "c3", "f4", "b4", "f6", "e7", "c7", "b6"}},
		{BISHOP, WHITE, "c1", []string{"b2", "a3", "d2", "e3", "f4", "g5", "h6"}},
		{ROOK, WHITE, "a1", []string{"b1", "c1", "d1", "e1", "f1", "g1", "h1", "a2", "a3", "a4", "a5", "a6", "a7", "a8"}},
		{QUEEN, WHITE, "d1", []string{
			"c1", "b1", "a1", "c2", "b3", "a4", "d2", "d3", "d4", "d5", "d6", "d7", "d8", "e2", "f3", "g4", "h5", "e1", "f1", "g1", "h1",
		}},
		{KING, WHITE, "e1", []string{"d1", "d2", "e2", "f2", "f1"}},
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
