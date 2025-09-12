package grpc

import (
	"context"
	"errors"
	"mmoviecom/gen"
	"mmoviecom/metadata/internal/controller/metadata"
	"mmoviecom/metadata/pkg/model"
	"mmoviecom/pkg/logging"
	"mmoviecom/pkg/metrics"

	"github.com/uber-go/tally/v6"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler deefines a movie metadata gRPC handler.
type Handler struct {
	gen.UnimplementedMetadataServiceServer
	ctrl               *metadata.Controller
	logger             *zap.Logger
	getMetadataMetrics *metrics.EndpointMetrics
	putMetadataMetrics *metrics.EndpointMetrics
}

// New creates a new movie metadata gRPC handler.
func New(ctrl *metadata.Controller, logger *zap.Logger, scope tally.Scope) *Handler {
	logger = logger.With(
		zap.String(logging.FieldComponent, "handler"),
		zap.String(logging.FieldType, "grpc"),
	)
	return &Handler{
		ctrl:               ctrl,
		logger:             logger,
		getMetadataMetrics: metrics.NewEndpointMetrics(scope, "GetMetadata"),
		putMetadataMetrics: metrics.NewEndpointMetrics(scope, "PutMetadata"),
	}
}

// GetMetadata returns movie metadata by id.
func (h *Handler) GetMetadata(ctx context.Context, req *gen.GetMetadataRequest) (*gen.GetMetadataResponse, error) {
	h.getMetadataMetrics.Calls.Inc(1)
	if req == nil || req.MovieId == "" {
		h.getMetadataMetrics.InvalidArgumentErrors.Inc(1)
		return nil, status.Error(codes.InvalidArgument, "nil req or empty id")
	}
	m, err := h.ctrl.Get(ctx, req.MovieId)
	if err != nil && errors.Is(err, metadata.ErrNotFound) {
		h.getMetadataMetrics.NotFoundErrors.Inc(1)
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		h.getMetadataMetrics.InternalErrors.Inc(1)
		return nil, status.Error(codes.Internal, err.Error())
	}
	h.getMetadataMetrics.Successes.Inc(1)
	return &gen.GetMetadataResponse{Metadata: model.MetadataToProto(m)}, nil
}

// PutMetadata receives movie metadata and stores it.
func (h *Handler) PutMetadata(ctx context.Context, req *gen.PutMetadataRequest) (*gen.PutMetadataResponse, error) {
	h.putMetadataMetrics.Calls.Inc(1)
	if req == nil || req.Metadata == nil {
		h.putMetadataMetrics.InvalidArgumentErrors.Inc(1)
		return nil, status.Error(codes.InvalidArgument, "nil req or metadata")
	}
	if err := h.ctrl.Put(ctx, req.Metadata.Id, model.MetadataFromProto(req.Metadata)); err != nil {
		h.putMetadataMetrics.InternalErrors.Inc(1)
		return nil, status.Error(codes.Internal, err.Error())
	}
	h.putMetadataMetrics.Successes.Inc(1)
	return &gen.PutMetadataResponse{}, nil
}
