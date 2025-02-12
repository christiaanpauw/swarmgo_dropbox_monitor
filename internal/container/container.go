package container

import (
	"context"
	"fmt"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agent"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agents"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/analysis"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/config"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/core"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/db"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/interfaces"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/scheduler"
)

// Container represents the application container
type Container struct {
	*lifecycle.BaseComponent
	config        *config.Config
	dropboxClient interfaces.DropboxClient
	notifier      notify.Notifier
	reportingAgent agents.ReportingAgent
	scheduler     *scheduler.Scheduler
	agentManager  agents.AgentManager
}

// NewContainer creates a new container
func NewContainer(cfg *config.Config) (*Container, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create dropbox client
	dropboxClient, err := dropbox.NewDropboxClient(cfg.DropboxToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create dropbox client: %w", err)
	}

	return NewContainerWithClient(cfg, dropboxClient)
}

// NewContainerWithClient creates a new container with a provided Dropbox client
func NewContainerWithClient(cfg *config.Config, dropboxClient interfaces.DropboxClient) (*Container, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create notifier
	notifier := notify.NewEmailNotifier(cfg.EmailConfig)

	// Create content analyzer
	contentAnalyzer := analysis.NewContentAnalyzer()

	// Create database connection
	dbConn, err := db.NewDB(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// Create database agent
	dbAgent, err := db.NewDatabaseAgent(dbConn)
	if err != nil {
		return nil, fmt.Errorf("failed to create database agent: %w", err)
	}

	// Create state manager
	stateManager := core.NewStateManager(cfg.State.Path)

	// Create reporting agent
	reportingAgent, err := agents.NewReportingAgent(notifier)
	if err != nil {
		return nil, fmt.Errorf("failed to create reporting agent: %w", err)
	}

	// Create scheduler
	scheduler, err := scheduler.NewScheduler(dropboxClient, reportingAgent, cfg.PollInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	// Create agent manager dependencies
	agentDeps := agents.AgentManagerDeps{
		FileChangeAgent:  agents.NewFileChangeAgent(dropboxClient, stateManager, cfg.Monitoring.Path),
		ContentAnalyzer:  contentAnalyzer,
		DatabaseAgent:    dbAgent,
		ReportingAgent:   reportingAgent,
		Notifier:        notifier,
	}

	// Create agent manager
	agentManager := agents.NewAgentManager(agentDeps)

	// Create container
	container := &Container{
		BaseComponent: lifecycle.NewBaseComponent("Container"),
		config:        cfg,
		dropboxClient: dropboxClient,
		notifier:      notifier,
		reportingAgent: reportingAgent,
		scheduler:     scheduler,
		agentManager:  agentManager,
	}

	container.SetState(lifecycle.StateInitialized)
	return container, nil
}

// NewContainerWithMocks creates a new container with provided mock dependencies
func NewContainerWithMocks(cfg *config.Config, dropboxClient interfaces.DropboxClient, reportingAgent agents.ReportingAgent, fileChangeAgent agent.FileChangeAgent, databaseAgent agents.DatabaseAgent, scheduler *scheduler.Scheduler) (*Container, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create agent manager dependencies
	agentDeps := agents.AgentManagerDeps{
		FileChangeAgent: fileChangeAgent,
		ContentAnalyzer: analysis.NewContentAnalyzer(),
		DatabaseAgent:   databaseAgent,
		ReportingAgent:  reportingAgent,
		Notifier:       notify.NewEmailNotifier(cfg.EmailConfig),
	}

	// Create agent manager
	agentManager := agents.NewAgentManager(agentDeps)

	// Create container
	container := &Container{
		BaseComponent: lifecycle.NewBaseComponent("Container"),
		config:        cfg,
		dropboxClient: dropboxClient,
		reportingAgent: reportingAgent,
		scheduler:     scheduler,
		agentManager:  agentManager,
	}

	container.SetState(lifecycle.StateInitialized)
	return container, nil
}

// GetAgentManager returns the agent manager instance
func (c *Container) GetAgentManager() agents.AgentManager {
	return c.agentManager
}

// GetHealthChecker returns the health checker instance
func (c *Container) GetHealthChecker() *lifecycle.BaseComponent {
	return c.BaseComponent
}

// GetNotifier returns the notifier instance
func (c *Container) GetNotifier() notify.Notifier {
	return c.notifier
}

// Start starts all components in the container
func (c *Container) Start(ctx context.Context) error {
	if err := c.DefaultStart(ctx); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	if err := c.agentManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start agent manager: %w", err)
	}

	if err := c.scheduler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	return nil
}

// Stop stops all components in the container
func (c *Container) Stop(ctx context.Context) error {
	if err := c.DefaultStop(ctx); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	if err := c.scheduler.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop scheduler: %w", err)
	}

	if err := c.agentManager.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop agent manager: %w", err)
	}

	return nil
}

// Health checks the health of all components in the container
func (c *Container) Health(ctx context.Context) error {
	if err := c.DefaultHealth(ctx); err != nil {
		return err
	}

	if err := c.agentManager.Health(ctx); err != nil {
		return fmt.Errorf("agent manager health check failed: %w", err)
	}

	if err := c.scheduler.Health(ctx); err != nil {
		return fmt.Errorf("scheduler health check failed: %w", err)
	}

	return nil
}
