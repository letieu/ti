package main

import (
	"context"
	"fmt"
	"os"

	"github.com/letieu/ti/internal/llm"
	"github.com/letieu/ti/internal/llm/antigravity"
	"github.com/letieu/ti/internal/message"
)

func main() {
	an := antigravity.New(
		antigravity.GeminiOptions{
			APIKey:      os.Getenv("GEMINI_API_KEY"),
			ProjectID:   os.Getenv("GEMINI_PROJECT_ID"),
			Model:       "gemini-3-flash",
			Temperature: 0.7,
			MaxTokens:   1024,
			IncludeThoughts: true,
		},
	)

	ch, err := an.Stream(context.TODO(), llm.LlmContext{
		SystemPrompt: "you are a coding assistant",
		Messages: []message.Message{
			&message.UserText{
				Text: "hi",
			},
		},
	})

	if err != nil {
		fmt.Printf("Err %v", err)
		return
	}

	for e := range ch {
		fmt.Printf("%T %v \n", e, e)
	}
}
