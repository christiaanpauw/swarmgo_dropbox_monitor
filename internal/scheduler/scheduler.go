package scheduler

import (
	"context"
	"fmt"
	"log"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/report"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	client   *dropbox.DropboxClient
	notifier notify.Notifier
	cron     *cron.Cron
}

// NewScheduler creates a new scheduler instance
func NewScheduler(client *dropbox.DropboxClient, notifier notify.Notifier) *Scheduler {
	return &Scheduler{
		client:   client,
		notifier: notifier,
		cron:     cron.New(),
	}
}

// Start begins the scheduled tasks
func (s *Scheduler) Start() error {
	// Schedule daily check at midnight
	_, err := s.cron.AddFunc("@midnight", func() {
		ctx := context.Background()
		changes, err := s.client.GetChangesLast24Hours(ctx)
		if err != nil {
			log.Printf("Error getting changes: %v", err)
			return
		}

		// Convert changes to strings
		var changeStrings []string
		for _, change := range changes {
			changeStrings = append(changeStrings, fmt.Sprintf("%s (modified at %s)", change.PathLower, change.ServerModified))
		}

		// Generate and send report
		reportData := report.Generate(changeStrings)
		if err := reportData.SendReport(); err != nil {
			log.Printf("Error sending email report: %v", err)
		}
	})

	if err != nil {
		return fmt.Errorf("error adding cron job: %v", err)
	}

	s.cron.Start()
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
}
