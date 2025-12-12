package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ==================== DaemonSet Handlers ====================

func (m *MCPServer) handleListDaemonSets(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling list daemonsets request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)

	if clusterID == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id is required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	daemonSets, err := m.k8sUC.ListDaemonSets(ctx, clusterID, namespace)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list daemonsets: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("⚙️ Found %d DaemonSets in namespace '%s':\n\n", len(daemonSets), namespace)
	for i, ds := range daemonSets {
		summary += fmt.Sprintf("%d. %s - Ready: %d/%d, Available: %d/%d\n",
			i+1, ds["name"],
			ds["number_ready"], ds["current_number_scheduled"],
			ds["number_available"], ds["desired_number_scheduled"])
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"count":      len(daemonSets),
		"daemonsets": daemonSets,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetDaemonSet(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling get daemonset request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	daemonSetName, _ := args["daemonset_name"].(string)

	if clusterID == "" || daemonSetName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and daemonset_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	daemonSetInfo, err := m.k8sUC.GetDaemonSet(ctx, clusterID, namespace, daemonSetName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get daemonset: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("⚙️ DaemonSet '%s' in namespace '%s':\n\n", daemonSetName, namespace)
	summary += fmt.Sprintf("Desired: %d nodes\n", daemonSetInfo["desired_number_scheduled"])
	summary += fmt.Sprintf("Current: %d scheduled\n", daemonSetInfo["current_number_scheduled"])
	summary += fmt.Sprintf("Ready: %d\n", daemonSetInfo["number_ready"])
	summary += fmt.Sprintf("Available: %d\n", daemonSetInfo["number_available"])
	summary += fmt.Sprintf("Updated: %d\n", daemonSetInfo["updated_number_scheduled"])

	if misscheduled, ok := daemonSetInfo["number_misscheduled"].(int32); ok && misscheduled > 0 {
		summary += fmt.Sprintf("⚠️  Misscheduled: %d\n", misscheduled)
	}

	summary += fmt.Sprintf("Update Strategy: %s\n", daemonSetInfo["update_strategy"])
	summary += fmt.Sprintf("Created: %s\n", daemonSetInfo["created"])

	if containers, ok := daemonSetInfo["containers"].([]map[string]any); ok && len(containers) > 0 {
		summary += fmt.Sprintf("\nContainers (%d):\n", len(containers))
		for i, container := range containers {
			summary += fmt.Sprintf("%d. %s - Image: %s\n", i+1, container["name"], container["image"])
		}
	}

	if nodeSelector, ok := daemonSetInfo["node_selector"].(map[string]string); ok && len(nodeSelector) > 0 {
		summary += "\nNode Selector:\n"
		for k, v := range nodeSelector {
			summary += fmt.Sprintf("  %s: %s\n", k, v)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(daemonSetInfo))},
		},
	}, daemonSetInfo, nil
}

func (m *MCPServer) handleRestartDaemonSet(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling restart daemonset request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	daemonSetName, _ := args["daemonset_name"].(string)

	if clusterID == "" || daemonSetName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and daemonset_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	err := m.k8sUC.RestartDaemonSet(ctx, clusterID, namespace, daemonSetName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to restart daemonset: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id":     clusterID,
		"namespace":      namespace,
		"daemonset_name": daemonSetName,
		"status":         "restarting",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" DaemonSet '%s/%s' restart initiated. Pods on all nodes will be recreated.",
				namespace, daemonSetName)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleDeleteDaemonSet(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling delete daemonset request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	daemonSetName, _ := args["daemonset_name"].(string)

	if clusterID == "" || daemonSetName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and daemonset_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	err := m.k8sUC.DeleteDaemonSet(ctx, clusterID, namespace, daemonSetName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to delete daemonset: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id":     clusterID,
		"namespace":      namespace,
		"daemonset_name": daemonSetName,
		"status":         "deleted",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" DaemonSet '%s/%s' deleted successfully. All pods will be terminated.",
				namespace, daemonSetName)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetDaemonSetPods(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling get daemonset pods request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	daemonSetName, _ := args["daemonset_name"].(string)

	if clusterID == "" || daemonSetName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and daemonset_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	pods, err := m.k8sUC.GetDaemonSetPods(ctx, clusterID, namespace, daemonSetName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get daemonset pods: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("📦 Found %d pods for DaemonSet '%s' in namespace '%s':\n\n",
		len(pods), daemonSetName, namespace)

	for i, pod := range pods {
		summary += fmt.Sprintf("%d. %s\n", i+1, pod["name"])
		summary += fmt.Sprintf("   Node: %s, Status: %s\n", pod["node"], pod["status"])
		if hostIP, ok := pod["host_ip"].(string); ok && hostIP != "" {
			summary += fmt.Sprintf("   Host IP: %s", hostIP)
			if podIP, ok := pod["pod_ip"].(string); ok && podIP != "" {
				summary += fmt.Sprintf(", Pod IP: %s", podIP)
			}
			summary += "\n"
		}
		summary += "\n"
	}

	resultData := map[string]any{
		"cluster_id":     clusterID,
		"namespace":      namespace,
		"daemonset_name": daemonSetName,
		"pod_count":      len(pods),
		"pods":           pods,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}
