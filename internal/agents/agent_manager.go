package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/analysis"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	coremodels "github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting/models"
)

// Agent represents a generic agent interface
type Agent interface {
	Execute(ctx context.Context) error
}

// AgentManager manages the lifecycle of all agents
type AgentManager struct {
	FileChangeAgent  FileChangeAgent
	ContentAnalyzer  analysis.ContentAnalyzer
	DatabaseAgent    DatabaseAgent
	ReportingAgent   ReportingAgent
	Notifier         notify.Notifier

	pollInterval time.Duration
	maxRetries   int
	retryDelay   time.Duration

	mu        sync.RWMutex
	isRunning bool
	stopCh    chan struct{}
}

// AgentManagerConfig holds configuration for the agent manager
type AgentManagerConfig struct {
	PollInterval time.Duration
	MaxRetries   int
	RetryDelay   time.Duration
}

// AgentManagerDeps holds dependencies for the agent manager
type AgentManagerDeps struct {
	FileChangeAgent  FileChangeAgent
	ContentAnalyzer  analysis.ContentAnalyzer
	DatabaseAgent    DatabaseAgent
	ReportingAgent   ReportingAgent
	Notifier         notify.Notifier
}

// NewAgentManager creates a new agent manager with the provided dependencies
func NewAgentManager(cfg AgentManagerConfig, deps AgentManagerDeps) *AgentManager {
	return &AgentManager{
		FileChangeAgent:  deps.FileChangeAgent,
		ContentAnalyzer:  deps.ContentAnalyzer,
		DatabaseAgent:    deps.DatabaseAgent,
		ReportingAgent:   deps.ReportingAgent,
		Notifier:         deps.Notifier,
		pollInterval:     cfg.PollInterval,
		maxRetries:       cfg.MaxRetries,
		retryDelay:       cfg.RetryDelay,
		stopCh:           make(chan struct{}),
	}
}

// Start starts all agents
func (am *AgentManager) Start(ctx context.Context) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.isRunning {
		return fmt.Errorf("agent manager already running")
	}

	// Validate dependencies
	if err := am.validateDependencies(); err != nil {
		return fmt.Errorf("validate dependencies: %w", err)
	}

	// Start reporting agent
	if err := am.ReportingAgent.Start(ctx); err != nil {
		return fmt.Errorf("start reporting agent: %w", err)
	}

	// Verify reporting agent is running
	if am.ReportingAgent.State() != lifecycle.StateRunning {
		return fmt.Errorf("reporting agent failed to start")
	}

	// Start polling loop
	go am.poll(ctx)

	am.isRunning = true
	return nil
}

// Stop stops all agents
func (am *AgentManager) Stop(ctx context.Context) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if !am.isRunning {
		return nil
	}

	// Signal polling loop to stop
	close(am.stopCh)

	// Stop reporting agent
	if err := am.ReportingAgent.Stop(ctx); err != nil {
		return fmt.Errorf("stop reporting agent: %w", err)
	}

	// Close database connection
	if am.DatabaseAgent != nil {
		if err := am.DatabaseAgent.Close(); err != nil {
			return fmt.Errorf("close database: %w", err)
		}
	}

	am.isRunning = false
	return nil
}

// Execute executes all agents once
func (am *AgentManager) Execute(ctx context.Context) error {
	// Get file changes
	changes, err := am.FileChangeAgent.GetChanges(ctx)
	if err != nil {
		return fmt.Errorf("get file changes: %w", err)
	}

	// Process each change
	for _, change := range changes {
		// Get file content
		content, err := am.FileChangeAgent.GetFileContent(ctx, change.Path)
		if err != nil {
			return fmt.Errorf("get file content: %w", err)
		}

		// Create file content model
		fileContent := &coremodels.FileContent{
			Path:        change.Path,
			ContentType: change.Extension,
			Size:        int64(len(content)),
			IsBinary:    false, // TODO: Implement proper binary detection
			ContentHash: "", // TODO: Implement content hash
		}

		// Store file content
		if err := am.DatabaseAgent.StoreFileContent(ctx, fileContent); err != nil {
			return fmt.Errorf("store file content: %w", err)
		}

		// Convert to reporting model
		reportChanges := []models.FileChange{
			{
				Path:      change.Path,
				Extension: change.Extension,
				Directory: change.Directory,
				ModTime:   change.ModTime,
			},
		}

		// Check reporting agent state
		if am.ReportingAgent.State() != lifecycle.StateRunning {
			return fmt.Errorf("reporting agent not running")
		}

		// Generate report with retries
		for attempt := 1; attempt <= am.maxRetries; attempt++ {
			err = am.ReportingAgent.GenerateReport(ctx, reportChanges)
			if err == nil {
				break
			}
			if attempt < am.maxRetries {
				time.Sleep(am.retryDelay)
			}
		}
		if err != nil {
			return fmt.Errorf("generate report: %w", err)
		}
	}

	return nil
}

// poll continuously executes agents at the configured interval
func (am *AgentManager) poll(ctx context.Context) {
	ticker := time.NewTicker(am.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-am.stopCh:
			return
		case <-ticker.C:
			if err := am.Execute(ctx); err != nil {
				// TODO: Add proper error handling/logging
				fmt.Printf("Error executing agents: %v\n", err)
			}
		}
	}
}

// validateDependencies ensures all required dependencies are set
func (am *AgentManager) validateDependencies() error {
	if am.FileChangeAgent == nil {
		return fmt.Errorf("file change agent is required")
	}
	if am.ContentAnalyzer == nil {
		return fmt.Errorf("content analyzer is required")
	}
	if am.DatabaseAgent == nil {
		return fmt.Errorf("database agent is required")
	}
	if am.ReportingAgent == nil {
		return fmt.Errorf("reporting agent is required")
	}
	if am.Notifier == nil {
		return fmt.Errorf("notifier is required")
	}
	return nil
}

// IsRunning returns true if the agent manager is running
func (am *AgentManager) IsRunning() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.isRunning
}
