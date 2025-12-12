package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	// Import domain package
	// Giả định mustMarshalJSON và logger được định nghĩa
)

func (m *MCPServer) handleListNodes(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	// Logger is assumed to be available
	// m.logger.Info("Handling list nodes request", "args", args)

	clusterID, _ := args["cluster_id"].(string)

	if clusterID == "" {
		return errorResultNode("cluster_id is required"), nil, nil
	}

	// CHÚ Ý: Use Case giờ trả về []domain.Node
	nodes, err := m.k8sUC.ListNodes(ctx, clusterID)
	if err != nil {
		return errorResultNode(fmt.Sprintf("Failed to list nodes: %v", err)), nil, nil
	}

	summary := fmt.Sprintf(" Found %d nodes in cluster '%s':\n\n", len(nodes), clusterID)

	// Sử dụng các trường dữ liệu trực tiếp từ domain.Node struct
	for i, node := range nodes {
		summary += fmt.Sprintf("%d. %s (Status: %s, Role: %s, IP: %s)\n",
			i+1, node.Name, node.Status, node.Roles, node.InternalIP)
	}

	// Gửi struct domain.Node đi, nó sẽ được marshal thành JSON
	resultData := map[string]any{
		"cluster_id": clusterID,
		"count":      len(nodes),
		"nodes":      nodes, // Trả về slice of domain.Node
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetNodeMetrics(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	// m.logger.Info("Handling get node metrics request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	nodeName, _ := args["node_name"].(string)

	if clusterID == "" || nodeName == "" {
		return errorResultNode("cluster_id and node_name are required"), nil, nil
	}

	// CHÚ Ý: Use Case giờ trả về domain.NodeMetrics
	metrics, err := m.k8sUC.GetNodeMetrics(ctx, clusterID, nodeName)
	if err != nil {
		return errorResultNode(fmt.Sprintf("Failed to get node metrics: %v", err)), nil, nil
	}

	summary := fmt.Sprintf(" Resource metrics for Node '%s':\n\n", nodeName)

	// Sử dụng các trường dữ liệu trực tiếp từ domain.NodeMetrics struct
	summary += fmt.Sprintf("--- Capacity (Total Resources) ---\n")
	summary += fmt.Sprintf("CPU: %s\n", metrics.Capacity["cpu"])
	summary += fmt.Sprintf("Memory: %s\n", metrics.Capacity["memory"])
	summary += fmt.Sprintf("Max Pods: %s\n", metrics.Capacity["pods"])

	summary += fmt.Sprintf("\n--- Allocatable (Available for Pods) ---\n")
	summary += fmt.Sprintf("CPU: %s\n", metrics.Allocatable["cpu"])
	summary += fmt.Sprintf("Memory: %s\n", metrics.Allocatable["memory"])
	summary += fmt.Sprintf("Max Pods: %s\n", metrics.Allocatable["pods"])

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(metrics))}, // Marshal domain.NodeMetrics
		},
	}, metrics, nil
}

func errorResultNode(errMsg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %s", errMsg)}},
		IsError: true,
	}
}

// handleApplyTaintToNode adds or removes a taint on a Kubernetes Node.
func (m *MCPServer) handleApplyTaintToNode(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	clusterID, _ := args["cluster_id"].(string)
	nodeName, _ := args["node_name"].(string)
	taintKey, _ := args["taint_key"].(string)
	action, _ := args["action"].(string) // "add" or "remove"

	if clusterID == "" || nodeName == "" || taintKey == "" || action == "" {
		err := fmt.Errorf("cluster_id, node_name, taint_key (e.g., 'key=value:NoSchedule'), and action ('add' or 'remove') are required")
		return errorResult(err), nil, err
	}

	action = strings.ToLower(action)
	if action != "add" && action != "remove" {
		err := fmt.Errorf("invalid action '%s'. Must be 'add' or 'remove'", action)
		return errorResult(err), nil, err
	}

	updatedNode, err := m.k8sUC.ApplyTaintToNode(ctx, clusterID, nodeName, taintKey, action)
	if err != nil {
		return errorResult(fmt.Errorf("failed to %s taint on node %s: %w", action, nodeName, err)), nil, err
	}

	taintCount := len(updatedNode.Taints)
	taintSummary := fmt.Sprintf("%d Taint(s):\n", taintCount)
	for i, t := range updatedNode.Taints {
		taintSummary += fmt.Sprintf("  - %s=%s:%s\n", t.Key, t.Value, t.Effect)
		if i >= 5 && taintCount > 6 {
			taintSummary += fmt.Sprintf("  ...and %d more.\n", taintCount-i-1)
			break
		}
	}

	summary := fmt.Sprintf(" Successfully **%s** Taint '%s' on Node '%s'.\n",
		action, taintKey, nodeName)
	summary += fmt.Sprintf("Node is now **%s**.\n",
		func() string {
			if updatedNode.Unschedulable {
				return "Unschedulable"
			}
			return "Schedulable"
		}())
	summary += "\n--- Current Taints ---\n" + taintSummary

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(updatedNode))},
		},
	}, updatedNode, nil
}
