package grpc

import (
	"context"
	"errors"
	"mmoviecom/gen"
	"mmoviecom/metadata/internal/controller/metadata"
	"mmoviecom/metadata/pkg/model"
	"mmoviecom/pkg/logging"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler deefines a movie metadata gRPC handler.
type Handler struct {
	gen.UnimplementedMetadataServiceServer
	ctrl   *metadata.Controller
	logger *zap.Logger
}

// New creates a new movie metadata gRPC handler.
func New(ctrl *metadata.Controller, logger *zap.Logger) *Handler {
	logger = logger.With(
		zap.String(logging.FieldComponent, "handler"),
		zap.String(logging.FieldType, "grpc"),
	)
	return &Handler{ctrl: ctrl}
}

// GetMetadata returns movie metadata by id.
func (h *Handler) GetMetadata(ctx context.Context, req *gen.GetMetadataRequest) (*gen.GetMetadataResponse, error) {
	if req == nil || req.MovieId == "" {
		return nil, status.Error(codes.InvalidArgument, "nil req or empty id")
	}
	m, err := h.ctrl.Get(ctx, req.MovieId)
	if err != nil && errors.Is(err, metadata.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.GetMetadataResponse{Metadata: model.MetadataToProto(m)}, nil
}

// PutMetadata receives movie metadata and stores it.
func (h *Handler) PutMetadata(ctx context.Context, req *gen.PutMetadataRequest) (*gen.PutMetadataResponse, error) {
	if req == nil || req.Metadata == nil {
		return nil, status.Error(codes.InvalidArgument, "nil req or metadata")
	}
	if err := h.ctrl.Put(ctx, req.Metadata.Id, model.MetadataFromProto(req.Metadata)); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &gen.PutMetadataResponse{}, nil
}
