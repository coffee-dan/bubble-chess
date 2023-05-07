package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	color "github.com/fatih/color"
)

type model struct {
	board [8][8]rune
}

func initialModel() model {
	return model{
		board: [8][8]rune{
			{'r', 'n', 'b', 'q', 'k', 'b', 'n', 'r'},
			{'p', 'p', 'p', 'p', 'p', 'p', 'p', 'p'},
			{'.', '.', '.', '.', '.', '.', '.', '.'},
			{'.', '.', '.', '.', '.', '.', '.', '.'},
			{'.', '.', '.', '.', '.', '.', '.', '.'},
			{'.', '.', '.', '.', '.', '.', '.', '.'},
			{'P', 'P', 'P', 'P', 'P', 'P', 'P', 'P'},
			{'R', 'N', 'B', 'Q', 'K', 'B', 'N', 'R'},
		},
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func viewBoard(board [8][8]rune) string {
	board_string := ""
	black_bg := color.New(color.BgBlack).SprintFunc()
	white_bg := color.New(color.BgWhite).SprintFunc()
	is_white := true
	for _, rank := range board {
		for _, square := range rank {
			square_string := fmt.Sprintf("%c ", square)
			if is_white {
				board_string += fmt.Sprintf(white_bg(square_string))
			} else {
				board_string += fmt.Sprintf(black_bg(square_string))
			}
			is_white = !is_white
		}
		is_white = !is_white
		board_string += "\n"
	}
	return board_string
}

func (m model) View() string {
	// The header
	s := "You gais like ganoo chese\n\n"
	s += viewBoard(m.board)
	s += "\nPress q to quit.\n"

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
