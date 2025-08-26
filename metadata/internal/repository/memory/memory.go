package memory

import (
	"context"
	"mmoviecom/metadata/internal/repository"
	"mmoviecom/metadata/pkg/model"
	"sync"
)

// Repository defines a memory movie metadata repository.
type Repository struct {
	sync.RWMutex
	data map[string]*model.Metadata
}

// New creates new memory repository.
func New() *Repository {
	return &Repository{data: map[string]*model.Metadata{}}
}

// Get retrieves movie metadata by movie id.
func (r *Repository) Get(_ context.Context, id string) (*model.Metadata, error) {
	r.RLock()
	defer r.RUnlock()
	m, ok := r.data[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return m, nil
}

// Put adds movie metadata for a given movie id.
func (r *Repository) Put(_ context.Context, _ string, m *model.Metadata) error {
	r.Lock()
	defer r.Unlock()
	r.data[m.ID] = m
	return nil
}
