package domain

import "time"

type ConfigMapName string

type ConfigMap struct {
	Name       ConfigMapName     `json:"name"`
	Namespace  Namespace         `json:"namespace"`
	Data       map[string]string `json:"data"`
	BinaryData map[string][]byte `json:"binary_data,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

type ConfigMapCreateOptions struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Data      map[string]string `json:"data"`
	Labels    map[string]string `json:"labels,omitempty"`
}
