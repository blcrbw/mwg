package testutil

import (
	"log"
	"mmoviecom/gen"
	"mmoviecom/rating/internal/controller/rating"
	"mmoviecom/rating/internal/handler/grpc"
	"mmoviecom/rating/internal/ingester/kafka"
	"mmoviecom/rating/internal/repository/memory"
)

func NewTestRatingGRPCServer() gen.RatingServiceServer {
	r := memory.New()

	ingester, err := kafka.NewIngester("localhost", "rating", "ratings")
	if err != nil {
		log.Fatalf("Failed to initialize ingester: %v", err)
	}
	ctrl := rating.New(r, ingester)
	return grpc.New(ctrl)
}
