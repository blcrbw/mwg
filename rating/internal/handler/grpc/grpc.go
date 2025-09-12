package grpc

import (
	"context"
	"errors"
	"mmoviecom/gen"
	"mmoviecom/pkg/logging"
	"mmoviecom/pkg/metrics"
	"mmoviecom/rating/internal/controller/rating"
	"mmoviecom/rating/pkg/model"

	"github.com/uber-go/tally/v6"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler define a gRPC rating API handler.
type Handler struct {
	gen.UnimplementedRatingServiceServer
	svc                        *rating.Controller
	logger                     *zap.Logger
	getAggregatedRatingMetrics *metrics.EndpointMetrics
	putRatingMetrics           *metrics.EndpointMetrics
}

// New creates a new rating gRPC handler.
func New(svc *rating.Controller, logger *zap.Logger, scope tally.Scope) *Handler {
	logger = logger.With(
		zap.String(logging.FieldComponent, "handler"),
		zap.String(logging.FieldType, "grpc"),
	)
	return &Handler{
		svc:                        svc,
		logger:                     logger,
		getAggregatedRatingMetrics: metrics.NewEndpointMetrics(scope, "GetAggregatedRating"),
		putRatingMetrics:           metrics.NewEndpointMetrics(scope, "PutRating"),
	}
}

// GetAggregatedRating returns the aggregated rating for a record.
func (h *Handler) GetAggregatedRating(ctx context.Context, req *gen.GetAggregatedRatingRequest) (*gen.GetAggregatedRatingResponse, error) {
	h.getAggregatedRatingMetrics.Calls.Inc(1)
	if req == nil || req.RecordId == "" || req.RecordType == "" {
		h.getAggregatedRatingMetrics.InvalidArgumentErrors.Inc(1)
		return nil, status.Error(codes.InvalidArgument, "nil req or empty id/type")
	}
	v, err := h.svc.GetAggregatedRating(ctx, model.RecordId(req.RecordId), model.RecordType(req.RecordType))
	if err != nil && errors.Is(err, rating.ErrNotFound) {
		h.getAggregatedRatingMetrics.NotFoundErrors.Inc(1)
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		h.getAggregatedRatingMetrics.InternalErrors.Inc(1)
		return nil, status.Error(codes.Internal, err.Error())
	}
	h.getAggregatedRatingMetrics.Successes.Inc(1)
	return &gen.GetAggregatedRatingResponse{RatingValue: v}, nil
}

// PutRating writes a rating for a given record.
func (h *Handler) PutRating(ctx context.Context, req *gen.PutRatingRequest) (*gen.PutRatingResponse, error) {
	h.putRatingMetrics.Calls.Inc(1)
	if req == nil || req.RecordId == "" || req.RecordType == "" || req.UserId == "" || req.Token == "" {
		h.putRatingMetrics.InvalidArgumentErrors.Inc(1)
		return nil, status.Error(codes.InvalidArgument, "nil req or empty user id or record id/type or token")
	}
	record := model.Rating{UserId: model.UserId(req.UserId), Value: model.RatingValue(req.RatingValue)}
	if err := h.svc.ValidateToken(ctx, req.GetToken(), &record); err != nil {
		h.putRatingMetrics.InvalidArgumentErrors.Inc(1)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := h.svc.PutRating(ctx, model.RecordId(req.RecordId), model.RecordType(req.RecordType), &record); err != nil {
		h.putRatingMetrics.InternalErrors.Inc(1)
		return nil, status.Error(codes.Internal, err.Error())
	}
	h.putRatingMetrics.Successes.Inc(1)
	return &gen.PutRatingResponse{}, nil
}
