package memory

import (
	"context"
	"mmoviecom/pkg/logging"
	"mmoviecom/rating/internal/repository"
	"mmoviecom/rating/pkg/model"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

const tracerID = "rating-repository-memory"

// Repository defines a rating repository.
type Repository struct {
	data   map[model.RecordType]map[model.RecordId][]model.Rating
	logger *zap.Logger
}

// New creates a new memory repository.
func New(logger *zap.Logger) *Repository {
	logger = logger.With(
		zap.String(logging.FieldComponent, "repository"),
		zap.String(logging.FieldType, "memory"),
	)
	return &Repository{data: map[model.RecordType]map[model.RecordId][]model.Rating{}, logger: logger}
}

// Get retrieves all ratings for a given record.
func (r *Repository) Get(ctx context.Context, recordId model.RecordId, recordType model.RecordType) ([]model.Rating, error) {
	_, span := otel.Tracer(tracerID).Start(ctx, "Repository/Get")
	defer span.End()
	if _, ok := r.data[recordType]; !ok {
		return nil, repository.ErrNotFound
	}
	if ratings, ok := r.data[recordType][recordId]; !ok || len(ratings) == 0 {
		return nil, repository.ErrNotFound
	}
	return r.data[recordType][recordId], nil
}

// Put adds a rating for a given record.
func (r *Repository) Put(ctx context.Context, recordId model.RecordId, recordType model.RecordType, rating *model.Rating) error {
	_, span := otel.Tracer(tracerID).Start(ctx, "Repository/Put")
	defer span.End()
	if _, ok := r.data[recordType]; !ok {
		r.data[recordType] = map[model.RecordId][]model.Rating{}
	}
	r.data[recordType][recordId] = append(r.data[recordType][recordId], *rating)
	return nil
}
