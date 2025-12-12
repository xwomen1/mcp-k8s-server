package usecase

import (
	"context"
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListEvents lists events in a namespace, optionally filtered by an involved object.
func (uc *K8sUseCase) ListEvents(ctx context.Context, clusterID, namespace, involvedKind, involvedName string) ([]domain.Event, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	opts := metav1.ListOptions{}

	// Construct Field Selector if filtering by object is requested
	if involvedKind != "" && involvedName != "" {
		// Event API uses "involvedObject.name" and "involvedObject.kind" for filtering
		opts.FieldSelector = fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=%s", involvedName, involvedKind)
	}

	eventList, err := client.CoreV1().Events(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list events in namespace %s: %w", namespace, err)
	}

	events := make([]domain.Event, 0, len(eventList.Items))
	for _, event := range eventList.Items {
		events = append(events, convertK8sEventToDomain(event))
	}

	return events, nil
}

// Helper: convertK8sEventToDomain converts a k8s Event API object to a domain.Event struct.
func convertK8sEventToDomain(k8sEvent corev1.Event) domain.Event {
	// Determine the source component and host
	sourceComponent := k8sEvent.Source.Component
	sourceHost := k8sEvent.Source.Host

	// Event source can be empty, especially in recent k8s versions or custom controllers
	if sourceComponent == "" {
		sourceComponent = "N/A"
	}
	if sourceHost == "" {
		sourceHost = "N/A"
	}

	return domain.Event{
		Name:            domain.EventName(k8sEvent.Name),
		Namespace:       domain.Namespace(k8sEvent.Namespace),
		Type:            domain.EventType(k8sEvent.Type),
		Reason:          k8sEvent.Reason,
		Message:         k8sEvent.Message,
		SourceComponent: sourceComponent,
		SourceHost:      sourceHost,
		InvolvedKind:    k8sEvent.InvolvedObject.Kind,
		InvolvedName:    k8sEvent.InvolvedObject.Name,
		InvolvedUID:     string(k8sEvent.InvolvedObject.UID),
		Count:           k8sEvent.Count,
		FirstTimestamp:  k8sEvent.FirstTimestamp.Time,
		LastTimestamp:   k8sEvent.LastTimestamp.Time,
	}
}
