package grpc

import (
	"context"
	"mmoviecom/gen"
	"mmoviecom/internal/grpcutil"
	"mmoviecom/metadata/pkg/model"
	"mmoviecom/pkg/discovery"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// Gateway defines a movie metadata gRPC gateway.
type Gateway struct {
	registry discovery.Registry
	creds    credentials.TransportCredentials
}

// New creates a new gRPC gateway for a movie metadata service.
func New(registry discovery.Registry, creds credentials.TransportCredentials) *Gateway {
	return &Gateway{registry: registry, creds: creds}
}

// Get returns movie metadata by a movie id.
func (g *Gateway) Get(ctx context.Context, id string) (*model.Metadata, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry, g.creds)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := gen.NewMetadataServiceClient(conn)
	const maxRetries = 5
	for i := 0; i < maxRetries; i++ {
		resp, err := client.GetMetadata(ctx, &gen.GetMetadataRequest{MovieId: id})
		if err != nil {
			if shouldRetry(err) {
				time.Sleep(1 * time.Second)
				continue
			}
			return nil, err
		}
		return model.MetadataFromProto(resp.GetMetadata()), nil
	}
	return nil, err
}

// Put stores movie metadata by a movie id.
func (g *Gateway) Put(ctx context.Context, metadata *model.Metadata) error {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry, g.creds)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := gen.NewMetadataServiceClient(conn)
	_, err = client.PutMetadata(ctx, &gen.PutMetadataRequest{Metadata: model.MetadataToProto(metadata)})
	if err != nil {
		return err
	}
	return nil
}

func shouldRetry(err error) bool {
	e, ok := status.FromError(err)
	if !ok {
		return false
	}
	return e.Code() == codes.DeadlineExceeded || e.Code() == codes.ResourceExhausted || e.Code() == codes.Unavailable
}
