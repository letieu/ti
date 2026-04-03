package messages

import (
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	messages []string
}

func (m Model) Init() tea.Cmd {
	return nil
}

func New() Model {
	return Model{
		messages: []string{"Welcome"},
	}
}
