package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/your-org/mcp-k8s-server/internal/domain"
	"github.com/your-org/mcp-k8s-server/internal/infrastructure"
	"github.com/your-org/mcp-k8s-server/internal/usecase"
)

type MCPServer struct {
	server    *mcp.Server
	clusterUC *usecase.ClusterUseCase
	k8sUC     *usecase.K8sUseCase
	logger    infrastructure.Logger
}

func NewMCPServer(
	clusterUC *usecase.ClusterUseCase,
	k8sUC *usecase.K8sUseCase,
	logger infrastructure.Logger,
) (*MCPServer, error) {

	impl := &mcp.Implementation{
		Name:    "k8s-mcp-server",
		Version: "1.0.0",
	}

	server := mcp.NewServer(impl, nil)
	mcpServer := &MCPServer{
		server:    server,
		clusterUC: clusterUC,
		k8sUC:     k8sUC,
		logger:    logger,
	}

	mcpServer.setupTools()
	return mcpServer, nil
}

func (m *MCPServer) setupTools() {

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_apply_yaml",
		Description: "Apply K8s resources. Use 'dry_run: true' to validate YAML without creating resources.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id":    map[string]any{"type": "string"},
				"yaml_body":     map[string]any{"type": "string"},
				"field_manager": map[string]any{"type": "string"},
				"dry_run": map[string]any{
					"type":        "boolean",
					"description": "If true, only validate the object without persisting it.",
				},
			},
			"required": []string{"cluster_id", "yaml_body"},
		},
	}, m.handleApplyYAML)

	// 2. Tool Port Forward
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_port_forward",
		Description: "Manage port forwarding. To STOP, call this tool with pod_name and cluster_id.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id":  map[string]any{"type": "string"},
				"namespace":   map[string]any{"type": "string"},
				"pod_name":    map[string]any{"type": "string"},
				"remote_port": map[string]any{"type": "number"},
				"local_port":  map[string]any{"type": "number"},
				"action":      map[string]any{"type": "string", "enum": []string{"start", "stop"}, "default": "start"},
			},
			"required": []string{"cluster_id", "namespace", "pod_name"},
		},
	}, m.handlePortForward)
	//  tool k8s_webhook_mutating_get
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_webhook_mutating_get",
		Description: "Get detailed configuration of a specific Mutating Webhook Configuration by name.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id":   map[string]any{"type": "string", "description": "ID of the cluster"},
				"webhook_name": map[string]any{"type": "string", "description": "Name of the Mutating Webhook Configuration."},
			},
			"required": []string{"cluster_id", "webhook_name"},
		},
	}, m.handleGetMutatingWebhook)

	//  tool k8s_webhook_validating_get
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_webhook_validating_get",
		Description: "Get detailed configuration of a specific Validating Webhook Configuration by name.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id":   map[string]any{"type": "string", "description": "ID of the cluster"},
				"webhook_name": map[string]any{"type": "string", "description": "Name of the Validating Webhook Configuration."},
			},
			"required": []string{"cluster_id", "webhook_name"},
		},
	}, m.handleGetValidatingWebhook)
	//  tool k8s_webhook_mutating_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_webhook_mutating_list",
		Description: "List all Mutating Webhook Configurations (used to change resources before validation) in the cluster.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{"type": "string", "description": "ID of the cluster"},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListMutatingWebhooks)

	//  tool k8s_webhook_validating_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_webhook_validating_list",
		Description: "List all Validating Webhook Configurations (used to enforce policy rules) in the cluster.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{"type": "string", "description": "ID of the cluster"},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListValidatingWebhooks)
	//  tool k8s_rbac_clusterrole_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_rbac_clusterrole_list",
		Description: "List all ClusterRoles (global authorization policies) in the cluster.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{"type": "string", "description": "ID of the cluster"},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListClusterRoles)

	//  tool k8s_event_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_event_list",
		Description: "List recent events in a Kubernetes namespace, optionally filtered by a specific involved object (Pod, Deployment, etc.).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id":    map[string]any{"type": "string", "description": "ID of the cluster"},
				"namespace":     map[string]any{"type": "string", "description": "Namespace to list events from", "default": "default"},
				"involved_kind": map[string]any{"type": "string", "description": "Optional: Kind of the object (e.g., 'Pod', 'Deployment') to filter events for."},
				"involved_name": map[string]any{"type": "string", "description": "Optional: Name of the object to filter events for."},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListEvents)

	// Register k8s_hpa_list tool
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_hpa_list",
		Description: "List all Horizontal Pod Autoscalers (HPA) in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{"type": "string", "description": "ID of the cluster"},
				"namespace":  map[string]any{"type": "string", "description": "Namespace to list HPAs from", "default": "default"},
			},
			"required": []string{"cluster_id", "namespace"},
		},
	}, m.handleListHPAs)

	// Register k8s_hpa_get tool
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_hpa_get",
		Description: "Get detailed information about a specific HPA, including current status and metrics.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{"type": "string", "description": "ID of the cluster"},
				"namespace":  map[string]any{"type": "string", "description": "Namespace of the HPA", "default": "default"},
				"hpa_name":   map[string]any{"type": "string", "description": "Name of the Horizontal Pod Autoscaler"},
			},
			"required": []string{"cluster_id", "namespace", "hpa_name"},
		},
	}, m.handleGetHPA)

	// Register k8s_hpa_delete tool
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_hpa_delete",
		Description: "Delete a Horizontal Pod Autoscaler (HPA) by name and namespace.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{"type": "string", "description": "ID of the cluster"},
				"namespace":  map[string]any{"type": "string", "description": "Namespace of the HPA to delete", "default": "default"},
				"hpa_name":   map[string]any{"type": "string", "description": "Name of the Horizontal Pod Autoscaler to delete"},
			},
			"required": []string{"cluster_id", "namespace", "hpa_name"},
		},
	}, m.handleDeleteHPA)
	//  tool k8s_quota_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_quota_list",
		Description: "List all ResourceQuotas in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{"type": "string", "description": "ID of the cluster"},
				"namespace":  map[string]any{"type": "string", "description": "Namespace to list ResourceQuotas from", "default": "default"},
			},
			"required": []string{"cluster_id", "namespace"},
		},
	}, m.handleListResourceQuotas)

	//  tool k8s_quota_get
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_quota_get",
		Description: "Get detailed information about a specific ResourceQuota",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{"type": "string", "description": "ID of the cluster"},
				"namespace":  map[string]any{"type": "string", "description": "Namespace of the ResourceQuota", "default": "default"},
				"quota_name": map[string]any{"type": "string", "description": "Name of the ResourceQuota"},
			},
			"required": []string{"cluster_id", "namespace", "quota_name"},
		},
	}, m.handleGetResourceQuota)

	//  tool k8s_limitrange_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_limitrange_list",
		Description: "List all LimitRanges in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{"type": "string", "description": "ID of the cluster"},
				"namespace":  map[string]any{"type": "string", "description": "Namespace to list LimitRanges from", "default": "default"},
			},
			"required": []string{"cluster_id", "namespace"},
		},
	}, m.handleListLimitRanges)

	//  tool k8s_limitrange_get
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_limitrange_get",
		Description: "Get detailed information about a specific LimitRange",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id":       map[string]any{"type": "string", "description": "ID of the cluster"},
				"namespace":        map[string]any{"type": "string", "description": "Namespace of the LimitRange", "default": "default"},
				"limit_range_name": map[string]any{"type": "string", "description": "Name of the LimitRange"},
			},
			"required": []string{"cluster_id", "namespace", "limit_range_name"},
		},
	}, m.handleGetLimitRange)

	// ==================== Node & Resource Tools ====================
	// Đăng ký tool k8s_node_taint_apply
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_node_taint_apply",
		Description: "Add or remove a Taint on a specific Kubernetes Node to control Pod scheduling (Taints must be in 'key=value:effect' format, e.g., 'gpu=true:NoSchedule').",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{"type": "string", "description": "ID of the cluster."},
				"node_name":  map[string]any{"type": "string", "description": "Name of the Node to modify."},
				"taint_key":  map[string]any{"type": "string", "description": "The Taint to apply/remove, in 'key=value:Effect' or 'key:Effect' format (Effect: NoSchedule, PreferNoSchedule, or NoExecute)."},
				"action":     map[string]any{"type": "string", "description": "Action to perform: 'add' or 'remove'."},
			},
			"required": []string{"cluster_id", "node_name", "taint_key", "action"},
		},
	}, m.handleApplyTaintToNode)
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_node_list",
		Description: "List all Kubernetes nodes in the cluster and their basic status",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace (mặc dù Node là tài nguyên Cluster-scoped, namespace vẫn có thể được dùng cho ngữ cảnh)",
					"default":     "default",
				},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListNodes)

	// register tool k8s_node_get_metrics
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_node_get_metrics",
		Description: "Get resource capacity and allocatable metrics (CPU, Memory, Pods) for a specific node",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"node_name": map[string]any{
					"type":        "string",
					"description": "Name of the Node",
				},
			},
			"required": []string{"cluster_id", "node_name"},
		},
	}, m.handleGetNodeMetrics)
	// register tool k8s_job_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_job_list",
		Description: "List all Jobs in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace to list Jobs from",
					"default":     "default",
				},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListJobs)

	// register tool k8s_job_get
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_job_get",
		Description: "Get detailed information about a specific Job",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the Job",
					"default":     "default",
				},
				"job_name": map[string]any{
					"type":        "string",
					"description": "Name of the Job",
				},
			},
			"required": []string{"cluster_id", "job_name"},
		},
	}, m.handleGetJob)

	// register tool k8s_job_create
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_job_create",
		Description: "Create a new Job in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace to create the Job in",
					"default":     "default",
				},
				"job_name": map[string]any{
					"type":        "string",
					"description": "Name of the Job",
				},
				"image": map[string]any{
					"type":        "string",
					"description": "Container image to run",
				},
				"command": map[string]any{
					"type":        "array",
					"description": "Command to run (optional)",
					"items":       map[string]any{"type": "string"},
				},
				"args": map[string]any{
					"type":        "array",
					"description": "Arguments for the command (optional)",
					"items":       map[string]any{"type": "string"},
				},
				"completions": map[string]any{
					"type":        "number",
					"description": "Number of successful completions required",
					"default":     1,
				},
				"parallelism": map[string]any{
					"type":        "number",
					"description": "Number of pods to run in parallel",
					"default":     1,
				},
				"backoff_limit": map[string]any{
					"type":        "number",
					"description": "Number of retries before considering Job failed",
					"default":     6,
				},
				"restart_policy": map[string]any{
					"type":        "string",
					"description": "Restart policy (OnFailure, Never)",
					"default":     "OnFailure",
				},
				"env": map[string]any{
					"type":        "object",
					"description": "Environment variables",
				},
				"labels": map[string]any{
					"type":        "object",
					"description": "Labels to apply to the Job",
				},
			},
			"required": []string{"cluster_id", "job_name", "image"},
		},
	}, m.handleCreateJob)

	// register tool k8s_job_delete
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_job_delete",
		Description: "Delete a Job from a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the Job",
					"default":     "default",
				},
				"job_name": map[string]any{
					"type":        "string",
					"description": "Name of the Job to delete",
				},
			},
			"required": []string{"cluster_id", "job_name"},
		},
	}, m.handleDeleteJob)

	// register tool k8s_job_get_logs
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_job_get_logs",
		Description: "Get logs from a Job's pod",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the Job",
					"default":     "default",
				},
				"job_name": map[string]any{
					"type":        "string",
					"description": "Name of the Job",
				},
				"tail_lines": map[string]any{
					"type":        "number",
					"description": "Number of lines to tail",
					"default":     100,
				},
			},
			"required": []string{"cluster_id", "job_name"},
		},
	}, m.handleGetJobLogs)

	// ==================== CronJob Tools ====================

	// register tool k8s_cronjob_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_cronjob_list",
		Description: "List all CronJobs in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace to list CronJobs from",
					"default":     "default",
				},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListCronJobs)

	// register tool k8s_cronjob_get
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_cronjob_get",
		Description: "Get detailed information about a specific CronJob",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the CronJob",
					"default":     "default",
				},
				"cronjob_name": map[string]any{
					"type":        "string",
					"description": "Name of the CronJob",
				},
			},
			"required": []string{"cluster_id", "cronjob_name"},
		},
	}, m.handleGetCronJob)

	// register tool k8s_cronjob_create
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_cronjob_create",
		Description: "Create a new CronJob in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace to create the CronJob in",
					"default":     "default",
				},
				"cronjob_name": map[string]any{
					"type":        "string",
					"description": "Name of the CronJob",
				},
				"schedule": map[string]any{
					"type":        "string",
					"description": "Cron schedule expression (e.g., '0 * * * *' for hourly)",
				},
				"image": map[string]any{
					"type":        "string",
					"description": "Container image to run",
				},
				"command": map[string]any{
					"type":        "array",
					"description": "Command to run (optional)",
					"items":       map[string]any{"type": "string"},
				},
				"args": map[string]any{
					"type":        "array",
					"description": "Arguments for the command (optional)",
					"items":       map[string]any{"type": "string"},
				},
				"suspend": map[string]any{
					"type":        "boolean",
					"description": "Whether the CronJob is suspended",
					"default":     false,
				},
				"concurrency_policy": map[string]any{
					"type":        "string",
					"description": "Allow, Forbid, or Replace",
					"default":     "Allow",
				},
				"successful_jobs_history_limit": map[string]any{
					"type":        "number",
					"description": "Number of successful jobs to keep",
					"default":     3,
				},
				"failed_jobs_history_limit": map[string]any{
					"type":        "number",
					"description": "Number of failed jobs to keep",
					"default":     1,
				},
				"restart_policy": map[string]any{
					"type":        "string",
					"description": "Restart policy (OnFailure, Never)",
					"default":     "OnFailure",
				},
				"env": map[string]any{
					"type":        "object",
					"description": "Environment variables",
				},
				"labels": map[string]any{
					"type":        "object",
					"description": "Labels to apply to the CronJob",
				},
			},
			"required": []string{"cluster_id", "cronjob_name", "schedule", "image"},
		},
	}, m.handleCreateCronJob)

	// register tool k8s_cronjob_delete
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_cronjob_delete",
		Description: "Delete a CronJob from a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the CronJob",
					"default":     "default",
				},
				"cronjob_name": map[string]any{
					"type":        "string",
					"description": "Name of the CronJob to delete",
				},
			},
			"required": []string{"cluster_id", "cronjob_name"},
		},
	}, m.handleDeleteCronJob)

	// register tool k8s_cronjob_suspend
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_cronjob_suspend",
		Description: "Suspend or resume a CronJob",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the CronJob",
					"default":     "default",
				},
				"cronjob_name": map[string]any{
					"type":        "string",
					"description": "Name of the CronJob",
				},
				"suspend": map[string]any{
					"type":        "boolean",
					"description": "True to suspend, false to resume",
				},
			},
			"required": []string{"cluster_id", "cronjob_name", "suspend"},
		},
	}, m.handleSuspendCronJob)

	// register tool k8s_cronjob_trigger
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_cronjob_trigger",
		Description: "Manually trigger a CronJob to run immediately",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the CronJob",
					"default":     "default",
				},
				"cronjob_name": map[string]any{
					"type":        "string",
					"description": "Name of the CronJob to trigger",
				},
			},
			"required": []string{"cluster_id", "cronjob_name"},
		},
	}, m.handleTriggerCronJob)

	// register tool k8s_statefulset_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_statefulset_list",
		Description: "List all StatefulSets in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace to list StatefulSets from",
					"default":     "default",
				},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListStatefulSets)

	// register tool k8s_statefulset_get
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_statefulset_get",
		Description: "Get detailed information about a specific StatefulSet",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the StatefulSet",
					"default":     "default",
				},
				"statefulset_name": map[string]any{
					"type":        "string",
					"description": "Name of the StatefulSet",
				},
			},
			"required": []string{"cluster_id", "statefulset_name"},
		},
	}, m.handleGetStatefulSet)

	// register tool k8s_statefulset_scale
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_statefulset_scale",
		Description: "Scale a StatefulSet to a specific number of replicas",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the StatefulSet",
					"default":     "default",
				},
				"statefulset_name": map[string]any{
					"type":        "string",
					"description": "Name of the StatefulSet",
				},
				"replicas": map[string]any{
					"type":        "number",
					"description": "Desired number of replicas",
					"minimum":     0,
				},
			},
			"required": []string{"cluster_id", "statefulset_name", "replicas"},
		},
	}, m.handleScaleStatefulSet)

	// register tool k8s_statefulset_restart
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_statefulset_restart",
		Description: "Restart a StatefulSet by adding a restart annotation",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the StatefulSet",
					"default":     "default",
				},
				"statefulset_name": map[string]any{
					"type":        "string",
					"description": "Name of the StatefulSet",
				},
			},
			"required": []string{"cluster_id", "statefulset_name"},
		},
	}, m.handleRestartStatefulSet)

	// register tool k8s_statefulset_delete
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_statefulset_delete",
		Description: "Delete a StatefulSet from a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the StatefulSet",
					"default":     "default",
				},
				"statefulset_name": map[string]any{
					"type":        "string",
					"description": "Name of the StatefulSet to delete",
				},
			},
			"required": []string{"cluster_id", "statefulset_name"},
		},
	}, m.handleDeleteStatefulSet)

	// ==================== DaemonSet Tools ====================

	// register tool k8s_daemonset_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_daemonset_list",
		Description: "List all DaemonSets in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace to list DaemonSets from",
					"default":     "default",
				},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListDaemonSets)

	// register tool k8s_daemonset_get
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_daemonset_get",
		Description: "Get detailed information about a specific DaemonSet",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the DaemonSet",
					"default":     "default",
				},
				"daemonset_name": map[string]any{
					"type":        "string",
					"description": "Name of the DaemonSet",
				},
			},
			"required": []string{"cluster_id", "daemonset_name"},
		},
	}, m.handleGetDaemonSet)

	// register tool k8s_daemonset_restart
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_daemonset_restart",
		Description: "Restart a DaemonSet by adding a restart annotation",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the DaemonSet",
					"default":     "default",
				},
				"daemonset_name": map[string]any{
					"type":        "string",
					"description": "Name of the DaemonSet",
				},
			},
			"required": []string{"cluster_id", "daemonset_name"},
		},
	}, m.handleRestartDaemonSet)

	// register tool k8s_daemonset_delete
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_daemonset_delete",
		Description: "Delete a DaemonSet from a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the DaemonSet",
					"default":     "default",
				},
				"daemonset_name": map[string]any{
					"type":        "string",
					"description": "Name of the DaemonSet to delete",
				},
			},
			"required": []string{"cluster_id", "daemonset_name"},
		},
	}, m.handleDeleteDaemonSet)

	// register tool k8s_daemonset_get_pods
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_daemonset_get_pods",
		Description: "Get all pods managed by a DaemonSet",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the DaemonSet",
					"default":     "default",
				},
				"daemonset_name": map[string]any{
					"type":        "string",
					"description": "Name of the DaemonSet",
				},
			},
			"required": []string{"cluster_id", "daemonset_name"},
		},
	}, m.handleGetDaemonSetPods)
	// register tool k8s_configmap_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_configmap_list",
		Description: "List all ConfigMaps in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace to list ConfigMaps from",
					"default":     "default",
				},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListConfigMaps)

	// register tool k8s_configmap_get
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_configmap_get",
		Description: "Get detailed information about a specific ConfigMap",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the ConfigMap",
					"default":     "default",
				},
				"configmap_name": map[string]any{
					"type":        "string",
					"description": "Name of the ConfigMap",
				},
			},
			"required": []string{"cluster_id", "configmap_name"},
		},
	}, m.handleGetConfigMap)

	// register tool k8s_configmap_create
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_configmap_create",
		Description: "Create a new ConfigMap in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace to create the ConfigMap in",
					"default":     "default",
				},
				"configmap_name": map[string]any{
					"type":        "string",
					"description": "Name of the ConfigMap",
				},
				"data": map[string]any{
					"type":        "object",
					"description": "Key-value pairs for the ConfigMap data",
				},
				"labels": map[string]any{
					"type":        "object",
					"description": "Labels to apply to the ConfigMap",
				},
			},
			"required": []string{"cluster_id", "configmap_name", "data"},
		},
	}, m.handleCreateConfigMap)

	// register tool k8s_configmap_delete
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_configmap_delete",
		Description: "Delete a ConfigMap from a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the ConfigMap",
					"default":     "default",
				},
				"configmap_name": map[string]any{
					"type":        "string",
					"description": "Name of the ConfigMap to delete",
				},
			},
			"required": []string{"cluster_id", "configmap_name"},
		},
	}, m.handleDeleteConfigMap)

	// ==================== Secret Tools ====================

	// register tool k8s_secret_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_secret_list",
		Description: "List all Secrets in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace to list Secrets from",
					"default":     "default",
				},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListSecrets)

	// register tool k8s_secret_get
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_secret_get",
		Description: "Get information about a specific Secret (keys only, not values)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the Secret",
					"default":     "default",
				},
				"secret_name": map[string]any{
					"type":        "string",
					"description": "Name of the Secret",
				},
			},
			"required": []string{"cluster_id", "secret_name"},
		},
	}, m.handleGetSecret)

	// register tool k8s_secret_create
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_secret_create",
		Description: "Create a new Secret in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace to create the Secret in",
					"default":     "default",
				},
				"secret_name": map[string]any{
					"type":        "string",
					"description": "Name of the Secret",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "Type of the Secret (Opaque, kubernetes.io/tls, etc.)",
					"default":     "Opaque",
				},
				"string_data": map[string]any{
					"type":        "object",
					"description": "Key-value pairs for the Secret data (will be base64 encoded)",
				},
				"labels": map[string]any{
					"type":        "object",
					"description": "Labels to apply to the Secret",
				},
			},
			"required": []string{"cluster_id", "secret_name", "string_data"},
		},
	}, m.handleCreateSecret)

	// register tool k8s_secret_delete
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_secret_delete",
		Description: "Delete a Secret from a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the Secret",
					"default":     "default",
				},
				"secret_name": map[string]any{
					"type":        "string",
					"description": "Name of the Secret to delete",
				},
			},
			"required": []string{"cluster_id", "secret_name"},
		},
	}, m.handleDeleteSecret)

	// register tool k8s_service_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_service_list",
		Description: "List all services in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace to list services from",
					"default":     "default",
				},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListServices)
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_service_get",
		Description: "Get detailed information about a Kubernetes service",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the service",
					"default":     "default",
				},
				"service_name": map[string]any{
					"type":        "string",
					"description": "Name of the service",
				},
			},
			"required": []string{"cluster_id", "service_name"},
		},
	}, m.handleGetService)
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_service_delete",
		Description: "Delete a Kubernetes service by name and namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the service",
					"default":     "default",
				},
				"service_name": map[string]any{
					"type":        "string",
					"description": "Name of the service to delete",
				},
			},
			"required": []string{"cluster_id", "service_name"},
		},
	}, m.handleDeleteService)

	// --- Ingress Tools ---

	//  register tool k8s_ingress_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_ingress_list",
		Description: "List all ingresses in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace to list ingresses from",
					"default":     "default",
				},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListIngresses)

	//  register tool k8s_ingress_get
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_ingress_get",
		Description: "Get detailed information about a Kubernetes ingress",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the ingress",
					"default":     "default",
				},
				"ingress_name": map[string]any{
					"type":        "string",
					"description": "Name of the ingress",
				},
			},
			"required": []string{"cluster_id", "ingress_name"},
		},
	}, m.handleGetIngress)

	//  register tool k8s_ingress_delete
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_ingress_delete",
		Description: "Delete a Kubernetes ingress by name and namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the ingress",
					"default":     "default",
				},
				"ingress_name": map[string]any{
					"type":        "string",
					"description": "Name of the ingress to delete",
				},
			},
			"required": []string{"cluster_id", "ingress_name"},
		},
	}, m.handleDeleteIngress)
	// register tool k8s_cluster_register
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_cluster_register",
		Description: "Register a new Kubernetes cluster with kubeconfig",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "Unique identifier for the cluster",
				},
				"kubeconfig_path": map[string]any{
					"type":        "string",
					"description": "Path to kubeconfig file",
				},
				"kubeconfig_data": map[string]any{
					"type":        "string",
					"description": "Base64 encoded kubeconfig data",
				},
				"context": map[string]any{
					"type":        "string",
					"description": "Kubernetes context to use",
				},
				"in_cluster": map[string]any{
					"type":        "boolean",
					"description": "Use in-cluster config",
					"default":     false,
				},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleClusterRegister)

	// register tool k8s_pod_get_logs
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_pod_get_logs",
		Description: "Get logs from a pod in a Kubernetes cluster",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the pod",
					"default":     "default",
				},
				"pod_name": map[string]any{
					"type":        "string",
					"description": "Name of the pod",
				},
				"tail_lines": map[string]any{
					"type":        "number",
					"description": "Number of lines to tail",
					"default":     100,
				},
			},
			"required": []string{"cluster_id", "pod_name"},
		},
	}, m.handleGetPodLogs)

	// register tool k8s_deployment_scale
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_deployment_scale",
		Description: "Scale a deployment in a Kubernetes cluster",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the deployment",
					"default":     "default",
				},
				"deployment_name": map[string]any{
					"type":        "string",
					"description": "Name of the deployment",
				},
				"replicas": map[string]any{
					"type":        "number",
					"description": "Number of replicas",
					"minimum":     0,
				},
			},
			"required": []string{"cluster_id", "deployment_name", "replicas"},
		},
	}, m.handleScaleDeployment)

	// register tool k8s_pod_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_pod_list",
		Description: "List pods in a Kubernetes namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace to list pods from",
					"default":     "default",
				},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListPods)

	// register tool k8s_deployment_get_info
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_deployment_get_info",
		Description: "Get detailed information about a Kubernetes deployment",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Namespace of the deployment",
					"default":     "default",
				},
				"deployment_name": map[string]any{
					"type":        "string",
					"description": "Name of the deployment",
				},
			},
			"required": []string{"cluster_id", "deployment_name"},
		},
	}, m.handleGetDeploymentInfo)

	// register tool k8s_namespace_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_namespace_list",
		Description: "List all namespaces in a Kubernetes cluster",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListNamespaces)

	// register tool k8s_namespace_get
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_namespace_get",
		Description: "Get detailed information about a specific namespace",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Name of the namespace",
				},
			},
			"required": []string{"cluster_id", "namespace"},
		},
	}, m.handleGetNamespace)

	// register tool k8s_namespace_create
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_namespace_create",
		Description: "Create a new namespace in a Kubernetes cluster",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Name of the namespace to create",
				},
			},
			"required": []string{"cluster_id", "namespace"},
		},
	}, m.handleCreateNamespace)

	// register tool k8s_namespace_delete
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_namespace_delete",
		Description: "Delete a namespace from a Kubernetes cluster",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
				"namespace": map[string]any{
					"type":        "string",
					"description": "Name of the namespace to delete",
				},
			},
			"required": []string{"cluster_id", "namespace"},
		},
	}, m.handleDeleteNamespace)

	// register tool k8s_persistentvolume_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_persistentvolume_list",
		Description: "List all PersistentVolumes in a Kubernetes cluster",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListPersistentVolumes)

	// register tool k8s_storageclass_list
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "k8s_storageclass_list",
		Description: "List all StorageClasses in a Kubernetes cluster",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{
					"type":        "string",
					"description": "ID of the cluster",
				},
			},
			"required": []string{"cluster_id"},
		},
	}, m.handleListStorageClasses)
}

