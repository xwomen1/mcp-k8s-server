// internal/infrastructure/persistence/cluster_repository.go
package persistence

import (
	"fmt"
	"sync"

	"github.com/your-org/mcp-k8s-server/internal/domain"
)

type ClusterRepository interface {
	Save(cluster *domain.Cluster) error
	Get(id string) (*domain.Cluster, error)
	GetByName(name string) (*domain.Cluster, error)
	List() ([]*domain.Cluster, error)
	Delete(id string) error
}

type InMemoryClusterRepository struct {
	clusters map[string]*domain.Cluster
	mu       sync.RWMutex
}

func NewInMemoryClusterRepository() *InMemoryClusterRepository {
	return &InMemoryClusterRepository{
		clusters: make(map[string]*domain.Cluster),
	}
}

func (r *InMemoryClusterRepository) Save(cluster *domain.Cluster) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clusters[string(cluster.ID)] = cluster
	return nil
}

func (r *InMemoryClusterRepository) Get(id string) (*domain.Cluster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cluster, exists := r.clusters[id]
	if !exists {
		return nil, fmt.Errorf("cluster not found: %s", id)
	}
	return cluster, nil
}

func (r *InMemoryClusterRepository) GetByName(name string) (*domain.Cluster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, cluster := range r.clusters {
		if cluster.Name == name {
			return cluster, nil
		}
	}
	return nil, fmt.Errorf("cluster not found with name: %s", name)
}

func (r *InMemoryClusterRepository) List() ([]*domain.Cluster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	clusters := make([]*domain.Cluster, 0, len(r.clusters))
	for _, cluster := range r.clusters {
		clusters = append(clusters, cluster)
	}
	return clusters, nil
}

func (r *InMemoryClusterRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.clusters[id]; !exists {
		return fmt.Errorf("cluster not found: %s", id)
	}
	delete(r.clusters, id)
	return nil
}
