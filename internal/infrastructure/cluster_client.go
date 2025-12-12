package infrastructure

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ClusterClient interface {
	RegisterCluster(ctx context.Context, clusterID domain.ClusterID, config domain.ClusterConfig) error
	GetClusterStatus(ctx context.Context, clusterID domain.ClusterID) (*domain.ClusterStatus, error)
	ListClusters(ctx context.Context) ([]domain.Cluster, error)
	DeleteCluster(ctx context.Context, clusterID domain.ClusterID) error
	GetPodLogs(ctx context.Context, clusterID domain.ClusterID, namespace domain.Namespace, podName domain.PodName, options domain.LogOptions) (domain.PodLogs, error)
	GetPodStatus(ctx context.Context, clusterID domain.ClusterID, namespace domain.Namespace, podName domain.PodName) (*domain.PodStatus, error)
	ListPods(ctx context.Context, clusterID domain.ClusterID, namespace domain.Namespace) ([]domain.Pod, error)
	GetClusterClient(clusterID domain.ClusterID) (kubernetes.Interface, error)
}

type ClusterManager struct {
	clusters map[domain.ClusterID]*ClusterContext
	mu       sync.RWMutex
	logger   Logger
}

type ClusterContext struct {
	Config     domain.ClusterConfig
	ClientSet  kubernetes.Interface
	RestConfig *rest.Config
	LastUsed   time.Time
}

var (
	clusterManager *ClusterManager
	once           sync.Once
)

func GetClusterManager(logger Logger) *ClusterManager {
	once.Do(func() {
		clusterManager = NewClusterClient(logger).(*ClusterManager)
	})
	return clusterManager
}

func NewClusterClient(logger Logger) ClusterClient {
	return &ClusterManager{
		clusters: make(map[domain.ClusterID]*ClusterContext),
		logger:   logger,
	}
}
func (cm *ClusterManager) GetClusterClient(clusterID domain.ClusterID) (kubernetes.Interface, error) {
	cm.mu.RLock()
	clusterCtx, exists := cm.clusters[clusterID]
	cm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("cluster not found: %s", clusterID)
	}

	clusterCtx.LastUsed = time.Now()
	return clusterCtx.ClientSet, nil
}

func (cm *ClusterManager) RegisterCluster(ctx context.Context, clusterID domain.ClusterID, config domain.ClusterConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.logger.Info("Registering cluster", "clusterID", clusterID)

	// Load kubeconfig
	restConfig, err := LoadKubeconfig(config)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	cm.clusters[clusterID] = &ClusterContext{
		Config:     config,
		ClientSet:  clientset,
		RestConfig: restConfig,
	}

	return nil
}

func (cm *ClusterManager) GetClusterStatus(ctx context.Context, clusterID domain.ClusterID) (*domain.ClusterStatus, error) {
	cm.mu.RLock()
	clusterCtx, exists := cm.clusters[clusterID]
	cm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("cluster not found: %s", clusterID)
	}

	// Test connection
	_, err := clusterCtx.ClientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cluster: %w", err)
	}

	status := domain.ClusterStatusActive
	return &status, nil
}

func (cm *ClusterManager) ListClusters(ctx context.Context) ([]domain.Cluster, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var clusters []domain.Cluster
	for clusterID, clusterCtx := range cm.clusters {
		// For each cluster, check status
		status, err := cm.GetClusterStatus(ctx, clusterID)
		if err != nil {
			// If there's an error, set status to error
			errorStatus := domain.ClusterStatusError
			status = &errorStatus
		}
		clusters = append(clusters, domain.Cluster{
			ID:     clusterID,
			Config: clusterCtx.Config,
			Status: *status,
		})
	}
	return clusters, nil
}

func (cm *ClusterManager) DeleteCluster(ctx context.Context, clusterID domain.ClusterID) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.clusters, clusterID)
	return nil
}

func (cm *ClusterManager) GetPodLogs(ctx context.Context, clusterID domain.ClusterID, namespace domain.Namespace, podName domain.PodName, options domain.LogOptions) (domain.PodLogs, error) {
	cm.mu.RLock()
	clusterCtx, exists := cm.clusters[clusterID]
	cm.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("cluster not found: %s", clusterID)
	}

	req := clusterCtx.ClientSet.CoreV1().Pods(string(namespace)).GetLogs(string(podName), &v1.PodLogOptions{
		TailLines: options.TailLines,
	})

	podLogs, err := req.Do(ctx).Raw()
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}

	return domain.PodLogs(podLogs), nil
}

func (cm *ClusterManager) GetPodStatus(ctx context.Context, clusterID domain.ClusterID, namespace domain.Namespace, podName domain.PodName) (*domain.PodStatus, error) {
	cm.mu.RLock()
	clusterCtx, exists := cm.clusters[clusterID]
	cm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("cluster not found: %s", clusterID)
	}

	pod, err := clusterCtx.ClientSet.CoreV1().Pods(string(namespace)).Get(ctx, string(podName), metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	// Convert to domain.PodStatus
	var startTime string
	if pod.Status.StartTime != nil {
		startTime = pod.Status.StartTime.Format(time.RFC3339)
	}

	status := &domain.PodStatus{
		Phase:     domain.PodPhase(pod.Status.Phase),
		Message:   pod.Status.Message,
		Reason:    pod.Status.Reason,
		StartTime: startTime,
	}

	return status, nil
}

func (cm *ClusterManager) ListPods(ctx context.Context, clusterID domain.ClusterID, namespace domain.Namespace) ([]domain.Pod, error) {
	client, err := cm.GetClusterClient(clusterID)
	if err != nil {
		return nil, err
	}

	podList, err := client.CoreV1().Pods(string(namespace)).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var pods []domain.Pod
	for _, p := range podList.Items {
		pods = append(pods, domain.Pod{
			Name:      domain.PodName(p.Name),
			Namespace: domain.Namespace(p.Namespace),
			ClusterID: clusterID,
			Status: domain.PodStatus{
				Phase:   domain.PodPhase(p.Status.Phase),
				Message: p.Status.Message,
				Reason:  p.Status.Reason,
				StartTime: func() string {
					if p.Status.StartTime != nil {
						return p.Status.StartTime.Format(time.RFC3339)
					}
					return ""
				}(),
			},
			Labels: p.Labels,
		})
	}

	return pods, nil
}

func (cm *ClusterManager) CloseAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for clusterID := range cm.clusters {
		delete(cm.clusters, clusterID)
		cm.logger.Info("Closed cluster connection", "clusterID", clusterID)
	}
}
