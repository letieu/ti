package ui

import (
	"bytes"
	"os/exec"
	"strings"

	"charm.land/bubbles/v2/cursor"
	tea "charm.land/bubbletea/v2"
	"github.com/letieu/ti/internal/agent"
	"github.com/letieu/ti/internal/ui/event"
	"github.com/letieu/ti/internal/ui/fzf"
	"github.com/letieu/ti/internal/ui/input"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		inputCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case input.TriggerFzfMsg:
		var stdout bytes.Buffer
		c := exec.Command("fzf", "--height", "50%")
		c.Stdout = &stdout
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			if err != nil {
				return nil
			}
			return fzf.FileSelectedMsg{Path: strings.TrimSpace(stdout.String())}
		})
	case fzf.FileSelectedMsg:
		m.input.InsertFile(msg.Path)
		return m, nil
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
		return m, func() tea.Msg {
			return event.UserMsg{Text: msg.Text}
		}

	case event.UserMsg:
		m.messages.AddUserMsg(msg.Text)
		ch, _ := m.agent.Chat(msg.Text)
		return m, waitForEvent(ch)

	case event.StreamMsg:
		m.messages.AddAgentStream(msg.Ev.Data)
		return m, waitForEvent(msg.Ch)

	case event.StreamDoneMsg:
		m.messages.AddAgentStream("DONE")
		return m, nil
	}

	return m, tea.Batch(inputCmd)
}

func waitForEvent(ch <-chan agent.AgentEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return event.StreamDoneMsg{}
		}

		return event.StreamMsg{Ev: ev, Ch: ch}
	}
}
