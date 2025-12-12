package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	"github.com/your-org/mcp-k8s-server/internal/infrastructure"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sUseCase struct {
	clusterRepo    domain.ClusterRepository
	clusterManager *infrastructure.ClusterManager
	logger         infrastructure.Logger
}

func NewK8sUseCase(
	clusterRepo domain.ClusterRepository,
	clusterManager *infrastructure.ClusterManager,
	logger infrastructure.Logger,
) *K8sUseCase {
	return &K8sUseCase{
		clusterRepo:    clusterRepo,
		clusterManager: clusterManager,
		logger:         logger,
	}
}

func (uc *K8sUseCase) GetPodLogs(ctx context.Context, clusterID, namespace, podName string, tailLines int64) (string, error) {
	logs, err := uc.clusterManager.GetPodLogs(
		ctx,
		domain.ClusterID(clusterID),
		domain.Namespace(namespace),
		domain.PodName(podName),
		domain.LogOptions{TailLines: &tailLines},
	)

	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}

	return string(logs), nil
}

func (uc *K8sUseCase) ScaleDeployment(ctx context.Context, clusterID, namespace, deploymentName string, replicas int32) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	deployment.Spec.Replicas = &replicas
	_, err = client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) ListPods(ctx context.Context, clusterID, namespace string) ([]domain.Pod, error) {
	pods, err := uc.clusterManager.ListPods(
		ctx,
		domain.ClusterID(clusterID),
		domain.Namespace(namespace),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	return pods, nil
}

func (uc *K8sUseCase) GetDeploymentInfo(ctx context.Context, clusterID, namespace, deploymentName string) (map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// Extract container information
	containers := make([]map[string]any, 0, len(deployment.Spec.Template.Spec.Containers))
	for _, container := range deployment.Spec.Template.Spec.Containers {
		containers = append(containers, map[string]any{
			"name":  container.Name,
			"image": container.Image,
		})
	}

	// Build deployment info
	info := map[string]any{
		"name":               deployment.Name,
		"namespace":          deployment.Namespace,
		"replicas_desired":   *deployment.Spec.Replicas,
		"replicas_ready":     deployment.Status.ReadyReplicas,
		"replicas_available": deployment.Status.AvailableReplicas,
		"replicas_updated":   deployment.Status.UpdatedReplicas,
		"strategy":           deployment.Spec.Strategy.Type,
		"containers":         containers,
	}

	return info, nil
}

func (uc *K8sUseCase) ListNamespaces(ctx context.Context, clusterID string) ([]map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	namespaceList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]map[string]any, 0, len(namespaceList.Items))
	for _, ns := range namespaceList.Items {
		namespaces = append(namespaces, map[string]any{
			"name":   ns.Name,
			"status": string(ns.Status.Phase),
		})
	}

	return namespaces, nil
}

func (uc *K8sUseCase) GetNamespace(ctx context.Context, clusterID, namespace string) (map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	ns, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	info := map[string]any{
		"name":    ns.Name,
		"status":  string(ns.Status.Phase),
		"created": ns.CreationTimestamp.Format("2006-01-02 15:04:05"),
		"labels":  ns.Labels,
	}

	return info, nil
}

func (uc *K8sUseCase) CreateNamespace(ctx context.Context, clusterID, namespace string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	_, err = client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) DeleteNamespace(ctx context.Context, clusterID, namespace string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	err = client.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) ListPersistentVolumes(ctx context.Context, clusterID string) ([]map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	pvList, err := client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list persistent volumes: %w", err)
	}

	pvs := make([]map[string]any, 0, len(pvList.Items))
	for _, pv := range pvList.Items {
		capacity := ""
		if storage, ok := pv.Spec.Capacity[v1.ResourceStorage]; ok {
			capacity = storage.String()
		}

		storageClass := ""
		if pv.Spec.StorageClassName != "" {
			storageClass = pv.Spec.StorageClassName
		}

		pvs = append(pvs, map[string]any{
			"name":           pv.Name,
			"capacity":       capacity,
			"status":         string(pv.Status.Phase),
			"storage_class":  storageClass,
			"reclaim_policy": string(pv.Spec.PersistentVolumeReclaimPolicy),
		})
	}

	return pvs, nil
}

