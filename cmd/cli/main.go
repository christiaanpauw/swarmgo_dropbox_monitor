package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/core"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/report"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or error loading it: %v", err)
	}

	// Parse command line flags
	checkNow := flag.Bool("check-now", false, "Run the Dropbox check immediately")
	quickCheck := flag.Bool("quick", false, "Check only last 10 minutes of changes (for testing)")
	lastHour := flag.Bool("last-hour", false, "Check only last hour of changes")
	last24Hours := flag.Bool("last-24h", false, "Check changes in the last 24 hours")
	dbConnStr := flag.String("db", "", "PostgreSQL connection string (optional, will use environment variables if not provided)")
	flag.Parse()

	fmt.Println("\nStarting Dropbox Monitor...")

	// Initialize the monitor
	if *dbConnStr == "" {
		*dbConnStr = os.Getenv("DROPBOX_MONITOR_DB")
	}
	monitor, err := core.NewMonitor(*dbConnStr)
	if err != nil {
		log.Fatalf("Error initializing monitor: %v", err)
	}
	defer monitor.Close()

	log.Println("✅ Successfully connected to Dropbox and Database!")

	if *checkNow {
		// Run ad-hoc check
		log.Println("Running ad-hoc Dropbox check...")
		runCheck(monitor, *quickCheck, *lastHour, *last24Hours)
	} else {
		// Set up scheduler
		c := cron.New()
		_, err := c.AddFunc("@midnight", func() { runCheck(monitor, false, false, true) })
		if err != nil {
			log.Fatalf("Error setting up scheduler: %v", err)
		}

		// Start scheduler
		c.Start()
		log.Printf("Scheduler started. Will check Dropbox at midnight.")

		// Keep the program running
		select {}
	}
}

func runCheck(monitor *core.Monitor, quickCheck bool, lastHour bool, last24Hours bool) {
	var changes []string
	var err error
	var since time.Time
	var period string

	if quickCheck {
		// Get changes from last 10 minutes
		since = time.Now().Add(-10 * time.Minute)
		period = "10min"
		log.Printf("Checking for changes in the last 10 minutes (since %v)...", since.Format(time.RFC3339))
		changes, err = monitor.DropboxClient.GetChangesLast10Minutes()
	} else if lastHour {
		// Get changes from last hour
		since = time.Now().Add(-1 * time.Hour)
		period = "hour"
		log.Printf("Checking for changes in the last hour (since %v)...", since.Format(time.RFC3339))
		changes, err = monitor.DropboxClient.GetChanges(since)
	} else if last24Hours {
		// Get changes from last 24 hours
		since = time.Now().Add(-24 * time.Hour)
		period = "day"
		log.Printf("Checking for changes in the last 24 hours (since %v)...", since.Format(time.RFC3339))
		changes, err = monitor.DropboxClient.GetChanges(since)
	} else {
		// Get changes from last 24 hours by default
		since = time.Now().Add(-24 * time.Hour)
		period = "day"
		log.Printf("Checking for changes in the last 24 hours (since %v)...", since.Format(time.RFC3339))
		changes, err = monitor.DropboxClient.GetChanges(since)
	}

	if err != nil {
		log.Printf("Error getting changes: %v", err)
		return
	}

	// Store changes in the database
	for _, change := range changes {
		err := monitor.DBAgent.StoreFileChange(change, time.Now(), nil)
		if err != nil {
			log.Printf("Error storing file change in database: %v", err)
		}
	}

	// Generate and send narrative report
	if err := report.SendNarrativeReport(changes, period, since); err != nil {
		log.Printf("Error sending narrative report: %v", err)
	} else {
		log.Printf("✅ Narrative report sent successfully")
	}
}
