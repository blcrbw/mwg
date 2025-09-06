package rating

import (
	"context"
	"errors"
	"fmt"
	"mmoviecom/rating/internal/repository"
	"mmoviecom/rating/pkg/model"
)

// ErrNotFound returned when no ratings are found for a record.
var ErrNotFound = errors.New("rating not found for a record")
var ErrTokenIsEmpty = errors.New("token is empty")

type ratingRepository interface {
	Get(ctx context.Context, recordId model.RecordId, recordType model.RecordType) ([]model.Rating, error)
	Put(ctx context.Context, recordId model.RecordId, recordType model.RecordType, record *model.Rating) error
}

type ratingIngester interface {
	Ingest(ctx context.Context) (chan model.RatingEvent, error)
}

type AuthGateway interface {
	ValidateToken(ctx context.Context, token string) (string, error)
}

// Controller defines a rating service controller.
type Controller struct {
	repo     ratingRepository
	ingester ratingIngester
	auth     AuthGateway
}

// New creates a rating service controller.
func New(repo ratingRepository, ingester ratingIngester, auth AuthGateway) *Controller {
	return &Controller{repo: repo, ingester: ingester, auth: auth}
}

// GetAggregatedRating returns the aggregated rating for a
// record or ErrNotFound if there are no ratings for it.
func (c *Controller) GetAggregatedRating(ctx context.Context, recordId model.RecordId, recordType model.RecordType) (float64, error) {
	ratings, err := c.repo.Get(ctx, recordId, recordType)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return 0, ErrNotFound
	} else if err != nil {
		return 0, err
	} else if len(ratings) == 0 {
		return 0, ErrNotFound
	}
	sum := float64(0)
	for _, r := range ratings {
		sum += float64(r.Value)
	}
	return sum / float64(len(ratings)), nil
}

// PutRating writes a rating for a given record.
func (c *Controller) PutRating(ctx context.Context, recordId model.RecordId, recordType model.RecordType, record *model.Rating) error {
	return c.repo.Put(ctx, recordId, recordType, record)
}

// ValidateToken validates token, get user id from token and compares with record one.
func (c *Controller) ValidateToken(ctx context.Context, token string, record *model.Rating) error {
	if token == "" {
		return ErrTokenIsEmpty
	}
	user, err := c.auth.ValidateToken(ctx, token)
	if err != nil {
		return err
	}
	if user == "" || user != string(record.UserId) {
		return errors.New("incorrect token")
	}
	return nil
}

// StartIngestion starts the ingestion of rating events.
func (c *Controller) StartIngestion(ctx context.Context) error {
	ch, err := c.ingester.Ingest(ctx)
	if err != nil {
		return err
	}
	for e := range ch {
		fmt.Printf("Consume a message: %v\n", e)
		if err := c.PutRating(ctx, model.RecordId(e.RecordId), model.RecordType(e.RecordType), &model.Rating{
			UserId: e.UserId,
			Value:  e.Value,
		}); err != nil {
			return err
		}
	}
	return nil
}
