package di

import (
	"fmt"
	"sync"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agents"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/analysis"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/config"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
)

// Container manages application-wide dependencies
type Container struct {
	mu sync.RWMutex

	config *config.Config

	// Agents and core components
	agentManager    *agents.AgentManager
	fileChangeAgent agents.FileChangeAgent
	databaseAgent   agents.DatabaseAgent
	reportingAgent  agents.ReportingAgent
	contentAnalyzer analysis.ContentAnalyzer
	notifier        notify.Notifier
}

// NewContainer creates a new DI container with the provided configuration
func NewContainer(cfg *config.Config) (*Container, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	c := &Container{
		config: cfg,
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

	// Initialize content analyzer
	c.contentAnalyzer = analysis.NewContentAnalyzer()

	// Initialize agents
	c.fileChangeAgent, err = agents.NewFileChangeAgent(c.config.DropboxAccessToken)
	if err != nil {
		return fmt.Errorf("failed to create file change agent: %w", err)
	}

	c.databaseAgent, err = agents.NewDatabaseAgent()
	if err != nil {
		return fmt.Errorf("failed to create database agent: %w", err)
	}

	c.reportingAgent = agents.NewReportingAgent(c.notifier)

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

// Close cleans up any resources held by the container
func (c *Container) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	// Close database agent if it implements io.Closer
	if closer, ok := c.databaseAgent.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close database agent: %w", err))
		}
	}

	// Close notifier if it implements io.Closer
	if closer, ok := c.notifier.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close notifier: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing container: %v", errs)
	}

	return nil
}
