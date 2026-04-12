package tool

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

const (
	DEFAULT_MAX_LINES = 2000
	DEFAULT_MAX_BYTES = 50 * 1024 // 50 Kb
)

type ReadTool struct{}

func (t ReadTool) Execute(ctx context.Context, input D) (D, error) {
	path, ok := input["path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid field: path. %v", input)
	}
	offset, _ := input["offset"].(float64)
	limit, _ := input["limit"].(float64)

	content, err := readFile(path, int(offset), int(limit))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return D{
		"content": content,
	}, nil
}

func (t ReadTool) Desc() *ToolDescription {
	return &ToolDescription{
		Name:        "read",
		Description: fmt.Sprintf("Read the contents of a file. Supports text files, text files, output is truncated to %d lines or %dKB (whichever is hit first). Use offset/limit for large files.", DEFAULT_MAX_LINES, DEFAULT_MAX_BYTES),
		Parameters: D{
			"type": "object",
			"properties": D{
				"label": D{
					"type":        "string",
					"description": "Brief description of what you're reading and why (shown to user)",
				},
				"path": D{
					"type":        "string",
					"description": "Path to the file to read (relative or absolute)",
				},
				"offset": D{
					"type":        "number",
					"description": "Line number to start reading from (1-indexed)",
				},
				"limit": D{
					"type":        "number",
					"description": "Maximum number of lines to read",
				},
			},
			"required": []string{"label", "path"},
		},
	}
}

func readFile(path string, offset, limit int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if limit <= 0 {
		limit = DEFAULT_MAX_LINES
	}

	scanner := bufio.NewScanner(f)
	var (
		lines      []string
		lineNum    int
		totalBytes int
	)

	for scanner.Scan() {
		lineNum++

		if offset > 0 && lineNum < offset {
			continue
		}

		line := scanner.Text()
		totalBytes += len(line) + 1 // +1 for newline

		if totalBytes > DEFAULT_MAX_BYTES {
			lines = append(lines, fmt.Sprintf("... truncated: exceeded %dKB limit", DEFAULT_MAX_BYTES/1024))
			break
		}

		lines = append(lines, line)

		if len(lines) >= limit {
			lines = append(lines, fmt.Sprintf("... truncated: exceeded %d line limit", limit))
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}
