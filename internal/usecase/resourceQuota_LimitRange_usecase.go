package usecase

import (
	"context"
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (uc *K8sUseCase) ListResourceQuotas(ctx context.Context, clusterID, namespace string) ([]map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	quotaList, err := client.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list resource quotas in namespace %s: %w", namespace, err)
	}

	quotas := make([]map[string]any, 0, len(quotaList.Items))
	for _, quota := range quotaList.Items {
		quotas = append(quotas, map[string]any{
			"name":      quota.Name,
			"namespace": namespace,
			"status":    formatResourceQuotaStatus(quota.Status),
			"age":       quota.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
		})
	}

	return quotas, nil
}

func (uc *K8sUseCase) GetResourceQuotaDetail(ctx context.Context, clusterID, namespace, name string) (map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	quota, err := client.CoreV1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get resource quota %s/%s: %w", namespace, name, err)
	}

	return map[string]any{
		"name":        quota.Name,
		"namespace":   namespace,
		"hard_limits": formatResourceList(quota.Spec.Hard),
		"used":        formatResourceList(quota.Status.Used),
		"allowed":     formatResourceList(quota.Status.Hard),
		"age":         quota.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
		"raw_object":  quota,
	}, nil
}

func (uc *K8sUseCase) ListLimitRanges(ctx context.Context, clusterID, namespace string) ([]map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	limitRangeList, err := client.CoreV1().LimitRanges(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list limit ranges in namespace %s: %w", namespace, err)
	}

	ranges := make([]map[string]any, 0, len(limitRangeList.Items))
	for _, lr := range limitRangeList.Items {
		ranges = append(ranges, map[string]any{
			"name":         lr.Name,
			"namespace":    namespace,
			"limits_count": len(lr.Spec.Limits),
			"age":          lr.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
		})
	}

	return ranges, nil
}

func (uc *K8sUseCase) GetLimitRangeDetail(ctx context.Context, clusterID, namespace, name string) (map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	lr, err := client.CoreV1().LimitRanges(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get limit range %s/%s: %w", namespace, name, err)
	}

	formattedLimits := []map[string]any{}
	for _, limit := range lr.Spec.Limits {
		formattedLimits = append(formattedLimits, map[string]any{
			"type":                    limit.Type,
			"max":                     formatResourceList(limit.Max),
			"min":                     formatResourceList(limit.Min),
			"default":                 formatResourceList(limit.Default),
			"default_request":         formatResourceList(limit.DefaultRequest),
			"max_limit_request_ratio": formatResourceList(limit.MaxLimitRequestRatio),
		})
	}

	return map[string]any{
		"name":       lr.Name,
		"namespace":  namespace,
		"limits":     formattedLimits,
		"raw_object": lr,
	}, nil
}

func formatResourceList(list corev1.ResourceList) map[string]string {
	formatted := make(map[string]string)
	for name, quantity := range list {

		formatted[string(name)] = quantity.String()
	}
	return formatted
}

func formatResourceQuotaStatus(status corev1.ResourceQuotaStatus) map[string]map[string]string {
	result := make(map[string]map[string]string)
	result["used"] = formatResourceList(status.Used)
	result["hard"] = formatResourceList(status.Hard)
	return result
}
