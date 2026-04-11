package main

import (
	"context"
	"fmt"

	"github.com/letieu/ti/internal/llm"
	"github.com/letieu/ti/internal/llm/antigravity"
)

func main() {
	an := antigravity.New(
		antigravity.GeminiOptions{
			APIKey:      "xx",
			ProjectID:   "xx",
			Model:       "gemini-3-flash",
			Temperature: 0.7,
			MaxTokens:   500,
		},
	)

	ch, err := an.Stream(context.TODO(), llm.LlmContext{
		SystemPrompt: "you are tieu",
	})

	if err != nil {
		fmt.Printf("Err %v", err)
		return
	}

	for e := range ch {
		fmt.Printf("e %v", e)
	}
}
