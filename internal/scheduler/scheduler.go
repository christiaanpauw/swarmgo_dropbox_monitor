package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agents"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/interfaces"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// Scheduler manages periodic execution of file change detection and reporting
type Scheduler struct {
	*lifecycle.BaseComponent
	client        interfaces.DropboxClient
	reportingAgent agents.ReportingAgent
	interval      time.Duration
	stopCh        chan struct{}
}

// NewScheduler creates a new scheduler
func NewScheduler(client interfaces.DropboxClient, reportingAgent agents.ReportingAgent, interval time.Duration) (*Scheduler, error) {
	if client == nil {
		return nil, fmt.Errorf("client cannot be nil")
	}
	if reportingAgent == nil {
		return nil, fmt.Errorf("reporting agent cannot be nil")
	}
	if interval <= 0 {
		return nil, fmt.Errorf("interval must be greater than 0")
	}

	scheduler := &Scheduler{
		BaseComponent:  lifecycle.NewBaseComponent("Scheduler"),
		client:        client,
		reportingAgent: reportingAgent,
		interval:      interval,
		stopCh:        make(chan struct{}),
	}
	scheduler.SetState(lifecycle.StateInitialized)
	return scheduler, nil
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	if err := s.DefaultStart(ctx); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	go s.run(ctx)

	s.SetState(lifecycle.StateRunning)
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	close(s.stopCh)
	s.SetState(lifecycle.StateStopped)
	return nil
}

// Health checks the health of the scheduler
func (s *Scheduler) Health(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	// Check reporting agent health
	if err := s.reportingAgent.Health(ctx); err != nil {
		return fmt.Errorf("reporting agent unhealthy: %w", err)
	}

	return nil
}

// Initialize initializes the scheduler
func (s *Scheduler) Initialize(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	// Validate dependencies are ready
	if s.client == nil {
		return fmt.Errorf("dropbox client not initialized")
	}
	if s.reportingAgent == nil {
		return fmt.Errorf("reporting agent not initialized")
	}

	s.SetState(lifecycle.StateInitialized)
	return nil
}

// run executes the scheduler loop
func (s *Scheduler) run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.execute(ctx); err != nil {
				fmt.Printf("Error executing scheduled task: %v\n", err)
			}
		}
	}
}

// execute performs a single execution of the scheduler
func (s *Scheduler) execute(ctx context.Context) error {
	// Get file changes from Dropbox
	changes, err := s.client.GetChanges(ctx)
	if err != nil {
		return fmt.Errorf("failed to get file changes: %w", err)
	}

	if len(changes) == 0 {
		return nil // No changes to report
	}

	// Convert to models.FileChange
	fileChanges := make([]models.FileChange, len(changes))
	for i, change := range changes {
		fileChanges[i] = models.FileChange{
			Path:      change.Path,
			Size:      change.Size,
			Modified:  change.Modified,
			IsDeleted: change.IsDeleted,
		}
	}

	// Generate report
	if err := s.reportingAgent.GenerateReport(ctx, fileChanges); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	return nil
}
