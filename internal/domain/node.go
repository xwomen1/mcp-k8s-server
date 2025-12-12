// internal/domain/node.go
package domain

import "time"

type NodeName string

type Node struct {
	Name          NodeName          `json:"name"`
	ClusterID     ClusterID         `json:"cluster_id"`
	Unschedulable bool              `json:"unschedulable"`
	Taints        []Taint           `json:"taints,omitempty"`
	Capacity      map[string]string `json:"capacity"`
	Allocatable   map[string]string `json:"allocatable"`
	Status        string            `json:"status"` // Ready, NotReady, Unknown
	Roles         string            `json:"roles"`  // Control-Plane, Worker, None
	Version       string            `json:"version"`
	InternalIP    string            `json:"internal_ip"`
	ExternalIP    string            `json:"external_ip"`
	Labels        map[string]string `json:"labels,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}

type NodeMetrics struct {
	NodeName    NodeName          `json:"node_name"`
	Capacity    map[string]string `json:"capacity"`
	Allocatable map[string]string `json:"allocatable"`
	Labels      map[string]string `json:"labels"`
}

type Taint struct {
	Key       string    `json:"key"`
	Value     string    `json:"value,omitempty"`
	Effect    string    `json:"effect"` // NoSchedule, PreferNoSchedule, NoExecute
	TimeAdded time.Time `json:"time_added,omitempty"`
}

type Toleration struct {
	Key      string `json:"key,omitempty"`
	Operator string `json:"operator,omitempty"` // Equal, Exists
	Value    string `json:"value,omitempty"`
	Effect   string `json:"effect,omitempty"`
	Seconds  *int64 `json:"seconds,omitempty"` // seconds for NoExecute
}
