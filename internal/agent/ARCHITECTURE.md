# LLM Architecture Guide

This guide explains the flexible LLM architecture that allows you to use different LLM providers per chat interaction.

## Architecture Overview

The agent is now designed to be **LLM-agnostic**. Instead of having a fixed LLM tied to an agent instance, each chat interaction can use a different LLM provider.

### Key Components:

1. **Agent** (`internal/agent/agent.go`) - Manages conversation memory and orchestrates chat flows
2. **LLM Interface** (`internal/llm/llm.go`) - Defines the contract all LLM providers must implement
3. **LLM Providers** (e.g., `internal/llm/antigravity/`) - Specific implementations for different LLM services
4. **Provider Factory** (`internal/llm/provider/provider.go`) - Helps create and configure LLM instances

## Design Philosophy

**Agent focuses on conversation management, not LLM specifics:**
- Maintains conversation history (memory)
- Orchestrates message flow
- Handles system prompts
- **Does NOT** own a specific LLM instance

**LLM providers are injected per chat:**
- Each `Chat()` call receives an LLM instance
- Easy to switch between providers mid-conversation
- Different chats can use different LLMs

## Basic Usage

### Simple Example

```go
import (
    "context"
    "github.com/letieu/ti/internal/agent"
    "github.com/letieu/ti/internal/llm/provider"
)

func main() {
    // Create agent (no LLM attached)
    ag := agent.NewAgent()
    
    // Create LLM provider
    llm := provider.DefaultAntigravity()
    
    // Start chat with this LLM
    ctx := context.Background()
    ch, _ := ag.Chat(ctx, "Hello!", agent.ChatOptions{
        LLM: llm,
    })
    
    // Process events
    for event := range ch {
        // Handle events...
    }
}
```

### Using Different LLMs Per Chat

```go
import (
    "github.com/letieu/ti/internal/agent"
    "github.com/letieu/ti/internal/llm/antigravity"
    "github.com/letieu/ti/internal/llm/provider"
)

func example() {
    ag := agent.NewAgent()
    
    // First message uses Antigravity with Gemini
    gemini := provider.NewAntigravity(antigravity.GeminiOptions{
        APIKey:    "your-key",
        ProjectID: "your-project",
        Model:     "gemini-3-flash",
    })
    
    ag.Chat(ctx, "Analyze this data...", agent.ChatOptions{
        LLM: gemini,
    })
    
    // Second message uses Antigravity with Claude
    claude := provider.NewAntigravity(antigravity.GeminiOptions{
        APIKey:    "your-key",
        ProjectID: "your-project",
        Model:     "claude-sonnet-4-6",
    })
    
    ag.Chat(ctx, "Continue the analysis...", agent.ChatOptions{
        LLM: claude,
    })
    
    // The conversation memory is preserved across different LLMs!
}
```

### Customizing System Prompt Per Chat

```go
// Agent has a default system prompt
ag := agent.NewAgent()
ag.SystemPrompt = "You are a helpful assistant"

// Use default system prompt
ag.Chat(ctx, "Hello", agent.ChatOptions{
    LLM: llm,
})

// Override system prompt for specific chat
ag.Chat(ctx, "Write code", agent.ChatOptions{
    LLM:          llm,
    SystemPrompt: "You are an expert programmer",
})
```

## LLM Provider Configuration

### Using the Provider Factory

```go
import "github.com/letieu/ti/internal/llm/provider"

// Quick default configuration
llm := provider.DefaultAntigravity()

// Custom configuration
llm := provider.NewAntigravity(antigravity.GeminiOptions{
    APIKey:      "your-api-key",
    ProjectID:   "your-project-id",
    Model:       "gemini-3-flash",
    Temperature: 0.9,
    MaxTokens:   2000,
})

// Generic provider creation
llm, err := provider.New(provider.Config{
    Type: provider.TypeAntigravity,
    Antigravity: &antigravity.GeminiOptions{
        APIKey:    "your-key",
        ProjectID: "your-project",
        Model:     "gemini-3-flash",
    },
})
```

### Loading from Environment

```go
import (
    "os"
    "github.com/letieu/ti/internal/llm/antigravity"
    "github.com/letieu/ti/internal/llm/provider"
)

func createLLMFromEnv() llm.Lmm {
    return provider.NewAntigravity(antigravity.GeminiOptions{
        APIKey:      os.Getenv("GEMINI_API_KEY"),
        ProjectID:   os.Getenv("GEMINI_PROJECT_ID"),
        Model:       getEnvOrDefault("GEMINI_MODEL", "gemini-3-flash"),
        Temperature: 0.7,
    })
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

## Adding New LLM Providers

### Step 1: Implement the Interface

Create a new package (e.g., `internal/llm/openai/`):

```go
package openai

