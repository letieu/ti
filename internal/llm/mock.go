package llm

import (
	"context"
	"time"

	"github.com/letieu/ti/internal/auth"
	"github.com/letieu/ti/internal/llm/event"
)

type MockStreamer struct {
	Response string
}

func NewMockWithCreds(creds auth.OAuthCredentials) (Streamer, error) {
	return MockStreamer{"Hello"}, nil
}

func (m MockStreamer) Stream(ctx context.Context, llmContext LlmContext) (<-chan event.Event, error) {
	events := make(chan event.Event)

	go func() {
		defer close(events)

		events <- event.Start{}
		events <- event.TextStart{}

		for _, char := range m.Response {
			select {
			case <-ctx.Done():
				return
			case events <- event.TextDelta{Delta: string(char)}:
				time.Sleep(10 * time.Millisecond)
			}
		}

		events <- event.TextEnd{}
		events <- event.End{}
	}()

	return events, nil
}
