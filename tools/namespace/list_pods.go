package namespace

import (
	"context"
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	"github.com/your-org/mcp-k8s-server/internal/usecase"
)

type ListPodsTool struct {
	ClusterUseCase *usecase.ClusterUseCase
}

// Constructor
func NewListPodsTool(clusterUC *usecase.ClusterUseCase) *ListPodsTool {
	return &ListPodsTool{
		ClusterUseCase: clusterUC,
	}
}

// Tool interface methods
func (t *ListPodsTool) Name() string {
	return "k8s_pod_list"
}

func (t *ListPodsTool) Description() string {
	return "List all pods in a given namespace"
}

// Schema method (trả về input JSON schema)
func (t *ListPodsTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"cluster_id": map[string]string{
				"type": "string",
			},
			"namespace": map[string]string{
				"type": "string",
			},
		},
		"required": []string{"cluster_id", "namespace"},
	}
}

// Execute method
func (t *ListPodsTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	clusterID, ok := input["cluster_id"].(string)
	if !ok || clusterID == "" {
		return nil, fmt.Errorf("cluster_id missing or invalid")
	}

	namespace, ok := input["namespace"].(string)
	if !ok || namespace == "" {
		return nil, fmt.Errorf("namespace missing or invalid")
	}

	return t.ClusterUseCase.ListPods(ctx, clusterID, domain.Namespace(namespace))
}
