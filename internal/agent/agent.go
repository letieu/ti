package agent

import (
	"context"
	"errors"

	"github.com/letieu/ti/internal/llm"
	"github.com/letieu/ti/internal/llm/event"
)

var ErrNoLLMProvided = errors.New("no LLM provided for chat")

type Agent struct {
	Memory       []llm.Message
	SystemPrompt string
}

// NewAgent creates a new agent
func NewAgent() Agent {
	return Agent{
		SystemPrompt: "you are Nam",
	}
}

// ChatOptions contains options for a chat interaction
type ChatOptions struct {
	LLM          llm.Lmm
	SystemPrompt string // Optional: override agent's default system prompt
}

// Chat starts a conversation with the given LLM
// Each chat can use a different LLM provider
func (a *Agent) Chat(ctx context.Context, input string, opts ChatOptions) (<-chan event.Event, error) {
	if opts.LLM == nil {
		return nil, ErrNoLLMProvided
	}

	a.addUserMsg(input)
	ch := make(chan event.Event)

	systemPrompt := a.SystemPrompt
	if opts.SystemPrompt != "" {
		systemPrompt = opts.SystemPrompt
	}

	llmContext := llm.LlmContext{
		SystemPrompt: systemPrompt,
		Messages:     a.Memory,
	}

	go a.mainLoop(ctx, ch, opts.LLM, llmContext)
	return ch, nil
}

func (a *Agent) addUserMsg(text string) {
	a.Memory = append(a.Memory, llm.Message{Role: "user", Text: text})
}

func (a *Agent) mainLoop(ctx context.Context, ch chan event.Event, lmm llm.Lmm, llmContext llm.LlmContext) {
	defer close(ch)
	stream := lmm.Stream(ctx, llmContext)

	for e := range stream {
		select {
		case <-ctx.Done():
			return
		case ch <- e:
		}
	}
}
