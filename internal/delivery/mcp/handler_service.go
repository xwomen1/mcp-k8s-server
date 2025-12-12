package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (m *MCPServer) handleListServices(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling list services request", "args", args)

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

	services, err := m.k8sUC.ListServices(ctx, clusterID, namespace)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list services: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("🌐 Found %d services in namespace '%s':\n\n", len(services), namespace)
	for i, svc := range services {
		summary += fmt.Sprintf("%d. %s - Type: %s, ClusterIP: %s\n",
			i+1, svc["name"], svc["type"], svc["cluster_ip"])
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"count":      len(services),
		"services":   services,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetService(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling get service request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	serviceName, _ := args["service_name"].(string)

	if clusterID == "" || serviceName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and service_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	serviceInfo, err := m.k8sUC.GetService(ctx, clusterID, namespace, serviceName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get service: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("🌐 Service '%s' in namespace '%s':\n\n", serviceName, namespace)
	summary += fmt.Sprintf("Type: %s\n", serviceInfo["type"])
	summary += fmt.Sprintf("ClusterIP: %s\n", serviceInfo["cluster_ip"])

	if ports, ok := serviceInfo["ports"].([]map[string]any); ok {
		summary += fmt.Sprintf("\nPorts (%d):\n", len(ports))
		for i, port := range ports {
			summary += fmt.Sprintf("%d. %s: %v -> %v\n",
				i+1, port["name"], port["port"], port["target_port"])
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(serviceInfo))},
		},
	}, serviceInfo, nil
}

func (m *MCPServer) handleDeleteService(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling delete service request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	serviceName, _ := args["service_name"].(string)

	if clusterID == "" || serviceName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and service_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	// m.k8sUC.DeleteService is assumed to be implemented
	err := m.k8sUC.DeleteService(ctx, clusterID, namespace, serviceName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to delete service '%s': %v", serviceName, err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf(" Service '%s' in namespace '%s' on cluster '%s' deleted successfully.", serviceName, namespace, clusterID)

	resultData := map[string]any{
		"cluster_id":   clusterID,
		"namespace":    namespace,
		"service_name": serviceName,
		"status":       "deleted",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

// --- Ingress Handlers ---
// Ingress is an API object that manages external access to the services in a cluster, typically HTTP.

func (m *MCPServer) handleListIngresses(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling list ingresses request", "args", args)

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

	// m.k8sUC.ListIngresses is assumed to be implemented
	ingresses, err := m.k8sUC.ListIngresses(ctx, clusterID, namespace)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list ingresses: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("🚪 Found %d ingresses in namespace '%s':\n\n", len(ingresses), namespace)
	for i, ing := range ingresses {
		// Assuming ing is a map[string]any containing name and host
		summary += fmt.Sprintf("%d. %s - Host: %s\n",
			i+1, ing["name"], ing["host"])
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"count":      len(ingresses),
		"ingresses":  ingresses,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetIngress(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling get ingress request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	ingressName, _ := args["ingress_name"].(string)

	if clusterID == "" || ingressName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and ingress_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	// m.k8sUC.GetIngress is assumed to be implemented
	ingressInfo, err := m.k8sUC.GetIngress(ctx, clusterID, namespace, ingressName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get ingress: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("🚪 Ingress '%s' in namespace '%s':\n\n", ingressName, namespace)
	summary += fmt.Sprintf("Class: %s\n", ingressInfo["ingress_class"])
	summary += fmt.Sprintf("Default Backend: %s:%v\n", ingressInfo["default_backend_service"], ingressInfo["default_backend_port"])

	if rules, ok := ingressInfo["rules"].([]map[string]any); ok {
		summary += fmt.Sprintf("\nRules (%d):\n", len(rules))
		for i, rule := range rules {
			summary += fmt.Sprintf("%d. Host: %s\n", i+1, rule["host"])
			if paths, ok := rule["paths"].([]map[string]any); ok {
				for _, path := range paths {
					// Assuming path is map[string]any with path and backend_service
					summary += fmt.Sprintf("   Path: %s -> Service: %s:%v\n",
						path["path"], path["backend_service"], path["backend_port"])
				}
			}
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(ingressInfo))},
		},
	}, ingressInfo, nil
}

func (m *MCPServer) handleDeleteIngress(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling delete ingress request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	ingressName, _ := args["ingress_name"].(string)

	if clusterID == "" || ingressName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and ingress_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	// m.k8sUC.DeleteIngress is assumed to be implemented
	err := m.k8sUC.DeleteIngress(ctx, clusterID, namespace, ingressName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to delete ingress '%s': %v", ingressName, err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf(" Ingress '%s' in namespace '%s' on cluster '%s' deleted successfully.", ingressName, namespace, clusterID)

	resultData := map[string]any{
		"cluster_id":   clusterID,
		"namespace":    namespace,
		"ingress_name": ingressName,
		"status":       "deleted",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

// NOTE: mustMarshalJSON and MCPServer/m.k8sUC implementation details are omitted for brevity.