func (m *MCPServer) handleClusterRegister(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling cluster register request", "args", args)

	// Parse arguments từ args (không phải req.Arguments)
	clusterID, _ := args["cluster_id"].(string)
	kubeconfigPath, _ := args["kubeconfig_path"].(string)
	kubeconfigData, _ := args["kubeconfig_data"].(string)
	contextName, _ := args["context"].(string)

	var inCluster bool
	if ic, ok := args["in_cluster"].(bool); ok {
		inCluster = ic
	}

	// Validate required fields
	if clusterID == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id is required"}, // ĐÚNG: dùng &mcp.TextContent{Text:
			},
			IsError: true,
		}, nil, nil
	}

	if kubeconfigPath == "" && kubeconfigData == "" && !inCluster {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: must provide either kubeconfig_path, kubeconfig_data, or set in_cluster=true"},
			},
			IsError: true,
		}, nil, nil
	}

	config := domain.ClusterConfig{
		KubeconfigPath: kubeconfigPath,
		KubeconfigData: []byte(kubeconfigData),
		Context:        contextName,
		InCluster:      inCluster,
	}

	err := m.clusterUC.RegisterCluster(ctx, domain.ClusterID(clusterID), config)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to register cluster: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"status":     "registered",
		"config": map[string]any{
			"has_kubeconfig_path": kubeconfigPath != "",
			"has_kubeconfig_data": len(kubeconfigData) > 0,
			"in_cluster":          inCluster,
			"context":             contextName,
		},
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" Cluster '%s' registered successfully", clusterID)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))}, // Dùng &mcp.TextContent{Text: cho JSON
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetPodLogs(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling get pod logs request", "args", args)

	// Lấy arguments từ args
	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	podName, _ := args["pod_name"].(string)

	var tailLines int64 = 100
	if tl, ok := args["tail_lines"].(float64); ok {
		tailLines = int64(tl)
	}

	// Validate required fields
	if clusterID == "" || podName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and pod_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	logs, err := m.k8sUC.GetPodLogs(ctx, clusterID, namespace, podName, tailLines)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get pod logs: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: logs,
			},
		},
		IsError: false,
	}, nil, nil
}

