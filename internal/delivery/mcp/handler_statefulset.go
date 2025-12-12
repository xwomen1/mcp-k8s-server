package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ==================== StatefulSet Handlers ====================

func (m *MCPServer) handleListStatefulSets(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling list statefulsets request", "args", args)

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

	statefulSets, err := m.k8sUC.ListStatefulSets(ctx, clusterID, namespace)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list statefulsets: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("📊 Found %d StatefulSets in namespace '%s':\n\n", len(statefulSets), namespace)
	for i, sts := range statefulSets {
		summary += fmt.Sprintf("%d. %s - Replicas: %d/%d (ready/desired), Service: %s\n",
			i+1, sts["name"], sts["ready_replicas"], sts["replicas"], sts["service_name"])
	}

	resultData := map[string]any{
		"cluster_id":   clusterID,
		"namespace":    namespace,
		"count":        len(statefulSets),
		"statefulsets": statefulSets,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetStatefulSet(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling get statefulset request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	statefulSetName, _ := args["statefulset_name"].(string)

	if clusterID == "" || statefulSetName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and statefulset_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	statefulSetInfo, err := m.k8sUC.GetStatefulSet(ctx, clusterID, namespace, statefulSetName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get statefulset: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("📊 StatefulSet '%s' in namespace '%s':\n\n", statefulSetName, namespace)
	summary += fmt.Sprintf("Replicas: %d/%d (ready/desired)\n",
		statefulSetInfo["replicas_ready"], statefulSetInfo["replicas_desired"])
	summary += fmt.Sprintf("Current: %d, Updated: %d\n",
		statefulSetInfo["replicas_current"], statefulSetInfo["replicas_updated"])
	summary += fmt.Sprintf("Service Name: %s\n", statefulSetInfo["service_name"])
	summary += fmt.Sprintf("Update Strategy: %s\n", statefulSetInfo["update_strategy"])
	summary += fmt.Sprintf("Created: %s\n", statefulSetInfo["created"])

	if containers, ok := statefulSetInfo["containers"].([]map[string]any); ok && len(containers) > 0 {
		summary += fmt.Sprintf("\nContainers (%d):\n", len(containers))
		for i, container := range containers {
			summary += fmt.Sprintf("%d. %s - Image: %s\n", i+1, container["name"], container["image"])
		}
	}

	if volumeClaims, ok := statefulSetInfo["volume_claims"].([]map[string]any); ok && len(volumeClaims) > 0 {
		summary += fmt.Sprintf("\nVolume Claim Templates (%d):\n", len(volumeClaims))
		for i, pvc := range volumeClaims {
			summary += fmt.Sprintf("%d. %s - Storage: %s", i+1, pvc["name"], pvc["storage"])
			if sc, ok := pvc["storage_class"].(string); ok && sc != "" {
				summary += fmt.Sprintf(", StorageClass: %s", sc)
			}
			summary += "\n"
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(statefulSetInfo))},
		},
	}, statefulSetInfo, nil
}

func (m *MCPServer) handleScaleStatefulSet(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling scale statefulset request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	statefulSetName, _ := args["statefulset_name"].(string)

	var replicas int32 = 1
	if r, ok := args["replicas"].(float64); ok {
		replicas = int32(r)
	}

	if clusterID == "" || statefulSetName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and statefulset_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	if replicas < 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error: replicas must be non-negative, got %d", replicas)},
			},
			IsError: true,
		}, nil, nil
	}

	err := m.k8sUC.ScaleStatefulSet(ctx, clusterID, namespace, statefulSetName, replicas)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to scale statefulset: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id":       clusterID,
		"namespace":        namespace,
		"statefulset_name": statefulSetName,
		"replicas":         replicas,
		"status":           "scaled",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" StatefulSet '%s/%s' scaled to %d replicas",
				namespace, statefulSetName, replicas)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleRestartStatefulSet(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling restart statefulset request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	statefulSetName, _ := args["statefulset_name"].(string)

	if clusterID == "" || statefulSetName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and statefulset_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	err := m.k8sUC.RestartStatefulSet(ctx, clusterID, namespace, statefulSetName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to restart statefulset: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id":       clusterID,
		"namespace":        namespace,
		"statefulset_name": statefulSetName,
		"status":           "restarting",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" StatefulSet '%s/%s' restart initiated. Pods will be recreated one by one.",
				namespace, statefulSetName)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleDeleteStatefulSet(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling delete statefulset request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	statefulSetName, _ := args["statefulset_name"].(string)

	if clusterID == "" || statefulSetName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and statefulset_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	err := m.k8sUC.DeleteStatefulSet(ctx, clusterID, namespace, statefulSetName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to delete statefulset: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id":       clusterID,
		"namespace":        namespace,
		"statefulset_name": statefulSetName,
		"status":           "deleted",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" StatefulSet '%s/%s' deleted successfully. Note: PersistentVolumeClaims are NOT automatically deleted.",
				namespace, statefulSetName)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}
