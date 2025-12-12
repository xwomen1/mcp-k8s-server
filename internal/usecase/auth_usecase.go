package usecase

import (
	"context"
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListClusterRoles lists all ClusterRoles and converts them to domain.ClusterRole.
func (uc *K8sUseCase) ListClusterRoles(ctx context.Context, clusterID string) ([]domain.ClusterRole, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	crList, err := client.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ClusterRoles: %w", err)
	}

	clusterRoles := make([]domain.ClusterRole, 0, len(crList.Items))
	for _, cr := range crList.Items {
		clusterRoles = append(clusterRoles, convertK8sClusterRoleToDomain(cr))
	}

	return clusterRoles, nil
}

// Helper: convertK8sPolicyRuleToDomain converts k8s rules to domain rules.
func convertK8sPolicyRuleToDomain(k8sRules []rbacv1.PolicyRule) []domain.PolicyRule {
	domainRules := make([]domain.PolicyRule, 0, len(k8sRules))
	for _, r := range k8sRules {
		domainRules = append(domainRules, domain.PolicyRule{
			Verbs:           r.Verbs,
			APIGroups:       r.APIGroups,
			Resources:       r.Resources,
			ResourceNames:   r.ResourceNames,
			NonResourceURLs: r.NonResourceURLs,
		})
	}
	return domainRules
}

// Helper: convertK8sClusterRoleToDomain converts a k8s ClusterRole API object to a domain.ClusterRole struct.
func convertK8sClusterRoleToDomain(k8sCR rbacv1.ClusterRole) domain.ClusterRole {
	return domain.ClusterRole{
		Name:      domain.RoleName(k8sCR.Name),
		Rules:     convertK8sPolicyRuleToDomain(k8sCR.Rules),
		Labels:    k8sCR.Labels,
		CreatedAt: k8sCR.CreationTimestamp.Time,
	}
}
