package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/your-org/mcp-k8s-server/internal/domain"
	// Import domain package
	// "github.com/your-org/mcp-k8s-server/internal/domain"
)

func (m *MCPServer) handleListEvents(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	involvedKind, _ := args["involved_kind"].(string)
	involvedName, _ := args["involved_name"].(string)

	if clusterID == "" {
		return errorResult(fmt.Errorf("cluster_id is required")), nil, nil
	}
	if namespace == "" {
		namespace = "default" // Default namespace if not specified
	}

	events, err := m.k8sUC.ListEvents(ctx, clusterID, namespace, involvedKind, involvedName)
	if err != nil {
		return errorResult(fmt.Errorf("failed to list events: %w", err)), nil, nil
	}

	filterMsg := ""
	if involvedKind != "" && involvedName != "" {
		filterMsg = fmt.Sprintf(" for object %s/%s", involvedKind, involvedName)
	}

	summary := fmt.Sprintf("📢 Found %d Events in %s/%s%s:\n", len(events), clusterID, namespace, filterMsg)

	// Sort by LastTimestamp (latest first) for better readability
	// Note: We should ideally sort the domain slice, but for simplicity here we just display:

	for i, event := range events {
		// Calculate time elapsed since the last event
		timeAgo := time.Since(event.LastTimestamp).Round(time.Second)

		// Use emoji to highlight warnings
		prefix := "🔸"
		if event.Type == domain.EventTypeWarning {
			prefix = "⚠️"
		}

		summary += fmt.Sprintf("%s [%s, %s, %dx] %s: %s (on %s/%s) - %s ago\n",
			prefix,
			event.Type,
			event.Reason,
			event.Count,
			event.Message,
			event.InvolvedKind,
			event.InvolvedName,
			timeAgo,
		)

		// Limit summary output for extremely long lists
		if i >= 10 {
			summary += fmt.Sprintf("...and %d more events. See full JSON for details.\n", len(events)-i-1)
			break
		}
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"filter":     map[string]string{"kind": involvedKind, "name": involvedName},
		"events":     events,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}
