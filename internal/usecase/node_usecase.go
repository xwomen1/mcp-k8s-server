package usecase

import (
	"context"
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (uc *K8sUseCase) ListNodes(ctx context.Context, clusterID string) ([]domain.Node, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	nodeList, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodes := make([]domain.Node, 0, len(nodeList.Items))
	for _, node := range nodeList.Items {
		status := "Unknown"
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				status = string(condition.Status)
				break
			}
		}

		nodes = append(nodes, domain.Node{
			Name:       domain.NodeName(node.Name),
			Status:     status,
			Roles:      extractNodeRoles(node.Labels),
			Version:    node.Status.NodeInfo.KubeletVersion,
			InternalIP: getInternalIP(node.Status.Addresses),
			Labels:     node.Labels,
			CreatedAt:  node.CreationTimestamp.Time,
		})
	}

	return nodes, nil
}

func (uc *K8sUseCase) GetNodeMetrics(ctx context.Context, clusterID, nodeName string) (domain.NodeMetrics, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	// Return zero value of NodeMetrics and error
	if err != nil {
		return domain.NodeMetrics{}, fmt.Errorf("failed to get client: %w", err)
	}

	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return domain.NodeMetrics{}, fmt.Errorf("failed to get node: %w", err)
	}

	capacity := map[string]string{
		"cpu":    node.Status.Capacity.Cpu().String(),
		"memory": node.Status.Capacity.Memory().String(),
		"pods":   node.Status.Capacity.Pods().String(),
	}

	allocatable := map[string]string{
		"cpu":    node.Status.Allocatable.Cpu().String(),
		"memory": node.Status.Allocatable.Memory().String(),
		"pods":   node.Status.Allocatable.Pods().String(),
	}

	return domain.NodeMetrics{
		NodeName:    domain.NodeName(nodeName),
		Capacity:    capacity,
		Allocatable: allocatable,
		Labels:      node.Labels,
	}, nil
}

// getInternalIP remains the same helper function.
func getInternalIP(addresses []corev1.NodeAddress) string {
	for _, addr := range addresses {
		if addr.Type == corev1.NodeInternalIP {
			return addr.Address
		}
	}
	return "N/A"
}

// extractNodeRoles remains the same helper function.
func extractNodeRoles(labels map[string]string) string {
	roles := []string{}
	// Only check for standard roles for simplicity
	if _, ok := labels["node-role.kubernetes.io/control-plane"]; ok {
		roles = append(roles, "Control-Plane")
	}
	if _, ok := labels["node-role.kubernetes.io/master"]; ok { // for older clusters
		roles = append(roles, "Control-Plane")
	}
	// Worker is often implied if not Control-Plane, but we can check the label
	if _, ok := labels["node-role.kubernetes.io/worker"]; ok {
		roles = append(roles, "Worker")
	}

	if len(roles) == 0 {
		return "Worker" // Default assumption in many setups if no explicit label
	}

	// Join unique roles, or just return the main one. Let's return the main one:
	return roles[0]
}
