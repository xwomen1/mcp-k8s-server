package usecase

import (
	"context"
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/your-org/mcp-k8s-server/internal/domain"
)

func (uc *K8sUseCase) ListHPAs(ctx context.Context, clusterID, namespace string) ([]domain.HPA, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	hpaList, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list HPAs in namespace %s: %w", namespace, err)
	}

	hpas := make([]domain.HPA, 0, len(hpaList.Items))
	for _, hpa := range hpaList.Items {
		// Chuyển đổi từ k8s API object sang domain struct
		hpas = append(hpas, domain.HPA{
			Name:            domain.HPAName(hpa.Name),
			Namespace:       domain.Namespace(namespace),
			TargetKind:      hpa.Spec.ScaleTargetRef.Kind,
			TargetName:      hpa.Spec.ScaleTargetRef.Name,
			MinReplicas:     *hpa.Spec.MinReplicas,
			MaxReplicas:     hpa.Spec.MaxReplicas,
			CurrentReplicas: hpa.Status.CurrentReplicas,
			DesiredReplicas: hpa.Status.DesiredReplicas,

			CreatedAt: hpa.CreationTimestamp.Time,
		})
	}

	return hpas, nil
}

func (uc *K8sUseCase) GetHPADetail(ctx context.Context, clusterID, namespace, name string) (domain.HPA, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return domain.HPA{}, fmt.Errorf("failed to get client: %w", err)
	}

	hpa, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return domain.HPA{}, fmt.Errorf("failed to get HPA %s/%s: %w", namespace, name, err)
	}

	domainMetrics := convertHPAMetrics(hpa.Spec.Metrics)
	domainConditions := convertHPAConditions(hpa.Status.Conditions)

	return domain.HPA{
		Name:            domain.HPAName(hpa.Name),
		Namespace:       domain.Namespace(namespace),
		TargetKind:      hpa.Spec.ScaleTargetRef.Kind,
		TargetName:      hpa.Spec.ScaleTargetRef.Name,
		MinReplicas:     *hpa.Spec.MinReplicas,
		MaxReplicas:     hpa.Spec.MaxReplicas,
		CurrentReplicas: hpa.Status.CurrentReplicas,
		DesiredReplicas: hpa.Status.DesiredReplicas,
		Metrics:         domainMetrics,
		Conditions:      domainConditions,
		Labels:          hpa.Labels,
		CreatedAt:       hpa.CreationTimestamp.Time,
	}, nil
}

func convertHPAMetrics(k8sMetrics []autoscalingv2.MetricSpec) []domain.HPAMetric {
	domainMetrics := make([]domain.HPAMetric, 0, len(k8sMetrics))
	for _, metric := range k8sMetrics {
		dMetric := domain.HPAMetric{
			Type: domain.HPAMetricType(metric.Type),
		}

		switch metric.Type {
		case autoscalingv2.ResourceMetricSourceType:
			dMetric.ResourceName = string(metric.Resource.Name)
			if metric.Resource.Target.AverageUtilization != nil {
				dMetric.TargetValue = fmt.Sprintf("%d%%", *metric.Resource.Target.AverageUtilization)
				dMetric.TargetType = "AverageUtilization"
			} else if metric.Resource.Target.AverageValue != nil {
				dMetric.TargetValue = metric.Resource.Target.AverageValue.String()
				dMetric.TargetType = "AverageValue"
			}

		}
		domainMetrics = append(domainMetrics, dMetric)
	}
	return domainMetrics
}

func convertHPAConditions(k8sConditions []autoscalingv2.HorizontalPodAutoscalerCondition) []domain.HPACondition {
	domainConditions := make([]domain.HPACondition, 0, len(k8sConditions))
	for _, cond := range k8sConditions {
		domainConditions = append(domainConditions, domain.HPACondition{
			Type:           domain.HPAStatusType(cond.Type),
			Status:         string(cond.Status),
			Reason:         cond.Reason,
			Message:        cond.Message,
			LastTransition: cond.LastTransitionTime.Time,
		})
	}
	return domainConditions
}

func (uc *K8sUseCase) DeleteHPA(ctx context.Context, clusterID, namespace, name string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	err = client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete HPA %s/%s: %w", namespace, name, err)
	}

	return nil
}