func (uc *K8sUseCase) ListStorageClasses(ctx context.Context, clusterID string) ([]map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	scList, err := client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list storage classes: %w", err)
	}

	storageClasses := make([]map[string]any, 0, len(scList.Items))
	for _, sc := range scList.Items {
		reclaimPolicy := "Delete"
		if sc.ReclaimPolicy != nil {
			reclaimPolicy = string(*sc.ReclaimPolicy)
		}

		storageClasses = append(storageClasses, map[string]any{
			"name":           sc.Name,
			"provisioner":    sc.Provisioner,
			"reclaim_policy": reclaimPolicy,
		})
	}

	return storageClasses, nil
}

func (uc *K8sUseCase) ListConfigMaps(ctx context.Context, clusterID, namespace string) ([]map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	configMapList, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps: %w", err)
	}

	configMaps := make([]map[string]any, 0, len(configMapList.Items))
	for _, cm := range configMapList.Items {
		configMaps = append(configMaps, map[string]any{
			"name":       cm.Name,
			"namespace":  cm.Namespace,
			"data_count": len(cm.Data),
			"created":    cm.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	return configMaps, nil
}

func (uc *K8sUseCase) GetConfigMap(ctx context.Context, clusterID, namespace, configMapName string) (map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	cm, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap: %w", err)
	}

	info := map[string]any{
		"name":      cm.Name,
		"namespace": cm.Namespace,
		"data":      cm.Data,
		"labels":    cm.Labels,
		"created":   cm.CreationTimestamp.Format("2006-01-02 15:04:05"),
	}

	return info, nil
}

func (uc *K8sUseCase) CreateConfigMap(ctx context.Context, clusterID string, options domain.ConfigMapCreateOptions) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.Name,
			Namespace: options.Namespace,
			Labels:    options.Labels,
		},
		Data: options.Data,
	}

	_, err = client.CoreV1().ConfigMaps(options.Namespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create configmap: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) DeleteConfigMap(ctx context.Context, clusterID, namespace, configMapName string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	err = client.CoreV1().ConfigMaps(namespace).Delete(ctx, configMapName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete configmap: %w", err)
	}

	return nil
}

// ==================== Secret Methods ====================

