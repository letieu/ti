package llm

import (
	"github.com/letieu/ti/internal/message"
	"github.com/letieu/ti/internal/tool"
)

type LlmContext struct {
	SystemPrompt string
	Messages     []message.Message
	Tools        []tool.ToolDescription
}
