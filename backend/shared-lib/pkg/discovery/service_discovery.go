package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ServiceInstance represents a registered service instance
type ServiceInstance struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	Tags        []string          `json:"tags"`
	Meta        map[string]string `json:"meta"`
	HealthCheck string            `json:"health_check"`
	Status      ServiceStatus     `json:"status"`
	LastSeen    time.Time         `json:"last_seen"`
}

// ServiceStatus represents the health status of a service
type ServiceStatus string

const (
	StatusHealthy   ServiceStatus = "healthy"
	StatusUnhealthy ServiceStatus = "unhealthy"
	StatusUnknown   ServiceStatus = "unknown"
)

// ServiceRegistry interface for service registration and discovery
type ServiceRegistry interface {
	// Register registers a service instance
	Register(ctx context.Context, instance *ServiceInstance) error

	// Deregister removes a service instance
	Deregister(ctx context.Context, instanceID string) error

	// Discover finds healthy instances of a service
	Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error)

	// HealthCheck updates the health status of an instance
	HealthCheck(ctx context.Context, instanceID string, status ServiceStatus) error
}

// InMemoryRegistry implements ServiceRegistry in memory (for development)
type InMemoryRegistry struct {
	instances map[string]*ServiceInstance
	mu        sync.RWMutex
}

// NewInMemoryRegistry creates a new in-memory registry
func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		instances: make(map[string]*ServiceInstance),
	}
}

func (r *InMemoryRegistry) Register(ctx context.Context, instance *ServiceInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance.LastSeen = time.Now()
	instance.Status = StatusHealthy
	r.instances[instance.ID] = instance
	return nil
}

func (r *InMemoryRegistry) Deregister(ctx context.Context, instanceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.instances, instanceID)
	return nil
}

func (r *InMemoryRegistry) Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var healthy []*ServiceInstance
	for _, instance := range r.instances {
		if instance.Name == serviceName && instance.Status == StatusHealthy {
			healthy = append(healthy, instance)
		}
	}
	return healthy, nil
}

func (r *InMemoryRegistry) HealthCheck(ctx context.Context, instanceID string, status ServiceStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if instance, exists := r.instances[instanceID]; exists {
		instance.Status = status
		instance.LastSeen = time.Now()
	}
	return nil
}

// ServiceDiscoveryClient provides client-side service discovery
type ServiceDiscoveryClient struct {
	registry ServiceRegistry
	cache    map[string][]*ServiceInstance
	cacheTTL time.Duration
	cacheAt  map[string]time.Time
	mu       sync.RWMutex
}

// NewServiceDiscoveryClient creates a new discovery client
func NewServiceDiscoveryClient(registry ServiceRegistry, cacheTTL time.Duration) *ServiceDiscoveryClient {
	return &ServiceDiscoveryClient{
		registry: registry,
		cache:    make(map[string][]*ServiceInstance),
		cacheTTL: cacheTTL,
		cacheAt:  make(map[string]time.Time),
	}
}

// GetServiceURL returns a URL for a service (with load balancing)
func (c *ServiceDiscoveryClient) GetServiceURL(ctx context.Context, serviceName string) (string, error) {
	instances, err := c.getInstances(ctx, serviceName)
	if err != nil {
		return "", err
	}

	if len(instances) == 0 {
		return "", fmt.Errorf("no healthy instances found for service: %s", serviceName)
	}

	// Simple round-robin (in production, use more sophisticated LB)
	instance := instances[time.Now().UnixNano()%int64(len(instances))]
	return fmt.Sprintf("http://%s:%d", instance.Host, instance.Port), nil
}

func (c *ServiceDiscoveryClient) getInstances(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	c.mu.RLock()
	if cached, exists := c.cache[serviceName]; exists {
		if time.Since(c.cacheAt[serviceName]) < c.cacheTTL {
			c.mu.RUnlock()
			return cached, nil
		}
	}
	c.mu.RUnlock()

	// Fetch from registry
	instances, err := c.registry.Discover(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	// Update cache
	c.mu.Lock()
	c.cache[serviceName] = instances
	c.cacheAt[serviceName] = time.Now()
	c.mu.Unlock()

	return instances, nil
}

// ConsulConfig holds Consul configuration
type ConsulConfig struct {
	Address    string
	Datacenter string
	Token      string
}

// ConsulRegistry implements ServiceRegistry using Consul
// Note: This is a stub - in production, use the official Consul client
type ConsulRegistry struct {
	config *ConsulConfig
}

// NewConsulRegistry creates a new Consul-based registry
func NewConsulRegistry(config *ConsulConfig) *ConsulRegistry {
	return &ConsulRegistry{config: config}
}

func (r *ConsulRegistry) Register(ctx context.Context, instance *ServiceInstance) error {
	// In production, use: github.com/hashicorp/consul/api
	// client.Agent().ServiceRegister(&api.AgentServiceRegistration{...})
	return nil
}

func (r *ConsulRegistry) Deregister(ctx context.Context, instanceID string) error {
	// In production: client.Agent().ServiceDeregister(instanceID)
	return nil
}

func (r *ConsulRegistry) Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	// In production: client.Health().Service(serviceName, "", true, nil)
	return nil, nil
}

func (r *ConsulRegistry) HealthCheck(ctx context.Context, instanceID string, status ServiceStatus) error {
	// In production: client.Agent().PassTTL or FailTTL
	return nil
}
