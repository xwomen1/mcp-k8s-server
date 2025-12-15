package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (m *MCPServer) handleApplyYAML(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	clusterID, _ := args["cluster_id"].(string)
	yamlBody, _ := args["yaml_body"].(string)

	dryRun, _ := args["dry_run"].(bool)

	manager, ok := args["field_manager"].(string)
	if !ok || manager == "" {
		manager = "mcp-k8s-assistant"
	}

	result, err := m.k8sUC.ApplyYAML(ctx, clusterID, yamlBody, manager, dryRun)
	if err != nil {
		return errorResult(err), nil, err
	}

	statusEmoji := "✅"
	actionText := "Applied"
	if dryRun {
		statusEmoji = " [DRY-RUN]"
		actionText = "validated (no changes made)"
	}

	summary := fmt.Sprintf("%s %s %s: %s in namespace %s",
		statusEmoji, actionText, result.Kind, result.Name, result.Namespace)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
		},
	}, result, nil
}

func (m *MCPServer) handlePortForward(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	clusterID, _ := args["cluster_id"].(string)
	ns, _ := args["namespace"].(string)
	pod, _ := args["pod_name"].(string)

	action, _ := args["action"].(string)
	if action == "" {
		action = "start"
	}

	var lPort, rPort int
	if val, ok := args["local_port"].(float64); ok {
		lPort = int(val)
	}
	if val, ok := args["remote_port"].(float64); ok {
		rPort = int(val)
	}

	res, err := m.k8sUC.PortForward(ctx, clusterID, ns, pod, lPort, rPort, action)
	if err != nil {
		return errorResult(err), nil, err
	}

	var summary string
	if action == "stop" {
		summary = fmt.Sprintf(" Port-forward for pod %s has been terminated.", pod)
	} else {
		summary = fmt.Sprintf(" Tunnel Active! Access here: %s", res.URL)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: summary}},
	}, res, nil
}
