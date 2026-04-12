package cli

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/letieu/ti/internal/agent"
	"github.com/letieu/ti/internal/logger"
)

type Cli struct {
	agent       agent.Agent
	ctx         context.Context
	input       []rune
	term        *Terminal
	authManager *AuthManager
	llmManager  *LlmManager
	stopSpinner chan bool
	provider    string
}

func New() Cli {
	t, _ := NewTerminal()

	a := agent.NewAgent("you are a coding agent, in cli. For formatting the response, please do not use too much markdown format due to we are in cli, so, for heading, use icon instead of markdown", getDefaultTools())
	if os.Getenv("TI_MEMORY_DUMP") == "true" {
		if path := os.Getenv("TI_MEMORY_DUMP_PATH"); path != "" {
			a.DebugMemoryDumpPath = path
		} else {
			a.DebugMemoryDumpPath = "memory_dump.json"
		}
	}

	authManager, _ := NewAuthManager()
	llmManager := NewLlmManager()

	return Cli{
		term:        t,
		agent:       a,
		ctx:         context.Background(),
		authManager: authManager,
		llmManager:  llmManager,
		stopSpinner: make(chan bool),
		provider:    "mock",
	}
}

func (c *Cli) startSpinner() {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-c.stopSpinner:
			fmt.Print("\r          \r") // Clear spinner
			return
		default:
			fmt.Printf("\r%s ", spinner[i])
			i = (i + 1) % len(spinner)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *Cli) Run() {
	logger.Log.Info("Starting CLI application")
	c.term.EnableRaw()
	defer c.term.Restore()

	buf := make([]byte, 1)

	c.renderInput()

	for {
		os.Stdin.Read(buf)
		b := buf[0]

		switch b {

		case '@':
			logger.Log.Debug("FZF triggered")
			selected, _ := c.triggerFzf()
			formated := fmt.Sprintf("\033[32m%s\033[0m", selected)
			c.input = append(c.input, []rune(formated)...)
			c.renderInput()

		// ENTER
		case '\r', '\n':
			fmt.Print("\r\n")

			line := string(c.input)
			logger.Log.Debug("User input received", "input", line)

			c.input = []rune{}
			if strings.HasPrefix(line, "/") {
				c.handleCmd(line)
				c.renderInput()
			} else {
				c.handleChat(line)
				c.renderInput()
			}

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
	logger.Log.Debug("Handling chat", "input", line)
	if line == "" {
		return
	}

	creds, err := c.authManager.GetCreds(c.provider)
	if err != nil {
		logger.Log.Error("Failed to get credentials", "error", err)
		fmt.Printf("err %v \n", err)
		return
	}

	err = c.llmManager.SetProvider(c.provider, creds)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting LLM provider: %v\n", err)
		return
	}

	ch, _ := c.agent.Chat(c.ctx, line, c.llmManager)

	go c.startSpinner()
	isSpinning := true

	for ev := range ch {
		if isSpinning {
			c.stopSpinner <- true
			isSpinning = false
		}
		switch e := ev.(type) {
		case agent.TextStart:
			// Could add visual indicator that agent is starting to respond

		case agent.TextDelta:
			fmt.Print(e.Delta)

		case agent.TextEnd:
			fmt.Print("\n")

		case agent.ThinkingStart:
			fmt.Print("\033[90m") // Gray color for thinking
			fmt.Print("[thinking] ")

		case agent.ThinkingDelta:
			fmt.Print(e.Delta)

		case agent.ThinkingEnd:
			fmt.Print("\033[0m") // Reset color
			fmt.Println()

		case agent.ToolCallRequest:
			label := e.Name
			if args, ok := e.Aggs.(map[string]any); ok {
				if labelStr, ok := args["label"].(string); ok && labelStr != "" {
					label = labelStr
				}
			}
			fmt.Printf("\033[36m🛠️  [%s: %s]\033[0m", e.Name, label) // Cyan

		case agent.ToolCallResult:
			diff, _ := e.Result["diff"].(string)
			if diff != "" {
				fmt.Printf("\n%s", diff)
			}

		case agent.Error:
			fmt.Printf("\033[31mError: %s\033[0m\n", e.Msg) // Red

		case agent.Done:
		}
	}

	if isSpinning {
		c.stopSpinner <- true
	}
}

func (c *Cli) handleCmd(command string) {
	logger.Log.Debug("Handling command", "command", command)

	// Parse command and arguments
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]

	switch cmd {
	case "/login":
		logger.Log.Info("Initiating login")
		c.authManager.Login(c.provider)

	case "/set-provider":
		if len(parts) < 2 {
			fmt.Println("Usage: /set-provider <provider>")
			fmt.Printf("Available providers: %s\n", strings.Join(c.llmManager.Providers(), ", "))
			return
		}

		provider := parts[1]
		providers := c.llmManager.Providers()

		valid := slices.Contains(providers, provider)

		if !valid {
			fmt.Printf("Invalid provider: %s\n", provider)
			fmt.Printf("Available providers: %s\n", strings.Join(providers, ", "))
			return
		}

		c.provider = provider
		fmt.Printf("Provider set to: %s\n", provider)
		logger.Log.Info("Provider changed", "provider", provider)

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
	}
}

func (c *Cli) renderInput() {
	fmt.Print("\r")
	fmt.Print("› ")

	fmt.Print(string(c.input))
	fmt.Print(" ")

	// move cursor back
	fmt.Print("\r")
	fmt.Print("› ")
	fmt.Print(string(c.input))
}
