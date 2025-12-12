package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/your-org/mcp-k8s-server/internal/domain"
)

// ==================== Job Handlers ====================

func (m *MCPServer) handleListJobs(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling list jobs request", "args", args)

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

	jobs, err := m.k8sUC.ListJobs(ctx, clusterID, namespace)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list jobs: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("⚡ Found %d Jobs in namespace '%s':\n\n", len(jobs), namespace)
	for i, job := range jobs {
		summary += fmt.Sprintf("%d. %s - Status: %s, Completions: %s, Duration: %s\n",
			i+1, job["name"], job["status"], job["completions"], job["duration"])
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"count":      len(jobs),
		"jobs":       jobs,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetJob(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling get job request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	jobName, _ := args["job_name"].(string)

	if clusterID == "" || jobName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and job_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	jobInfo, err := m.k8sUC.GetJob(ctx, clusterID, namespace, jobName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get job: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("⚡ Job '%s' in namespace '%s':\n\n", jobName, namespace)
	summary += fmt.Sprintf("Completions: %d (desired)\n", jobInfo["completions"])
	summary += fmt.Sprintf("Parallelism: %d\n", jobInfo["parallelism"])
	summary += fmt.Sprintf("Backoff Limit: %d\n", jobInfo["backoff_limit"])
	summary += fmt.Sprintf("Active: %d, Succeeded: %d, Failed: %d\n",
		jobInfo["active"], jobInfo["succeeded"], jobInfo["failed"])

	if startTime, ok := jobInfo["start_time"].(string); ok && startTime != "" {
		summary += fmt.Sprintf("Start Time: %s\n", startTime)
	}
	if completionTime, ok := jobInfo["completion_time"].(string); ok && completionTime != "" {
		summary += fmt.Sprintf("Completion Time: %s\n", completionTime)
	}
	if duration, ok := jobInfo["duration"].(string); ok && duration != "" {
		summary += fmt.Sprintf("Duration: %s\n", duration)
	}
	summary += fmt.Sprintf("Created: %s\n", jobInfo["created"])

	if containers, ok := jobInfo["containers"].([]map[string]any); ok && len(containers) > 0 {
		summary += fmt.Sprintf("\nContainers (%d):\n", len(containers))
		for i, container := range containers {
			summary += fmt.Sprintf("%d. %s - Image: %s\n", i+1, container["name"], container["image"])
			if cmd, ok := container["command"].([]string); ok && len(cmd) > 0 {
				summary += fmt.Sprintf("   Command: %v\n", cmd)
			}
			if args, ok := container["args"].([]string); ok && len(args) > 0 {
				summary += fmt.Sprintf("   Args: %v\n", args)
			}
		}
	}

	if conditions, ok := jobInfo["conditions"].([]map[string]any); ok && len(conditions) > 0 {
		summary += "\nConditions:\n"
		for _, condition := range conditions {
			summary += fmt.Sprintf("  %s: %s", condition["type"], condition["status"])
			if reason, ok := condition["reason"].(string); ok && reason != "" {
				summary += fmt.Sprintf(" - %s", reason)
			}
			summary += "\n"
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(jobInfo))},
		},
	}, jobInfo, nil
}

func (m *MCPServer) handleCreateJob(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling create job request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	jobName, _ := args["job_name"].(string)
	image, _ := args["image"].(string)

	if clusterID == "" || jobName == "" || image == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id, job_name, and image are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	// Parse command
	var command []string
	if cmdRaw, ok := args["command"].([]any); ok {
		for _, c := range cmdRaw {
			if str, ok := c.(string); ok {
				command = append(command, str)
			}
		}
	}

	// Parse args
	var jobArgs []string
	if argsRaw, ok := args["args"].([]any); ok {
		for _, a := range argsRaw {
			if str, ok := a.(string); ok {
				jobArgs = append(jobArgs, str)
			}
		}
	}

	// Parse env
	env := make(map[string]string)
	if envRaw, ok := args["env"].(map[string]any); ok {
		for k, v := range envRaw {
			if strVal, ok := v.(string); ok {
				env[k] = strVal
			}
		}
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

	options := domain.JobCreateOptions{
		Name:      jobName,
		Namespace: namespace,
		Image:     image,
		Command:   command,
		Args:      jobArgs,
		Env:       env,
		Labels:    labels,
	}

	// Parse optional fields
	if completions, ok := args["completions"].(float64); ok {
		c := int32(completions)
		options.Completions = &c
	}
	if parallelism, ok := args["parallelism"].(float64); ok {
		p := int32(parallelism)
		options.Parallelism = &p
	}
	if backoffLimit, ok := args["backoff_limit"].(float64); ok {
		b := int32(backoffLimit)
		options.BackoffLimit = &b
	}
	if restartPolicy, ok := args["restart_policy"].(string); ok {
		options.RestartPolicy = restartPolicy
	}

	err := m.k8sUC.CreateJob(ctx, clusterID, options)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create job: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"job_name":   jobName,
		"image":      image,
		"status":     "created",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" Job '%s' created successfully in namespace '%s'",
				jobName, namespace)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleDeleteJob(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling delete job request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	jobName, _ := args["job_name"].(string)

	if clusterID == "" || jobName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and job_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	err := m.k8sUC.DeleteJob(ctx, clusterID, namespace, jobName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to delete job: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"job_name":   jobName,
		"status":     "deleted",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Job '%s' deleted successfully from namespace '%s'",
				jobName, namespace)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetJobLogs(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling get job logs request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	jobName, _ := args["job_name"].(string)

	var tailLines int64 = 100
	if tl, ok := args["tail_lines"].(float64); ok {
		tailLines = int64(tl)
	}

	if clusterID == "" || jobName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and job_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	logs, err := m.k8sUC.GetJobLogs(ctx, clusterID, namespace, jobName, tailLines)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get job logs: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("📋 Logs from Job '%s' in namespace '%s' (last %d lines):\n\n%s",
				jobName, namespace, tailLines, logs)},
		},
	}, logs, nil
}
