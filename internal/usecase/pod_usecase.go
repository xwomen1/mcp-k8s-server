package usecase

import (
	"context"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	"github.com/your-org/mcp-k8s-server/internal/infrastructure"
)

type PodUseCase struct {
	clusterClient infrastructure.ClusterClient
	logger        infrastructure.Logger
}

func NewPodUseCase(logger infrastructure.Logger) *PodUseCase {
	return &PodUseCase{
		clusterClient: infrastructure.GetClusterManager(logger),
		logger:        logger,
	}
}

func (uc *PodUseCase) GetPodLogs(ctx context.Context, clusterID domain.ClusterID, namespace domain.Namespace, podName domain.PodName, options domain.LogOptions) (domain.PodLogs, error) {
	uc.logger.Info("Getting pod logs", "clusterID", clusterID, "namespace", namespace, "podName", podName)
	return uc.clusterClient.GetPodLogs(ctx, clusterID, namespace, podName, options)
}

func (uc *PodUseCase) GetPodStatus(ctx context.Context, clusterID domain.ClusterID, namespace domain.Namespace, podName domain.PodName) (*domain.PodStatus, error) {
	return uc.clusterClient.GetPodStatus(ctx, clusterID, namespace, podName)
}

func (uc *PodUseCase) ListPods(ctx context.Context, clusterID domain.ClusterID, namespace domain.Namespace) ([]domain.Pod, error) {
	return uc.clusterClient.ListPods(ctx, clusterID, namespace)
}
