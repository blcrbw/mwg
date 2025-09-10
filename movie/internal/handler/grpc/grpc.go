package grpc

import (
	"context"
	"errors"
	"mmoviecom/gen"
	"mmoviecom/metadata/pkg/model"
	"mmoviecom/movie/internal/controller/movie"
	"mmoviecom/pkg/logging"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler defines movie GRPC handler.
type Handler struct {
	gen.UnimplementedMovieServiceServer
	ctrl   *movie.Controller
	logger *zap.Logger
}

// New creates a new movie gRPC handler.
func New(ctrl *movie.Controller, logger *zap.Logger) *Handler {
	logger = logger.With(
		zap.String(logging.FieldComponent, "handler"),
		zap.String(logging.FieldType, "grpc"),
	)
	return &Handler{ctrl: ctrl, logger: logger}
}

// GetMovieDetails returns movie details by id.
func (h *Handler) GetMovieDetails(ctx context.Context, req *gen.GetMovieDetailsRequest) (*gen.GetMovieDetailsResponse, error) {
	if req == nil || req.MovieId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty id")
	}
	m, err := h.ctrl.Get(ctx, req.MovieId)
	if err != nil && errors.Is(err, movie.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	var rating float64
	if m.Rating != nil {
		rating = *m.Rating
	}
	return &gen.GetMovieDetailsResponse{
		MovieDetails: &gen.MovieDetails{
			Metadata: model.MetadataToProto(&m.Metadata),
			Rating:   rating,
		},
	}, nil
}
