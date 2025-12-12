package infrastructure

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/your-org/mcp-k8s-server/internal/domain"
)

type Persistence interface {
	SaveCluster(ctx context.Context, cluster domain.Cluster) error
	LoadCluster(ctx context.Context, clusterID domain.ClusterID) (*domain.Cluster, error)
	LoadAllClusters(ctx context.Context) ([]domain.Cluster, error)
	DeleteCluster(ctx context.Context, clusterID domain.ClusterID) error
}

type FilePersistence struct {
	dataDir string
	logger  Logger
}

func NewFilePersistence(dataDir string, logger Logger) Persistence {
	if dataDir == "" {
		dataDir = "./data"
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		logger.Error("Failed to create data directory", "error", err)
	}
	return &FilePersistence{
		dataDir: dataDir,
		logger:  logger,
	}
}

func (fp *FilePersistence) SaveCluster(ctx context.Context, cluster domain.Cluster) error {
	filePath := filepath.Join(fp.dataDir, string(cluster.ID)+".json")
	data, err := json.MarshalIndent(cluster, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func (fp *FilePersistence) LoadCluster(ctx context.Context, clusterID domain.ClusterID) (*domain.Cluster, error) {
	filePath := filepath.Join(fp.dataDir, string(clusterID)+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cluster domain.Cluster
	if err := json.Unmarshal(data, &cluster); err != nil {
		return nil, err
	}

	return &cluster, nil
}

func (fp *FilePersistence) LoadAllClusters(ctx context.Context) ([]domain.Cluster, error) {
	files, err := os.ReadDir(fp.dataDir)
	if err != nil {
		return nil, err
	}

	var clusters []domain.Cluster
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		clusterID := domain.ClusterID(file.Name()[:len(file.Name())-5]) // Remove .json extension
		cluster, err := fp.LoadCluster(ctx, clusterID)
		if err != nil {
			fp.logger.Error("Failed to load cluster", "clusterID", clusterID, "error", err)
			continue
		}
		clusters = append(clusters, *cluster)
	}

	return clusters, nil
}

func (fp *FilePersistence) DeleteCluster(ctx context.Context, clusterID domain.ClusterID) error {
	filePath := filepath.Join(fp.dataDir, string(clusterID)+".json")
	return os.Remove(filePath)
}
