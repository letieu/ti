package cli

import "github.com/letieu/ti/internal/tool"

func getDefaultTools() []tool.Tool {
	return []tool.Tool{
		tool.ReadTool{},
		tool.WriteTool{},
		tool.EditTool{},
		tool.BashTool{},
	}
}
