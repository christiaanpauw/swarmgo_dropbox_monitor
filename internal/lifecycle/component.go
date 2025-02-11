package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// State represents the current state of a component
type State int

const (
	// StateUninitialized is the default state of a component
	StateUninitialized State = iota
	// StateInitialized means the component has been initialized but not started
	StateInitialized
	// StateStarting means the component is in the process of starting
	StateStarting
	// StateRunning means the component is running normally
	StateRunning
	// StateStopping means the component is in the process of stopping
	StateStopping
	// StateStopped means the component has been stopped
	StateStopped
	// StateFailed means the component has encountered an error
	StateFailed
)

// String returns a string representation of the state
func (s State) String() string {
	switch s {
	case StateUninitialized:
		return "Uninitialized"
	case StateInitialized:
		return "Initialized"
	case StateStarting:
		return "Starting"
	case StateRunning:
		return "Running"
	case StateStopping:
		return "Stopping"
	case StateStopped:
		return "Stopped"
	case StateFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// Component represents a component that can be started and stopped
type Component interface {
	// Start starts the component
	Start(context.Context) error
	// Stop stops the component
	Stop(context.Context) error
	// State returns the current state of the component
	State() State
	// Health returns an error if the component is unhealthy
	Health(context.Context) error
}

// BaseComponent provides a base implementation of the Component interface
type BaseComponent struct {
	mu    sync.RWMutex
	state State
	name  string
}

// NewBaseComponent creates a new BaseComponent
func NewBaseComponent(name string) *BaseComponent {
	return &BaseComponent{
		name:  name,
		state: StateUninitialized,
	}
}

// SetState sets the component state with proper locking
func (c *BaseComponent) SetState(s State) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = s
}

// State returns the current state of the component
func (c *BaseComponent) State() State {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// Name returns the component name
func (c *BaseComponent) Name() string {
	return c.name
}

// DefaultStart provides a default implementation of Start
func (c *BaseComponent) DefaultStart(ctx context.Context) error {
	if c.State() != StateInitialized {
		return fmt.Errorf("component %s must be in Initialized state to start, current state: %s", c.Name(), c.State())
	}
	c.SetState(StateStarting)
	c.SetState(StateRunning)
	return nil
}

// DefaultStop provides a default implementation of Stop
func (c *BaseComponent) DefaultStop(ctx context.Context) error {
	if c.State() != StateRunning {
		return fmt.Errorf("component %s must be in Running state to stop, current state: %s", c.Name(), c.State())
	}
	c.SetState(StateStopping)
	c.SetState(StateStopped)
	return nil
}

// DefaultHealth provides a default implementation of Health
func (c *BaseComponent) DefaultHealth(ctx context.Context) error {
	if c.State() != StateRunning {
		return fmt.Errorf("component %s is not running, current state: %s", c.Name(), c.State())
	}
	return nil
}

// StartupSequence runs a standard startup sequence for a component
func StartupSequence(ctx context.Context, component Component, timeout time.Duration) error {
	// Create a context with timeout for startup
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Start the component
	if err := component.Start(ctx); err != nil {
		return fmt.Errorf("failed to start component: %w", err)
	}

	// Wait for component to be healthy
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("startup sequence timed out")
		case <-ticker.C:
			if err := component.Health(ctx); err == nil {
				return nil
			}
		}
	}
}

// ShutdownSequence runs a standard shutdown sequence for a component
func ShutdownSequence(ctx context.Context, component Component, timeout time.Duration) error {
	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Stop the component
	if err := component.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop component: %w", err)
	}

	// Wait for component to be stopped
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("shutdown sequence timed out")
		case <-ticker.C:
			if component.State() == StateStopped {
				return nil
			}
		}
	}
}
