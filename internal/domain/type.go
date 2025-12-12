// internal/domain/types.go
package domain

import "time"

type ClusterMetadata struct {
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
}

type ContainerStatus struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restart_count"`
	Image        string `json:"image"`
	State        string `json:"state"`
}
