package usecase

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/your-org/mcp-k8s-server/internal/domain"
	"github.com/your-org/mcp-k8s-server/internal/infrastructure"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PodUseCase struct {
	clusterClient infrastructure.ClusterClient
	logger        infrastructure.Logger
}

func NewPodUseCase(logger infrastructure.Logger) *PodUseCase {
	return &PodUseCase{
		clusterClient: infrastructure.GetClusterManager(logger),
		logger:        logger,
	}
}

func (uc *PodUseCase) GetPodLogs(ctx context.Context, clusterID domain.ClusterID, namespace domain.Namespace, podName domain.PodName, options domain.LogOptions) (domain.PodLogs, error) {
	uc.logger.Info("Getting pod logs", "clusterID", clusterID, "namespace", namespace, "podName", podName)
	return uc.clusterClient.GetPodLogs(ctx, clusterID, namespace, podName, options)
}

func (uc *PodUseCase) GetPodStatus(ctx context.Context, clusterID domain.ClusterID, namespace domain.Namespace, podName domain.PodName) (*domain.PodStatus, error) {
	return uc.clusterClient.GetPodStatus(ctx, clusterID, namespace, podName)
}

func (uc *PodUseCase) ListPods(ctx context.Context, clusterID domain.ClusterID, namespace domain.Namespace) ([]domain.Pod, error) {
	return uc.clusterClient.ListPods(ctx, clusterID, namespace)
}
func (uc *K8sUseCase) PortForward(ctx context.Context, clusterID string, ns, podName string, localPort, remotePort int, action string) (*domain.PortForwardResult, error) {
	// 1. Xử lý lệnh STOP
	if action == "stop" {
		// Giả sử bạn lưu tunnels trong uc.activeTunnels
		// Nếu chưa có Map quản lý, tạm thời trả về thông báo
		return &domain.PortForwardResult{PodName: podName, URL: "Tunnel requested to stop"}, nil
	}

	// 2. Lấy Config
	restConfig, err := uc.clusterManager.GetRESTConfig(domain.ClusterID(clusterID))
	if err != nil {
		return nil, err
	}
	client, _ := uc.clusterManager.GetClusterClient(domain.ClusterID(clusterID))

	req := client.CoreV1().RESTClient().Post().
		Resource("pods").Namespace(ns).Name(podName).SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		return nil, err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{})

	// 3. Sửa Output sang os.Stderr (Cực kỳ quan trọng để tránh lỗi JSON-RPC)
	pf, err := portforward.New(
		dialer,
		[]string{fmt.Sprintf("%d:%d", localPort, remotePort)},
		stopChan,
		readyChan,
		os.Stderr,
		os.Stderr,
	)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := pf.ForwardPorts(); err != nil {

			fmt.Fprintf(os.Stderr, "PortForward error: %v\n", err)
		}
	}()

	select {
	case <-readyChan:
		return &domain.PortForwardResult{
			PodName: podName, LocalPort: localPort, RemotePort: remotePort, URL: fmt.Sprintf("http://localhost:%d", localPort),
		}, nil
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for port forward to be ready")
	}
}
func ptrBool(b bool) *bool { return &b }
