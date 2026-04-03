package input

import (
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Model struct {
	textarea textarea.Model
}

func New() Model {
	ta := textarea.New()
	ta.Placeholder = ""
	ta.SetVirtualCursor(true)
	ta.Focus()

	ta.SetPromptFunc(2, func(p textarea.PromptInfo) string {
		if p.LineNumber == 0 {
			return "❯ "
		}
		return " "
	})

	ta.CharLimit = 1000
	ta.SetHeight(3)

	s := ta.Styles()
	s.Focused.CursorLine = lipgloss.NewStyle()
	ta.SetStyles(s)

	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetKeys("shift+enter")

	return Model{
		textarea: ta,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m *Model) SetWidth(w int) {
	m.textarea.SetWidth(w)
}

func (m *Model) InsertFile(path string) {
	m.textarea.InsertString("@" + path)
}
