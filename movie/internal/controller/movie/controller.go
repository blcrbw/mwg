package movie

import (
	"context"
	"errors"
	metadatamodel "mmoviecom/metadata/pkg/model"
	"mmoviecom/movie/internal/gateway"
	"mmoviecom/movie/pkg/model"
	ratingmodel "mmoviecom/rating/pkg/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrNotFound is returned when the movie metadata not found.
var ErrNotFound = errors.New("movie metadata not found")

type ratingGateway interface {
	GetAggregatedRating(ctx context.Context, recordId ratingmodel.RecordId, recordType ratingmodel.RecordType) (float64, error)
	PutRating(ctx context.Context, recordId ratingmodel.RecordId, recordType ratingmodel.RecordType, rating *ratingmodel.Rating) error
}

type metadataGateway interface {
	Get(ctx context.Context, id string) (*metadatamodel.Metadata, error)
}

// Controller defines a movie service controller.
type Controller struct {
	ratingGateway   ratingGateway
	metadataGateway metadataGateway
}

// New creates a movie service controller.
func New(gateway ratingGateway, metadataGateway metadataGateway) *Controller {
	return &Controller{gateway, metadataGateway}
}

// Get returns the movie details including the aggregated rating and movie metadata.
func (c *Controller) Get(ctx context.Context, id string) (*model.MovieDetails, error) {
	metadata, err := c.metadataGateway.Get(ctx, id)
	if err != nil && errors.Is(err, gateway.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	details := &model.MovieDetails{Metadata: *metadata}
	rating, err := c.ratingGateway.GetAggregatedRating(ctx, ratingmodel.RecordId(id), ratingmodel.RecordTypeMovie)
	if err != nil && errors.Is(err, errors.New("rating not found for a record")) {
		// ok
	} else if err != nil && errors.Is(err, status.Errorf(codes.NotFound, "rating not found for a record")) {
		// ok
	} else if err != nil {
		return nil, err
	} else {
		details.Rating = &rating
	}
	return details, nil
}
