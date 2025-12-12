package cluster

import (
	"context"
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	"github.com/your-org/mcp-k8s-server/internal/usecase"
)

type RegisterTool struct {
	clusterUseCase *usecase.ClusterUseCase
}

func NewRegisterTool(clusterUseCase *usecase.ClusterUseCase) *RegisterTool {
	return &RegisterTool{
		clusterUseCase: clusterUseCase,
	}
}

func (t *RegisterTool) Name() string {
	return "k8s_cluster_register"
}

func (t *RegisterTool) Description() string {
	return "Register a Kubernetes cluster with kubeconfig"
}

func (t *RegisterTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"cluster_id": map[string]interface{}{
				"type":        "string",
				"description": "Unique identifier for the cluster",
			},
			"kubeconfig_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to kubeconfig file",
			},
			"kubeconfig_data": map[string]interface{}{
				"type":        "string",
				"description": "Base64 encoded kubeconfig data",
			},
			"context": map[string]interface{}{
				"type":        "string",
				"description": "Kubernetes context to use",
			},
		},
		"required": []string{"cluster_id"},
	}
}

func (t *RegisterTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	clusterIDStr, ok := args["cluster_id"].(string)
	if !ok {
		return nil, fmt.Errorf("cluster_id is required and must be a string")
	}

	config := domain.ClusterConfig{}
	if kubeconfigPath, ok := args["kubeconfig_path"].(string); ok {
		config.KubeconfigPath = kubeconfigPath
	}
	if kubeconfigData, ok := args["kubeconfig_data"].(string); ok {
		config.KubeconfigData = []byte(kubeconfigData)
	}
	if contextStr, ok := args["context"].(string); ok {
		config.Context = contextStr
	}

	if config.KubeconfigPath == "" && len(config.KubeconfigData) == 0 {
		return nil, fmt.Errorf("either kubeconfig_path or kubeconfig_data must be provided")
	}

	err := t.clusterUseCase.RegisterCluster(ctx, domain.ClusterID(clusterIDStr), config)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":     "registered",
		"cluster_id": clusterIDStr,
	}, nil
}

