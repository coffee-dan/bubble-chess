package main

import (
	"fmt"
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
