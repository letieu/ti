package ui

import (
	"charm.land/bubbles/v2/cursor"
	tea "charm.land/bubbletea/v2"
	"github.com/letieu/ti/internal/ui/input"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		messagesCmd tea.Cmd
		inputCmd    tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.SetWidth(m.width)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		default:
			m.input, inputCmd = m.input.Update(msg)
			return m, inputCmd
		}

	case cursor.BlinkMsg:
		m.input, inputCmd = m.input.Update(msg)
		return m, inputCmd

	case input.SubmitMsg:
		m.messages, messagesCmd = m.messages.Update(msg)
		return m, messagesCmd
	}

	m.messages, messagesCmd = m.messages.Update(msg)
	return m, tea.Batch(inputCmd, messagesCmd)
}
