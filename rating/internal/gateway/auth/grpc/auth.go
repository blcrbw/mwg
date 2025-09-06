package grpc

import (
	"context"
	"mmoviecom/gen"
	"mmoviecom/internal/grpcutil"
	"mmoviecom/pkg/discovery"

	"google.golang.org/grpc/credentials"
)

type Gateway struct {
	registry discovery.Registry
	creds    credentials.TransportCredentials
}

func New(registry discovery.Registry, creds credentials.TransportCredentials) *Gateway {
	return &Gateway{registry: registry, creds: creds}
}

func (g *Gateway) ValidateToken(ctx context.Context, token string) (string, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "auth", g.registry, g.creds)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	client := gen.NewAuthServiceClient(conn)
	resp, err := client.ValidateToken(ctx, &gen.ValidateTokenRequest{Token: token})
	if err != nil {
		return "", err
	}
	return resp.GetUsername(), nil
}
