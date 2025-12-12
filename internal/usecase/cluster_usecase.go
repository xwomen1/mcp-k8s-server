// internal/usecase/cluster_usecase.go
package usecase

import (
	"context"
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	"github.com/your-org/mcp-k8s-server/internal/infrastructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterUseCase struct {
	clusterManager *infrastructure.ClusterManager
	clusterRepo    domain.ClusterRepository
	logger         infrastructure.Logger
}

func NewClusterUseCase(
	clusterManager *infrastructure.ClusterManager,
	clusterRepo domain.ClusterRepository,
	logger infrastructure.Logger,
) *ClusterUseCase {
	return &ClusterUseCase{
		clusterManager: clusterManager,
		clusterRepo:    clusterRepo,
		logger:         logger,
	}
}

func (uc *ClusterUseCase) ScaleDeployment(ctx context.Context, clusterID, namespace, deploymentName string, replicas int32) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	deployment.Spec.Replicas = &replicas
	_, err = client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}

	return nil
}

func (uc *ClusterUseCase) GetPodLogs(ctx context.Context, clusterID, namespace, podName string, tailLines int64) (string, error) {
	logs, err := uc.clusterManager.GetPodLogs(
		ctx,
		domain.ClusterID(clusterID),
		domain.Namespace(namespace),
		domain.PodName(podName),
		domain.LogOptions{TailLines: &tailLines},
	)

	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}

	return string(logs), nil
}

func (uc *ClusterUseCase) RegisterCluster(ctx context.Context, clusterID domain.ClusterID, config domain.ClusterConfig) error {
	uc.logger.Info("Registering cluster", "clusterID", clusterID)

	// First save to repository
	cluster := &domain.Cluster{
		ID:     clusterID,
		Config: config,
		Status: domain.ClusterStatusActive,
	}

	if err := uc.clusterRepo.Save(cluster); err != nil {
		return fmt.Errorf("failed to save cluster: %w", err)
	}

	// Then register with cluster manager
	if err := uc.clusterManager.RegisterCluster(ctx, clusterID, config); err != nil {
		// Rollback repository save
		_ = uc.clusterRepo.Delete(clusterID)
		return fmt.Errorf("failed to register cluster: %w", err)
	}

	return nil
}

func (uc *ClusterUseCase) ListPods(ctx context.Context, clusterID string, namespace domain.Namespace) ([]domain.Pod, error) {
	return uc.clusterManager.ListPods(ctx, domain.ClusterID(clusterID), namespace)
}

func (uc *ClusterUseCase) GetClusterStatus(ctx context.Context, clusterID domain.ClusterID) (*domain.ClusterStatus, error) {
	return uc.clusterManager.GetClusterStatus(ctx, clusterID)
}

func (uc *ClusterUseCase) ListClusters(ctx context.Context) ([]domain.Cluster, error) {
	return uc.clusterManager.ListClusters(ctx)
}

func (uc *ClusterUseCase) DeleteCluster(ctx context.Context, clusterID domain.ClusterID) error {
	return uc.clusterManager.DeleteCluster(ctx, clusterID)
}
