package agent

type AgentEventType string

const (
	AgentEventToken AgentEventType = "token"
	AgentEventTool  AgentEventType = "tool"
	AgentEventDone  AgentEventType = "done"
)

type AgentEvent struct {
	Type AgentEventType
	Data string
}
