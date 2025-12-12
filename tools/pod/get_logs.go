package pod

import (
	"context"
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	"github.com/your-org/mcp-k8s-server/internal/usecase"
)

type GetLogsTool struct {
	podUseCase *usecase.PodUseCase
}

func NewGetLogsTool(podUseCase *usecase.PodUseCase) *GetLogsTool {
	return &GetLogsTool{
		podUseCase: podUseCase,
	}
}

func (t *GetLogsTool) Name() string {
	return "k8s_pod_get_logs"
}

func (t *GetLogsTool) Description() string {
	return "Get logs from a Kubernetes pod in a specific namespace"
}

func (t *GetLogsTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"cluster_id": map[string]interface{}{
				"type":        "string",
				"description": "The cluster ID",
			},
			"namespace": map[string]interface{}{
				"type":        "string",
				"description": "The namespace name",
			},
			"pod_name": map[string]interface{}{
				"type":        "string",
				"description": "The pod name",
			},
			"tail_lines": map[string]interface{}{
				"type":        "integer",
				"description": "Number of lines from the end to retrieve",
			},
			"container": map[string]interface{}{
				"type":        "string",
				"description": "Container name (optional)",
			},
		},
		"required": []string{"cluster_id", "namespace", "pod_name"},
	}
}

func (t *GetLogsTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	clusterIDStr, ok := args["cluster_id"].(string)
	if !ok {
		return nil, fmt.Errorf("cluster_id is required and must be a string")
	}

	namespaceStr, ok := args["namespace"].(string)
	if !ok {
		return nil, fmt.Errorf("namespace is required and must be a string")
	}

	podNameStr, ok := args["pod_name"].(string)
	if !ok {
		return nil, fmt.Errorf("pod_name is required and must be a string")
	}

	options := domain.LogOptions{}
	if tailLines, ok := args["tail_lines"].(float64); ok {
		lines := int64(tailLines)
		options.TailLines = &lines
	}
	if container, ok := args["container"].(string); ok {
		options.Container = container
	}

	logs, err := t.podUseCase.GetPodLogs(
		ctx,
		domain.ClusterID(clusterIDStr),
		domain.Namespace(namespaceStr),
		domain.PodName(podNameStr),
		options,
	)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"logs": string(logs),
	}, nil
}
