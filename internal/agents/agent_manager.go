package agents

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agent"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/analysis"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
)

// AgentManagerDeps holds dependencies for the agent manager
type AgentManagerDeps struct {
	FileChangeAgent  agent.FileChangeAgent
	ContentAnalyzer  analysis.ContentAnalyzer
	DatabaseAgent    agent.DatabaseAgent
	ReportingAgent   agent.ReportingAgent
	Notifier         notify.Notifier
}

// AgentManagerConfig holds configuration for the agent manager
type AgentManagerConfig struct {
}

// AgentManager defines the interface for agent coordination
type AgentManager interface {
	lifecycle.Component
	Initialize(ctx context.Context) error
	GetFileChangeAgent() agent.FileChangeAgent
}

// AgentManagerImpl implements the AgentManager interface
type AgentManagerImpl struct {
	*lifecycle.BaseComponent
	deps   AgentManagerDeps
	config AgentManagerConfig
	stopCh chan struct{}
	mu     sync.RWMutex
}

// NewAgentManager creates a new agent manager
func NewAgentManager(deps AgentManagerDeps) AgentManager {
	am := &AgentManagerImpl{
		BaseComponent: lifecycle.NewBaseComponent("AgentManager"),
		deps:         deps,
		stopCh:       make(chan struct{}),
	}
	am.SetState(lifecycle.StateInitialized)
	return am
}

// Start starts all agents
func (am *AgentManagerImpl) Start(ctx context.Context) error {
	if err := am.DefaultStart(ctx); err != nil {
		return err
	}

	log.Printf("🚀 Starting AgentManager...")

	// Check that all agents are initialized
	if am.deps.FileChangeAgent.State() != lifecycle.StateInitialized {
		return fmt.Errorf("file change agent not initialized")
	}
	if am.deps.DatabaseAgent.State() != lifecycle.StateInitialized {
		return fmt.Errorf("database agent not initialized")
	}
	if am.deps.ReportingAgent.State() != lifecycle.StateInitialized {
		return fmt.Errorf("reporting agent not initialized")
	}

	// Start file change monitoring
	if err := am.deps.FileChangeAgent.Start(ctx); err != nil {
		am.SetState(lifecycle.StateFailed)
		return fmt.Errorf("failed to start file change agent: %w", err)
	}

	// Start database agent
	if err := am.deps.DatabaseAgent.Start(ctx); err != nil {
		am.SetState(lifecycle.StateFailed)
		return fmt.Errorf("failed to start database agent: %w", err)
	}

	// Start reporting agent
	if err := am.deps.ReportingAgent.Start(ctx); err != nil {
		am.SetState(lifecycle.StateFailed)
		return fmt.Errorf("failed to start reporting agent: %w", err)
	}

	// Check that all agents are running
	if am.deps.FileChangeAgent.State() != lifecycle.StateRunning {
		am.SetState(lifecycle.StateFailed)
		return fmt.Errorf("file change agent failed to start")
	}
	if am.deps.DatabaseAgent.State() != lifecycle.StateRunning {
		am.SetState(lifecycle.StateFailed)
		return fmt.Errorf("database agent failed to start")
	}
	if am.deps.ReportingAgent.State() != lifecycle.StateRunning {
		am.SetState(lifecycle.StateFailed)
		return fmt.Errorf("reporting agent failed to start")
	}

	// Set state to running after all agents are started and verified
	am.SetState(lifecycle.StateRunning)
	return nil
}

// Stop stops all agents
func (am *AgentManagerImpl) Stop(ctx context.Context) error {
	log.Printf("🛑 Stopping AgentManager...")

	if err := am.DefaultStop(ctx); err != nil {
		return err
	}

	am.SetState(lifecycle.StateStopping)

	// Stop file change monitoring
	if err := am.deps.FileChangeAgent.Stop(ctx); err != nil {
		am.SetState(lifecycle.StateFailed)
		return fmt.Errorf("failed to stop file change agent: %w", err)
	}

	// Stop database agent
	if err := am.deps.DatabaseAgent.Stop(ctx); err != nil {
		am.SetState(lifecycle.StateFailed)
		return fmt.Errorf("failed to stop database agent: %w", err)
	}

	// Stop reporting agent
	if err := am.deps.ReportingAgent.Stop(ctx); err != nil {
		am.SetState(lifecycle.StateFailed)
		return fmt.Errorf("failed to stop reporting agent: %w", err)
	}

	am.SetState(lifecycle.StateStopped)
	return nil
}

// Health checks the health of all agents
func (am *AgentManagerImpl) Health(ctx context.Context) error {
	if err := am.DefaultHealth(ctx); err != nil {
		return err
	}

	// Check that all agents are running
	if am.deps.FileChangeAgent.State() != lifecycle.StateRunning {
		return fmt.Errorf("file change agent not running")
	}
	if am.deps.DatabaseAgent.State() != lifecycle.StateRunning {
		return fmt.Errorf("database agent not running")
	}
	if am.deps.ReportingAgent.State() != lifecycle.StateRunning {
		return fmt.Errorf("reporting agent not running")
	}

	// Check file change agent health
	if err := am.deps.FileChangeAgent.Health(ctx); err != nil {
		return fmt.Errorf("file change agent unhealthy: %w", err)
	}

	// Check database agent health
	if err := am.deps.DatabaseAgent.Health(ctx); err != nil {
		return fmt.Errorf("database agent unhealthy: %w", err)
	}

	// Check reporting agent health
	if err := am.deps.ReportingAgent.Health(ctx); err != nil {
		return fmt.Errorf("reporting agent unhealthy: %w", err)
	}

	return nil
}

// Initialize performs any necessary initialization before starting the agents
func (am *AgentManagerImpl) Initialize(ctx context.Context) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Validate required dependencies
	if am.deps.FileChangeAgent == nil {
		return fmt.Errorf("FileChangeAgent is required")
	}
	if am.deps.DatabaseAgent == nil {
		return fmt.Errorf("DatabaseAgent is required")
	}
	if am.deps.ReportingAgent == nil {
		return fmt.Errorf("ReportingAgent is required")
	}

	// Set state to initialized after validation
	am.SetState(lifecycle.StateInitialized)
	return nil
}

// GetFileChangeAgent returns the file change agent
func (am *AgentManagerImpl) GetFileChangeAgent() agent.FileChangeAgent {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.deps.FileChangeAgent
}
