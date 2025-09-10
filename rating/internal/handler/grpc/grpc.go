package grpc

import (
	"context"
	"errors"
	"mmoviecom/gen"
	"mmoviecom/pkg/logging"
	"mmoviecom/rating/internal/controller/rating"
	"mmoviecom/rating/pkg/model"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler define a gRPC rating API handler.
type Handler struct {
	gen.UnimplementedRatingServiceServer
	svc    *rating.Controller
	logger *zap.Logger
}

// New creates a new rating gRPC handler.
func New(svc *rating.Controller, logger *zap.Logger) *Handler {
	logger = logger.With(
		zap.String(logging.FieldComponent, "handler"),
		zap.String(logging.FieldType, "grpc"),
	)
	return &Handler{svc: svc, logger: logger}
}

// GetAggregatedRating returns the aggregated rating for a record.
func (h *Handler) GetAggregatedRating(ctx context.Context, req *gen.GetAggregatedRatingRequest) (*gen.GetAggregatedRatingResponse, error) {
	if req == nil || req.RecordId == "" || req.RecordType == "" {
		return nil, status.Error(codes.InvalidArgument, "nil req or empty id/type")
	}
	v, err := h.svc.GetAggregatedRating(ctx, model.RecordId(req.RecordId), model.RecordType(req.RecordType))
	if err != nil && errors.Is(err, rating.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.GetAggregatedRatingResponse{RatingValue: v}, nil
}

// PutRating writes a rating for a given record.
func (h *Handler) PutRating(ctx context.Context, req *gen.PutRatingRequest) (*gen.PutRatingResponse, error) {
	if req == nil || req.RecordId == "" || req.RecordType == "" || req.UserId == "" || req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "nil req or empty user id or record id/type or token")
	}
	record := model.Rating{UserId: model.UserId(req.UserId), Value: model.RatingValue(req.RatingValue)}
	if err := h.svc.ValidateToken(ctx, req.GetToken(), &record); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := h.svc.PutRating(ctx, model.RecordId(req.RecordId), model.RecordType(req.RecordType), &record); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.PutRatingResponse{}, nil
}