func (m *MCPServer) handleScaleDeployment(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling scale deployment request", "args", args)

	// Lấy arguments từ args
	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	deploymentName, _ := args["deployment_name"].(string)

	var replicas int32 = 1
	if r, ok := args["replicas"].(float64); ok {
		replicas = int32(r)
	}

	// Validate required fields
	if clusterID == "" || deploymentName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and deployment_name are required"},
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

	err := m.k8sUC.ScaleDeployment(ctx, clusterID, namespace, deploymentName, replicas)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to scale deployment: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"deployment": deploymentName,
		"namespace":  namespace,
		"replicas":   replicas,
		"status":     "scaled",
		"cluster_id": clusterID,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" Deployment '%s/%s' scaled to %d replicas", namespace, deploymentName, replicas)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleListPods(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling list pods request", "args", args)

	// Lấy arguments từ args
	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)

	// Validate required field
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

	pods, err := m.k8sUC.ListPods(ctx, clusterID, namespace)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list pods: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	// Format pods for display
	var podList []map[string]any
	for _, pod := range pods {
		podList = append(podList, map[string]any{
			"name":      string(pod.Name),
			"namespace": string(pod.Namespace),
			"status":    string(pod.Status.Phase),
			"cluster":   string(pod.ClusterID),
		})
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"pod_count":  len(podList),
		"pods":       podList,
	}

	// Create a readable text summary
	summary := fmt.Sprintf(" Found %d pods in namespace '%s':\n\n", len(podList), namespace)
	for i, pod := range podList {
		summary += fmt.Sprintf("%d. %s - Status: %s\n", i+1, pod["name"], pod["status"])
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetDeploymentInfo(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling get deployment info request", "args", args)

	// Parse arguments
	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)
	deploymentName, _ := args["deployment_name"].(string)

	// Validate required fields
	if clusterID == "" || deploymentName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and deployment_name are required"},
			},
			IsError: true,
		}, nil, nil
	}

	if namespace == "" {
		namespace = "default"
	}

	// Get deployment info from use case
	deploymentInfo, err := m.k8sUC.GetDeploymentInfo(ctx, clusterID, namespace, deploymentName)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get deployment info: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	// Format the response
	summary := fmt.Sprintf(" Deployment '%s' in namespace '%s':\n\n", deploymentName, namespace)
	summary += fmt.Sprintf("Replicas: %d/%d (desired/ready)\n", deploymentInfo["replicas_ready"], deploymentInfo["replicas_desired"])
	summary += fmt.Sprintf("Available: %d\n", deploymentInfo["replicas_available"])
	summary += fmt.Sprintf("Updated: %d\n", deploymentInfo["replicas_updated"])
	summary += fmt.Sprintf("Strategy: %s\n", deploymentInfo["strategy"])

	if containers, ok := deploymentInfo["containers"].([]map[string]any); ok {
		summary += fmt.Sprintf("\nContainers (%d):\n", len(containers))
		for i, container := range containers {
			summary += fmt.Sprintf("%d. %s - Image: %s\n", i+1, container["name"], container["image"])
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(deploymentInfo))},
		},
	}, deploymentInfo, nil
}

