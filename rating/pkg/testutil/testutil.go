package testutil

import (
	"log"
	"mmoviecom/gen"
	"mmoviecom/pkg/discovery"
	"mmoviecom/rating/internal/controller/rating"
	authgateway "mmoviecom/rating/internal/gateway/auth/grpc"
	"mmoviecom/rating/internal/handler/grpc"
	"mmoviecom/rating/internal/ingester/kafka"
	"mmoviecom/rating/internal/repository/memory"

	"google.golang.org/grpc/credentials/insecure"
)

func NewTestRatingGRPCServer(registry discovery.Registry) gen.RatingServiceServer {
	r := memory.New()

	ingester, err := kafka.NewIngester("localhost", "rating", "ratings")
	if err != nil {
		log.Fatalf("Failed to initialize ingester: %v", err)
	}

	auth := authgateway.New(registry, insecure.NewCredentials())
	ctrl := rating.New(r, ingester, auth)
	return grpc.New(ctrl)
}
