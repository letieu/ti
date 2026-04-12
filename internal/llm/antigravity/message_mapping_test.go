package antigravity

import (
	"testing"

	"github.com/letieu/ti/internal/message"
)

func TestMapMessagesToRequestWithPointers(t *testing.T) {
	messages := []message.Message{
		&message.UserText{Text: "Hello"},
		&message.ModelText{Text: "Hi there"},
		&message.ModelThought{Text: "Let me think"},
		&message.UserText{Text: "How are you?"},
	}

	contents := mapMessagesToRequestContents(messages)

	if len(contents) != 4 {
		t.Errorf("Expected 4 contents, got %d", len(contents))
	}

	// Verify first message (UserText)
	if contents[0]["role"] != message.UserRole {
		t.Errorf("Expected first message role to be UserRole, got %v", contents[0]["role"])
	}
	parts0 := contents[0]["parts"].([]map[string]any)
	if parts0[0]["text"] != "Hello" {
		t.Errorf("Expected first message text to be 'Hello', got %v", parts0[0]["text"])
	}

	// Verify second message (ModelText)
	if contents[1]["role"] != message.ModelRole {
		t.Errorf("Expected second message role to be ModelRole, got %v", contents[1]["role"])
	}
	parts1 := contents[1]["parts"].([]map[string]any)
	if parts1[0]["text"] != "Hi there" {
		t.Errorf("Expected second message text to be 'Hi there', got %v", parts1[0]["text"])
	}

	// Verify third message (ModelThought)
	if contents[2]["role"] != message.ModelRole {
		t.Errorf("Expected third message role to be ModelRole, got %v", contents[2]["role"])
	}
	parts2 := contents[2]["parts"].([]map[string]any)
	if parts2[0]["text"] != "Let me think" {
		t.Errorf("Expected third message text to be 'Let me think', got %v", parts2[0]["text"])
	}
	if parts2[0]["thought"] != true {
		t.Errorf("Expected third message to have thought=true, got %v", parts2[0]["thought"])
	}

	// Verify fourth message (UserText)
	if contents[3]["role"] != message.UserRole {
		t.Errorf("Expected fourth message role to be UserRole, got %v", contents[3]["role"])
	}
	parts3 := contents[3]["parts"].([]map[string]any)
	if parts3[0]["text"] != "How are you?" {
		t.Errorf("Expected fourth message text to be 'How are you?', got %v", parts3[0]["text"])
	}
}

func TestMapMessagesToRequestWithToolMessages(t *testing.T) {
	messages := []message.Message{
		&message.UserText{Text: "Search for golang"},
		&message.ModelToolRequest{
			Name:      "search",
			Args:      map[string]string{"query": "golang"},
			Id:        "tool_123",
			Signature: "sig_abc",
		},
		&message.UserToolResult{
			Name:     "search",
			Response: map[string]string{"result": "Go is a programming language"},
		},
	}

	contents := mapMessagesToRequestContents(messages)

	if len(contents) != 3 {
		t.Errorf("Expected 3 contents, got %d", len(contents))
	}

	// Verify tool request
	parts1 := contents[1]["parts"].([]map[string]any)
	if parts1[0]["thoughtSignature"] != "sig_abc" {
		t.Errorf("Expected thoughtSignature to be 'sig_abc', got %v", parts1[0]["thoughtSignature"])
	}
	funcCall := parts1[0]["functionCall"].(map[string]any)
	if funcCall["name"] != "search" {
		t.Errorf("Expected function name to be 'search', got %v", funcCall["name"])
	}
	if funcCall["id"] != "tool_123" {
		t.Errorf("Expected function id to be 'tool_123', got %v", funcCall["id"])
	}

	// Verify tool result
	parts2 := contents[2]["parts"].([]map[string]any)
	funcResp := parts2[0]["functionResponse"].(map[string]any)
	if funcResp["name"] != "search" {
		t.Errorf("Expected function response name to be 'search', got %v", funcResp["name"])
	}
}
