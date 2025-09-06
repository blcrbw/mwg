package testutil

import (
	"mmoviecom/gen"
	"mmoviecom/movie/internal/controller/movie"
	metadatagateway "mmoviecom/movie/internal/gateway/metadata/grpc"
	ratinggateway "mmoviecom/movie/internal/gateway/rating/grpc"
	"mmoviecom/movie/internal/handler/grpc"
	"mmoviecom/pkg/discovery"

	"google.golang.org/grpc/credentials/insecure"
)

func NewTestMovieGRPCServer(registry discovery.Registry) gen.MovieServiceServer {
	m := metadatagateway.New(registry, insecure.NewCredentials())
	r := ratinggateway.New(registry, insecure.NewCredentials())
	ctrl := movie.New(r, m)
	return grpc.New(ctrl)
}