func (m *MCPServer) handleListNamespaces(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling list namespaces request", "args", args)

	clusterID, _ := args["cluster_id"].(string)

	if clusterID == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id is required"},
			},
			IsError: true,
		}, nil, nil
	}

	namespaces, err := m.k8sUC.ListNamespaces(ctx, clusterID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list namespaces: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf(" Found %d namespaces:\n\n", len(namespaces))
	for i, ns := range namespaces {
		summary += fmt.Sprintf("%d. %s - Status: %s\n", i+1, ns["name"], ns["status"])
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"count":      len(namespaces),
		"namespaces": namespaces,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleGetNamespace(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling get namespace request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)

	if clusterID == "" || namespace == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and namespace are required"},
			},
			IsError: true,
		}, nil, nil
	}

	namespaceInfo, err := m.k8sUC.GetNamespace(ctx, clusterID, namespace)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get namespace: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf(" Namespace '%s':\n\n", namespace)
	summary += fmt.Sprintf("Status: %s\n", namespaceInfo["status"])
	summary += fmt.Sprintf("Created: %s\n", namespaceInfo["created"])

	if labels, ok := namespaceInfo["labels"].(map[string]string); ok && len(labels) > 0 {
		summary += "\nLabels:\n"
		for k, v := range labels {
			summary += fmt.Sprintf("  %s: %s\n", k, v)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(namespaceInfo))},
		},
	}, namespaceInfo, nil
}

