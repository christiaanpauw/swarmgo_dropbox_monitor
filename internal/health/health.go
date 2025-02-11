package health

import (
	"context"
	"sync"
	"time"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusStarting  Status = "starting"
	StatusStopped   Status = "stopped"
)

// Check represents a health check function
type Check func(ctx context.Context) error

// Component represents a component that can be health checked
type Component struct {
	Name   string
	Check  Check
	Status Status
	Error  error
}

// HealthChecker manages health checks for components
type HealthChecker struct {
	components map[string]*Component
	interval   time.Duration
	mu         sync.RWMutex
	stopCh     chan struct{}
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(interval time.Duration) *HealthChecker {
	return &HealthChecker{
		components: make(map[string]*Component),
		interval:   interval,
		stopCh:     make(chan struct{}),
	}
}

// RegisterComponent registers a component for health checking
func (h *HealthChecker) RegisterComponent(name string, check Check) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.components[name] = &Component{
		Name:   name,
		Check:  check,
		Status: StatusStarting,
	}
}

// Start starts the health checker
func (h *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopCh:
			return
		case <-ticker.C:
			h.checkAll(ctx)
		}
	}
}

// Stop stops the health checker
func (h *HealthChecker) Stop() {
	close(h.stopCh)
}

// checkAll checks all registered components
func (h *HealthChecker) checkAll(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, component := range h.components {
		if err := component.Check(ctx); err != nil {
			component.Status = StatusUnhealthy
			component.Error = err
		} else {
			component.Status = StatusHealthy
			component.Error = nil
		}
	}
}

// GetStatus returns the status of all components
func (h *HealthChecker) GetStatus() map[string]Status {
	h.mu.RLock()
	defer h.mu.RUnlock()

	status := make(map[string]Status)
	for name, component := range h.components {
		status[name] = component.Status
	}
	return status
}

// GetErrors returns any errors from unhealthy components
func (h *HealthChecker) GetErrors() map[string]error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	errors := make(map[string]error)
	for name, component := range h.components {
		if component.Error != nil {
			errors[name] = component.Error
		}
	}
	return errors
}

// IsHealthy returns true if all components are healthy
func (h *HealthChecker) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, component := range h.components {
		if component.Status != StatusHealthy {
			return false
		}
	}
	return true
}
