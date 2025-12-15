package usecase

import (
	"context"
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

// kubectl apply
func (uc *K8sUseCase) ApplyYAML(ctx context.Context, clusterID string, yamlBody string, manager string, dryRun bool) (*domain.ApplyResult, error) {
	restConfig, err := uc.clusterManager.GetRESTConfig(domain.ClusterID(clusterID))
	if err != nil {
		return nil, err
	}

	dynClient, _ := dynamic.NewForConfig(restConfig)
	discClient, _ := discovery.NewDiscoveryClientForConfig(restConfig)

	// Decode YAML
	obj := &unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode([]byte(yamlBody), nil, obj)
	if err != nil {
		return nil, fmt.Errorf("decode YAML failed: %w", err)
	}

	// Mapping GVR
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discClient))
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = dynClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		dr = dynClient.Resource(mapping.Resource)
	}

	data, _ := obj.MarshalJSON()
	_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: manager,
		Force:        ptrBool(true),
	})
	if err != nil {
		return nil, err
	}

	patchOptions := metav1.PatchOptions{
		FieldManager: manager,
		Force:        ptrBool(true),
	}

	if dryRun {
		patchOptions.DryRun = []string{metav1.DryRunAll}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to apply (dry-run=%v): %w", dryRun, err)
	}

	return &domain.ApplyResult{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Kind:      obj.GetKind(),
		DryRun:    dryRun,
	}, nil

}
