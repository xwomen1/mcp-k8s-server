// internal/domain/webhook.go
package domain

import "time"

type WebhookName string

type WebhookRule struct {
	Operations []string `json:"operations"` // CREATE, UPDATE, DELETE, CONNECT
	APIGroups  []string `json:"api_groups"`
	Resources  []string `json:"resources"`
	Scope      string   `json:"scope"` // Cluster, Namespaced
}

type WebhookClientConfig struct {
	Service string `json:"service"`
	Path    string `json:"path"`
}

type ValidatingWebhook struct {
	Name          WebhookName         `json:"name"`
	WebhooksCount int                 `json:"webhooks_count"`
	ClientConfig  WebhookClientConfig `json:"client_config"`
	FailurePolicy string              `json:"failure_policy"`
	Rules         []WebhookRule       `json:"rules"`
	Labels        map[string]string   `json:"labels,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
}

type MutatingWebhook struct {
	Name          WebhookName         `json:"name"`
	WebhooksCount int                 `json:"webhooks_count"`
	ClientConfig  WebhookClientConfig `json:"client_config"`
	FailurePolicy string              `json:"failure_policy"`
	Rules         []WebhookRule       `json:"rules"`
	Labels        map[string]string   `json:"labels,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
}
