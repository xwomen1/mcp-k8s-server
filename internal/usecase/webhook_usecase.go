package usecase

import (
	"context"
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// --- Validating Webhooks ---

// ListValidatingWebhooks lists all ValidatingWebhookConfigurations.
func (uc *K8sUseCase) ListValidatingWebhooks(ctx context.Context, clusterID string) ([]domain.ValidatingWebhook, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	list, err := client.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ValidatingWebhookConfigurations: %w", err)
	}

	webhooks := make([]domain.ValidatingWebhook, 0, len(list.Items))
	for _, item := range list.Items {
		webhooks = append(webhooks, convertValidatingWebhookToDomain(item))
	}
	return webhooks, nil
}

// GetValidatingWebhook retrieves a specific ValidatingWebhookConfiguration.
func (uc *K8sUseCase) GetValidatingWebhook(ctx context.Context, clusterID, name string) (domain.ValidatingWebhook, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return domain.ValidatingWebhook{}, fmt.Errorf("failed to get client: %w", err)
	}

	item, err := client.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return domain.ValidatingWebhook{}, fmt.Errorf("failed to get ValidatingWebhookConfiguration %s: %w", name, err)
	}

	return convertValidatingWebhookToDomain(*item), nil
}

// DeleteValidatingWebhook deletes a ValidatingWebhookConfiguration.
func (uc *K8sUseCase) DeleteValidatingWebhook(ctx context.Context, clusterID, name string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}
	return client.AdmissionregistrationV1().ValidatingWebhookConfigurations().Delete(ctx, name, metav1.DeleteOptions{})
}

// --- Mutating Webhooks ---

// ListMutatingWebhooks lists all MutatingWebhookConfigurations.
func (uc *K8sUseCase) ListMutatingWebhooks(ctx context.Context, clusterID string) ([]domain.MutatingWebhook, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	list, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list MutatingWebhookConfigurations: %w", err)
	}

	webhooks := make([]domain.MutatingWebhook, 0, len(list.Items))
	for _, item := range list.Items {
		webhooks = append(webhooks, convertMutatingWebhookToDomain(item))
	}
	return webhooks, nil
}

// GetMutatingWebhook retrieves a specific MutatingWebhookConfiguration.
func (uc *K8sUseCase) GetMutatingWebhook(ctx context.Context, clusterID, name string) (domain.MutatingWebhook, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return domain.MutatingWebhook{}, fmt.Errorf("failed to get client: %w", err)
	}

	item, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return domain.MutatingWebhook{}, fmt.Errorf("failed to get MutatingWebhookConfiguration %s: %w", name, err)
	}

	return convertMutatingWebhookToDomain(*item), nil
}

// DeleteMutatingWebhook deletes a MutatingWebhookConfiguration.
func (uc *K8sUseCase) DeleteMutatingWebhook(ctx context.Context, clusterID, name string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}
	return client.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(ctx, name, metav1.DeleteOptions{})
}

// --- Helper Functions ---

func convertWebhookRules(rules []admissionv1.RuleWithOperations) []domain.WebhookRule {
	domainRules := make([]domain.WebhookRule, 0)

	if len(rules) == 0 {
		return domainRules
	}

	r := rules[0]
	// convert scopeValueto string
	scopeValue := ""
	if r.Scope != nil {
		scopeValue = string(*r.Scope)
	}

	domainRules = append(domainRules, domain.WebhookRule{
		Operations: convertAdmissionOperations(r.Operations),
		APIGroups:  r.APIGroups,
		Resources:  r.Resources,
		Scope:      scopeValue,
	})

	return domainRules
}

func convertAdmissionOperations(ops []admissionv1.OperationType) []string {
	strOps := make([]string, len(ops))
	for i, op := range ops {
		strOps[i] = string(op)
	}
	return strOps
}

func extractClientConfig(webhooks []admissionv1.MutatingWebhook) (domain.WebhookClientConfig, string, []admissionv1.RuleWithOperations) {
	if len(webhooks) == 0 {
		return domain.WebhookClientConfig{}, "", nil
	}

	// Use the first webhook for summary extraction
	firstWebhook := webhooks[0]

	config := domain.WebhookClientConfig{}
	if firstWebhook.ClientConfig.Service != nil {
		config.Service = fmt.Sprintf("%s/%s", firstWebhook.ClientConfig.Service.Namespace, firstWebhook.ClientConfig.Service.Name)
		config.Path = *firstWebhook.ClientConfig.Service.Path
	} else if firstWebhook.ClientConfig.URL != nil {
		config.Service = "External URL"
		config.Path = *firstWebhook.ClientConfig.URL
	}

	return config, string(*firstWebhook.FailurePolicy), firstWebhook.Rules
}

func convertMutatingWebhookToDomain(item admissionv1.MutatingWebhookConfiguration) domain.MutatingWebhook {
	config, policy, rules := extractClientConfig(item.Webhooks)

	return domain.MutatingWebhook{
		Name:          domain.WebhookName(item.Name),
		WebhooksCount: len(item.Webhooks),
		ClientConfig:  config,
		FailurePolicy: policy,
		Rules:         convertWebhookRules(rules),
		Labels:        item.Labels,
		CreatedAt:     item.CreationTimestamp.Time,
	}
}

func convertValidatingWebhookToDomain(item admissionv1.ValidatingWebhookConfiguration) domain.ValidatingWebhook {
	// Note: ValidatingWebhookConfiguration uses the same ClientConfig structure

	config := domain.WebhookClientConfig{}
	policy := ""
	rules := []admissionv1.RuleWithOperations{}

	if len(item.Webhooks) > 0 {
		firstWebhook := item.Webhooks[0]
		if firstWebhook.ClientConfig.Service != nil {
			config.Service = fmt.Sprintf("%s/%s", firstWebhook.ClientConfig.Service.Namespace, firstWebhook.ClientConfig.Service.Name)
			config.Path = *firstWebhook.ClientConfig.Service.Path
		} else if firstWebhook.ClientConfig.URL != nil {
			config.Service = "External URL"
			config.Path = *firstWebhook.ClientConfig.URL
		}
		policy = string(*firstWebhook.FailurePolicy)
		rules = firstWebhook.Rules
	}

	return domain.ValidatingWebhook{
		Name:          domain.WebhookName(item.Name),
		WebhooksCount: len(item.Webhooks),
		ClientConfig:  config,
		FailurePolicy: policy,
		Rules:         convertWebhookRules(rules),
		Labels:        item.Labels,
		CreatedAt:     item.CreationTimestamp.Time,
	}
}
