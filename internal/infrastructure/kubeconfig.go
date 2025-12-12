package infrastructure

import (
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func LoadKubeconfig(clusterConfig domain.ClusterConfig) (*rest.Config, error) {
	var clientConfig clientcmd.ClientConfig

	if clusterConfig.KubeconfigData != nil && len(clusterConfig.KubeconfigData) > 0 {
		// Load from in-memory data
		kubeConfig, err := clientcmd.Load(clusterConfig.KubeconfigData)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig data: %w", err)
		}

		overrides := &clientcmd.ConfigOverrides{}
		if clusterConfig.Context != "" {
			overrides.CurrentContext = clusterConfig.Context
		}

		clientConfig = clientcmd.NewDefaultClientConfig(*kubeConfig, overrides)
	} else if clusterConfig.KubeconfigPath != "" {
		// Load from file path
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		loadingRules.ExplicitPath = clusterConfig.KubeconfigPath

		overrides := &clientcmd.ConfigOverrides{}
		if clusterConfig.Context != "" {
			overrides.CurrentContext = clusterConfig.Context
		}

		clientConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
	} else {
		// Try in-cluster config
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("no kubeconfig provided and not running in-cluster: %w", err)
		}
		return restConfig, nil
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create client config: %w", err)
	}

	return restConfig, nil
}
