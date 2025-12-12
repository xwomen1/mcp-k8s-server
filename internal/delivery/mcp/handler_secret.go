package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/your-org/mcp-k8s-server/internal/domain"
)

// ==================== Secret Handlers ====================

func (m *MCPServer) handleListSecrets(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling list secrets request", "args", args)

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

	secrets, err := m.k8sUC.ListSecrets(ctx, clusterID, namespace)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list secrets: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf(" Found %d Secrets in namespace '%s':\n\n", len(secrets), namespace)
	for i, secret := range secrets {
		summary += fmt.Sprintf("%d. %s - Type: %s, Keys: %d, Created: %s\n",
			i+1, secret["name"], secret["type"], secret["data_count"], secret["created"])
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"count":      len(secrets),
		"secrets":    secrets,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetSecret(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling get secret request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	secretName, _ := args["secret_name"].(string)

	if clusterID == "" || secretName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and secret_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	secretInfo, err := m.k8sUC.GetSecret(ctx, clusterID, namespace, secretName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get secret: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf(" Secret '%s' in namespace '%s':\n\n", secretName, namespace)
	summary += fmt.Sprintf("Type: %s\n", secretInfo["type"])
	summary += fmt.Sprintf("Created: %s\n\n", secretInfo["created"])

	if dataKeys, ok := secretInfo["data_keys"].([]string); ok && len(dataKeys) > 0 {
		summary += fmt.Sprintf("Data Keys (%d):\n", len(dataKeys))
		for i, key := range dataKeys {
			summary += fmt.Sprintf("  %d. %s\n", i+1, key)
		}
		summary += "\n  Note: Secret values are not displayed for security reasons\n"
	}

	if labels, ok := secretInfo["labels"].(map[string]string); ok && len(labels) > 0 {
		summary += "\nLabels:\n"
		for k, v := range labels {
			summary += fmt.Sprintf("  %s: %s\n", k, v)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(secretInfo))},
		},
	}, secretInfo, nil
}

func (m *MCPServer) handleCreateSecret(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling create secret request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	secretName, _ := args["secret_name"].(string)
	secretType, _ := args["type"].(string)

	if clusterID == "" || secretName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and secret_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	if secretType == "" {
		secretType = "Opaque"
	}

	// Parse string_data
	stringData := make(map[string]string)
	if dataRaw, ok := args["string_data"].(map[string]any); ok {
		for k, v := range dataRaw {
			if strVal, ok := v.(string); ok {
				stringData[k] = strVal
			}
		}
	}

	if len(stringData) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: string_data is required and must contain at least one key-value pair"},
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

	options := domain.SecretCreateOptions{
		Name:       secretName,
		Namespace:  namespace,
		Type:       secretType,
		StringData: stringData,
		Labels:     labels,
	}

	err := m.k8sUC.CreateSecret(ctx, clusterID, options)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create secret: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id":  clusterID,
		"namespace":   namespace,
		"secret_name": secretName,
		"type":        secretType,
		"data_keys":   len(stringData),
		"status":      "created",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" Secret '%s' created successfully in namespace '%s' with %d data keys",
				secretName, namespace, len(stringData))},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleDeleteSecret(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling delete secret request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	secretName, _ := args["secret_name"].(string)

	if clusterID == "" || secretName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and secret_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	err := m.k8sUC.DeleteSecret(ctx, clusterID, namespace, secretName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to delete secret: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id":  clusterID,
		"namespace":   namespace,
		"secret_name": secretName,
		"status":      "deleted",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" Secret '%s' deleted successfully from namespace '%s'",
				secretName, namespace)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}
