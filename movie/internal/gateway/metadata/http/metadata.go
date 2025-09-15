package http

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"mmoviecom/metadata/pkg/model"
	"mmoviecom/movie/internal/gateway"
	"mmoviecom/pkg/discovery"
	"mmoviecom/pkg/logging"
	"net/http"

	"go.uber.org/zap"
)

// Gateway defines a movie metadata HTTP gateway.
type Gateway struct {
	registry discovery.Registry
	logger   *zap.Logger
}

// New creates a new HTTP gateway for a movie metadata service.
func New(registry discovery.Registry, logger *zap.Logger) *Gateway {
	logger = logger.With(
		zap.String(logging.FieldComponent, "metadata-gateway"),
		zap.String(logging.FieldType, "http"),
	)
	return &Gateway{registry: registry, logger: logger}
}

// Get gets movie metadata by a movie id.
func (g *Gateway) Get(ctx context.Context, id string) (*model.Metadata, error) {
	addrs, err := g.registry.ServiceAddresses(ctx, "metadata")
	if err != nil {
		return nil, err
	}
	url := "http://" + addrs[rand.Intn(len(addrs))] + "/metadata/"
	g.logger.Debug("Calling metadata service",
		zap.String("url", url),
		zap.String("method", "GET"),
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	values := req.URL.Query()
	values.Add("id", id)
	req.URL.RawQuery = values.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, gateway.ErrNotFound
	} else if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("non-2xx status code: %v", resp)
	}
	var v *model.Metadata
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}
