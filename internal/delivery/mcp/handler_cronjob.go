package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/your-org/mcp-k8s-server/internal/domain"
)

// ==================== CronJob Handlers ====================

func (m *MCPServer) handleListCronJobs(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling list cronjobs request", "args", args)

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

	cronJobs, err := m.k8sUC.ListCronJobs(ctx, clusterID, namespace)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list cronjobs: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("⏰ Found %d CronJobs in namespace '%s':\n\n", len(cronJobs), namespace)
	for i, cj := range cronJobs {
		statusIcon := ""
		if cj["suspend"].(bool) {
			statusIcon = "⏸️ "
		}
		summary += fmt.Sprintf("%d. %s %s - Schedule: %s, Active: %d, Last: %s\n",
			i+1, statusIcon, cj["name"], cj["schedule"], cj["active"], cj["last_schedule"])
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"count":      len(cronJobs),
		"cronjobs":   cronJobs,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetCronJob(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling get cronjob request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	cronJobName, _ := args["cronjob_name"].(string)

	if clusterID == "" || cronJobName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and cronjob_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	cronJobInfo, err := m.k8sUC.GetCronJob(ctx, clusterID, namespace, cronJobName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get cronjob: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	statusIcon := " Active"
	if cronJobInfo["suspend"].(bool) {
		statusIcon = "⏸️  Suspended"
	}

	summary := fmt.Sprintf("⏰ CronJob '%s' in namespace '%s':\n\n", cronJobName, namespace)
	summary += fmt.Sprintf("Status: %s\n", statusIcon)
	summary += fmt.Sprintf("Schedule: %s\n", cronJobInfo["schedule"])
	summary += fmt.Sprintf("Concurrency Policy: %s\n", cronJobInfo["concurrency_policy"])
	summary += fmt.Sprintf("Successful Jobs History Limit: %d\n", cronJobInfo["successful_jobs_history_limit"])
	summary += fmt.Sprintf("Failed Jobs History Limit: %d\n", cronJobInfo["failed_jobs_history_limit"])

	if lastSchedule, ok := cronJobInfo["last_schedule_time"].(string); ok && lastSchedule != "" {
		summary += fmt.Sprintf("Last Schedule Time: %s\n", lastSchedule)
	}
	if lastSuccessful, ok := cronJobInfo["last_successful_time"].(string); ok && lastSuccessful != "" {
		summary += fmt.Sprintf("Last Successful Time: %s\n", lastSuccessful)
	}

	summary += fmt.Sprintf("Active Jobs: %d\n", cronJobInfo["active_count"])
	if activeJobs, ok := cronJobInfo["active_jobs"].([]string); ok && len(activeJobs) > 0 {
		summary += "  Running:\n"
		for _, job := range activeJobs {
			summary += fmt.Sprintf("  - %s\n", job)
		}
	}

	summary += fmt.Sprintf("Created: %s\n", cronJobInfo["created"])

	if containers, ok := cronJobInfo["containers"].([]map[string]any); ok && len(containers) > 0 {
		summary += fmt.Sprintf("\nJob Template - Containers (%d):\n", len(containers))
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

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(cronJobInfo))},
		},
	}, cronJobInfo, nil
}

func (m *MCPServer) handleCreateCronJob(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling create cronjob request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	cronJobName, _ := args["cronjob_name"].(string)
	schedule, _ := args["schedule"].(string)
	image, _ := args["image"].(string)

	if clusterID == "" || cronJobName == "" || schedule == "" || image == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id, cronjob_name, schedule, and image are required"},
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

	options := domain.CronJobCreateOptions{
		Name:      cronJobName,
		Namespace: namespace,
		Schedule:  schedule,
		Image:     image,
		Command:   command,
		Args:      jobArgs,
		Env:       env,
		Labels:    labels,
	}

	// Parse optional fields
	if suspend, ok := args["suspend"].(bool); ok {
		options.Suspend = suspend
	}
	if concurrencyPolicy, ok := args["concurrency_policy"].(string); ok {
		options.ConcurrencyPolicy = concurrencyPolicy
	}
	if successLimit, ok := args["successful_jobs_history_limit"].(float64); ok {
		s := int32(successLimit)
		options.SuccessfulJobsHistoryLimit = &s
	}
	if failLimit, ok := args["failed_jobs_history_limit"].(float64); ok {
		f := int32(failLimit)
		options.FailedJobsHistoryLimit = &f
	}
	if restartPolicy, ok := args["restart_policy"].(string); ok {
		options.RestartPolicy = restartPolicy
	}

	err := m.k8sUC.CreateCronJob(ctx, clusterID, options)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create cronjob: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id":   clusterID,
		"namespace":    namespace,
		"cronjob_name": cronJobName,
		"schedule":     schedule,
		"image":        image,
		"status":       "created",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" CronJob '%s' created successfully in namespace '%s' with schedule '%s'",
				cronJobName, namespace, schedule)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleDeleteCronJob(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling delete cronjob request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	cronJobName, _ := args["cronjob_name"].(string)

	if clusterID == "" || cronJobName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and cronjob_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	err := m.k8sUC.DeleteCronJob(ctx, clusterID, namespace, cronJobName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to delete cronjob: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id":   clusterID,
		"namespace":    namespace,
		"cronjob_name": cronJobName,
		"status":       "deleted",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" CronJob '%s' deleted successfully from namespace '%s'",
				cronJobName, namespace)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleSuspendCronJob(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling suspend cronjob request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	cronJobName, _ := args["cronjob_name"].(string)
	suspend, _ := args["suspend"].(bool)

	if clusterID == "" || cronJobName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id, cronjob_name, and suspend are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	err := m.k8sUC.SuspendCronJob(ctx, clusterID, namespace, cronJobName, suspend)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to update cronjob: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	action := "resumed"
	icon := "▶️ "
	if suspend {
		action = "suspended"
		icon = "⏸️ "
	}

	resultData := map[string]any{
		"cluster_id":   clusterID,
		"namespace":    namespace,
		"cronjob_name": cronJobName,
		"suspend":      suspend,
		"status":       action,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("%s CronJob '%s' %s successfully in namespace '%s'",
				icon, cronJobName, action, namespace)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleTriggerCronJob(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling trigger cronjob request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	cronJobName, _ := args["cronjob_name"].(string)

	if clusterID == "" || cronJobName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and cronjob_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	err := m.k8sUC.TriggerCronJob(ctx, clusterID, namespace, cronJobName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to trigger cronjob: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id":   clusterID,
		"namespace":    namespace,
		"cronjob_name": cronJobName,
		"status":       "triggered",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" CronJob '%s' triggered manually. A new Job has been created.",
				cronJobName)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}
