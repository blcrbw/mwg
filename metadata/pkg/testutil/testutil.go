package testutil

import (
	"mmoviecom/gen"
	"mmoviecom/metadata/internal/controller/metadata"
	"mmoviecom/metadata/internal/handler/grpc"
	"mmoviecom/metadata/internal/repository/memory"
)

func NewTestMetadataGRPCServer() gen.MetadataServiceServer {
	r := memory.New()
	c := memory.New()
	ctrl := metadata.New(r, c)
	return grpc.New(ctrl)
}
