package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

type Colors map[string]lipgloss.Color

var colors = Colors{
	"primary": lipgloss.Color("#D19A66"),
}

func getTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(colors["primary"])
}

func main() {
	bail(GetUI().Start())
	fmt.Printf("\n   %s\n", getTextStyle().Render("✦ Done!"))
}

func bail(err error) {
	if err != nil {
		fmt.Printf("so ... that busted")
		os.Exit(1)
	}
}
