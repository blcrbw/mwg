package grpc

import (
	"context"
	"mmoviecom/gen"
	"mmoviecom/internal/grpcutil"
	"mmoviecom/metadata/pkg/model"
	"mmoviecom/pkg/discovery"
)

// Gateway defines a movie metadata gRPC gateway.
type Gateway struct {
	registry discovery.Registry
}

// New creates a new gRPC gateway for a movie metadata service.
func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry: registry}
}

// Get returns movie metadata by a movie id.
func (g *Gateway) Get(ctx context.Context, id string) (*model.Metadata, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := gen.NewMetadataServiceClient(conn)
	resp, err := client.GetMetadata(ctx, &gen.GetMetadataRequest{MovieId: id})
	if err != nil {
		return nil, err
	}
	return model.MetadataFromProto(resp.Metadata), nil
}

// Put stores movie metadata by a movie id.
func (g *Gateway) Put(ctx context.Context, metadata *model.Metadata) error {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry)
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
