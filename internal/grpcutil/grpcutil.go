package grpcutil

import (
	"context"
	"math/rand"
	"mmoviecom/pkg/discovery"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ServiceConnection attempts to select a random service instance
// and returns a gRPC connection to it.
func ServiceConnection(ctx context.Context, serviceName string, registry discovery.Registry, creds credentials.TransportCredentials) (*grpc.ClientConn, error) {
	addrs, err := registry.ServiceAddresses(ctx, serviceName)
	if err != nil {
		return nil, err
	}
	return grpc.NewClient(addrs[rand.Intn(len(addrs))], grpc.WithTransportCredentials(creds))
}
