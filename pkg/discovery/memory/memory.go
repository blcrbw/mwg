package memory

import (
	"context"
	"errors"
	"log"
	"mmoviecom/pkg/discovery"
	"sync"
	"time"
)

// Registry defines an in-memory service registry.
type Registry struct {
	sync.RWMutex
	ServiceAddrs map[string]map[string]*serviceInstance
}

// serviceInstance defines a service instance record in the registry.
type serviceInstance struct {
	hostPort   string
	lastActive time.Time
}

// NewRegistry creates a new in-memory service registry instance.
func NewRegistry() *Registry {
	return &Registry{
		ServiceAddrs: make(map[string]map[string]*serviceInstance),
	}
}

// Register creates a service record in the registry.
func (r *Registry) Register(_ context.Context, instanceID string, serviceName string, hostPort string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.ServiceAddrs[serviceName]; !ok {
		r.ServiceAddrs[serviceName] = make(map[string]*serviceInstance)
	}
	r.ServiceAddrs[serviceName][instanceID] = &serviceInstance{hostPort: hostPort, lastActive: time.Now()}
	return nil
}

// Deregister removes a service record from the registry.
func (r *Registry) Deregister(_ context.Context, instanceID string, serviceName string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.ServiceAddrs[serviceName]; !ok {
		return nil
	}
	delete(r.ServiceAddrs[serviceName], instanceID)
	return nil
}

// ReportHealthyState is a push mechanism for reporting healthy state to the registry.
func (r *Registry) ReportHealthyState(instanceID string, serviceName string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.ServiceAddrs[serviceName]; !ok {
		return errors.New("service is not registered yet")
	}
	if _, ok := r.ServiceAddrs[serviceName][instanceID]; !ok {
		return errors.New("instance " + instanceID + " of service " + serviceName + " is not registered yet")
	}
	r.ServiceAddrs[serviceName][instanceID].lastActive = time.Now()
	return nil
}

func (r *Registry) ServiceAddresses(_ context.Context, serviceName string) ([]string, error) {
	r.RLock()
	defer r.RUnlock()
	if len(r.ServiceAddrs[serviceName]) == 0 {
		return nil, discovery.ErrNotFound
	}
	var res []string
	for instanceId, i := range r.ServiceAddrs[serviceName] {
		if i.lastActive.Before(time.Now().Add(-5 * time.Second)) {
			log.Println("Instance " + instanceId + " of service " + serviceName + " is not active, skipping")
			continue
		}
		res = append(res, i.hostPort)
	}
	return res, nil
}
