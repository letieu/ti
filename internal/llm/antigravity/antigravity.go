package antigravity

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/letieu/ti/internal/llm"
	"github.com/letieu/ti/internal/llm/event"
	"github.com/letieu/ti/internal/logger"
	"github.com/letieu/ti/internal/message"
	"github.com/letieu/ti/internal/tool"
)

type Antigravity struct {
	option GeminiOptions
}

type GeminiOptions struct {
	APIKey          string
	ProjectID       string
	Model           string
	Temperature     float64
	MaxTokens       int
	IncludeThoughts bool
	ThinkingLevel   string
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("status %d: %s", e.StatusCode, e.Message)
}

// New creates a new Antigravity LLM with the provided options
func New(opts GeminiOptions) Antigravity {
	// Apply defaults if not provided
	if opts.Temperature == 0 {
		opts.Temperature = 0.7
	}
	if opts.MaxTokens == 0 {
		opts.MaxTokens = 500
	}
	if opts.Model == "" {
		opts.Model = "gemini-3-flash"
	}

	return Antigravity{
		option: opts,
	}
}

func (a Antigravity) GetName() string {
	return "Antigravity"
}

func (a Antigravity) Stream(
	ctx context.Context, llmContext llm.LlmContext) (<-chan event.Event, error) {
	if len(llmContext.Messages) == 0 {
		return nil, fmt.Errorf("Should have messages")
	}

	eventCh := make(chan event.Event, 32)
	go a.streamEvents(ctx, eventCh, llmContext)
	return eventCh, nil
}

func (a Antigravity) streamEvents(ctx context.Context, eventCh chan<- event.Event, llmContext llm.LlmContext) {
	defer close(eventCh)

	logger.Log.Debug("Starting stream events", "model", a.option.Model)

	eventCh <- event.Start{}

	rawCh := make(chan RawPart, 32)
	errCh := make(chan error, 1)

	go a.runStreamProvider(ctx, rawCh, errCh, llmContext)
	a.adaptRawPartsToEvents(ctx, rawCh, eventCh, errCh)
}

func (a Antigravity) runStreamProvider(ctx context.Context, rawCh chan<- RawPart, errCh chan<- error, llmContext llm.LlmContext) {
	defer close(rawCh)
	err := streamFromAPI(ctx, a.option, llmContext, rawCh)
	if err != nil {
		errCh <- err
	}
}

type partType string

const (
	partTypeNone     partType = ""
	partTypeText     partType = "text"
	partTypeThinking partType = "thinking"
	partTypeFunction partType = "function"
)

func (a Antigravity) adaptRawPartsToEvents(ctx context.Context, rawCh <-chan RawPart, eventCh chan<- event.Event, errCh <-chan error) {
	currentPart := partTypeNone

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug("Context done in adaptRawPartsToEvents")
			a.endCurrentPart(currentPart, eventCh)
			return

		case err := <-errCh:
			if err != nil {
				logger.Log.Error("Error received from stream provider", "error", err)
				var apiErr *APIError

				if errors.As(err, &apiErr) {
					code := "UNKNOWN"
					if apiErr.StatusCode == 401 {
						code = "UNAUTHENTICATED"
					}

					eventCh <- event.Error{
						Type: "stream",
						Msg:  apiErr.Message,
						Code: code,
					}
				} else {
					eventCh <- event.Error{
						Type: "stream",
						Msg:  err.Error(),
						Code: "UNKNOWN",
					}
				}
			}

		case part, ok := <-rawCh:
			if !ok {
				// Stream finished, end current part
				logger.Log.Debug("Raw parts channel closed, ending stream")
				a.endCurrentPart(currentPart, eventCh)
				eventCh <- event.End{}
				return
			}

			// Determine the type of this part
			newPartType := a.detectPartType(part)

			// Handle state transitions
			if newPartType != currentPart {
				// End the previous part
				a.endCurrentPart(currentPart, eventCh)
				// Start the new part
				a.startNewPart(part, newPartType, eventCh)
				logger.Log.Debug("Part type changed", "from", currentPart, "to", newPartType)
				currentPart = newPartType
			}

			// Send delta for the current part
			a.sendPartDelta(part, currentPart, eventCh)
		}
	}
}

