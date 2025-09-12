package testutil

import (
	"mmoviecom/gen"
	"mmoviecom/pkg/discovery"
	"mmoviecom/pkg/logging"
	"mmoviecom/rating/internal/controller/rating"
	authgateway "mmoviecom/rating/internal/gateway/auth/grpc"
	"mmoviecom/rating/internal/handler/grpc"
	"mmoviecom/rating/internal/ingester/kafka"
	"mmoviecom/rating/internal/repository/memory"

	"github.com/uber-go/tally/v6"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials/insecure"
)

func NewTestRatingGRPCServer(registry discovery.Registry, logger *zap.Logger, scope tally.Scope) gen.RatingServiceServer {
	logger = logger.With(
		zap.String(logging.FieldService, "rating"),
	)
	r := memory.New(logger)

	ingester, err := kafka.NewIngester("localhost", "rating", "ratings", logger)
	if err != nil {
		logger.Fatal("Failed to initialize ingester", zap.Error(err))
	}

	auth := authgateway.New(registry, insecure.NewCredentials(), logger)
	ctrl := rating.New(r, ingester, auth, logger)
	return grpc.New(ctrl, logger, scope)
}
