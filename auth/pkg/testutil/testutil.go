package testutil

import (
	"mmoviecom/auth/internal/handler/grpc"
	"mmoviecom/gen"
)

func NewTestAuthGRPCServer() gen.AuthServiceServer {
	return grpc.New(func() []byte {
		return []byte("test-secret")
	})
}
