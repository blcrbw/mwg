package testutil

import (
	"mmoviecom/gen"
	"mmoviecom/metadata/internal/controller/metadata"
	"mmoviecom/metadata/internal/handler/grpc"
	"mmoviecom/metadata/internal/repository/memory"
	"mmoviecom/pkg/logging"

	"github.com/uber-go/tally/v6"
	"go.uber.org/zap"
)

func NewTestMetadataGRPCServer(logger *zap.Logger, scope tally.Scope) gen.MetadataServiceServer {
	logger = logger.With(
		zap.String(logging.FieldService, "metadata"),
	)
	r := memory.New(logger)
	c := memory.New(logger)
	ctrl := metadata.New(r, c, logger)
	return grpc.New(ctrl, logger, scope)
}
