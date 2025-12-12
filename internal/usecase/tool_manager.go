package usecase

import (
	"context"
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/infrastructure"
)

type Tool interface {
	Name() string
	Description() string
	Schema() map[string]interface{}
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

type ToolManager struct {
	tools  map[string]Tool
	logger infrastructure.Logger
}

func NewToolManager(logger infrastructure.Logger) *ToolManager {
	return &ToolManager{
		tools:  make(map[string]Tool),
		logger: logger,
	}
}

func (tm *ToolManager) RegisterTool(tool Tool) {
	tm.tools[tool.Name()] = tool
	tm.logger.Info("Registered tool", "name", tool.Name())
}

func (tm *ToolManager) ListTools() []map[string]interface{} {
	var result []map[string]interface{}
	for _, tool := range tm.tools {
		result = append(result, map[string]interface{}{
			"name":        tool.Name(),
			"description": tool.Description(),
			"inputSchema": tool.Schema(),
		})
	}
	return result
}

func (tm *ToolManager) ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	tool, exists := tm.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return tool.Execute(ctx, args)
}
