package ui

import (
	"github.com/letieu/ti/internal/agent"
	"github.com/letieu/ti/internal/ui/input"
	"github.com/letieu/ti/internal/ui/messages"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	width  int
	height int

	messages messages.Model
	input    input.Model
	agent    agent.Agent
}

func InitialModel() Model {
	return Model{
		messages: messages.New(),
		input:    input.New(),
		agent: agent.NewAgent(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
