package cli

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/letieu/ti/internal/logger"
)

type CommandFunc func(args []string)

func (c *Cli) initCommands() {
	c.commands = map[string]CommandFunc{
		"/login":        c.cmdLogin,
		"/get-provider": c.cmdGetProvider,
		"/set-provider": c.cmdSetProvider,
	}
}

func (c *Cli) CommandList() string {
	keys := make([]string, 0, len(c.commands))
	for k := range c.commands {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return strings.Join(keys, "\n");
}

func (c *Cli) handleCmd(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]
	args := parts[1:]

	handler, ok := c.commands[cmd]
	if !ok {
		fmt.Printf("Unknown command: %s\n", cmd)
		return
	}

	handler(args)
}

func (c *Cli) cmdLogin(args []string) {
	logger.Log.Info("Initiating login")
	c.authManager.Login(c.provider)
}

func (c *Cli) cmdGetProvider(args []string) {
	fmt.Printf("Provider: %s, Model: %s\n", c.provider, c.model)
}

func (c *Cli) cmdSetProvider(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: /set-provider <provider> <model>")
		fmt.Printf("Available providers: %s\n", strings.Join(c.llmManager.Providers(), ", "))
		return
	}

	provider := args[0]
	model := args[1]

	providers := c.llmManager.Providers()
	if !slices.Contains(providers, provider) {
		fmt.Printf("Invalid provider: %s\n", provider)
		fmt.Printf("Available providers: %s\n", strings.Join(providers, ", "))
		return
	}

	c.provider = provider
	c.model = model

	fmt.Printf("Provider set to: %s\n", provider)
	fmt.Printf("Model set to: %s\n", model)

	logger.Log.Info("Provider changed", "provider", provider)
}
