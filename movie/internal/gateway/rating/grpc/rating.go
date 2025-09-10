package grpc

import (
	"context"
	"mmoviecom/gen"
	"mmoviecom/internal/grpcutil"
	"mmoviecom/pkg/discovery"
	"mmoviecom/pkg/logging"
	"mmoviecom/rating/pkg/model"

	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

// Gateway defines an gRPC getaway for a rating service.
type Gateway struct {
	registry discovery.Registry
	creds    credentials.TransportCredentials
	logger   *zap.Logger
}

// New creates a new gPRC gateway for a rating service.
func New(registry discovery.Registry, creds credentials.TransportCredentials, logger *zap.Logger) *Gateway {
	logger = logger.With(
		zap.String(logging.FieldComponent, "rating-gateway"),
		zap.String(logging.FieldType, "grpc"),
	)
	return &Gateway{registry: registry, creds: creds, logger: logger}
}

// GetAggregatedRating returns the aggregated rating for a
// record or ErrNotFound if there are no rating for it.
func (g *Gateway) GetAggregatedRating(ctx context.Context, recordID model.RecordId, recordType model.RecordType) (float64, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "rating", g.registry, g.creds)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	client := gen.NewRatingServiceClient(conn)
	resp, err := client.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{RecordId: string(recordID), RecordType: string(recordType)})
	if err != nil {
		return 0, err
	}
	return resp.RatingValue, nil
}

func (g *Gateway) PutRating(ctx context.Context, recordId model.RecordId, recordType model.RecordType, rating *model.Rating, token string) error {
	conn, err := grpcutil.ServiceConnection(ctx, "rating", g.registry, g.creds)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := gen.NewRatingServiceClient(conn)
	_, err = client.PutRating(ctx, &gen.PutRatingRequest{RecordId: string(recordId), RecordType: string(recordType), UserId: string(rating.UserId), RatingValue: int32(rating.Value), Token: token})
	if err != nil {
		return err
	}
	return nil
}
