package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (m *MCPServer) handleListHPAs(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)

	nodes, err := m.k8sUC.ListHPAs(ctx, clusterID, namespace)
	if err != nil {
		return errorResult(fmt.Errorf("failed to list HPAs: %w", err)), nil, nil
	}

	summary := fmt.Sprintf("Found %d HorizontalPodAutoscaler(s) in %s/%s:\n", len(nodes), clusterID, namespace)

	for i, hpa := range nodes {
		summary += fmt.Sprintf("%d. Name: %s (Target: %s/%s, Pods: %d-%d)\n",
			i+1,
			hpa.Name,
			hpa.TargetKind,
			hpa.TargetName,
			hpa.MinReplicas,
			hpa.MaxReplicas,
		)
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"hpas":       nodes, // nodes là []domain.HPA
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetHPA(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	name, _ := args["hpa_name"].(string)

	hpa, err := m.k8sUC.GetHPADetail(ctx, clusterID, namespace, name)
	if err != nil {
		return errorResult(fmt.Errorf("failed to get HPA: %w", err)), nil, nil
	}

	summary := fmt.Sprintf("HPA Details (%s/%s):\n", namespace, name)

	summary += fmt.Sprintf("Target: %s/%s (Pods: %d-%d, Current: %d, Desired: %d)\n",
		hpa.TargetKind,
		hpa.TargetName,
		hpa.MinReplicas,
		hpa.MaxReplicas,
		hpa.CurrentReplicas,
		hpa.DesiredReplicas,
	)

	metrics := hpa.Metrics
	if len(metrics) > 0 {
		summary += "\nMetrics:\n"

		for _, m := range metrics {
			summary += fmt.Sprintf("- Type: %s, Resource: %s, Target: %s\n",
				m.Type,
				m.ResourceName,
				m.TargetValue,
			)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(hpa))},
		},
	}, hpa, nil
}

func (m *MCPServer) handleDeleteHPA(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, map[string]any, error) {
	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	name, _ := args["hpa_name"].(string)

	err := m.k8sUC.DeleteHPA(ctx, clusterID, namespace, name)
	if err != nil {
		return errorResult(fmt.Errorf("failed to delete HPA: %w", err)), nil, err
	}

	successMsg := fmt.Sprintf("HorizontalPodAutoscaler %s/%s deleted successfully.", namespace, name)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: successMsg},
		},
	}, map[string]any{"status": "success", "message": successMsg}, nil
}
