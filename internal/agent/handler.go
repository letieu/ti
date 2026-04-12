package agent

import (
	"fmt"

	"github.com/letieu/ti/internal/llm/event"
	"github.com/letieu/ti/internal/logger"
	"github.com/letieu/ti/internal/message"
)

// eventHandler manages state during event processing
type eventHandler struct {
	agent             *Agent
	eventChan         chan Event
	textContent       string
	thinkingContent   string
	currentTextMsg    *message.ModelText
	currentThoughtMsg *message.ModelThought
}

func newEventHandler(agent *Agent, ch chan Event) *eventHandler {
	return &eventHandler{
		agent:     agent,
		eventChan: ch,
	}
}

func (h *eventHandler) handleTextStart() {
	h.textContent = ""
	newTextMsg := &message.ModelText{Text: ""}
	h.agent.Memory = append(h.agent.Memory, newTextMsg)
	h.currentTextMsg = newTextMsg
	logger.Log.Debug("Started text message in memory")
	h.eventChan <- TextStart{}
}

func (h *eventHandler) handleTextDelta(delta string) {
	h.textContent += delta
	// Update the last message in memory
	if len(h.agent.Memory) > 0 {
		if textMsg, ok := h.agent.Memory[len(h.agent.Memory)-1].(*message.ModelText); ok {
			textMsg.Text += delta
			logger.Log.Debug("Updated text in memory", "currentLength", len(textMsg.Text))
		}
	}
	h.eventChan <- TextDelta{Delta: delta}
}

func (h *eventHandler) handleTextEnd() {
	h.eventChan <- TextEnd{Content: h.textContent}
	logger.Log.Debug(fmt.Sprintf("Completed text %v", h.agent.Memory[len(h.agent.Memory)-1]))
	h.currentTextMsg = nil
}

func (h *eventHandler) handleThinkingStart() {
	h.thinkingContent = ""
	newThoughtMsg := &message.ModelThought{Text: ""}
	h.agent.Memory = append(h.agent.Memory, newThoughtMsg)
	h.currentThoughtMsg = newThoughtMsg
	logger.Log.Debug("Started thought message in memory")
	h.eventChan <- ThinkingStart{}
}

func (h *eventHandler) handleThinkingDelta(delta string) {
	h.thinkingContent += delta
	// Update the last message in memory
	if len(h.agent.Memory) > 0 {
		if thoughtMsg, ok := h.agent.Memory[len(h.agent.Memory)-1].(*message.ModelThought); ok {
			thoughtMsg.Text += delta
			logger.Log.Debug("Updated thought in memory", "currentLength", len(thoughtMsg.Text))
		}
	}
	h.eventChan <- ThinkingDelta{Delta: delta}
}

func (h *eventHandler) handleThinkingEnd() {
	logger.Log.Debug("Completed thought message in memory", "finalLength", len(h.thinkingContent))
	h.eventChan <- ThinkingEnd{Content: h.thinkingContent}
	h.currentThoughtMsg = nil
}

func (h *eventHandler) handleError(err event.Error) {
	h.eventChan <- Error{
		Type: err.Type,
		Msg:  err.Msg,
		Code: err.Code,
	}
}

func (h *eventHandler) handleEnd() {
	h.eventChan <- Done{}
}

func (h *eventHandler) processEvent(ev event.Event) message.ModelToolRequest {
	logger.Log.Debug(fmt.Sprintf("Stream Event: %T %v", ev, ev))

	switch e := ev.(type) {
	case event.TextStart:
		h.handleTextStart()
	case event.TextDelta:
		h.handleTextDelta(e.Delta)
	case event.TextEnd:
		h.handleTextEnd()
	case event.ThinkingStart:
		h.handleThinkingStart()
	case event.ThinkingDelta:
		h.handleThinkingDelta(e.Delta)
	case event.ThinkingEnd:
		h.handleThinkingEnd()
	case event.FunctionStart:
		return h.handleFunction(e)
	case event.Error:
		h.handleError(e)
	case event.End:
		h.handleEnd()
		// default:
		// 	fmt.Printf("%T %v \n", ev, ev)
	}

	return message.ModelToolRequest{}
}

func (h *eventHandler) handleFunction(ev event.FunctionStart) message.ModelToolRequest {
	newTool := message.ModelToolRequest{
		Name:      ev.Name,
		Id:        ev.Id,
		Args:      ev.Args,
		Signature: ev.ThoughtSignature,
	}
	h.eventChan <- ToolCallRequest{
		Name: ev.Name,
		Aggs: ev.Args,
	}

	return newTool
}
