package ui

import (
	tea "charm.land/bubbletea/v2"
)

func (m Model) View() tea.View {
	content := m.messages.View()
	content += "\n"
	content += m.input.View()

	v := tea.NewView(content)
	v.AltScreen = false
	return v
}
