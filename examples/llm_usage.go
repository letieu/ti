package main

// This file demonstrates various ways to use the flexible LLM architecture

import (
	"context"
	"fmt"

	"github.com/letieu/ti/internal/agent"
	"github.com/letieu/ti/internal/llm"
	"github.com/letieu/ti/internal/llm/antigravity"
	"github.com/letieu/ti/internal/llm/event"
	"github.com/letieu/ti/internal/llm/provider"
)

// Example 1: Basic usage with default LLM
func exampleBasicUsage() {
	ag := agent.NewAgent()
	llm := provider.DefaultAntigravity()

	ctx := context.Background()
	ch, _ := ag.Chat(ctx, "Hello, how are you?", agent.ChatOptions{
		LLM: llm,
	})

	for ev := range ch {
		switch e := ev.(type) {
		case event.TextDelta:
			fmt.Print(e.Delta)
		}
	}
}

// Example 2: Custom Antigravity configuration
func exampleCustomAntigravity() {
	ag := agent.NewAgent()

	llm := provider.NewAntigravity(antigravity.GeminiOptions{
		APIKey:      "your-api-key",
		ProjectID:   "your-project-id",
		Model:       "gemini-3-flash",
		Temperature: 0.9,
		MaxTokens:   2000,
	})

	ctx := context.Background()
	ch, _ := ag.Chat(ctx, "Write a poem", agent.ChatOptions{
		LLM: llm,
	})

	handleChatEvents(ch)
}

// Example 3: Using different LLMs in the same conversation
func exampleMultipleLLMs() {
	ag := agent.NewAgent()
	ctx := context.Background()

	// First message with Gemini
	gemini := provider.NewAntigravity(antigravity.GeminiOptions{
		Model: "gemini-3-flash",
	})

	ch1, _ := ag.Chat(ctx, "What's 2+2?", agent.ChatOptions{
		LLM: gemini,
	})
	handleChatEvents(ch1)

	// Second message with Claude (same conversation memory!)
	claude := provider.NewAntigravity(antigravity.GeminiOptions{
		Model: "claude-sonnet-4-6",
	})

	ch2, _ := ag.Chat(ctx, "Now multiply that by 3", agent.ChatOptions{
		LLM: claude,
	})
	handleChatEvents(ch2)

	// The agent maintains conversation history across different LLMs
}

// Example 4: Override system prompt per chat
func exampleSystemPromptOverride() {
	ag := agent.NewAgent()
	ag.SystemPrompt = "You are a helpful assistant"

	llm := provider.DefaultAntigravity()
	ctx := context.Background()

	// Use default system prompt
	ch1, _ := ag.Chat(ctx, "Hello", agent.ChatOptions{
		LLM: llm,
	})
	handleChatEvents(ch1)

	// Override for a specific chat
	ch2, _ := ag.Chat(ctx, "Write some Go code", agent.ChatOptions{
		LLM:          llm,
		SystemPrompt: "You are an expert Go programmer",
	})
	handleChatEvents(ch2)
}

// Example 5: Dynamic model selection based on task
func exampleDynamicModelSelection(taskType string, input string) {
	ag := agent.NewAgent()
	ctx := context.Background()

	var lmm llm.Lmm

	switch taskType {
	case "code":
		// Use a fast model for code
		lmm = provider.NewAntigravity(antigravity.GeminiOptions{
			Model:       "gemini-3-flash",
			Temperature: 0.3, // Lower temperature for deterministic code
		})

	case "creative":
		// Use a more creative model
		lmm = provider.NewAntigravity(antigravity.GeminiOptions{
			Model:       "claude-sonnet-4-6",
			Temperature: 0.9, // Higher temperature for creativity
		})

	case "analysis":
		// Use a powerful model for analysis
		lmm = provider.NewAntigravity(antigravity.GeminiOptions{
			Model:     "gemini-pro",
			MaxTokens: 4000, // More tokens for detailed analysis
		})

	default:
		lmm = provider.DefaultAntigravity()
	}

	ch, _ := ag.Chat(ctx, input, agent.ChatOptions{
		LLM: lmm,
	})
	handleChatEvents(ch)
}

// Example 6: Load configuration from environment
func exampleLoadFromEnv() {
	ag := agent.NewAgent()

	// In real code, you'd use os.Getenv()
	llm := provider.NewAntigravity(antigravity.GeminiOptions{
		APIKey:      getEnv("GEMINI_API_KEY", ""),
		ProjectID:   getEnv("GEMINI_PROJECT_ID", ""),
		Model:       getEnv("GEMINI_MODEL", "gemini-3-flash"),
		Temperature: 0.7,
	})

	ctx := context.Background()
	ch, _ := ag.Chat(ctx, "Hello from env config!", agent.ChatOptions{
		LLM: llm,
	})
	handleChatEvents(ch)
}

// Example 7: Using provider factory with generic config
func exampleProviderFactory() {
	ag := agent.NewAgent()

	llm, err := provider.New(provider.Config{
		Type: provider.TypeAntigravity,
		Antigravity: &antigravity.GeminiOptions{
			APIKey:    "your-key",
			ProjectID: "your-project",
			Model:     "gemini-3-flash",
		},
	})

	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	ch, _ := ag.Chat(ctx, "Hello via factory!", agent.ChatOptions{
		LLM: llm,
	})
	handleChatEvents(ch)
}

// Helper functions

func handleChatEvents(ch <-chan event.Event) {
	for ev := range ch {
		switch e := ev.(type) {
		case event.Start:
			fmt.Println("Chat started")
		case event.TextStart:
			fmt.Print("Response: ")
		case event.TextDelta:
			fmt.Print(e.Delta)
		case event.TextEnd:
			fmt.Println()
		case event.Error:
			fmt.Printf("Error: %s\n", e.Msg)
		case event.End:
			fmt.Println("Chat ended")
		}
	}
}

func getEnv(key, defaultValue string) string {
	// In real code: return os.Getenv(key) or defaultValue
	return defaultValue
}

func main() {
	// Run examples
	fmt.Println("Example 1: Basic usage")
	exampleBasicUsage()

	fmt.Println("\nExample 2: Custom Antigravity")
	exampleCustomAntigravity()

	fmt.Println("\nExample 4: System prompt override")
	exampleSystemPromptOverride()

	fmt.Println("\nExample 5: Dynamic model selection")
	exampleDynamicModelSelection("code", "Write a Go function")

	fmt.Println("\nExample 6: Load from environment")
	exampleLoadFromEnv()

	fmt.Println("\nExample 7: Provider factory")
	exampleProviderFactory()
}
