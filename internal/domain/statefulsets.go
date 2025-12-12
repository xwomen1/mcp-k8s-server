// internal/domain/statefulset.go
package domain

import "time"

type StatefulSetName string

type StatefulSetStatus string

const (
	StatefulSetStatusReady       StatefulSetStatus = "Ready"
	StatefulSetStatusProgressing StatefulSetStatus = "Progressing"
	StatefulSetStatusFailed      StatefulSetStatus = "Failed"
)

type StatefulSet struct {
	Name            StatefulSetName   `json:"name"`
	Namespace       Namespace         `json:"namespace"`
	Replicas        int32             `json:"replicas"`
	ReadyReplicas   int32             `json:"ready_replicas"`
	CurrentReplicas int32             `json:"current_replicas"`
	UpdatedReplicas int32             `json:"updated_replicas"`
	Status          StatefulSetStatus `json:"status"`
	ServiceName     string            `json:"service_name"`
	Labels          map[string]string `json:"labels,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
}

type StatefulSetScaleOptions struct {
	Replicas int32 `json:"replicas"`
}

type StatefulSetCreateOptions struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Replicas    int32             `json:"replicas"`
	ServiceName string            `json:"service_name"`
	Labels      map[string]string `json:"labels,omitempty"`
}
