package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mmoviecom/gen"
	"mmoviecom/metadata/pkg/model"
	"mmoviecom/movie/internal/controller/movie"
	"mmoviecom/pkg/logging"
	"mmoviecom/pkg/metrics"
	"os"

	"github.com/uber-go/tally/v6"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler defines movie GRPC handler.
type Handler struct {
	gen.UnimplementedMovieServiceServer
	ctrl                   *movie.Controller
	logger                 *zap.Logger
	getMovieDetailsMetrics *metrics.EndpointMetrics
}

// New creates a new movie gRPC handler.
func New(ctrl *movie.Controller, logger *zap.Logger, scope tally.Scope) *Handler {
	logger = logger.With(
		zap.String(logging.FieldComponent, "handler"),
		zap.String(logging.FieldType, "grpc"),
	)
	return &Handler{
		ctrl:                   ctrl,
		logger:                 logger,
		getMovieDetailsMetrics: metrics.NewEndpointMetrics(scope, "GetMovieDetails"),
	}
}

// GetMovieDetails returns movie details by id.
func (h *Handler) GetMovieDetails(ctx context.Context, req *gen.GetMovieDetailsRequest) (*gen.GetMovieDetailsResponse, error) {
	h.getMovieDetailsMetrics.Calls.Inc(1)
	if req == nil || req.MovieId == "" {
		h.getMovieDetailsMetrics.InvalidArgumentErrors.Inc(1)
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty id")
	}
	m, err := h.ctrl.Get(ctx, req.MovieId)
	if err != nil && errors.Is(err, movie.ErrNotFound) {
		h.getMovieDetailsMetrics.NotFoundErrors.Inc(1)

		return nil, status.Errorf(codes.NotFound, "not found")
	} else if err != nil {
		h.getMovieDetailsMetrics.InternalErrors.Inc(1)
		h.logger.Error("Cannot get movie details", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
	var rating float64
	if m.Rating != nil {
		rating = *m.Rating
	}
	h.getMovieDetailsMetrics.Successes.Inc(1)
	return &gen.GetMovieDetailsResponse{
		MovieDetails: &gen.MovieDetails{
			Metadata: model.MetadataToProto(&m.Metadata),
			Rating:   rating,
		},
	}, nil
}

// UploadFile handles streaming file upload.
func (h *Handler) UploadFile(stream gen.MovieService_UploadFileServer) error {
	var filename string
	var file *os.File
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&gen.UploadResponse{
				Message: fmt.Sprintf("File %s uploaded successfully", filename),
			})
		}
		if err != nil {
			return err
		}
		if file == nil {
			filename = req.GetFilename()
			file, err = os.Create(filename)
			if err != nil {
				return err
			}
			defer func() {
				if err := file.Close(); err != nil {
					h.logger.Error("Cannot close file", zap.Error(err))
				}
			}()
		}
		if _, err := file.Write(req.Chunk); err != nil {
			return err
		}
	}
}
