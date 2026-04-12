package tool

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	DEFAULT_TIMEOUT = 30 * time.Second
)

type BashTool struct{}

func (t BashTool) Execute(ctx context.Context, input D) (D, error) {
	command, ok := input["command"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid field: command. %v", input)
	}

	// Optional timeout parameter
	timeout := DEFAULT_TIMEOUT
	if timeoutSecs, ok := input["timeout"].(float64); ok && timeoutSecs > 0 {
		timeout = time.Duration(timeoutSecs) * time.Second
	}

	// Optional working directory
	workdir, _ := input["workdir"].(string)

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute the command
	cmd := exec.CommandContext(execCtx, "bash", "-c", command)
	if workdir != "" {
		cmd.Dir = workdir
	}

	output, err := cmd.CombinedOutput()
	result := D{
		"stdout": string(output),
	}

	if err != nil {
		// Check if it was a timeout
		if execCtx.Err() == context.DeadlineExceeded {
			result["exit_code"] = -1
			return result, nil
		}
		// Include exit code information
		if exitErr, ok := err.(*exec.ExitError); ok {
			result["exit_code"] = exitErr.ExitCode()
			return result, nil
		}
		result["exit_code"] = -1
		result["error"] = err.Error()
		return result, nil
	}

	result["exit_code"] = 0
	return result, nil
}

func (t BashTool) Desc() *ToolDescription {
	return &ToolDescription{
		Name:        "bash",
		Description: "Execute a bash command and return its output. Supports timeouts and custom working directories.",
		Parameters: D{
			"type": "object",
			"properties": D{
				"label": D{
					"type":        "string",
					"description": "Brief description of what this command does (shown to user)",
				},
				"command": D{
					"type":        "string",
					"description": "The bash command to execute",
				},
				"timeout": D{
					"type":        "number",
					"description": fmt.Sprintf("Timeout in seconds (default: %d)", int(DEFAULT_TIMEOUT.Seconds())),
				},
				"workdir": D{
					"type":        "string",
					"description": "Working directory to execute the command in (optional)",
				},
			},
			"required": []string{"label", "command"},
		},
	}
}

// Helper function to sanitize command output
func sanitizeOutput(output string) string {
	return strings.TrimSpace(output)
}
