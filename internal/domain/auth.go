// internal/domain/auth.go
package domain

import "time"

type RoleName string
type SubjectName string

type PolicyRule struct {
	Verbs           []string `json:"verbs"`
	APIGroups       []string `json:"api_groups"`
	Resources       []string `json:"resources"`
	ResourceNames   []string `json:"resource_names,omitempty"`
	NonResourceURLs []string `json:"non_resource_urls,omitempty"`
}

type Subject struct {
	Kind      string    `json:"kind"`
	Name      string    `json:"name"`
	Namespace Namespace `json:"namespace,omitempty"`
}

type Role struct {
	Name      RoleName          `json:"name"`
	Namespace Namespace         `json:"namespace"`
	Rules     []PolicyRule      `json:"rules"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type ClusterRole struct {
	Name      RoleName          `json:"name"`
	Rules     []PolicyRule      `json:"rules"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type RoleBinding struct {
	Name        RoleName          `json:"name"`
	Namespace   Namespace         `json:"namespace"`
	RoleRefName string            `json:"role_ref_name"`
	RoleRefKind string            `json:"role_ref_kind"`
	Subjects    []Subject         `json:"subjects"`
	Labels      map[string]string `json:"labels,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

type ClusterRoleBinding struct {
	Name        RoleName          `json:"name"`
	RoleRefName string            `json:"role_ref_name"`
	RoleRefKind string            `json:"role_ref_kind"`
	Subjects    []Subject         `json:"subjects"`
	Labels      map[string]string `json:"labels,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}
