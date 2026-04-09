package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/letieu/ti/internal/agent"
	"github.com/letieu/ti/internal/llm/antigravity"
	"github.com/letieu/ti/internal/llm/event"
)

type Cli struct {
	agent       agent.Agent
	ctx         context.Context
	input       []rune
	term        *Terminal
	authManager *AuthManager
}

func New() Cli {
	t, _ := NewTerminal()

	a := agent.NewAgent()
	authManager, _ := NewAuthManager()

	return Cli{
		term:        t,
		agent:       a,
		ctx:         context.Background(),
		authManager: authManager,
	}
}

func (c *Cli) Run() {
	c.term.EnableRaw()
	defer c.term.Restore()

	buf := make([]byte, 1)

	c.renderInput()

	for {
		os.Stdin.Read(buf)
		b := buf[0]

		switch b {

		case '@':
			selected, _ := c.triggerFzf()
			formated := fmt.Sprintf("\033[32m%s\033[0m", selected)
			c.input = append(c.input, []rune(formated)...)
			c.renderInput()

		// ENTER
		case '\r', '\n':
			fmt.Print("\r\n")

			line := string(c.input)

			c.input = []rune{}
			c.handleChat(line)
			c.renderInput()

		// BACKSPACE
		case 127:
			if len(c.input) > 0 {
				c.input = c.input[:len(c.input)-1]
				c.renderInput()
			}

		// NORMAL CHAR
		default:
			c.input = append(c.input, rune(b))
			fmt.Printf("%c", b)
		}
	}
}

func (c *Cli) handleChat(line string) {
	oauth, err := c.authManager.GetCreds("antigravity")
	if err != nil {
		fmt.Printf("err %v \n", err)
		return
	}

	llm := antigravity.New(
		antigravity.GeminiOptions{
			APIKey:      oauth.Access,
			ProjectID:   oauth.Metadata["project_id"],
			Model:       "gemini-3-flash",
			Temperature: 0.7,
			MaxTokens:   500,
		},
	)

	ch, _ := c.agent.Chat(c.ctx, line, agent.ChatOptions{
		LLM: llm,
	})

	for ev := range ch {
		switch e := ev.(type) {
		case event.Start:
		case event.TextDelta:
			fmt.Print(e.Delta)
		case event.Error:
			fmt.Println(e.Msg)
		case event.End:
			fmt.Print("\n")
		}
	}
}

func (c *Cli) renderInput() {
	fmt.Print("\r")
	fmt.Print("> ")

	fmt.Print(string(c.input))
	fmt.Print(" ")

	// move cursor back
	fmt.Print("\r")
	fmt.Print("> ")
	fmt.Print(string(c.input))
}
