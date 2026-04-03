package messages

import "strings"

func (m Model) View() string {
	messagesView := strings.Join(m.messages, "\n")
	return messagesView
}
