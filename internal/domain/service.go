package domain

type ServiceName string

type ServiceType string

const (
	ServiceTypeClusterIP    ServiceType = "ClusterIP"
	ServiceTypeNodePort     ServiceType = "NodePort"
	ServiceTypeLoadBalancer ServiceType = "LoadBalancer"
)

type Service struct {
	Name      ServiceName       `json:"name"`
	Namespace Namespace         `json:"namespace"`
	Type      ServiceType       `json:"type"`
	ClusterIP string            `json:"cluster_ip"`
	Ports     []ServicePort     `json:"ports"`
	Labels    map[string]string `json:"labels"`
}

type ServicePort struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"target_port"`
	NodePort   int32  `json:"node_port,omitempty"`
}
