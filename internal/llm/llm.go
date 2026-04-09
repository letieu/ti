package llm

import (
	"context"

	"github.com/letieu/ti/internal/llm/event"
)

type LlmContext struct {
	SystemPrompt string
	Messages     []Message
}

type Message struct {
	Role string
	Text string
}

type Lmm interface {
	GetName() string
	Stream(ctx context.Context, llmContext LlmContext) <-chan event.Event
	// TODO: should have somehow set API key, payload header, config, ...
}
