package domain

type DeploymentName string

type DeploymentStatus string

const (
	DeploymentStatusAvailable      DeploymentStatus = "Available"
	DeploymentStatusProgressing    DeploymentStatus = "Progressing"
	DeploymentStatusReplicaFailure DeploymentStatus = "ReplicaFailure"
)

type Deployment struct {
	Name          DeploymentName    `json:"name"`
	Namespace     Namespace         `json:"namespace"`
	Replicas      int32             `json:"replicas"`
	ReadyReplicas int32             `json:"ready_replicas"`
	Status        DeploymentStatus  `json:"status"`
	Labels        map[string]string `json:"labels,omitempty"`
}

type ScaleOptions struct {
	Replicas int32 `json:"replicas"`
}
