package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/your-org/mcp-k8s-server/internal/domain"
)

// ==================== ConfigMap Handlers ====================

func (m *MCPServer) handleListConfigMaps(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling list configmaps request", "args", args)

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

	configMaps, err := m.k8sUC.ListConfigMaps(ctx, clusterID, namespace)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list configmaps: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("📋 Found %d ConfigMaps in namespace '%s':\n\n", len(configMaps), namespace)
	for i, cm := range configMaps {
		summary += fmt.Sprintf("%d. %s - Data keys: %d, Created: %s\n",
			i+1, cm["name"], cm["data_count"], cm["created"])
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"count":      len(configMaps),
		"configmaps": configMaps,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetConfigMap(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling get configmap request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	configMapName, _ := args["configmap_name"].(string)

	if clusterID == "" || configMapName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and configmap_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	configMapInfo, err := m.k8sUC.GetConfigMap(ctx, clusterID, namespace, configMapName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get configmap: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("📋 ConfigMap '%s' in namespace '%s':\n\n", configMapName, namespace)
	summary += fmt.Sprintf("Created: %s\n\n", configMapInfo["created"])

	if data, ok := configMapInfo["data"].(map[string]string); ok && len(data) > 0 {
		summary += fmt.Sprintf("Data (%d keys):\n", len(data))
		for key, value := range data {
			// Truncate long values
			displayValue := value
			if len(value) > 100 {
				displayValue = value[:100] + "... (truncated)"
			}
			summary += fmt.Sprintf("  %s: %s\n", key, displayValue)
		}
	}

	if labels, ok := configMapInfo["labels"].(map[string]string); ok && len(labels) > 0 {
		summary += "\nLabels:\n"
		for k, v := range labels {
			summary += fmt.Sprintf("  %s: %s\n", k, v)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(configMapInfo))},
		},
	}, configMapInfo, nil
}

func (m *MCPServer) handleCreateConfigMap(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling create configmap request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	configMapName, _ := args["configmap_name"].(string)

	if clusterID == "" || configMapName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and configmap_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	// Parse data
	data := make(map[string]string)
	if dataRaw, ok := args["data"].(map[string]any); ok {
		for k, v := range dataRaw {
			if strVal, ok := v.(string); ok {
				data[k] = strVal
			}
		}
	}

	if len(data) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: data is required and must contain at least one key-value pair"},
			},
			IsError: true,
		}, nil, nil
	}

	// Parse labels
	labels := make(map[string]string)
	if labelsRaw, ok := args["labels"].(map[string]any); ok {
		for k, v := range labelsRaw {
			if strVal, ok := v.(string); ok {
				labels[k] = strVal
			}
		}
	}

	options := domain.ConfigMapCreateOptions{
		Name:      configMapName,
		Namespace: namespace,
		Data:      data,
		Labels:    labels,
	}

	err := m.k8sUC.CreateConfigMap(ctx, clusterID, options)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create configmap: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id":     clusterID,
		"namespace":      namespace,
		"configmap_name": configMapName,
		"data_keys":      len(data),
		"status":         "created",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" ConfigMap '%s' created successfully in namespace '%s' with %d data keys",
				configMapName, namespace, len(data))},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleDeleteConfigMap(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling delete configmap request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	configMapName, _ := args["configmap_name"].(string)

	if clusterID == "" || configMapName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and configmap_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	err := m.k8sUC.DeleteConfigMap(ctx, clusterID, namespace, configMapName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to delete configmap: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id":     clusterID,
		"namespace":      namespace,
		"configmap_name": configMapName,
		"status":         "deleted",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" ConfigMap '%s' deleted successfully from namespace '%s'",
				configMapName, namespace)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}
