package testutil

import (
	"mmoviecom/auth/internal/handler/grpc"
	"mmoviecom/gen"

	"github.com/uber-go/tally/v6"
	"go.uber.org/zap"
)

func NewTestAuthGRPCServer(scope tally.Scope, logger *zap.Logger) gen.AuthServiceServer {
	return grpc.New(func() []byte {
		return []byte("test-secret")
	}, scope, logger)
}
