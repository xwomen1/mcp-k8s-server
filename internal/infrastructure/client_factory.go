// infrastructure/client_factory.go
package infrastructure

import (
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClient tạo client kubernetes mới từ config
func NewClient(config domain.ClusterConfig, context string) (kubernetes.Interface, *rest.Config, error) {
	// Tạo rest config từ kubeconfig
	restConfig, err := loadKubeconfigWithContext(config, context)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Tạo clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return clientset, restConfig, nil
}

func loadKubeconfigWithContext(config domain.ClusterConfig, contextOverride string) (*rest.Config, error) {
	// Load kubeconfig file
	loader := clientcmd.NewDefaultClientConfigLoadingRules()
	if config.KubeconfigPath != "" {
		loader.ExplicitPath = config.KubeconfigPath
	}

	// Apply context override if provided
	overrides := &clientcmd.ConfigOverrides{}
	if contextOverride != "" {
		overrides.CurrentContext = contextOverride
	} else if config.Context != "" {
		overrides.CurrentContext = config.Context
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader,
		overrides,
	)

	return clientConfig.ClientConfig()
}

type ClientFactory interface {
	CreateClient(config domain.ClusterConfig, context string) (kubernetes.Interface, error)
}

type K8sClientFactory struct {
	logger Logger
}

func NewClientFactory(logger Logger) *K8sClientFactory {
	return &K8sClientFactory{logger: logger}
}

func (f *K8sClientFactory) CreateClient(config domain.ClusterConfig, context string) (kubernetes.Interface, error) {
	f.logger.Info("Creating Kubernetes client", "config", config, "context", context)

	var restConfig *rest.Config
	var err error

	if config.InCluster {
		// Use in-cluster config
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
		}
	} else {
		// Use kubeconfig
		restConfig, err = f.loadKubeconfig(config, context)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return clientset, nil
}

func (f *K8sClientFactory) loadKubeconfig(config domain.ClusterConfig, context string) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if config.KubeconfigPath != "" {
		loadingRules.ExplicitPath = config.KubeconfigPath
		f.logger.Debug("Using kubeconfig", "path", config.KubeconfigPath)
	}

	overrides := &clientcmd.ConfigOverrides{}

	// Priority: passed context > config context
	if context != "" {
		overrides.CurrentContext = context
	} else if config.Context != "" {
		overrides.CurrentContext = config.Context
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		overrides,
	)

	return clientConfig.ClientConfig()
}
