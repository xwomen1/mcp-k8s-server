package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
)

func (uc *K8sUseCase) ApplyTaintToNode(ctx context.Context, clusterID, nodeName, taintKey, action string) (domain.Node, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return domain.Node{}, fmt.Errorf("failed to get client: %w", err)
	}

	taint, err := parseTaintKey(taintKey)
	if err != nil {
		return domain.Node{}, err
	}

	var patchData []byte
	var currentTaints []corev1.Taint

	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return domain.Node{}, fmt.Errorf("failed to get node: %w", err)
	}
	currentTaints = node.Spec.Taints

	if action == "add" {

		exists := false
		for _, t := range currentTaints {
			if t.Key == taint.Key && t.Value == taint.Value && t.Effect == taint.Effect {
				exists = true
				break
			}
		}
		if !exists {
			currentTaints = append(currentTaints, taint)
		}

		patchStructure := map[string]any{
			"spec": map[string]any{
				"taints": currentTaints,
			},
		}
		patchData, err = json.Marshal(patchStructure)
		if err != nil {
			return domain.Node{}, fmt.Errorf("failed to marshal patch data: %w", err)
		}

	} else if action == "remove" {

		newTaints := make([]corev1.Taint, 0)
		removed := false
		for _, t := range currentTaints {
			if !(t.Key == taint.Key && t.Effect == taint.Effect) {
				newTaints = append(newTaints, t)
			} else {
				removed = true
			}
		}

		if !removed {
			return domain.Node{}, fmt.Errorf("taint '%s' not found on node '%s'", taintKey, nodeName)
		}

		patchStructure := map[string]any{
			"spec": map[string]any{
				"taints": newTaints,
			},
		}
		patchData, err = json.Marshal(patchStructure)
		if err != nil {
			return domain.Node{}, fmt.Errorf("failed to marshal patch data: %w", err)
		}

	} else {
		return domain.Node{}, fmt.Errorf("invalid action '%s', must be 'add' or 'remove'", action)
	}

	patchedNode, err := client.CoreV1().Nodes().Patch(ctx, nodeName, types.StrategicMergePatchType, patchData, metav1.PatchOptions{})
	if err != nil {
		return domain.Node{}, fmt.Errorf("failed to patch node %s: %w", nodeName, err)
	}

	return convertK8sNodeToDomain(*patchedNode), nil
}

// Helper: parseTaintKey
func parseTaintKey(taintKey string) (corev1.Taint, error) {
	parts := strings.Split(taintKey, ":")
	if len(parts) != 2 {
		return corev1.Taint{}, fmt.Errorf("invalid taint format. Expected 'key=value:effect' or 'key:effect'")
	}

	keyEffect := parts[0]
	effect := corev1.TaintEffect(parts[1])

	keyParts := strings.Split(keyEffect, "=")
	key := keyParts[0]
	value := ""
	if len(keyParts) == 2 {
		value = keyParts[1]
	}

	validEffects := map[corev1.TaintEffect]bool{
		corev1.TaintEffectNoSchedule:       true,
		corev1.TaintEffectPreferNoSchedule: true,
		corev1.TaintEffectNoExecute:        true,
	}
	if !validEffects[effect] {
		return corev1.Taint{}, fmt.Errorf("invalid taint effect: %s. Must be NoSchedule, PreferNoSchedule, or NoExecute", effect)
	}

	taintTime := metav1.Now()
	return corev1.Taint{
		Key:    key,
		Value:  value,
		Effect: effect,

		TimeAdded: &taintTime,
	}, nil
}

func convertK8sTaintsToDomain(k8sTaints []corev1.Taint) []domain.Taint {
	domainTaints := make([]domain.Taint, 0, len(k8sTaints))
	for _, t := range k8sTaints {
		domainTaints = append(domainTaints, domain.Taint{
			Key:       t.Key,
			Value:     t.Value,
			Effect:    string(t.Effect),
			TimeAdded: t.TimeAdded.Time,
		})
	}
	return domainTaints
}

// convertK8sNodeToDomain chuyển đổi k8s Node API object thành domain.Node.
// Hàm này phải được cập nhật để bao gồm cả trường Taints mới.
func convertK8sNodeToDomain(k8sNode corev1.Node) domain.Node {

	// Trích xuất IP
	internalIP := "N/A"
	externalIP := "N/A"
	for _, addr := range k8sNode.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			internalIP = addr.Address
		}
		if addr.Type == corev1.NodeExternalIP {
			externalIP = addr.Address
		}
	}

	// Trích xuất Roles
	roles := "" // Logic phức tạp hơn có thể cần thiết cho Roles

	// Kiểm tra nhãn thứ nhất: master
	_, isMaster := k8sNode.Labels["node-role.kubernetes.io/master"]
	// Kiểm tra nhãn thứ hai: control-plane
	_, isControlPlane := k8sNode.Labels["node-role.kubernetes.io/control-plane"]

	if isMaster || isControlPlane {
		roles = "ControlPlane"
	} else {
		roles = "Worker"
	}

	// Trích xuất Status (Ready/NotReady)
	status := "Unknown"
	for _, condition := range k8sNode.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				status = "Ready"
			} else {
				status = "NotReady"
			}
			break
		}
	}

	domainTaints := convertK8sTaintsToDomain(k8sNode.Spec.Taints)

	return domain.Node{
		Name: domain.NodeName(k8sNode.Name),

		Status:        status,
		Roles:         roles,
		InternalIP:    internalIP,
		ExternalIP:    externalIP,
		Unschedulable: k8sNode.Spec.Unschedulable,
		Taints:        domainTaints,
		Capacity:      formatResourceList(k8sNode.Status.Capacity),
		Allocatable:   formatResourceList(k8sNode.Status.Allocatable),
	}
}
