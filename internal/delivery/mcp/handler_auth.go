package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	// Import domain package
	// "github.com/your-org/mcp-k8s-server/internal/domain"
)

func (m *MCPServer) handleListClusterRoles(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	clusterID, _ := args["cluster_id"].(string)

	if clusterID == "" {
		err := fmt.Errorf("cluster_id is required")
		return errorResult(err), nil, err
	}

	clusterRoles, err := m.k8sUC.ListClusterRoles(ctx, clusterID)
	if err != nil {
		return errorResult(fmt.Errorf("failed to list ClusterRoles: %w", err)), nil, err
	}

	summary := fmt.Sprintf("🛡️ Found %d ClusterRoles in cluster '%s':\n", len(clusterRoles), clusterID)

	for i, cr := range clusterRoles {
		ruleCount := len(cr.Rules)
		summary += fmt.Sprintf("%d. Name: %s (Rules: %d)\n",
			i+1,
			cr.Name,
			ruleCount,
		)

		// Limit summary output
		if i >= 10 {
			summary += fmt.Sprintf("...and %d more ClusterRoles. Use dedicated tool to get details.\n", len(clusterRoles)-i-1)
			break
		}
	}

	resultData := map[string]any{
		"cluster_id":    clusterID,
		"count":         len(clusterRoles),
		"cluster_roles": clusterRoles,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}
