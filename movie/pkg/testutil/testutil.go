package testutil

import (
	"mmoviecom/gen"
	"mmoviecom/movie/internal/controller/movie"
	metadatagateway "mmoviecom/movie/internal/gateway/metadata/grpc"
	ratinggateway "mmoviecom/movie/internal/gateway/rating/grpc"
	"mmoviecom/movie/internal/handler/grpc"
	"mmoviecom/pkg/discovery"
	"mmoviecom/pkg/logging"

	"github.com/uber-go/tally/v6"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials/insecure"
)

func NewTestMovieGRPCServer(registry discovery.Registry, logger *zap.Logger, scope tally.Scope) gen.MovieServiceServer {
	logger = logger.With(
		zap.String(logging.FieldService, "movie"),
	)
	m := metadatagateway.New(registry, insecure.NewCredentials(), logger)
	r := ratinggateway.New(registry, insecure.NewCredentials(), logger)
	ctrl := movie.New(r, m, logger)
	return grpc.New(ctrl, logger, scope)
}
