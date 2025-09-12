package testutil

import (
	"mmoviecom/auth/internal/handler/grpc"
	"mmoviecom/gen"

	"github.com/uber-go/tally/v6"
)

func NewTestAuthGRPCServer(scope tally.Scope) gen.AuthServiceServer {
	return grpc.New(func() []byte {
		return []byte("test-secret")
	}, scope)
}
