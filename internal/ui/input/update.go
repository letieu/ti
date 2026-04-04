package input

import (
	tea "charm.land/bubbletea/v2"
	"strings"
)

type TriggerFzfMsg struct{}

type SubmitMsg struct {
	Text string
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "@":
			val := m.textarea.Value()
			if strings.HasSuffix(val, " ") || val == "" {
				return m, func() tea.Msg {
					return TriggerFzfMsg{}
				}
			}
		case "enter", "ctrl+m":
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