func (uc *K8sUseCase) ListSecrets(ctx context.Context, clusterID, namespace string) ([]map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	secretList, err := client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	secrets := make([]map[string]any, 0, len(secretList.Items))
	for _, secret := range secretList.Items {
		secrets = append(secrets, map[string]any{
			"name":       secret.Name,
			"namespace":  secret.Namespace,
			"type":       string(secret.Type),
			"data_count": len(secret.Data),
			"created":    secret.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	return secrets, nil
}

func (uc *K8sUseCase) GetSecret(ctx context.Context, clusterID, namespace, secretName string) (map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Convert secret data to base64 strings for display (keys only, not actual values for security)
	dataKeys := make([]string, 0, len(secret.Data))
	for key := range secret.Data {
		dataKeys = append(dataKeys, key)
	}

	info := map[string]any{
		"name":      secret.Name,
		"namespace": secret.Namespace,
		"type":      string(secret.Type),
		"data_keys": dataKeys,
		"labels":    secret.Labels,
		"created":   secret.CreationTimestamp.Format("2006-01-02 15:04:05"),
	}

	return info, nil
}

func (uc *K8sUseCase) CreateSecret(ctx context.Context, clusterID string, options domain.SecretCreateOptions) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	secretType := v1.SecretTypeOpaque
	if options.Type != "" {
		secretType = v1.SecretType(options.Type)
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.Name,
			Namespace: options.Namespace,
			Labels:    options.Labels,
		},
		Type:       secretType,
		StringData: options.StringData,
	}

	_, err = client.CoreV1().Secrets(options.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) DeleteSecret(ctx context.Context, clusterID, namespace, secretName string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	err = client.CoreV1().Secrets(namespace).Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) ListStatefulSets(ctx context.Context, clusterID, namespace string) ([]map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	statefulSetList, err := client.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulsets: %w", err)
	}

	statefulSets := make([]map[string]any, 0, len(statefulSetList.Items))
	for _, sts := range statefulSetList.Items {
		replicas := int32(0)
		if sts.Spec.Replicas != nil {
			replicas = *sts.Spec.Replicas
		}

		statefulSets = append(statefulSets, map[string]any{
			"name":             sts.Name,
			"namespace":        sts.Namespace,
			"replicas":         replicas,
			"ready_replicas":   sts.Status.ReadyReplicas,
			"current_replicas": sts.Status.CurrentReplicas,
			"service_name":     sts.Spec.ServiceName,
			"created":          sts.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	return statefulSets, nil
}

func (uc *K8sUseCase) GetStatefulSet(ctx context.Context, clusterID, namespace, statefulSetName string) (map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, statefulSetName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get statefulset: %w", err)
	}

	replicas := int32(0)
	if sts.Spec.Replicas != nil {
		replicas = *sts.Spec.Replicas
	}

	// Extract container information
	containers := make([]map[string]any, 0, len(sts.Spec.Template.Spec.Containers))
	for _, container := range sts.Spec.Template.Spec.Containers {
		containers = append(containers, map[string]any{
			"name":  container.Name,
			"image": container.Image,
		})
	}

	// Extract volume claim templates
	volumeClaims := make([]map[string]any, 0, len(sts.Spec.VolumeClaimTemplates))
	for _, pvc := range sts.Spec.VolumeClaimTemplates {
		storage := ""
		if storageQty, ok := pvc.Spec.Resources.Requests[v1.ResourceStorage]; ok {
			storage = storageQty.String()
		}

		volumeClaims = append(volumeClaims, map[string]any{
			"name":    pvc.Name,
			"storage": storage,
			"storage_class": func() string {
				if pvc.Spec.StorageClassName != nil {
					return *pvc.Spec.StorageClassName
				}
				return ""
			}(),
		})
	}

	info := map[string]any{
		"name":             sts.Name,
		"namespace":        sts.Namespace,
		"replicas_desired": replicas,
		"replicas_ready":   sts.Status.ReadyReplicas,
		"replicas_current": sts.Status.CurrentReplicas,
		"replicas_updated": sts.Status.UpdatedReplicas,
		"service_name":     sts.Spec.ServiceName,
		"update_strategy":  string(sts.Spec.UpdateStrategy.Type),
		"containers":       containers,
		"volume_claims":    volumeClaims,
		"labels":           sts.Labels,
		"created":          sts.CreationTimestamp.Format("2006-01-02 15:04:05"),
	}

	return info, nil
}

func (uc *K8sUseCase) ScaleStatefulSet(ctx context.Context, clusterID, namespace, statefulSetName string, replicas int32) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, statefulSetName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get statefulset: %w", err)
	}

	sts.Spec.Replicas = &replicas
	_, err = client.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale statefulset: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) DeleteStatefulSet(ctx context.Context, clusterID, namespace, statefulSetName string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	err = client.AppsV1().StatefulSets(namespace).Delete(ctx, statefulSetName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete statefulset: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) RestartStatefulSet(ctx context.Context, clusterID, namespace, statefulSetName string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, statefulSetName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get statefulset: %w", err)
	}

	// Add or update restart annotation
	if sts.Spec.Template.Annotations == nil {
		sts.Spec.Template.Annotations = make(map[string]string)
	}
	sts.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = client.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to restart statefulset: %w", err)
	}

	return nil
}

