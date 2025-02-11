package container

import (
	"context"
	"fmt"
	"sync"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agents"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/analysis"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/config"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/health"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
)

// Container manages application dependencies
type Container struct {
	config        *config.Config
	healthChecker *health.HealthChecker
	agentManager  *agents.AgentManager
	mu            sync.RWMutex
	isStarted     bool
}

// New creates a new dependency container
func New(cfg *config.Config) *Container {
	return &Container{
		config:        cfg,
		healthChecker: health.NewHealthChecker(cfg.HealthCheckInterval),
	}
}

// Start initializes and starts all components
func (c *Container) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isStarted {
		return fmt.Errorf("container already started")
	}

	// Initialize components
	if err := c.initComponents(); err != nil {
		return fmt.Errorf("initialize components: %w", err)
	}

	// Register health checks
	c.registerHealthChecks()

	// Start health checker
	go c.healthChecker.Start(ctx)

	c.isStarted = true
	return nil
}

// Stop stops all components
func (c *Container) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isStarted {
		return nil
	}

	// Stop health checker
	c.healthChecker.Stop()

	// Stop agent manager
	if c.agentManager != nil {
		if err := c.agentManager.Stop(ctx); err != nil {
			return fmt.Errorf("stop agent manager: %w", err)
		}
	}

	c.isStarted = false
	return nil
}

// initComponents initializes all components
func (c *Container) initComponents() error {
	// Create notifier
	notifier := notify.NewNotifier()

	// Create content analyzer
	contentAnalyzer := analysis.NewContentAnalyzer()

	// Create file change agent
	fileChangeAgent, err := agents.NewFileChangeAgent(c.config.DropboxAccessToken)
	if err != nil {
		return fmt.Errorf("create file change agent: %w", err)
	}

	// Create database agent
	databaseAgent, err := agents.NewDatabaseAgent()
	if err != nil {
		return fmt.Errorf("create database agent: %w", err)
	}

	// Create reporting agent
	reportingAgent := agents.NewReportingAgent(notifier)

	// Create agent manager
	c.agentManager = &agents.AgentManager{
		FileChangeAgent:  fileChangeAgent,
		ContentAnalyzer:  contentAnalyzer,
		DatabaseAgent:    databaseAgent,
		ReportingAgent:   reportingAgent,
		Notifier:        notifier,
	}

	return nil
}

// registerHealthChecks registers health checks for all components
func (c *Container) registerHealthChecks() {
	// Register database health check
	c.healthChecker.RegisterComponent("database", func(ctx context.Context) error {
		// TODO: Implement database health check
		return nil
	})

	// Register Dropbox health check
	c.healthChecker.RegisterComponent("dropbox", func(ctx context.Context) error {
		// TODO: Implement Dropbox health check
		return nil
	})

	// Register agent manager health check
	c.healthChecker.RegisterComponent("agent_manager", func(ctx context.Context) error {
		// TODO: Implement agent manager health check
		return nil
	})
}

// GetAgentManager returns the agent manager
func (c *Container) GetAgentManager() *agents.AgentManager {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.agentManager
}

// GetHealthChecker returns the health checker
func (c *Container) GetHealthChecker() *health.HealthChecker {
	return c.healthChecker
}

// IsHealthy returns true if all components are healthy
func (c *Container) IsHealthy() bool {
	return c.healthChecker.IsHealthy()
}
