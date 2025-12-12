package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ... imports

func (m *MCPServer) handleListResourceQuotas(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, map[string]any, error) {
	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)

	quotas, err := m.k8sUC.ListResourceQuotas(ctx, clusterID, namespace)
	if err != nil {
		return errorResult(fmt.Errorf("failed to list quotas: %w", err)), nil, err
	}

	summary := fmt.Sprintf(" Found %d ResourceQuota(s) in %s/%s:\n", len(quotas), clusterID, namespace)
	for i, q := range quotas {
		summary += fmt.Sprintf("%d. Name: %s\n", i+1, q["name"])
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"quotas":     quotas,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetResourceQuota(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, map[string]any, error) {
	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	name, _ := args["quota_name"].(string)

	quota, err := m.k8sUC.GetResourceQuotaDetail(ctx, clusterID, namespace, name)
	if err != nil {
		return errorResult(fmt.Errorf("failed to get quota: %w", err)), nil, err
	}

	summary := fmt.Sprintf("📈 Resource Quota Details (%s/%s):\n", namespace, name)
	summary += "--- Limits (Hard) ---\n"
	hardLimits, _ := quota["hard_limits"].(map[string]string)
	for res, limit := range hardLimits {
		summary += fmt.Sprintf("%s: %s\n", res, limit)
	}

	summary += "\n--- Used ---\n"
	used, _ := quota["used"].(map[string]string)
	for res, usage := range used {
		summary += fmt.Sprintf("%s: %s\n", res, usage)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(quota))},
		},
	}, quota, nil
}

func (m *MCPServer) handleListLimitRanges(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, map[string]any, error) {
	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)

	limitRanges, err := m.k8sUC.ListLimitRanges(ctx, clusterID, namespace)
	if err != nil {
		return errorResult(fmt.Errorf("failed to list limit ranges: %w", err)), nil, err
	}

	summary := fmt.Sprintf("⚖️ Found %d LimitRange(s) in %s/%s:\n", len(limitRanges), clusterID, namespace)
	for i, lr := range limitRanges {
		summary += fmt.Sprintf("%d. Name: %s (Limits: %d)\n", i+1, lr["name"], lr["limits_count"])
	}

	resultData := map[string]any{
		"cluster_id":   clusterID,
		"namespace":    namespace,
		"limit_ranges": limitRanges,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetLimitRange(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, map[string]any, error) {
	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	name, _ := args["limit_range_name"].(string)

	limitRange, err := m.k8sUC.GetLimitRangeDetail(ctx, clusterID, namespace, name)
	if err != nil {
		return errorResult(fmt.Errorf("failed to get limit range: %w", err)), nil, err
	}

	summary := fmt.Sprintf(" Limit Range Details (%s/%s):\n\n", namespace, name)

	limits, _ := limitRange["limits"].([]map[string]any)
	for i, limit := range limits {
		summary += fmt.Sprintf("--- Limit Set %d (Type: %s) ---\n", i+1, limit["type"])

		max, _ := limit["max"].(map[string]string)
		if len(max) > 0 {
			summary += "Max: "
			for res, val := range max {
				summary += fmt.Sprintf("%s=%s ", res, val)
			}
			summary += "\n"
		}

		def, _ := limit["default"].(map[string]string)
		if len(def) > 0 {
			summary += "Default: "
			for res, val := range def {
				summary += fmt.Sprintf("%s=%s ", res, val)
			}
			summary += "\n"
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(limitRange))},
		},
	}, limitRange, nil
}

// Giả định errorResult là một helper function để tạo phản hồi lỗi
func errorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
		IsError: true,
	}
}