// ==================== DaemonSet Methods ====================

func (uc *K8sUseCase) ListDaemonSets(ctx context.Context, clusterID, namespace string) ([]map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	daemonSetList, err := client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list daemonsets: %w", err)
	}

	daemonSets := make([]map[string]any, 0, len(daemonSetList.Items))
	for _, ds := range daemonSetList.Items {
		daemonSets = append(daemonSets, map[string]any{
			"name":                     ds.Name,
			"namespace":                ds.Namespace,
			"desired_number_scheduled": ds.Status.DesiredNumberScheduled,
			"current_number_scheduled": ds.Status.CurrentNumberScheduled,
			"number_ready":             ds.Status.NumberReady,
			"number_available":         ds.Status.NumberAvailable,
			"created":                  ds.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	return daemonSets, nil
}

func (uc *K8sUseCase) GetDaemonSet(ctx context.Context, clusterID, namespace, daemonSetName string) (map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, daemonSetName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get daemonset: %w", err)
	}

	// Extract container information
	containers := make([]map[string]any, 0, len(ds.Spec.Template.Spec.Containers))
	for _, container := range ds.Spec.Template.Spec.Containers {
		containers = append(containers, map[string]any{
			"name":  container.Name,
			"image": container.Image,
		})
	}

	// Extract node selector
	nodeSelector := ds.Spec.Template.Spec.NodeSelector

	info := map[string]any{
		"name":                     ds.Name,
		"namespace":                ds.Namespace,
		"desired_number_scheduled": ds.Status.DesiredNumberScheduled,
		"current_number_scheduled": ds.Status.CurrentNumberScheduled,
		"number_ready":             ds.Status.NumberReady,
		"number_available":         ds.Status.NumberAvailable,
		"number_misscheduled":      ds.Status.NumberMisscheduled,
		"updated_number_scheduled": ds.Status.UpdatedNumberScheduled,
		"update_strategy":          string(ds.Spec.UpdateStrategy.Type),
		"containers":               containers,
		"node_selector":            nodeSelector,
		"labels":                   ds.Labels,
		"created":                  ds.CreationTimestamp.Format("2006-01-02 15:04:05"),
	}

	return info, nil
}

func (uc *K8sUseCase) DeleteDaemonSet(ctx context.Context, clusterID, namespace, daemonSetName string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	err = client.AppsV1().DaemonSets(namespace).Delete(ctx, daemonSetName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete daemonset: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) RestartDaemonSet(ctx context.Context, clusterID, namespace, daemonSetName string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, daemonSetName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get daemonset: %w", err)
	}

	// Add or update restart annotation
	if ds.Spec.Template.Annotations == nil {
		ds.Spec.Template.Annotations = make(map[string]string)
	}
	ds.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = client.AppsV1().DaemonSets(namespace).Update(ctx, ds, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to restart daemonset: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) GetDaemonSetPods(ctx context.Context, clusterID, namespace, daemonSetName string) ([]map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Get DaemonSet to find label selector
	ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, daemonSetName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get daemonset: %w", err)
	}

	// List pods with matching labels
	labelSelector := metav1.FormatLabelSelector(ds.Spec.Selector)
	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	pods := make([]map[string]any, 0, len(podList.Items))
	for _, pod := range podList.Items {
		pods = append(pods, map[string]any{
			"name":      pod.Name,
			"namespace": pod.Namespace,
			"node":      pod.Spec.NodeName,
			"status":    string(pod.Status.Phase),
			"host_ip":   pod.Status.HostIP,
			"pod_ip":    pod.Status.PodIP,
		})
	}

	return pods, nil
}

func (uc *K8sUseCase) ListJobs(ctx context.Context, clusterID, namespace string) ([]map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	jobList, err := client.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	jobs := make([]map[string]any, 0, len(jobList.Items))
	for _, job := range jobList.Items {
		completions := int32(1)
		if job.Spec.Completions != nil {
			completions = *job.Spec.Completions
		}

		status := "Active"
		if job.Status.Succeeded > 0 {
			status = "Succeeded"
		} else if job.Status.Failed > 0 {
			status = "Failed"
		}

		duration := ""
		if job.Status.StartTime != nil {
			if job.Status.CompletionTime != nil {
				duration = job.Status.CompletionTime.Sub(job.Status.StartTime.Time).String()
			} else {
				duration = time.Since(job.Status.StartTime.Time).String()
			}
		}

		jobs = append(jobs, map[string]any{
			"name":        job.Name,
			"namespace":   job.Namespace,
			"completions": fmt.Sprintf("%d/%d", job.Status.Succeeded, completions),
			"status":      status,
			"active":      job.Status.Active,
			"succeeded":   job.Status.Succeeded,
			"failed":      job.Status.Failed,
			"duration":    duration,
			"created":     job.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	return jobs, nil
}

func (uc *K8sUseCase) GetJob(ctx context.Context, clusterID, namespace, jobName string) (map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	job, err := client.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	completions := int32(1)
	if job.Spec.Completions != nil {
		completions = *job.Spec.Completions
	}

	parallelism := int32(1)
	if job.Spec.Parallelism != nil {
		parallelism = *job.Spec.Parallelism
	}

	backoffLimit := int32(6)
	if job.Spec.BackoffLimit != nil {
		backoffLimit = *job.Spec.BackoffLimit
	}

	// Extract container information
	containers := make([]map[string]any, 0, len(job.Spec.Template.Spec.Containers))
	for _, container := range job.Spec.Template.Spec.Containers {
		containers = append(containers, map[string]any{
			"name":    container.Name,
			"image":   container.Image,
			"command": container.Command,
			"args":    container.Args,
		})
	}

	// Calculate duration
	duration := ""
	if job.Status.StartTime != nil {
		if job.Status.CompletionTime != nil {
			duration = job.Status.CompletionTime.Sub(job.Status.StartTime.Time).String()
		} else {
			duration = time.Since(job.Status.StartTime.Time).String()
		}
	}

	// Job conditions
	conditions := make([]map[string]any, 0, len(job.Status.Conditions))
	for _, condition := range job.Status.Conditions {
		conditions = append(conditions, map[string]any{
			"type":    string(condition.Type),
			"status":  string(condition.Status),
			"reason":  condition.Reason,
			"message": condition.Message,
		})
	}

	info := map[string]any{
		"name":          job.Name,
		"namespace":     job.Namespace,
		"completions":   completions,
		"parallelism":   parallelism,
		"backoff_limit": backoffLimit,
		"active":        job.Status.Active,
		"succeeded":     job.Status.Succeeded,
		"failed":        job.Status.Failed,
		"start_time": func() string {
			if job.Status.StartTime != nil {
				return job.Status.StartTime.Format(time.RFC3339)
			}
			return ""
		}(),
		"completion_time": func() string {
			if job.Status.CompletionTime != nil {
				return job.Status.CompletionTime.Format(time.RFC3339)
			}
			return ""
		}(),
		"duration":   duration,
		"containers": containers,
		"conditions": conditions,
		"labels":     job.Labels,
		"created":    job.CreationTimestamp.Format("2006-01-02 15:04:05"),
	}

	return info, nil
}

func (uc *K8sUseCase) CreateJob(ctx context.Context, clusterID string, options domain.JobCreateOptions) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	restartPolicy := v1.RestartPolicyOnFailure
	if options.RestartPolicy != "" {
		restartPolicy = v1.RestartPolicy(options.RestartPolicy)
	}

	// Build environment variables
	envVars := make([]v1.EnvVar, 0, len(options.Env))
	for k, v := range options.Env {
		envVars = append(envVars, v1.EnvVar{Name: k, Value: v})
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.Name,
			Namespace: options.Namespace,
			Labels:    options.Labels,
		},
		Spec: batchv1.JobSpec{
			Completions:  options.Completions,
			Parallelism:  options.Parallelism,
			BackoffLimit: options.BackoffLimit,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					RestartPolicy: restartPolicy,
					Containers: []v1.Container{
						{
							Name:    options.Name,
							Image:   options.Image,
							Command: options.Command,
							Args:    options.Args,
							Env:     envVars,
						},
					},
				},
			},
		},
	}

	_, err = client.BatchV1().Jobs(options.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) DeleteJob(ctx context.Context, clusterID, namespace, jobName string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	propagationPolicy := metav1.DeletePropagationBackground
	err = client.BatchV1().Jobs(namespace).Delete(ctx, jobName, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) GetJobLogs(ctx context.Context, clusterID, namespace, jobName string, tailLines int64) (string, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return "", fmt.Errorf("failed to get client: %w", err)
	}

	// Get job to find pod selector
	_, err = client.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get job: %w", err)
	}

	// List pods for this job
	labelSelector := metav1.FormatLabelSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{"job-name": jobName},
	})

	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list pods: %w", err)
	}

	if len(podList.Items) == 0 {
		return "No pods found for this job", nil
	}

	// Get logs from the first pod
	pod := podList.Items[0]
	req := client.CoreV1().Pods(namespace).GetLogs(pod.Name, &v1.PodLogOptions{
		TailLines: &tailLines,
	})

	logs, err := req.Do(ctx).Raw()
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}

	return string(logs), nil
}

