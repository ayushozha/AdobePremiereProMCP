// Package health provides health checking, circuit breaker, and HTTP endpoint
// functionality for monitoring the three backend services that the Go
// orchestrator depends on: the Rust media engine, the Python intelligence
// service, and the TypeScript Premiere bridge.
package health

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Status represents the health state of a backend service.
type Status int

const (
	// StatusHealthy means the service is responding normally.
	StatusHealthy Status = iota
	// StatusDegraded means the service is responding but with elevated latency
	// or intermittent errors.
	StatusDegraded
	// StatusUnhealthy means the service is not responding or has exceeded the
	// consecutive failure threshold.
	StatusUnhealthy
)

// String returns a human-readable label for the status.
func (s Status) String() string {
	switch s {
	case StatusHealthy:
		return "healthy"
	case StatusDegraded:
		return "degraded"
	case StatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// ServiceHealth holds the most recently observed health state for a single
// backend service.
type ServiceHealth struct {
	Name        string
	Status      Status
	LastCheck   time.Time
	LastError   error
	Latency     time.Duration
	Consecutive int // consecutive failures
}

// degradedLatency is the latency threshold above which a service is considered
// degraded rather than healthy.
const degradedLatency = 500 * time.Millisecond

// unhealthyThreshold is the number of consecutive failures after which a
// service is considered unhealthy.
const unhealthyThreshold = 3

// defaultServices are the three backends that every orchestrator deployment
// must communicate with.
var defaultServices = []string{
	"media-engine",
	"intelligence",
	"premiere-bridge",
}

// Checker periodically probes backend services and caches their health states.
type Checker struct {
	services map[string]*ServiceHealth
	probes   map[string]Probe
	logger   *zap.Logger
	mu       sync.RWMutex
}

// NewChecker creates a Checker pre-populated with entries for the three
// default services. Probes can be registered afterwards via RegisterProbe.
func NewChecker(logger *zap.Logger) *Checker {
	services := make(map[string]*ServiceHealth, len(defaultServices))
	for _, name := range defaultServices {
		services[name] = &ServiceHealth{
			Name:   name,
			Status: StatusUnhealthy, // unknown until first check
		}
	}
	return &Checker{
		services: services,
		probes:   make(map[string]Probe, len(defaultServices)),
		logger:   logger,
	}
}

// RegisterProbe associates a Probe function with a service name. The probe
// will be invoked on each health check cycle.
func (c *Checker) RegisterProbe(name string, p Probe) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.probes[name] = p
	if _, ok := c.services[name]; !ok {
		c.services[name] = &ServiceHealth{
			Name:   name,
			Status: StatusUnhealthy,
		}
	}
}

// Start launches a background goroutine that calls CheckAll at the given
// interval until ctx is cancelled. If interval is zero, a 30-second default
// is used.
func (c *Checker) Start(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	c.logger.Info("health checker starting", zap.Duration("interval", interval))

	// Run an initial check immediately so callers don't have to wait a full
	// interval for the first status.
	c.CheckAll(ctx)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("health checker stopped")
				return
			case <-ticker.C:
				c.CheckAll(ctx)
			}
		}
	}()
}

// Check probes a single service and updates its cached ServiceHealth.
func (c *Checker) Check(ctx context.Context, serviceName string) *ServiceHealth {
	c.mu.RLock()
	probe, hasProbe := c.probes[serviceName]
	c.mu.RUnlock()

	if !hasProbe {
		c.logger.Warn("no probe registered for service", zap.String("service", serviceName))
		return c.GetStatus(serviceName)
	}

	start := time.Now()
	err := probe(ctx)
	latency := time.Since(start)

	c.mu.Lock()
	defer c.mu.Unlock()

	sh, ok := c.services[serviceName]
	if !ok {
		// Should not happen if RegisterProbe was called, but be defensive.
		sh = &ServiceHealth{Name: serviceName}
		c.services[serviceName] = sh
	}

	sh.LastCheck = time.Now()
	sh.Latency = latency

	if err != nil {
		sh.Consecutive++
		sh.LastError = err
		if sh.Consecutive >= unhealthyThreshold {
			sh.Status = StatusUnhealthy
		} else {
			sh.Status = StatusDegraded
		}
		c.logger.Warn("service health check failed",
			zap.String("service", serviceName),
			zap.Int("consecutive_failures", sh.Consecutive),
			zap.Duration("latency", latency),
			zap.Error(err),
		)
	} else {
		sh.LastError = nil
		sh.Consecutive = 0
		if latency > degradedLatency {
			sh.Status = StatusDegraded
			c.logger.Warn("service responding slowly",
				zap.String("service", serviceName),
				zap.Duration("latency", latency),
			)
		} else {
			sh.Status = StatusHealthy
		}
	}
	return sh
}

// CheckAll probes every registered service in parallel and updates all cached
// health states.
func (c *Checker) CheckAll(ctx context.Context) {
	c.mu.RLock()
	names := make([]string, 0, len(c.services))
	for name := range c.services {
		names = append(names, name)
	}
	c.mu.RUnlock()

	var wg sync.WaitGroup
	wg.Add(len(names))
	for _, name := range names {
		go func(n string) {
			defer wg.Done()
			c.Check(ctx, n)
		}(name)
	}
	wg.Wait()
}

// GetStatus returns a snapshot of the cached health for a single service. It
// returns nil if the service name is unknown.
func (c *Checker) GetStatus(serviceName string) *ServiceHealth {
	c.mu.RLock()
	defer c.mu.RUnlock()
	sh, ok := c.services[serviceName]
	if !ok {
		return nil
	}
	// Return a copy so the caller cannot mutate internal state.
	cp := *sh
	return &cp
}

// GetAllStatuses returns a snapshot of every tracked service's health.
func (c *Checker) GetAllStatuses() map[string]*ServiceHealth {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]*ServiceHealth, len(c.services))
	for k, v := range c.services {
		cp := *v
		out[k] = &cp
	}
	return out
}

// IsReady reports whether the orchestrator is ready to serve requests. The
// minimum requirement is that the premiere-bridge service is healthy.
func (c *Checker) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	sh, ok := c.services["premiere-bridge"]
	if !ok {
		return false
	}
	return sh.Status == StatusHealthy
}
