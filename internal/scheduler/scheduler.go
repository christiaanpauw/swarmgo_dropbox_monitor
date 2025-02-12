package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agents"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting/models"
)

// Scheduler manages periodic execution of file change detection and reporting
type Scheduler struct {
	*lifecycle.BaseComponent
	client   *dropbox.DropboxClient
	notifier notify.Notifier
	reporter reporting.Reporter
	interval time.Duration
	stopCh   chan struct{}
}

// NewScheduler creates a new scheduler instance
func NewScheduler(client *dropbox.DropboxClient, notifier notify.Notifier, interval time.Duration) *Scheduler {
	return &Scheduler{
		BaseComponent: lifecycle.NewBaseComponent("Scheduler"),
		client:        client,
		notifier:      notifier,
		reporter:      reporting.NewReporter(notifier),
		interval:      interval,
		stopCh:        make(chan struct{}),
	}
}

// Start begins periodic execution
func (s *Scheduler) Start(ctx context.Context) error {
	if err := s.DefaultStart(ctx); err != nil {
		return err
	}

	go s.run(ctx)
	return nil
}

// Stop halts periodic execution
func (s *Scheduler) Stop(ctx context.Context) error {
	close(s.stopCh)
	return s.DefaultStop(ctx)
}

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
				log.Printf("Error executing scheduled task: %v", err)
			}
		}
	}
}

func (s *Scheduler) execute(ctx context.Context) error {
	// Get file changes
	changes, err := s.client.GetChangesLast24Hours(ctx)
	if err != nil {
		return err
	}

	if len(changes) == 0 {
		return nil
	}

	// Convert changes to reporting model
	reportChanges := make([]models.FileChange, 0, len(changes))
	for _, change := range changes {
		reportChanges = append(reportChanges, models.FileChange{
			Path:      change.PathLower,
			Extension: "",
			Size:      0,
			Modified:  change.ServerModified,
		})
	}

	// Generate and send report
	report, err := s.reporter.GenerateReport(ctx, reportChanges)
	if err != nil {
		return err
	}

	if err := s.reporter.SendReport(ctx, report); err != nil {
		return err
	}

	return nil
}
