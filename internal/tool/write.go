package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type WriteTool struct{}

func (t WriteTool) Execute(ctx context.Context, input D) (D, error) {
	path, ok := input["path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid field: path. %v", input)
	}

	content, ok := input["content"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid field: content. %v", input)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	// Write file (overwrite if exists)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}

	return D{
		"path":          path,
		"bytes_written": fileInfo.Size(),
		"success":       true,
	}, nil
}

func (t WriteTool) Desc() *ToolDescription {
	return &ToolDescription{
		Name:        "write",
		Description: "Write content to a file. Creates the file if it doesn't exist, creates parent directories if needed, and overwrites existing files.",
		Parameters: D{
			"type": "object",
			"properties": D{
				"label": D{
					"type":        "string",
					"description": "Brief description of what you're writing and why (shown to user)",
				},
				"path": D{
					"type":        "string",
					"description": "Path to the file to write (relative or absolute)",
				},
				"content": D{
					"type":        "string",
					"description": "Content to write to the file",
				},
			},
			"required": []string{"label", "path", "content"},
		},
	}
}
