package agent

// Event is the interface for all agent events sent to the CLI
type Event interface{}

// Text events - when the agent is generating a response
type TextStart struct{}
type TextDelta struct {
	Delta string
}
type TextEnd struct {
	Content string // Full accumulated text
}

// Thinking events - when the agent is processing/reasoning
type ThinkingStart struct{}
type ThinkingDelta struct {
	Delta string
}
type ThinkingEnd struct {
	Content string // Full thinking content
}

// Tool call events - when the agent is calling a tool
type ToolCallRequest struct {
	Name string
	Aggs any
}

// Tool result events - when a tool returns results
type ToolCallResult struct {
	ToolName string
	ToolID   string
	Result   map[string]any
}

// Error events
type Error struct {
	Type string
	Msg  string
	Code string
}

// Done event - signals the conversation turn is complete
type Done struct{}
