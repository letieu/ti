package input

import (
	tea "charm.land/bubbletea/v2"
)

type SubmitMsg struct {
	Text string
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			val := m.textarea.Value()
			m.textarea.Reset()
			return m, func() tea.Msg {
				return SubmitMsg{Text: val}
			}
		}
	}

	var taCmd tea.Cmd
	m.textarea, taCmd = m.textarea.Update(msg)
	return m, taCmd
}
