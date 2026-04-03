package messages

import (
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
)

var filePattern = regexp.MustCompile(`@[./\w-]+\.[a-zA-Z0-9]+`)
var fileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")) // Green for file paths

func (m Model) View() string {
	var coloredMessages []string
	for _, msg := range m.messages {
		coloredMsg := filePattern.ReplaceAllStringFunc(msg, func(path string) string {
			return fileStyle.Render(path)
		})
		coloredMessages = append(coloredMessages, coloredMsg)
	}
	messagesView := strings.Join(coloredMessages, "\n")
	return messagesView
}
