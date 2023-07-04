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

func (c *Chess) clearBoard() {
	for idx := range c.pieceBoard {
		c.pieceBoard[idx] = EMPTY
	}
	for idx := range c.colorBoard {
		c.colorBoard[idx] = EMPTY
	}
}

func (c *Chess) setPiece(piece int, color int, coord string) {
	idx := toIndex(coord)
	c.pieceBoard[idx] = piece
	c.colorBoard[idx] = color
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

func TestGenerateFuture(t *testing.T) {
	chess := New()
	chess.clearBoard()
	chess.setPiece(PAWN, WHITE, "e2")
	moves, _ := chess.generateFuture(WHITE)

	expectedMoves := []move{
		{
			from: toIndex("e2"),
			to:   toIndex("e3"),
		},
		{
			from: toIndex("e2"),
			to:   toIndex("e4"),
		},
	}

	if !equal(moves, expectedMoves) {
		fmt.Printf("%+v\n\n", move{})
		t.Errorf("Error, got: %v, want: %v", moves, expectedMoves)
	}

}