import (
    "context"
    "github.com/letieu/ti/internal/llm"
    "github.com/letieu/ti/internal/llm/event"
)

type OpenAI struct {
    options Options
}

type Options struct {
    APIKey      string
    Model       string
    Temperature float64
    MaxTokens   int
}

func New(opts Options) OpenAI {
    // Apply defaults
    if opts.Model == "" {
        opts.Model = "gpt-4"
    }
    return OpenAI{options: opts}
}

func (o OpenAI) GetName() string {
    return "OpenAI"
}

func (o OpenAI) Stream(ctx context.Context, llmContext llm.LlmContext) <-chan event.Event {
    eventCh := make(chan event.Event)
    // Implement OpenAI streaming...
    return eventCh
}
```

### Step 2: Add to Provider Factory

Update `internal/llm/provider/provider.go`:

```go
const (
    TypeAntigravity Type = "antigravity"
    TypeOpenAI      Type = "openai"  // Add new type
)

type Config struct {
    Type        Type
    Antigravity *antigravity.GeminiOptions
    OpenAI      *openai.Options  // Add new config
}

func New(cfg Config) (llm.Lmm, error) {
    switch cfg.Type {
    case TypeAntigravity:
        // ...existing code...
    case TypeOpenAI:
        if cfg.OpenAI == nil {
            return nil, fmt.Errorf("openai config required")
        }
        return openai.New(*cfg.OpenAI), nil
    // ...
    }
}

func NewOpenAI(opts openai.Options) llm.Lmm {
    return openai.New(opts)
}
```

### Step 3: Use It

```go
import (
    "github.com/letieu/ti/internal/llm/openai"
    "github.com/letieu/ti/internal/llm/provider"
)

llm := provider.NewOpenAI(openai.Options{
    APIKey: "sk-...",
    Model:  "gpt-4",
})

ag.Chat(ctx, "Hello", agent.ChatOptions{LLM: llm})
```

## Testing

### Mock LLM for Testing

```go
type MockLLM struct {
    Response string
}

func (m MockLLM) GetName() string {
    return "Mock"
}

func (m MockLLM) Stream(ctx context.Context, llmContext llm.LlmContext) <-chan event.Event {
    ch := make(chan event.Event)
    go func() {
        defer close(ch)
        ch <- event.Start{}
        ch <- event.TextStart{}
        ch <- event.TextDelta{Delta: m.Response}
        ch <- event.TextEnd{}
        ch <- event.End{}
    }()
    return ch
}

// In tests:
func TestAgent(t *testing.T) {
    ag := agent.NewAgent()
    mock := MockLLM{Response: "Test response"}
    
    ch, _ := ag.Chat(ctx, "test", agent.ChatOptions{LLM: mock})
    // Assert on events...
}
```

## Architecture Benefits

1. **Flexibility**: Use different LLMs for different tasks in the same conversation
2. **Testability**: Easy to inject mock LLMs for testing
3. **Separation of Concerns**: Agent manages conversation, LLMs handle generation
4. **Provider-Specific Options**: Each LLM can have its own configuration structure
5. **Runtime Switching**: Change LLMs dynamically based on user preferences or task requirements
6. **Memory Preservation**: Conversation history is maintained regardless of which LLM you use

## Real-World Scenarios

### Model Selection Based on Task

```go
func smartChat(ag *agent.Agent, task string, input string) {
    var llm llm.Lmm
    
    switch task {
    case "code":
        // Use a code-specialized model
        llm = provider.NewAntigravity(antigravity.GeminiOptions{
            Model: "gemini-code-pro",
        })
    case "creative":
        // Use a creative model
        llm = provider.NewAntigravity(antigravity.GeminiOptions{
            Model: "claude-sonnet-4-6",
            Temperature: 0.9,
        })
    default:
        llm = provider.DefaultAntigravity()
    }
    
    ag.Chat(ctx, input, agent.ChatOptions{LLM: llm})
}
```

### User-Selectable Providers

```go
func chatWithUserChoice(ag *agent.Agent, providerName string, input string) {
    var llm llm.Lmm
    
    switch providerName {
    case "gemini":
        llm = provider.NewAntigravity(antigravity.GeminiOptions{
            Model: "gemini-3-flash",
        })
    case "claude":
        llm = provider.NewAntigravity(antigravity.GeminiOptions{
            Model: "claude-sonnet-4-6",
        })
    default:
        llm = provider.DefaultAntigravity()
    }
    
    ag.Chat(ctx, input, agent.ChatOptions{LLM: llm})
}
```
