package ui

import (
	"github.com/letieu/ti/internal/ui/input"
	"github.com/letieu/ti/internal/ui/list"
	"github.com/letieu/ti/internal/ui/messages"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	suggestion list.Model

	width  int
	height int

	messages messages.Model
	input    input.Model
}

func InitialModel() Model {
	return Model{
		messages: messages.New(),
		input:    input.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
