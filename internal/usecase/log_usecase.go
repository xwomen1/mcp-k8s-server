// internal/usecase/log_usecase.go
package usecase

import (
	"context"
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	"github.com/your-org/mcp-k8s-server/internal/infrastructure"
	corev1 "k8s.io/api/core/v1"
)

type LogUseCase struct {
	clusterRepo    domain.ClusterRepository
	clientFactory  infrastructure.ClientFactory
	clusterManager *infrastructure.ClusterManager
}

func (uc *LogUseCase) GetPodLogs(ctx context.Context, clusterID domain.ClusterID, namespace domain.Namespace, podName domain.PodName, options domain.LogOptions) (domain.PodLogs, error) {

	if uc.clusterManager != nil {
		return uc.clusterManager.GetPodLogs(ctx, clusterID, namespace, podName, options)
	}

	cluster, err := uc.clusterRepo.FindByID(clusterID)
	if err != nil {
		return "", fmt.Errorf("cluster not found: %w", err)
	}

	client, err := uc.clientFactory.CreateClient(cluster.Config, cluster.Config.Context)
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}

	req := client.CoreV1().Pods(string(namespace)).GetLogs(string(podName), &corev1.PodLogOptions{
		TailLines: options.TailLines,
	})

	logs, err := req.Do(ctx).Raw()
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}

	return domain.PodLogs(logs), nil
}
