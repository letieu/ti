package event

import "github.com/letieu/ti/internal/agent"

type StreamMsg struct {
	Ev agent.AgentEvent
	Ch <-chan agent.AgentEvent
}

type StreamDoneMsg struct{}

type UserMsg struct {
	Text string
}