func (a Antigravity) detectPartType(part RawPart) partType {
	if _, hasText := part["text"]; hasText {
		if thought, _ := part["thought"]; thought == true {
			return partTypeThinking
		}

		return partTypeText
	}
	if _, hasThinking := part["thinking"]; hasThinking {
		return partTypeThinking
	}
	if _, hasFunction := part["functionCall"]; hasFunction {
		return partTypeFunction
	}
	return partTypeNone
}

func (a Antigravity) startNewPart(part RawPart, pt partType, eventCh chan<- event.Event) {
	logger.Log.Debug(fmt.Sprintf("Raw part: %v", part))

	switch pt {
	case partTypeText:
		eventCh <- event.TextStart{}
	case partTypeThinking:
		eventCh <- event.ThinkingStart{}
	case partTypeFunction:
		sig, _ := part["thoughtSignature"].(string)

		functionCall, _ := part["functionCall"].(map[string]any)

		id, _ := functionCall["id"].(string)
		name, _ := functionCall["name"].(string)
		args, _ := functionCall["args"].(map[string]any)

		eventCh <- event.FunctionStart{
			Id:               id,
			Name:             name,
			Args:             args,
			ThoughtSignature: sig,
		}
	}
}

func (a Antigravity) endCurrentPart(pt partType, eventCh chan<- event.Event) {
	switch pt {
	case partTypeText:
		eventCh <- event.TextEnd{}
	case partTypeThinking:
		eventCh <- event.ThinkingEnd{}
	case partTypeFunction:
		eventCh <- event.FunctionEnd{}
	}
}

func (a Antigravity) sendPartDelta(part RawPart, pt partType, eventCh chan<- event.Event) {
	switch pt {
	case partTypeText:
		if text, ok := part["text"].(string); ok {
			eventCh <- event.TextDelta{Delta: text}
		}
	case partTypeThinking:
		if thinking, ok := part["text"].(string); ok {
			eventCh <- event.ThinkingDelta{Delta: thinking}
		}
	case partTypeFunction:
		if funcCall, ok := part["functionCall"].(string); ok {
			eventCh <- event.FunctionDelta{Delta: funcCall}
		}
	}
}

type RawPart map[string]any

func streamFromAPI(
	ctx context.Context,
	opts GeminiOptions,
	llmContext llm.LlmContext,
	ch chan<- RawPart,
) error {
	// url := "https://cloudcode-pa.googleapis.com/v1internal:streamGenerateContent?alt=sse"
	url := "https://daily-cloudcode-pa.sandbox.googleapis.com/v1internal:streamGenerateContent?alt=sse"
	jsonBody := buildRequestBody(opts, llmContext)

	logger.Log.Debug("Preparing API request",
		"url", url,
		"model", opts.Model,
		"temperature", opts.Temperature,
		"maxTokens", opts.MaxTokens,
		"messageCount", len(llmContext.Messages),
	)

	// Log request body for debugging (be careful with sensitive data)
	logger.Log.Debug("Request body", "body", string(jsonBody))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		logger.Log.Error("Failed to create HTTP request", "error", err)
		return err
	}

	req.Header.Set("Authorization", "Bearer "+opts.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("User-Agent", "antigravity/1.18.4 darwin/arm64")

	logger.Log.Debug("Sending API request")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log.Error("HTTP request failed", "error", err)
		return err
	}
	defer res.Body.Close()

	logger.Log.Debug("Received API response", "statusCode", res.StatusCode)

	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		logger.Log.Error("API error response",
			"statusCode", res.StatusCode,
			"response", string(b),
		)
		return &APIError{
			StatusCode: res.StatusCode,
			Message:    string(b),
		}
	}

	return readSSEStream(ctx, res.Body, ch)
}

func readSSEStream(ctx context.Context, body io.Reader, ch chan<- RawPart) error {
	reader := bufio.NewReader(body)
	logger.Log.Debug("Starting to read SSE stream")
	partCount := 0

	for {
		resp, err := parseSSELine(reader)
		if err != nil {
			if err == io.EOF {
				logger.Log.Debug("SSE stream ended", "partsReceived", partCount)
				return nil
			}
			if ctx.Err() != nil {
				logger.Log.Debug("Context cancelled", "error", ctx.Err())
				return ctx.Err()
			}
			logger.Log.Error("Error parsing SSE line", "error", err)
			return err
		}

		parts := extractPartsFromResponse(resp)
		if len(parts) > 0 {
			partCount += len(parts)
			logger.Log.Debug("Extracted parts from response", "count", len(parts))
		}
		if err := sendPartsToChannel(ctx, parts, ch); err != nil {
			logger.Log.Error("Error sending parts to channel", "error", err)
			return err
		}
	}
}

