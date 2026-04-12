package cli

import (
	"context"
	"fmt"

	"github.com/letieu/ti/internal/auth"
	"github.com/letieu/ti/internal/llm"
	"github.com/letieu/ti/internal/llm/antigravity"
	"github.com/letieu/ti/internal/llm/event"
)

type LlmManager struct {
	streamer  llm.Streamer
	providers []string
}

func NewLlmManager() *LlmManager {
	providers := []string{"antigravity", "mock"}
	return &LlmManager{providers: providers}
}

// SetProvider configures the LLM provider with credentials
func (m *LlmManager) SetProvider(provider string, model string, creds auth.OAuthCredentials) error {
	switch provider {
	case "antigravity":
		projectID := creds.Metadata["project_id"]
		if projectID == "" {
			return fmt.Errorf("project_id not found in credentials metadata")
		}

		m.streamer = antigravity.New(antigravity.GeminiOptions{
			APIKey:          creds.Access,
			ProjectID:       projectID,
			Model:           model,
			Temperature:     0.7,
			MaxTokens:       2048,
			IncludeThoughts: true,
		})
		return nil

	case "mock":
		m.streamer = llm.MockStreamer{Response: "Hello from mock LLM"}
		return nil

	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

// Providers returns the list of available providers
func (m *LlmManager) Providers() []string {
	return m.providers
}

// Stream delegates to the configured streamer
func (m *LlmManager) Stream(ctx context.Context, llmContext llm.LlmContext) (<-chan event.Event, error) {
	if m.streamer == nil {
		return nil, fmt.Errorf("no LLM provider configured")
	}
	return m.streamer.Stream(ctx, llmContext)
}
