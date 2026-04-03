package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/letieu/ti/internal/ui"
)

func main() {
	p := tea.NewProgram(ui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
}
