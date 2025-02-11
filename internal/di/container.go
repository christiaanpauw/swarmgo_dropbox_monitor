package di

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agents"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/analysis"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/config"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
)

const (
	defaultStartupTimeout  = 30 * time.Second
	defaultShutdownTimeout = 30 * time.Second
)

// Container manages application-wide dependencies
type Container struct {
	*lifecycle.BaseComponent
	mu sync.RWMutex

	config *config.Config

	// Agents and core components
	agentManager    *agents.AgentManager
	fileChangeAgent agents.FileChangeAgent
	databaseAgent   agents.DatabaseAgent
	reportingAgent  agents.ReportingAgent
	contentAnalyzer analysis.ContentAnalyzer
	notifier        notify.Notifier

	// Track managed components for lifecycle management
	components []lifecycle.Component
}

// NewContainer creates a new DI container with the provided configuration
func NewContainer(cfg *config.Config) (*Container, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	c := &Container{
		BaseComponent: lifecycle.NewBaseComponent("Container"),
		config:        cfg,
	}

	// Initialize dependencies in the correct order
	if err := c.initDependencies(); err != nil {
		return nil, fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	return c, nil
}

// initDependencies initializes all dependencies in the correct order
func (c *Container) initDependencies() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Initialize core components first
	var err error

	// Initialize notifier
	c.notifier = notify.NewNotifier()
	if lifecycleComponent, ok := c.notifier.(lifecycle.Component); ok {
		c.components = append(c.components, lifecycleComponent)
	}

	// Initialize content analyzer
	c.contentAnalyzer = analysis.NewContentAnalyzer()
	if lifecycleComponent, ok := c.contentAnalyzer.(lifecycle.Component); ok {
		c.components = append(c.components, lifecycleComponent)
	}

	// Initialize agents
	c.fileChangeAgent, err = agents.NewFileChangeAgent(c.config.DropboxAccessToken)
	if err != nil {
		return fmt.Errorf("failed to create file change agent: %w", err)
	}
	if lifecycleComponent, ok := c.fileChangeAgent.(lifecycle.Component); ok {
		c.components = append(c.components, lifecycleComponent)
	}

	c.databaseAgent, err = agents.NewDatabaseAgent()
	if err != nil {
		return fmt.Errorf("failed to create database agent: %w", err)
	}
	if lifecycleComponent, ok := c.databaseAgent.(lifecycle.Component); ok {
		c.components = append(c.components, lifecycleComponent)
	}

	c.reportingAgent = agents.NewReportingAgent(c.notifier)
	if lifecycleComponent, ok := c.reportingAgent.(lifecycle.Component); ok {
		c.components = append(c.components, lifecycleComponent)
	}

	// Initialize agent manager last since it depends on all other components
	deps := agents.AgentManagerDeps{
		FileChangeAgent:  c.fileChangeAgent,
		ContentAnalyzer:  c.contentAnalyzer,
		DatabaseAgent:    c.databaseAgent,
		ReportingAgent:   c.reportingAgent,
		Notifier:         c.notifier,
	}

	cfg := agents.AgentManagerConfig{
		PollInterval: c.config.PollInterval,
		MaxRetries:   c.config.MaxRetries,
		RetryDelay:   c.config.RetryDelay,
	}

	c.agentManager = agents.NewAgentManager(cfg, deps)
	if lifecycleComponent, ok := interface{}(c.agentManager).(lifecycle.Component); ok {
		c.components = append(c.components, lifecycleComponent)
	}

	c.SetState(lifecycle.StateInitialized)
	return nil
}

// Start starts all components in the container
func (c *Container) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.State() != lifecycle.StateInitialized {
		return fmt.Errorf("container must be in Initialized state to start, current state: %s", c.State())
	}

	c.SetState(lifecycle.StateStarting)

	// Start all components in order
	for _, component := range c.components {
		if err := lifecycle.StartupSequence(ctx, component, defaultStartupTimeout); err != nil {
			c.SetState(lifecycle.StateFailed)
			return fmt.Errorf("failed to start component: %w", err)
		}
	}

	c.SetState(lifecycle.StateRunning)
	return nil
}

// Stop stops all components in the container
func (c *Container) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.State() != lifecycle.StateRunning {
		return fmt.Errorf("container must be in Running state to stop, current state: %s", c.State())
	}

	c.SetState(lifecycle.StateStopping)

	// Stop components in reverse order
	for i := len(c.components) - 1; i >= 0; i-- {
		if err := lifecycle.ShutdownSequence(ctx, c.components[i], defaultShutdownTimeout); err != nil {
			c.SetState(lifecycle.StateFailed)
			return fmt.Errorf("failed to stop component: %w", err)
		}
	}

	c.SetState(lifecycle.StateStopped)
	return nil
}

// Health checks the health of all components
func (c *Container) Health(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.State() != lifecycle.StateRunning {
		return fmt.Errorf("container is not running, current state: %s", c.State())
	}

	for _, component := range c.components {
		if err := component.Health(ctx); err != nil {
			return fmt.Errorf("component health check failed: %w", err)
		}
	}

	return nil
}

// GetAgentManager returns the initialized agent manager
func (c *Container) GetAgentManager() *agents.AgentManager {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.agentManager
}

// GetNotifier returns the initialized notifier
func (c *Container) GetNotifier() notify.Notifier {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.notifier
}

// GetContentAnalyzer returns the initialized content analyzer
func (c *Container) GetContentAnalyzer() analysis.ContentAnalyzer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.contentAnalyzer
}

// Close is deprecated, use Stop instead
func (c *Container) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()
	return c.Stop(ctx)
}

func (c *Container) Startup(ctx context.Context) error {
	return c.Start(ctx)
}

func (c *Container) Shutdown(ctx context.Context) error {
	return c.Stop(ctx)
}

func (c *Container) HealthCheck(ctx context.Context) error {
	return c.Health(ctx)
}
