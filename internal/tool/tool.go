package tool

import "context"

type Tool interface {
	Desc() *ToolDescription
    Execute(ctx context.Context, input D) (D, error)
}

type ToolDescription struct {
	Name        string
	Description string
	Parameters  D
}

type D map[string]any
