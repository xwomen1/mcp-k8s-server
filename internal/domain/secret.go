// internal/domain/secret.go
package domain

import "time"

type SecretName string

type SecretType string

const (
	SecretTypeOpaque              SecretType = "Opaque"
	SecretTypeServiceAccountToken SecretType = "kubernetes.io/service-account-token"
	SecretTypeDockerConfigJson    SecretType = "kubernetes.io/dockerconfigjson"
	SecretTypeBasicAuth           SecretType = "kubernetes.io/basic-auth"
	SecretTypeSSHAuth             SecretType = "kubernetes.io/ssh-auth"
	SecretTypeTLS                 SecretType = "kubernetes.io/tls"
)

type Secret struct {
	Name       SecretName        `json:"name"`
	Namespace  Namespace         `json:"namespace"`
	Type       SecretType        `json:"type"`
	Data       map[string][]byte `json:"data"`
	StringData map[string]string `json:"string_data,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

type SecretCreateOptions struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Type       string            `json:"type"`
	StringData map[string]string `json:"string_data"`
	Labels     map[string]string `json:"labels,omitempty"`
}
