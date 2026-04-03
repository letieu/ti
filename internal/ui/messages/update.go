package messages

import (
	tea "charm.land/bubbletea/v2"
	"github.com/letieu/ti/internal/ui/input"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case input.SubmitMsg:
		m.messages = append(m.messages, "❯ "+msg.Text)
		m.messages = append(m.messages, "Not ok")
	}

	return m, nil
}
