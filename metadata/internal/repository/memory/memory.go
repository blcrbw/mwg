package memory

import (
	"context"
	"mmoviecom/metadata/internal/repository"
	"mmoviecom/metadata/pkg/model"
	"mmoviecom/pkg/logging"
	"sync"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

const tracerID = "metadata-repository-memory"

// Repository defines a memory movie metadata repository.
type Repository struct {
	sync.RWMutex
	data   map[string]*model.Metadata
	logger *zap.Logger
}

// New creates new memory repository.
func New(logger *zap.Logger) *Repository {
	logger = logger.With(
		zap.String(logging.FieldComponent, "repository"),
		zap.String(logging.FieldType, "memory"),
	)
	return &Repository{data: map[string]*model.Metadata{}, logger: logger}
}

// Get retrieves movie metadata by movie id.
func (r *Repository) Get(ctx context.Context, id string) (*model.Metadata, error) {
	_, span := otel.Tracer(tracerID).Start(ctx, "Repository/Get")
	defer span.End()
	r.RLock()
	defer r.RUnlock()
	m, ok := r.data[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return m, nil
}

// Put adds movie metadata for a given movie id.
func (r *Repository) Put(ctx context.Context, _ string, m *model.Metadata) error {
	_, span := otel.Tracer(tracerID).Start(ctx, "Repository/Put")
	defer span.End()
	r.Lock()
	defer r.Unlock()
	r.data[m.ID] = m
	return nil
}
