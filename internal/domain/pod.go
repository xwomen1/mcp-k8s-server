package domain

import "time"

type PodName string

type PodPhase string

const (
	PodPhasePending   PodPhase = "Pending"
	PodPhaseRunning   PodPhase = "Running"
	PodPhaseSucceeded PodPhase = "Succeeded"
	PodPhaseFailed    PodPhase = "Failed"
	PodPhaseUnknown   PodPhase = "Unknown"
)

type PodStatus struct {
	Phase      PodPhase       `json:"phase"`
	Message    string         `json:"message,omitempty"`
	Reason     string         `json:"reason,omitempty"`
	StartTime  string         `json:"start_time,omitempty"`
	Conditions []PodCondition `json:"conditions,omitempty"`
}

type PodCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type Pod struct {
	Name      PodName           `json:"name"`
	Namespace Namespace         `json:"namespace"`
	Status    PodStatus         `json:"status"`
	Labels    map[string]string `json:"labels,omitempty"`
	ClusterID ClusterID         `json:"clusterID,omitempty"`
	IP        string            `json:"ip"`
	CreatedAt time.Time         `json:"created_at"`
}

type PortForwardResult struct {
	PodName    string `json:"pod_name"`
	LocalPort  int    `json:"local_port"`
	RemotePort int    `json:"remote_port"`
	URL        string `json:"url"`
	Status     string `json:"status"`
}

type NamespaceResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
type LogOptions struct {
	TailLines *int64 `json:"tail_lines,omitempty"`
	Follow    bool   `json:"follow,omitempty"`
	Container string `json:"container,omitempty"`
	Previous  bool   `json:"previous,omitempty"`
}

type PodLogs string
