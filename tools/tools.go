package tools

import (
	"github.com/your-org/mcp-k8s-server/internal/usecase"
	"github.com/your-org/mcp-k8s-server/tools/cluster"
	"github.com/your-org/mcp-k8s-server/tools/deployment"
	"github.com/your-org/mcp-k8s-server/tools/namespace"
	"github.com/your-org/mcp-k8s-server/tools/pod"
)

// RegisterClusterTools registers cluster-related tools
func RegisterClusterTools(tm *usecase.ToolManager, clusterUseCase *usecase.ClusterUseCase) {
	registerTool := cluster.NewRegisterTool(clusterUseCase)
	tm.RegisterTool(registerTool)
}

// RegisterPodTools registers pod-related tools
func RegisterPodTools(tm *usecase.ToolManager, podUseCase *usecase.PodUseCase) {
	getLogsTool := pod.NewGetLogsTool(podUseCase)
	tm.RegisterTool(getLogsTool)
}

// RegisterDeploymentTools registers deployment-related tools
func RegisterDeploymentTools(tm *usecase.ToolManager) {
	scaleTool := deployment.NewScaleTool()
	tm.RegisterTool(scaleTool)
}

func RegisterNamespaceTools(tm *usecase.ToolManager, clusterUseCase *usecase.ClusterUseCase) {
	listPodsTool := namespace.NewListPodsTool(clusterUseCase)
	tm.RegisterTool(listPodsTool)
}