func sendPartsToChannel(ctx context.Context, parts []any, ch chan<- RawPart) error {
	for _, p := range parts {
		part, ok := p.(map[string]any)
		if !ok {
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- part:
		}
	}
	return nil
}

func parseSSELine(reader *bufio.Reader) (map[string]any, error) {
	line, err := reader.ReadString('\n')
	logger.Log.Debug(line)

	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "data:") {
		return map[string]any{}, nil
	}

	jsonStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
	if jsonStr == "" {
		return map[string]any{}, nil
	}

	var response map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		return nil, fmt.Errorf("invalid JSON in SSE data: %w", err)
	}

	return response, nil
}

func extractPartsFromResponse(resp map[string]any) []any {
	if len(resp) == 0 {
		return []any{}
	}

	response, ok := resp["response"].(map[string]any)
	if !ok {
		return []any{}
	}

	candidates, ok := response["candidates"].([]any)
	if !ok || len(candidates) == 0 {
		return []any{}
	}

	candidate, ok := candidates[0].(map[string]any)
	if !ok {
		return []any{}
	}

	content, ok := candidate["content"].(map[string]any)
	if !ok {
		return []any{}
	}

	parts, ok := content["parts"].([]any)
	if !ok {
		return []any{}
	}

	return parts
}

func buildRequestBody(opts GeminiOptions, llmContext llm.LlmContext) []byte {
	contents := mapMessagesToRequestContents(llmContext.Messages)
	tools := mapRequestTools(llmContext.Tools)

	body := map[string]any{
		"project":     opts.ProjectID,
		"model":       opts.Model,
		"userAgent":   "antigravity",
		"requestType": "agent",
		"request": map[string]any{
			"contents": contents,
			"tools":    tools,
			"systemInstruction": map[string]any{
				"role": message.UserRole,
				"parts": []map[string]string{
					{"text": llmContext.SystemPrompt},
				},
			},
			"generationConfig": map[string]any{
				"temperature":     defaultFloat(opts.Temperature, 0.7),
				"maxOutputTokens": defaultInt(opts.MaxTokens, 1024),
				"thinkingConfig": map[string]any{
					"includeThoughts": opts.IncludeThoughts,
					"thinkingLevel":   defaultString(opts.ThinkingLevel, "LOW"),
				},
			},
			"toolConfig": map[string]any{
				"functionCallingConfig": map[string]string{
					"mode": "AUTO",
				},
			},
		},
	}

	jsonBody, _ := json.Marshal(body)
	return jsonBody
}

func defaultFloat(v, def float64) float64 {
	if v == 0 {
		return def
	}
	return v
}

func defaultString(v, def string) string {
	if v == "" {
		return def
	}

	return v
}

func defaultInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}

func mapRequestTools(tools []tool.ToolDescription) []map[string]any {
	requestTools := []map[string]any{}

	for _, tool := range tools {
		toolDef := map[string]any{
			"name":        tool.Name,
			"description": tool.Description,
			"parameters":  tool.Parameters,
		}

		requestTools = append(requestTools, map[string]any{
			"functionDeclarations": toolDef,
		})
	}

	return requestTools
}

func mapMessagesToRequestContents(messages []message.Message) []map[string]any {
	contents := []map[string]any{}

	for _, msg := range messages {
		part := map[string]any{}

		switch m := msg.(type) {
		case *message.UserText:
			part = map[string]any{
				"text": m.Text,
			}
		case *message.ModelText:
			part = map[string]any{
				"text": m.Text,
			}
		case *message.ModelThought:
			part = map[string]any{
				"text": m.Text, "thought": true,
			}
		case *message.ModelToolRequest:
			part = map[string]any{
				"thoughtSignature": m.Signature,
				"functionCall": map[string]any{
					"name": m.Name,
					"args": m.Args, // TODO: map args
					"id":   m.Id,
				},
			}
		case *message.UserToolResult:
			part = map[string]any{
				"functionResponse": map[string]any{
					"name":     m.Name,
					"response": m.Response,
				},
			}
		}

		content := map[string]any{
			"role":  msg.GetRole(),
			"parts": []map[string]any{part},
		}

		contents = append(contents, content)
	}

	return contents
}