// ==================== CronJob Methods ====================

func (uc *K8sUseCase) ListCronJobs(ctx context.Context, clusterID, namespace string) ([]map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	cronJobList, err := client.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list cronjobs: %w", err)
	}

	cronJobs := make([]map[string]any, 0, len(cronJobList.Items))
	for _, cj := range cronJobList.Items {
		lastSchedule := ""
		if cj.Status.LastScheduleTime != nil {
			lastSchedule = cj.Status.LastScheduleTime.Format("2006-01-02 15:04:05")
		}

		status := "Active"
		if *cj.Spec.Suspend {
			status = "Suspended"
		}

		cronJobs = append(cronJobs, map[string]any{
			"name":          cj.Name,
			"namespace":     cj.Namespace,
			"schedule":      cj.Spec.Schedule,
			"suspend":       *cj.Spec.Suspend,
			"active":        len(cj.Status.Active),
			"last_schedule": lastSchedule,
			"status":        status,
			"created":       cj.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	return cronJobs, nil
}

func (uc *K8sUseCase) GetCronJob(ctx context.Context, clusterID, namespace, cronJobName string) (map[string]any, error) {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	cj, err := client.BatchV1().CronJobs(namespace).Get(ctx, cronJobName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get cronjob: %w", err)
	}

	// Extract container information
	containers := make([]map[string]any, 0, len(cj.Spec.JobTemplate.Spec.Template.Spec.Containers))
	for _, container := range cj.Spec.JobTemplate.Spec.Template.Spec.Containers {
		containers = append(containers, map[string]any{
			"name":    container.Name,
			"image":   container.Image,
			"command": container.Command,
			"args":    container.Args,
		})
	}

	// Active jobs
	activeJobs := make([]string, 0, len(cj.Status.Active))
	for _, activeJob := range cj.Status.Active {
		activeJobs = append(activeJobs, activeJob.Name)
	}

	info := map[string]any{
		"name":               cj.Name,
		"namespace":          cj.Namespace,
		"schedule":           cj.Spec.Schedule,
		"suspend":            *cj.Spec.Suspend,
		"concurrency_policy": string(cj.Spec.ConcurrencyPolicy),
		"successful_jobs_history_limit": func() int32 {
			if cj.Spec.SuccessfulJobsHistoryLimit != nil {
				return *cj.Spec.SuccessfulJobsHistoryLimit
			}
			return 3
		}(),
		"failed_jobs_history_limit": func() int32 {
			if cj.Spec.FailedJobsHistoryLimit != nil {
				return *cj.Spec.FailedJobsHistoryLimit
			}
			return 1
		}(),
		"last_schedule_time": func() string {
			if cj.Status.LastScheduleTime != nil {
				return cj.Status.LastScheduleTime.Format(time.RFC3339)
			}
			return ""
		}(),
		"last_successful_time": func() string {
			if cj.Status.LastSuccessfulTime != nil {
				return cj.Status.LastSuccessfulTime.Format(time.RFC3339)
			}
			return ""
		}(),
		"active_jobs":  activeJobs,
		"active_count": len(activeJobs),
		"containers":   containers,
		"labels":       cj.Labels,
		"created":      cj.CreationTimestamp.Format("2006-01-02 15:04:05"),
	}

	return info, nil
}

func (uc *K8sUseCase) CreateCronJob(ctx context.Context, clusterID string, options domain.CronJobCreateOptions) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	restartPolicy := v1.RestartPolicyOnFailure
	if options.RestartPolicy != "" {
		restartPolicy = v1.RestartPolicy(options.RestartPolicy)
	}

	concurrencyPolicy := batchv1.AllowConcurrent
	if options.ConcurrencyPolicy != "" {
		concurrencyPolicy = batchv1.ConcurrencyPolicy(options.ConcurrencyPolicy)
	}

	// Build environment variables
	envVars := make([]v1.EnvVar, 0, len(options.Env))
	for k, v := range options.Env {
		envVars = append(envVars, v1.EnvVar{Name: k, Value: v})
	}

	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.Name,
			Namespace: options.Namespace,
			Labels:    options.Labels,
		},
		Spec: batchv1.CronJobSpec{
			Schedule:                   options.Schedule,
			Suspend:                    &options.Suspend,
			ConcurrencyPolicy:          concurrencyPolicy,
			SuccessfulJobsHistoryLimit: options.SuccessfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     options.FailedJobsHistoryLimit,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							RestartPolicy: restartPolicy,
							Containers: []v1.Container{
								{
									Name:    options.Name,
									Image:   options.Image,
									Command: options.Command,
									Args:    options.Args,
									Env:     envVars,
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = client.BatchV1().CronJobs(options.Namespace).Create(ctx, cronJob, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create cronjob: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) DeleteCronJob(ctx context.Context, clusterID, namespace, cronJobName string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	propagationPolicy := metav1.DeletePropagationBackground
	err = client.BatchV1().CronJobs(namespace).Delete(ctx, cronJobName, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to delete cronjob: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) SuspendCronJob(ctx context.Context, clusterID, namespace, cronJobName string, suspend bool) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	cj, err := client.BatchV1().CronJobs(namespace).Get(ctx, cronJobName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get cronjob: %w", err)
	}

	cj.Spec.Suspend = &suspend
	_, err = client.BatchV1().CronJobs(namespace).Update(ctx, cj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update cronjob: %w", err)
	}

	return nil
}

func (uc *K8sUseCase) TriggerCronJob(ctx context.Context, clusterID, namespace, cronJobName string) error {
	client, err := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	// Get CronJob
	cj, err := client.BatchV1().CronJobs(namespace).Get(ctx, cronJobName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get cronjob: %w", err)
	}

	// Create a Job from the CronJob template
	annotations := make(map[string]string)
	annotations["cronjob.kubernetes.io/instantiate"] = "manual"

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-manual-%d", cronJobName, time.Now().Unix()),
			Namespace:   namespace,
			Annotations: annotations,
			Labels:      cj.Spec.JobTemplate.Labels,
		},
		Spec: cj.Spec.JobTemplate.Spec,
	}

	_, err = client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to trigger cronjob: %w", err)
	}

	return nil
}
