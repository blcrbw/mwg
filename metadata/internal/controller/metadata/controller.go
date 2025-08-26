package metadata

import (
	"context"
	"errors"
	"fmt"
	"mmoviecom/metadata/internal/repository"
	"mmoviecom/metadata/pkg/model"
)

// ErrNotFound is returned when a requested record is not found.
var ErrNotFound = errors.New("not found")

type metadataRepository interface {
	Get(ctx context.Context, id string) (*model.Metadata, error)
	Put(ctx context.Context, id string, metadata *model.Metadata) error
}

// Controller defines a metadata service controller.
type Controller struct {
	repo  metadataRepository
	cache metadataRepository
}

// New creates a metadata service controller.
func New(repo metadataRepository, cache metadataRepository) *Controller {
	return &Controller{repo: repo, cache: cache}
}

func (c *Controller) Get(ctx context.Context, id string) (*model.Metadata, error) {
	cacheRes, err := c.cache.Get(ctx, id)
	if err == nil {
		fmt.Println("Returning metadata from a cache for " + id)
		return cacheRes, nil
	}

	res, err := c.repo.Get(ctx, id)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}

	if err := c.cache.Put(ctx, id, res); err != nil {
		fmt.Println("Error updating cache: " + err.Error())
	}

	return res, err
}
