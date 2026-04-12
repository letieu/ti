package llm

import (
	"context"

	"github.com/letieu/ti/internal/llm/event"
)

type Streamer interface {
	Stream(ctx context.Context, llmContext LlmContext) (<-chan event.Event, error)
}
