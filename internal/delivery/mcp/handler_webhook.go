package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	// Import domain package
)

// --- Mutating Webhooks Handler ---

func (m *MCPServer) handleListMutatingWebhooks(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	clusterID, _ := args["cluster_id"].(string)

	if clusterID == "" {
		err := fmt.Errorf("cluster_id is required")
		return errorResult(err), nil, err
	}

	webhooks, err := m.k8sUC.ListMutatingWebhooks(ctx, clusterID)
	if err != nil {
		return errorResult(fmt.Errorf("failed to list Mutating Webhooks: %w", err)), nil, err
	}

	summary := fmt.Sprintf("✂️ Found %d Mutating Webhook Configuration(s) in cluster '%s':\n", len(webhooks), clusterID)

	for i, wh := range webhooks {
		summary += fmt.Sprintf("%d. Name: %s (Webhooks: %d, Policy: %s, Target: %s)\n",
			i+1, wh.Name, wh.WebhooksCount, wh.FailurePolicy, wh.ClientConfig.Service)
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"webhooks":   webhooks,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

// --- Validating Webhooks Handler ---

func (m *MCPServer) handleListValidatingWebhooks(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	clusterID, _ := args["cluster_id"].(string)

	if clusterID == "" {
		err := fmt.Errorf("cluster_id is required")
		return errorResult(err), nil, err
	}

	webhooks, err := m.k8sUC.ListValidatingWebhooks(ctx, clusterID)
	if err != nil {
		return errorResult(fmt.Errorf("failed to list Validating Webhooks: %w", err)), nil, err
	}

	summary := fmt.Sprintf(" Found %d Validating Webhook Configuration(s) in cluster '%s':\n", len(webhooks), clusterID)

	for i, wh := range webhooks {
		summary += fmt.Sprintf("%d. Name: %s (Webhooks: %d, Policy: %s, Target: %s)\n",
			i+1, wh.Name, wh.WebhooksCount, wh.FailurePolicy, wh.ClientConfig.Service)
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"webhooks":   webhooks,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetMutatingWebhook(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	clusterID, _ := args["cluster_id"].(string)
	name, _ := args["webhook_name"].(string)

	if clusterID == "" || name == "" {
		err := fmt.Errorf("cluster_id and webhook_name are required")
		return errorResult(err), nil, err
	}

	webhook, err := m.k8sUC.GetMutatingWebhook(ctx, clusterID, name)
	if err != nil {
		return errorResult(fmt.Errorf("failed to get Mutating Webhook %s: %w", name, err)), nil, err
	}

	summary := fmt.Sprintf(" Mutating Webhook Details (%s):\n", webhook.Name)
	summary += fmt.Sprintf("Policy: %s\n", webhook.FailurePolicy)
	summary += fmt.Sprintf("Webhooks Count: %d\n", webhook.WebhooksCount)
	summary += fmt.Sprintf("Target Service: %s\n", webhook.ClientConfig.Service)
	summary += fmt.Sprintf("Path: %s\n", webhook.ClientConfig.Path)

	if len(webhook.Rules) > 0 {
		rule := webhook.Rules[0] // Lấy quy tắc đầu tiên để tóm tắt
		summary += "\n--- Admission Rules (Summary) ---\n"
		summary += fmt.Sprintf("Resources: %v\n", rule.Resources)
		summary += fmt.Sprintf("Operations: %v\n", rule.Operations)
		summary += fmt.Sprintf("Scope: %s\n", rule.Scope)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(webhook))},
		},
	}, webhook, nil
}

// handleGetValidatingWebhook retrieves and displays detailed information for a specific ValidatingWebhookConfiguration.
func (m *MCPServer) handleGetValidatingWebhook(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	clusterID, _ := args["cluster_id"].(string)
	name, _ := args["webhook_name"].(string)

	if clusterID == "" || name == "" {
		err := fmt.Errorf("cluster_id and webhook_name are required")
		return errorResult(err), nil, err
	}

	webhook, err := m.k8sUC.GetValidatingWebhook(ctx, clusterID, name)
	if err != nil {
		return errorResult(fmt.Errorf("failed to get Validating Webhook %s: %w", name, err)), nil, err
	}

	summary := fmt.Sprintf(" Validating Webhook Details (%s):\n", webhook.Name)
	summary += fmt.Sprintf("Policy: %s\n", webhook.FailurePolicy)
	summary += fmt.Sprintf("Webhooks Count: %d\n", webhook.WebhooksCount)
	summary += fmt.Sprintf("Target Service: %s\n", webhook.ClientConfig.Service)
	summary += fmt.Sprintf("Path: %s\n", webhook.ClientConfig.Path)

	if len(webhook.Rules) > 0 {
		rule := webhook.Rules[0] // Lấy quy tắc đầu tiên để tóm tắt
		summary += "\n--- Validation Rules (Summary) ---\n"
		summary += fmt.Sprintf("Resources: %v\n", rule.Resources)
		summary += fmt.Sprintf("Operations: %v\n", rule.Operations)
		summary += fmt.Sprintf("Scope: %s\n", rule.Scope)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(webhook))},
		},
	}, webhook, nil
}
