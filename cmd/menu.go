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
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "bubble-chess",
		Short: "A simple chess tui",
		Long: `Bubble Chess is a terminal UI chess game.
It is build with Bubbletea and is intended
to be feature complete by December 31 2023.`,

		Run: func(cmd *cobra.Command, args []string) {
			p := tea.NewProgram(New())

			if _, err := p.Run(); err != nil {
				fmt.Printf("Alas, there's been an error: %v", err)
				os.Exit(1)
			}
		},
	}
)

type Model struct {
	menuItems  []string
	menuCursor int
	err        error
}

func New() *Model {

	return &Model{
		menuItems:  []string{"Vs. Computer", "Credits"},
		menuCursor: 0,
		err:        nil,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyUp:
			if m.menuCursor < len(m.menuItems)-1 {
				m.menuCursor += 1
			} else {
				m.menuCursor = 0
			}
		case tea.KeyDown:
			if m.menuCursor > 0 {
				m.menuCursor -= 1
			} else {
				m.menuCursor = len(m.menuItems) - 1
			}
		}
		return m, nil
	}

	return m, nil
}

// TODO make a colors module
var (
	white   lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#FFFFFF", ANSI256: "15", ANSI: "15"}
	black   lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"}
	magenta lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#AF48B6", ANSI256: "13", ANSI: "5"}
	cyan    lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#4DA5C9", ANSI256: "14", ANSI: "6"}
	// green       lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#0dbc79", ANSI256: "2", ANSI: "2"}
	// brightgreen lipgloss.CompleteColor = lipgloss.CompleteColor{TrueColor: "#23d18b", ANSI256: "10", ANSI: "10"}
)

// 88888bo 888 888 88888bo 88888bo 888    d88888
// 888 888 888v888 888 888 888 888 888    8888<
// 88888<  8888888 88888<  88888<  888.d8 888888
// 888 888 8888888 888 888 888 888 888888 8888<
// 88888P" ?88888P 88888P" 88888P" 888888 ?88888

const chess string = ` ,d8888 d88 88b d88888  d88888  d88888
d888888 888v888 8888<  88888<  88888<
88888<  8888888 888888 8888888 8888888
?888888 888^888 8888<   >88888  >88888
 "Y8888 ?88 88? ?88888 88888P  88888P`

var chessTitle string = ""

var chessStyle = lipgloss.NewStyle().
	Background(cyan).
	Foreground(black)

var bannerStyle = lipgloss.NewStyle().
	Align(lipgloss.Left).
	Margin(1)

var selectedMenuItemStyle = lipgloss.NewStyle().
	Background(magenta).
	Foreground(white)

var menuListStyle = lipgloss.NewStyle().
	MarginRight(4)

func renderTitle() (title string) {
	if chessTitle == "" {
		var text string
		for _, ru := range chess {
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
		baseItm := fmt.Sprintf(" %s ", itm)
		if idx == m.menuCursor {
			menu += selectedMenuItemStyle.Render(baseItm)
		} else {
			menu += baseItm
		}
		menu += "\n"
	}
	return menuListStyle.Render(menu)
}

const sidebar = `bubble-chess 0.0.1`

func (m *Model) View() string {
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
