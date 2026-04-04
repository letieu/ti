package agent

import "time"

type Agent struct {
	Memory []string
}

func NewAgent() Agent {
	return Agent{}
}

func (a *Agent) Chat(input string) (<-chan AgentEvent, error) {
	ch := make(chan AgentEvent)
	go mainLoop(ch)
	return ch, nil
}

func mainLoop(ch chan AgentEvent) {
	defer close(ch)
	ch <- AgentEvent{Type: AgentEventToken, Data: "hihi1"}
	time.Sleep(1 * time.Second)

	ch <- AgentEvent{Type: AgentEventToken, Data: "hihi 2"}
	time.Sleep(1 * time.Second)

	ch <- AgentEvent{Data: "Done!", Type: AgentEventDone}
}
