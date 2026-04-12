package tool

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type EditTool struct{}

func (t EditTool) Execute(ctx context.Context, input D) (D, error) {
	path, ok := input["path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid field: path. %v", input)
	}

	oldContent, ok := input["old"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid field: old. %v", input)
	}

	newContent, ok := input["new"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid field: new. %v", input)
	}

	// Read current file content
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	currentContent := string(fileBytes)

	// Find and replace old content with new content
	if !strings.Contains(currentContent, oldContent) {
		return nil, fmt.Errorf("old content not found in file")
	}

	// Replace the old content with new content
	updatedContent := strings.Replace(currentContent, oldContent, newContent, 1)

	// Write the updated content back to file
	if err := os.WriteFile(path, []byte(updatedContent), 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	// Generate diff
	diff := generateDiff(currentContent, updatedContent)

	return D{
		"path":    path,
		"diff":    diff,
		"success": true,
	}, nil
}

func (t EditTool) Desc() *ToolDescription {
	return &ToolDescription{
		Name:        "edit",
		Description: "Edit a file by replacing old content with new content. Shows a diff of the changes made.",
		Parameters: D{
			"type": "object",
			"properties": D{
				"label": D{
					"type":        "string",
					"description": "Brief description of what you're editing and why (shown to user)",
				},
				"path": D{
					"type":        "string",
					"description": "Path to the file to edit (relative or absolute)",
				},
				"old": D{
					"type":        "string",
					"description": "The old content to find and replace",
				},
				"new": D{
					"type":        "string",
					"description": "The new content to replace with",
				},
			},
			"required": []string{"label", "path", "old", "new"},
		},
	}
}

// generateDiff creates a unified diff between old and new content using the diff command
func generateDiff(oldText, newText string) string {
	// Create temporary files for diff
	oldFile, err := os.CreateTemp("", "old-*.txt")
	if err != nil {
		return fmt.Sprintf("Error creating temp file: %v", err)
	}
	defer os.Remove(oldFile.Name())
	defer oldFile.Close()

	newFile, err := os.CreateTemp("", "new-*.txt")
	if err != nil {
		return fmt.Sprintf("Error creating temp file: %v", err)
	}
	defer os.Remove(newFile.Name())
	defer newFile.Close()

	// Write content to temp files
	if _, err := oldFile.WriteString(oldText); err != nil {
		return fmt.Sprintf("Error writing old content: %v", err)
	}
	if _, err := newFile.WriteString(newText); err != nil {
		return fmt.Sprintf("Error writing new content: %v", err)
	}

	// Close files before running diff
	oldFile.Close()
	newFile.Close()

	// Run diff command with color output
	cmd := exec.Command("diff", "-u", "--color=always", oldFile.Name(), newFile.Name())
	output, _ := cmd.CombinedOutput()

	// diff returns exit code 1 when files differ, which is expected
	if len(output) == 0 {
		return "No changes"
	}

	// Remove the temp file paths from output and keep only the actual diff
	lines := strings.Split(string(output), "\n")
	var result strings.Builder
	result.WriteString("\033[1mDiff:\033[0m\n")

	// Skip first 2 lines (---/+++ headers with temp file paths)
	for i, line := range lines {
		if i < 2 {
			continue
		}
		result.WriteString(line + "\n")
	}

	return result.String()
}
