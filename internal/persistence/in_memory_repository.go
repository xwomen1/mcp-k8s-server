// internal/infrastructure/persistence/in_memory_repository.go
package persistence

import (
	"fmt"
	"time"

	"github.com/your-org/mcp-k8s-server/internal/domain"
)

func (r *InMemoryClusterRepository) FindByID(id string) (*domain.Cluster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cluster, exists := r.clusters[id]
	if !exists {
		return nil, fmt.Errorf("cluster not found: %s", id)
	}

	return cluster, nil
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

	if _, exists := r.clusters[string(cluster.ID)]; !exists {
		return fmt.Errorf("cluster not found: %s", cluster.ID)
	}

	cluster.UpdatedAt = time.Now()
	r.clusters[string(cluster.ID)] = cluster

	return nil
}

func (r *InMemoryClusterRepository) SetActive(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Deactivate all clusters first
	for _, cluster := range r.clusters {
		cluster.IsActive = false
	}

	// Activate specified cluster
	if cluster, exists := r.clusters[id]; exists {
		cluster.IsActive = true
		cluster.UpdatedAt = time.Now()
	} else {
		return fmt.Errorf("cluster not found: %s", id)
	}

	return nil
}

func generateID() string {
	return fmt.Sprintf("cluster-%d", time.Now().UnixNano())
}
