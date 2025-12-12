// internal/infrastructure/repository.go
package infrastructure

import (
	"fmt"
	"sync"
	"time"

	"github.com/your-org/mcp-k8s-server/internal/domain"
)

type InMemoryClusterRepository struct {
	clusters map[domain.ClusterID]*domain.Cluster
	mu       sync.RWMutex
}

func NewInMemoryClusterRepository() *InMemoryClusterRepository {
	return &InMemoryClusterRepository{
		clusters: make(map[domain.ClusterID]*domain.Cluster),
	}
}

func (r *InMemoryClusterRepository) Save(cluster *domain.Cluster) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clusters[cluster.ID]; exists {
		return fmt.Errorf("cluster already exists: %s", cluster.ID)
	}

	now := time.Now()
	cluster.CreatedAt = now
	cluster.UpdatedAt = now
	r.clusters[cluster.ID] = cluster

	return nil
}
func (r *InMemoryClusterRepository) FindByName(name string) (*domain.Cluster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, cluster := range r.clusters {
		if cluster.Name == name {
			return cluster, nil
		}
	}

	return nil, fmt.Errorf("cluster not found with name: %s", name)
}

func (r *InMemoryClusterRepository) FindByID(id domain.ClusterID) (*domain.Cluster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cluster, exists := r.clusters[id]
	if !exists {
		return nil, fmt.Errorf("cluster not found: %s", id)
	}

	return cluster, nil
}

func (r *InMemoryClusterRepository) Get(clusterID string) (*domain.Cluster, error) {
	return r.FindByID(domain.ClusterID(clusterID))
}

func (r *InMemoryClusterRepository) FindAll() ([]*domain.Cluster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clusters := make([]*domain.Cluster, 0, len(r.clusters))
	for _, cluster := range r.clusters {
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

func (r *InMemoryClusterRepository) Update(cluster *domain.Cluster) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clusters[cluster.ID]; !exists {
		return fmt.Errorf("cluster not found: %s", cluster.ID)
	}

	cluster.UpdatedAt = time.Now()
	r.clusters[cluster.ID] = cluster

	return nil
}

func (r *InMemoryClusterRepository) Delete(id domain.ClusterID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clusters[id]; !exists {
		return fmt.Errorf("cluster not found: %s", id)
	}

	delete(r.clusters, id)
	return nil
}
