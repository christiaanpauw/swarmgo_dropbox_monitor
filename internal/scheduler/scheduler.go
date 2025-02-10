package scheduler

import (
	"fmt"
	"log"
	"time"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/report"
	"database/sql"
)

type Scheduler struct {
	client *dropbox.DropboxClient
	db     *sql.DB
}

// NewScheduler creates a new scheduler instance
func NewScheduler(client *dropbox.DropboxClient, db *sql.DB) *Scheduler {
	return &Scheduler{
		client: client,
		db:     db,
	}
}

// Start initializes the daily scheduler
func (s *Scheduler) Start() error {
	go func() {
		for {
			now := time.Now()
			nextRun := now.Truncate(24 * time.Hour).Add(24 * time.Hour) // Midnight next day
			waitTime := nextRun.Sub(now)

			log.Printf("Next check scheduled in %v", waitTime)
			time.Sleep(waitTime) // Wait until the next day

			log.Println("Checking Dropbox for changes...")

			// Step 1: Check Dropbox
			changes, err := s.client.GetChangesLast24Hours()
			if err != nil {
				log.Printf("Error checking Dropbox: %v", err)
				continue
			}

			// Step 2: Generate report
			var changeStrings []string
			for _, change := range changes {
				changeStrings = append(changeStrings, fmt.Sprintf("%s (modified at %s)", change.PathLower, change.ServerModified))
			}

			reportData := report.Generate(changeStrings)

			// Step 3: Send notification (using email)
			err = s.sendEmailReport(reportData)
			if err != nil {
				log.Printf("Error sending notification: %v", err)
			} else {
				log.Println("Report sent successfully!")
			}
		}
	}()

	return nil
}

// sendEmailReport sends the report via email
func (s *Scheduler) sendEmailReport(reportData *report.Report) error {
	// Send the email using the report package's SendReport method
	return reportData.SendReport()
}
