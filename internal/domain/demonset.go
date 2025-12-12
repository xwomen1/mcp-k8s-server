// internal/domain/daemonset.go
package domain

import "time"

type DaemonSetName string

type DaemonSetStatus string

const (
	DaemonSetStatusReady       DaemonSetStatus = "Ready"
	DaemonSetStatusProgressing DaemonSetStatus = "Progressing"
	DaemonSetStatusFailed      DaemonSetStatus = "Failed"
)

type DaemonSet struct {
	Name                   DaemonSetName     `json:"name"`
	Namespace              Namespace         `json:"namespace"`
	DesiredNumberScheduled int32             `json:"desired_number_scheduled"`
	CurrentNumberScheduled int32             `json:"current_number_scheduled"`
	NumberReady            int32             `json:"number_ready"`
	NumberAvailable        int32             `json:"number_available"`
	NumberMisscheduled     int32             `json:"number_misscheduled"`
	UpdatedNumberScheduled int32             `json:"updated_number_scheduled"`
	Status                 DaemonSetStatus   `json:"status"`
	Labels                 map[string]string `json:"labels,omitempty"`
	CreatedAt              time.Time         `json:"created_at"`
}

type DaemonSetCreateOptions struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels,omitempty"`
}
