package scheduler

import (
	"log"
	"time"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/report"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
)

// Start initializes the daily scheduler
func Start() error {
	go func() {
		for {
			now := time.Now()
			nextRun := now.Truncate(24 * time.Hour).Add(24 * time.Hour) // Midnight next day
			waitTime := nextRun.Sub(now)

			log.Printf("Next check scheduled in %v", waitTime)
			time.Sleep(waitTime) // Wait until the next day

			log.Println("Checking Dropbox for changes...")

			// Step 1: Check Dropbox
			changes, err := dropbox.CheckForChanges()
			if err != nil {
				log.Printf("Error checking Dropbox: %v", err)
				continue
			}

			// Step 2: Generate report
			reportData := report.Generate(changes)

			// Step 3: Send notification
			err = notify.Send(reportData)
			if err != nil {
				log.Printf("Error sending notification: %v", err)
			} else {
				log.Println("Report sent successfully!")
			}
		}
	}()

	return nil
}

