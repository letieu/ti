package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/letieu/ti/internal/llm"
	"github.com/letieu/ti/internal/logger"
	"github.com/letieu/ti/internal/message"
	"github.com/letieu/ti/internal/tool"
)

var ErrNoLLMProvided = errors.New("no LLM provided for chat")

type Agent struct {
	Memory              []message.Message
	Tools               map[string]tool.Tool
	SystemPrompt        string
	DebugMemoryDumpPath string
}

func NewAgent(systemPrompt string, tools []tool.Tool) Agent {
	toolsReg := make(map[string]tool.Tool)

	for _, tool := range tools {
		toolsReg[tool.Desc().Name] = tool
	}

	return Agent{
		SystemPrompt: systemPrompt,
		Tools:        toolsReg,
	}
}

func (a *Agent) dumpMemory() {
	if a.DebugMemoryDumpPath == "" {
		return
	}

	data, err := json.MarshalIndent(a.Memory, "", "  ")
	if err != nil {
		logger.Log.Error("Failed to marshal memory for debug dump", "error", err)
		return
	}

	err = os.WriteFile(a.DebugMemoryDumpPath, data, 0644)
	if err != nil {
		logger.Log.Error("Failed to write memory dump to file", "path", a.DebugMemoryDumpPath, "error", err)
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

	handler := newEventHandler(a, ch)

	// Agent loop: Keep calling LLM until it stops requesting tools
	for {
		if ctx.Err() != nil {
			logger.Log.Debug("Context cancelled, stopping main loop")
			break
		}

		logger.Log.Debug("____TURN___START__")
		logger.Log.Debug(fmt.Sprintf("mem %v", a.Memory))
		shouldContinue, err := a.processTurn(ctx, handler, lmm)
		logger.Log.Debug("____TURN___NED__")
		a.dumpMemory()

		if err != nil {
			logger.Log.Error("Error processing LLM turn", "error", err)
			ch <- Error{
				Type: "stream_error",
				Msg:  err.Error(),
				Code: "STREAM_FAILED",
			}
			break
		}

		if !shouldContinue {
			logger.Log.Debug("LLM finished without calling tools")
			break
		}

		logger.Log.Debug("Tool(s) executed, continuing agent loop")
	}

	logger.Log.Debug("Main loop completed", "finalMemorySize", len(a.Memory))
}

func (a *Agent) processTurn(ctx context.Context, handler *eventHandler, lmm Streamer) (bool, error) {
	stream, err := lmm.Stream(ctx, llm.LlmContext{
		SystemPrompt: a.SystemPrompt,
		Messages:     a.Memory,
		Tools:        DescList(a.Tools),
	})
	if err != nil {
		return false, err
	}

	// Process all events from this LLM turn
	toolCalled := false
	for ev := range stream {
		// Check for context cancellation
		if ctx.Err() != nil {
			logger.Log.Debug("Context cancelled during event processing")
			return false, ctx.Err()
		}

		toolRequest := handler.processEvent(ev)
		if toolRequest.Name != "" {
			logger.Log.Debug("Tool requested", "name", toolRequest.Name)
			err := a.handleToolRequest(ctx, handler, toolRequest)
			if err != nil {
				return false, err
			}

			toolCalled = true
		}
	}

	// Return whether we should continue the agent loop
	return toolCalled, nil
}

func (a *Agent) handleToolRequest(ctx context.Context, handler *eventHandler, toolRequest message.ModelToolRequest) error {
	a.Memory = append(a.Memory, &toolRequest)
	tool := a.Tools[toolRequest.Name]

	if tool.Desc().Name == "" {
		fmt.Printf("invalid tool")
		return fmt.Errorf("invalid tool")
	}

	result, err := tool.Execute(ctx, toolRequest.Args)
	if err != nil {
		return fmt.Errorf("fail to Execute tool %v", toolRequest)
	}

	a.Memory = append(a.Memory, &message.UserToolResult{
		Name:     toolRequest.Name,
		Response: result,
	})

	// Emit tool result event
	handler.eventChan <- ToolCallResult{
		ToolName: toolRequest.Name,
		ToolID:   toolRequest.Id,
		Result:   result,
	}

	return nil
}

func DescList(tools map[string]tool.Tool) []tool.ToolDescription {
	descs := make([]tool.ToolDescription, 0, len(tools))
	for _, t := range tools {
		descs = append(descs, *t.Desc())
	}
	return descs
}
