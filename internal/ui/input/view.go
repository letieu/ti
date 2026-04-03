package input

import (
	"regexp"

	"charm.land/lipgloss/v2"
)

var filePattern = regexp.MustCompile(`@[./\w-]+\.[a-zA-Z0-9]+`)
var fileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")) // Green for file paths

func (m Model) View() string {
	view := m.textarea.View()
	return filePattern.ReplaceAllStringFunc(view, func(path string) string {
		return fileStyle.Render(path)
	})
}
