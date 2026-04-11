package agent

import (
	"context"

	"github.com/letieu/ti/internal/llm"
	"github.com/letieu/ti/internal/llm/event"
)

type Streamer interface {
	Stream(ctx context.Context, llmContext llm.LlmContext) (<-chan event.Event, error)
}
