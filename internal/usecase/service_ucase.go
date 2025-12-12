package usecase

import (
	"context"
	"fmt"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (uc *K8sUseCase) ListServices(ctx context.Context, clusterID, namespace string) ([]map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	serviceList, err := client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	services := make([]map[string]any, 0, len(serviceList.Items))
	for _, svc := range serviceList.Items {
		services = append(services, map[string]any{
			"name":       svc.Name,
			"namespace":  svc.Namespace,
			"type":       string(svc.Spec.Type),
			"cluster_ip": svc.Spec.ClusterIP,
		})
	}

	return services, nil
}

func (uc *K8sUseCase) GetService(ctx context.Context, clusterID, namespace, serviceName string) (map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	svc, err := client.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	ports := make([]map[string]any, 0, len(svc.Spec.Ports))
	for _, port := range svc.Spec.Ports {
		ports = append(ports, map[string]any{
			"name":        port.Name,
			"port":        port.Port,
			"target_port": port.TargetPort.IntValue(),
			"node_port":   port.NodePort,
		})
	}

	return map[string]any{
		"name":       svc.Name,
		"namespace":  svc.Namespace,
		"type":       string(svc.Spec.Type),
		"cluster_ip": svc.Spec.ClusterIP,
		"ports":      ports,
		"labels":     svc.Labels,
	}, nil
}
func (uc *K8sUseCase) DeleteService(ctx context.Context, clusterID, namespace, serviceName string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	err = client.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil {
		// Xử lý trường hợp Service không tồn tại
		if errors.IsNotFound(err) {
			return nil // Coi là thành công nếu nó đã không tồn tại
		}
		return fmt.Errorf("failed to delete service '%s': %w", serviceName, err)
	}

	return nil
}

// --- Ingress Implementation ---

// ListIngresses liệt kê tất cả các Ingress trong một namespace.
func (uc *K8sUseCase) ListIngresses(ctx context.Context, clusterID, namespace string) ([]map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Sử dụng client.NetworkingV1() để tương tác với Ingress (APIs networking.k8s.io/v1)
	ingressList, err := client.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ingresses: %w", err)
	}

	ingresses := make([]map[string]any, 0, len(ingressList.Items))
	for _, ing := range ingressList.Items {
		host := ""
		if len(ing.Spec.Rules) > 0 {
			host = ing.Spec.Rules[0].Host // Lấy host đầu tiên làm đại diện
		}

		ingresses = append(ingresses, map[string]any{
			"name":          ing.Name,
			"namespace":     ing.Namespace,
			"ingress_class": ing.Spec.IngressClassName,
			"host":          host,
		})
	}

	return ingresses, nil
}

// GetIngress lấy thông tin chi tiết của một Ingress cụ thể.
func (uc *K8sUseCase) GetIngress(ctx context.Context, clusterID, namespace, ingressName string) (map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	ing, err := client.NetworkingV1().Ingresses(namespace).Get(ctx, ingressName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ingress: %w", err)
	}

	// Chuẩn bị thông tin về các Rules (đường dẫn và backend)
	rulesData := make([]map[string]any, 0, len(ing.Spec.Rules))
	for _, rule := range ing.Spec.Rules {
		pathsData := make([]map[string]any, 0)
		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				pathsData = append(pathsData, map[string]any{
					"path":            path.Path,
					"path_type":       string(*path.PathType),
					"backend_service": path.Backend.Service.Name,
					"backend_port":    path.Backend.Service.Port.Number,
				})
			}
		}

		rulesData = append(rulesData, map[string]any{
			"host":  rule.Host,
			"paths": pathsData,
		})
	}

	defaultBackendService := ""
	var defaultBackendPort int32 = 0
	if ing.Spec.DefaultBackend != nil && ing.Spec.DefaultBackend.Service != nil {
		defaultBackendService = ing.Spec.DefaultBackend.Service.Name
		defaultBackendPort = ing.Spec.DefaultBackend.Service.Port.Number
	}

	//
	return map[string]any{
		"name":                    ing.Name,
		"namespace":               ing.Namespace,
		"ingress_class":           ing.Spec.IngressClassName,
		"default_backend_service": defaultBackendService,
		"default_backend_port":    defaultBackendPort,
		"rules":                   rulesData,
		"labels":                  ing.Labels,
	}, nil
}

// DeleteIngress xóa một Ingress cụ thể trong namespace.
func (uc *K8sUseCase) DeleteIngress(ctx context.Context, clusterID, namespace, ingressName string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	err = client.NetworkingV1().Ingresses(namespace).Delete(ctx, ingressName, metav1.DeleteOptions{})
	if err != nil {
		// Xử lý trường hợp Ingress không tồn tại
		if errors.IsNotFound(err) {
			return nil // Coi là thành công nếu nó đã không tồn tại
		}
		return fmt.Errorf("failed to delete ingress '%s': %w", ingressName, err)
	}

	return nil
}
