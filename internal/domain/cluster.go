// internal/domain/cluster.go
package domain

import "time"

type ClusterID string

type ClusterConfig struct {
	KubeconfigPath string `json:"kubeconfig_path"`
	KubeconfigData []byte `json:"kubeconfig_data,omitempty"`
	Context        string `json:"context,omitempty"`
	InCluster      bool   `json:"in_cluster,omitempty"`
}

type ClusterStatus string

const (
	ClusterStatusActive  ClusterStatus = "active"
	ClusterStatusError   ClusterStatus = "error"
	ClusterStatusUnknown ClusterStatus = "unknown"
)

type Cluster struct {
	ID          ClusterID     `json:"id"`
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	Config      ClusterConfig `json:"config"`
	Status      ClusterStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	IsActive    bool          `json:"is_active"`
}

type ClusterInfo struct {
	ID          ClusterID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Status      ClusterStatus   `json:"status"`
	Config      ClusterConfig   `json:"config"`
	Metadata    ClusterMetadata `json:"metadata"`
}

type ClusterMetrics struct {
	NodeCount int       `json:"node_count"`
	PodCount  int       `json:"pod_count"`
	LastCheck time.Time `json:"last_check"`
}

type ClusterRepository interface {
	Save(cluster *Cluster) error
	FindByID(id ClusterID) (*Cluster, error)

	FindByName(name string) (*Cluster, error)
	FindAll() ([]*Cluster, error)
	Update(cluster *Cluster) error
	Delete(id ClusterID) error
	Get(clusterID string) (*Cluster, error)
}
