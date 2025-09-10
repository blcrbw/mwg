package consul

import (
	"context"
	"errors"
	"fmt"
	"mmoviecom/pkg/discovery"
	"mmoviecom/pkg/logging"
	"strconv"
	"strings"

	consul "github.com/hashicorp/consul/api"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

const tracerId = "discovery-consul"

// Registry defines a Consul-based service registry.
type Registry struct {
	client *consul.Client
	logger *zap.Logger
}

// NewRegistry creates a new Consul-based service registry instance.
func NewRegistry(addr string, logger *zap.Logger) (*Registry, error) {
	logger = logger.With(
		zap.String(logging.FieldComponent, "discovery"),
		zap.String(logging.FieldType, "consul"),
	)
	config := consul.DefaultConfig()
	config.Address = addr
	client, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Registry{client, logger}, nil
}

// Register creates a service record in the registry.
func (r *Registry) Register(ctx context.Context, instanceId string, serviceName string, hostPort string) error {
	_, span := otel.Tracer(tracerId).Start(ctx, "Register")
	defer span.End()
	parts := strings.Split(hostPort, ":")
	if len(parts) != 2 {
		return errors.New("hostPort must be in a form of <host>:<port>, example: localhost:8500")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}
	return r.client.Agent().ServiceRegister(&consul.AgentServiceRegistration{
		Address: parts[0],
		ID:      instanceId,
		Name:    serviceName,
		Port:    port,
		Check:   &consul.AgentServiceCheck{CheckID: instanceId, TTL: "5s"},
	})
}

// Deregister removes a service record from the registry.
func (r *Registry) Deregister(ctx context.Context, instanceId string, _ string) error {
	_, span := otel.Tracer(tracerId).Start(ctx, "Deregister")
	defer span.End()
	return r.client.Agent().ServiceDeregister(instanceId)
}

// ServiceAddresses returns the list of addresses of active instance of the given service.
func (r *Registry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	_, span := otel.Tracer(tracerId).Start(ctx, "ServiceAddresses")
	defer span.End()
	entries, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	} else if len(entries) == 0 {
		return nil, discovery.ErrNotFound
	}
	var res []string
	for _, e := range entries {
		res = append(res, fmt.Sprintf("%s:%d", e.Service.Address, e.Service.Port))
	}
	return res, nil
}

// ReportHealthyState is a push mechanism for reporting healthy state to the registry.
func (r *Registry) ReportHealthyState(instanceID string, _ string) error {
	_, span := otel.Tracer(tracerId).Start(context.Background(), "ReportHealthyState")
	defer span.End()
	return r.client.Agent().PassTTL(instanceID, "")
}
