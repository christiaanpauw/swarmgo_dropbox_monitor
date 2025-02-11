package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/core"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/report"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or error loading it: %v", err)
	}

	// Initialize the monitor
	dbConnStr := os.Getenv("DROPBOX_MONITOR_DB")
	dropboxToken := os.Getenv("DROPBOX_ACCESS_TOKEN")
	monitor, err := core.NewMonitor(dbConnStr, dropboxToken)
	if err != nil {
		log.Fatalf("Error initializing monitor: %v", err)
	}
	defer monitor.Close()

	// Create and set up application
	myApp := app.New()
	window := myApp.NewWindow("Dropbox Monitor")

	// Create status label
	statusLabel := widget.NewLabel("Status: Ready")

	// Create output text area
	output := widget.NewTextGrid()

	// Create time window selector
	timeSelect := widget.NewSelect([]string{"Last 10 minutes", "Last hour", "Last 24 hours"}, nil)
	timeSelect.SetSelected("Last 24 hours")

	// Function to perform the check
	performCheck := func(ctx context.Context) {
		statusLabel.SetText("Status: Checking for changes...")
		output.SetText("")
		
		var changes []*models.FileMetadata
		var err error

		// Get changes based on selected time window
		switch timeSelect.Selected {
		case "Last 10 minutes":
			changes, err = monitor.DropboxClient.GetChangesLast10Minutes(ctx)
		case "Last hour":
			changes, err = monitor.DropboxClient.GetChanges(ctx)
		default: // Last 24 hours
			changes, err = monitor.DropboxClient.GetChangesLast24Hours(ctx)
		}

		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Status: Error - %v", err))
			return
		}

		// Convert changes to strings for report
		var changeStrings []string
		for _, change := range changes {
			changeStrings = append(changeStrings, fmt.Sprintf("%s (modified at %s)", change.PathLower, change.ServerModified))
		}

		// Generate report
		reportData := report.Generate(changeStrings)
		output.SetText(reportData.FormatReport())
		statusLabel.SetText(fmt.Sprintf("Status: Found %d changes", len(changes)))
	}

	// Create Check Now button
	checkButton := widget.NewButton("Check Now", func() {
		ctx := context.Background()
		go performCheck(ctx)
	})

	// Create scheduler options
	var cronScheduler *cron.Cron
	var refreshTimer *time.Timer

	schedulerOptions := []string{"Manual", "Every 10 minutes", "Every hour", "At midnight"}
	schedulerSelect := widget.NewSelect(schedulerOptions, func(selected string) {
		// Stop existing schedulers
		if cronScheduler != nil {
			cronScheduler.Stop()
			cronScheduler = nil
		}
		if refreshTimer != nil {
			refreshTimer.Stop()
			refreshTimer = nil
		}

		// Set up new scheduler based on selection
		switch selected {
		case "At midnight":
			// Use cron scheduler for midnight runs
			cronScheduler = cron.New()
			_, err := cronScheduler.AddFunc("@midnight", func() {
				ctx := context.Background()
				performCheck(ctx)
			})
			if err != nil {
				dialog.ShowError(fmt.Errorf("Error setting up scheduler: %v", err), window)
				return
			}
			cronScheduler.Start()
			statusLabel.SetText("Status: Scheduled for midnight runs")

		case "Every 10 minutes":
			setupPeriodicCheck(10*time.Minute, &refreshTimer, performCheck)
			statusLabel.SetText("Status: Checking every 10 minutes")

		case "Every hour":
			setupPeriodicCheck(time.Hour, &refreshTimer, performCheck)
			statusLabel.SetText("Status: Checking every hour")

		default:
			statusLabel.SetText("Status: Manual mode")
		}
	})
	schedulerSelect.SetSelected("Manual")

	// Layout
	content := container.NewVBox(
		widget.NewLabel("Select Time Window:"),
		timeSelect,
		widget.NewLabel("Scheduler:"),
		schedulerSelect,
		checkButton,
		statusLabel,
		output,
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(600, 400))
	window.ShowAndRun()
}

func setupPeriodicCheck(duration time.Duration, timer **time.Timer, check func(context.Context)) {
	*timer = time.NewTimer(duration)
	go func() {
		for range (*timer).C {
			ctx := context.Background()
			check(ctx)
			(*timer).Reset(duration)
		}
	}()
}