func (m *MCPServer) handleCreateNamespace(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling create namespace request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)

	if clusterID == "" || namespace == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and namespace are required"},
			},
			IsError: true,
		}, nil, nil
	}

	err := m.k8sUC.CreateNamespace(ctx, clusterID, namespace)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create namespace: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"status":     "created",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(" Namespace '%s' created successfully", namespace)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleDeleteNamespace(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling delete namespace request", "args", args)

	clusterID, _ := args["cluster_id"].(string)
	namespace, _ := args["namespace"].(string)

	if clusterID == "" || namespace == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id and namespace are required"},
			},
			IsError: true,
		}, nil, nil
	}

	err := m.k8sUC.DeleteNamespace(ctx, clusterID, namespace)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to delete namespace: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	resultData := map[string]any{
		"cluster_id": clusterID,
		"namespace":  namespace,
		"status":     "deleted",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Namespace '%s' deleted successfully", namespace)},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleListPersistentVolumes(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling list persistent volumes request", "args", args)

	clusterID, _ := args["cluster_id"].(string)

	if clusterID == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id is required"},
			},
			IsError: true,
		}, nil, nil
	}

	pvs, err := m.k8sUC.ListPersistentVolumes(ctx, clusterID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list persistent volumes: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("💾 Found %d PersistentVolumes:\n\n", len(pvs))
	for i, pv := range pvs {
		summary += fmt.Sprintf("%d. %s - Capacity: %s, Status: %s, StorageClass: %s\n",
			i+1, pv["name"], pv["capacity"], pv["status"], pv["storage_class"])
	}

	resultData := map[string]any{
		"cluster_id":         clusterID,
		"count":              len(pvs),
		"persistent_volumes": pvs,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) handleListStorageClasses(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.logger.Info("Handling list storage classes request", "args", args)

	clusterID, _ := args["cluster_id"].(string)

	if clusterID == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: cluster_id is required"},
			},
			IsError: true,
		}, nil, nil
	}

	storageClasses, err := m.k8sUC.ListStorageClasses(ctx, clusterID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list storage classes: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	summary := fmt.Sprintf("📦 Found %d StorageClasses:\n\n", len(storageClasses))
	for i, sc := range storageClasses {
		summary += fmt.Sprintf("%d. %s - Provisioner: %s, ReclaimPolicy: %s\n",
			i+1, sc["name"], sc["provisioner"], sc["reclaim_policy"])
	}

	resultData := map[string]any{
		"cluster_id":      clusterID,
		"count":           len(storageClasses),
		"storage_classes": storageClasses,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(mustMarshalJSON(resultData))},
		},
	}, resultData, nil
}

func (m *MCPServer) Run(ctx context.Context) error {
	m.logger.Info("Starting MCP server")
	// Sử dụng &mcp.StdioTransport{} (ĐÚNG)
	transport := &mcp.StdioTransport{}
	return m.server.Run(ctx, transport)
}

func mustMarshalJSON(v any) []byte {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal JSON: %v", err)
		return []byte("{}")
	}
	return data
}
