package agent

import (
	"context"
	"errors"

	"github.com/letieu/ti/internal/llm"
	"github.com/letieu/ti/internal/logger"
	"github.com/letieu/ti/internal/message"
	"github.com/letieu/ti/internal/tool"
)

var ErrNoLLMProvided = errors.New("no LLM provided for chat")

type Agent struct {
	Memory       []message.Message
	Tools        []tool.Tool
	SystemPrompt string
}

// NewAgent creates a new agent
func NewAgent() Agent {
	return Agent{
		SystemPrompt: "you are Nam",
	}
}

// Chat starts a conversation with the given LLM
// Each chat can use a different LLM provider
func (a *Agent) Chat(ctx context.Context, input string, streamer Streamer) (<-chan Event, error) {
	logger.Log.Info("Starting chat", "inputLength", len(input), "memorySize", len(a.Memory))

	a.addUserMsg(input)
	ch := make(chan Event)

	go a.mainLoop(ctx, ch, streamer)
	return ch, nil
}

func (a *Agent) addUserMsg(text string) {
	logger.Log.Debug("Adding user message", "text", text)
	a.Memory = append(a.Memory, &message.UserText{Text: text})
}

func (a *Agent) mainLoop(ctx context.Context, ch chan Event, lmm Streamer) {
	defer close(ch)
	logger.Log.Debug("Starting main loop", "toolsCount", len(a.Tools))
	logger.Log.Debug("Starting main loop", "MessagesCount", len(a.Memory))
	logger.Log.Debug("Starting main loop", "Messages", a.Memory)

	stream, err := lmm.Stream(ctx, llm.LlmContext{
		SystemPrompt: a.SystemPrompt,
		Messages:     a.Memory,
		Tools:        a.Tools,
	})
	if err != nil {
		logger.Log.Error("Failed to create LLM stream", "error", err)
		ch <- Error{
			Type: "stream_error",
			Msg:  err.Error(),
			Code: "STREAM_FAILED",
		}
		return
	}

	handler := newEventHandler(a, ch)

	for ev := range stream {
		select {
		case <-ctx.Done():
			logger.Log.Debug("Context cancelled, stopping main loop")
			return
		default:
			handler.processEvent(ev)
		}
	}

	logger.Log.Debug("Main loop completed", "finalMemorySize", len(a.Memory))
}
