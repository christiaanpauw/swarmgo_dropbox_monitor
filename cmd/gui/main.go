package main

import (
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
	monitor, err := core.NewMonitor(dbConnStr)
	if err != nil {
		log.Fatalf("Error initializing monitor: %v", err)
	}
	defer monitor.Close()

	// Create GUI application
	myApp := app.New()
	window := myApp.NewWindow("Dropbox Monitor")

	// Create status label
	statusLabel := widget.NewLabel("Status: Ready")

	// Create output text area
	output := widget.NewMultiLineEntry()
	output.Wrapping = fyne.TextWrapWord
	output.MultiLine = true
	output.SetPlaceHolder("Welcome to Dropbox Monitor!\nClick 'Check Now' to check for file changes.")

	// Create time window selection
	timeSelect := widget.NewSelect([]string{
		"Last 10 minutes",
		"Last hour",
		"Last 24 hours",
	}, nil)
	timeSelect.SetSelected("Last 24 hours")

	// Function to perform the check
	performCheck := func() {
		statusLabel.SetText("Status: Checking for changes...")
		output.SetText("")
		
		var changes []string
		var err error

		// Get changes based on selected time window
		switch timeSelect.Selected {
		case "Last 10 minutes":
			metadata, err := monitor.DropboxClient.GetChangesLast10Minutes()
			if err == nil {
				for _, m := range metadata {
					changes = append(changes, fmt.Sprintf("%s (modified at %s)", m.PathLower, m.ServerModified))
				}
			}
		case "Last hour":
			since := time.Now().Add(-1 * time.Hour)
			metadata, err := monitor.DropboxClient.GetChanges(since)
			if err == nil {
				for _, m := range metadata {
					changes = append(changes, fmt.Sprintf("%s (modified at %s)", m.PathLower, m.ServerModified))
				}
			}
		default: // Last 24 hours
			metadata, err := monitor.DropboxClient.GetChangesLast24Hours()
			if err == nil {
				for _, m := range metadata {
					changes = append(changes, fmt.Sprintf("%s (modified at %s)", m.PathLower, m.ServerModified))
				}
			}
		}

		if err != nil {
			dialog.ShowError(fmt.Errorf("Error checking changes: %v", err), window)
			statusLabel.SetText("Status: Error occurred")
			return
		}

		// Generate report
		reportData := report.Generate(changes)
		reportText := reportData.FormatReport()

		// Update GUI
		output.SetText(reportText)
		statusLabel.SetText(fmt.Sprintf("Status: Found %d changes", len(changes)))

		// Send email report
		if err := reportData.SendReport(); err != nil {
			dialog.ShowError(fmt.Errorf("Error sending email report: %v", err), window)
		} else {
			dialog.ShowInformation("Success", "Email report sent successfully!", window)
		}
	}

	// Create Check Now button
	checkButton := widget.NewButton("Check Now", func() {
		go performCheck()
	})

	// Create scheduler options
	schedulerOptions := []string{"Manual", "Every 5 minutes", "Every 15 minutes", "Every 30 minutes", "Every hour", "Daily at midnight"}
	schedulerSelect := widget.NewSelect(schedulerOptions, nil)
	schedulerSelect.SetSelected("Manual")

	var refreshTimer *time.Timer
	var cronScheduler *cron.Cron

	schedulerSelect.OnChanged = func(interval string) {
		// Stop existing timers/schedulers
		if refreshTimer != nil {
			refreshTimer.Stop()
		}
		if cronScheduler != nil {
			cronScheduler.Stop()
		}

		if interval == "Manual" {
			statusLabel.SetText("Status: Manual mode")
			return
		}

		if interval == "Daily at midnight" {
			// Use cron scheduler for midnight runs
			cronScheduler = cron.New()
			_, err := cronScheduler.AddFunc("@midnight", func() {
				performCheck()
			})
			if err != nil {
				dialog.ShowError(fmt.Errorf("Error setting up scheduler: %v", err), window)
				return
			}
			cronScheduler.Start()
			statusLabel.SetText("Status: Scheduled to run daily at midnight")
			return
		}

		// Parse interval and create timer for other intervals
		var duration time.Duration
		switch interval {
		case "Every 5 minutes":
			duration = 5 * time.Minute
		case "Every 15 minutes":
			duration = 15 * time.Minute
		case "Every 30 minutes":
			duration = 30 * time.Minute
		case "Every hour":
			duration = time.Hour
		}

		refreshTimer = time.NewTimer(duration)
		go func() {
			for range refreshTimer.C {
				performCheck()
				refreshTimer.Reset(duration)
			}
		}()
		statusLabel.SetText(fmt.Sprintf("Status: Auto-refresh set to %s", interval))
	}

	// Create controls container
	controls := container.NewVBox(
		widget.NewLabel("Time Window:"),
		timeSelect,
		widget.NewLabel("Schedule:"),
		schedulerSelect,
		checkButton,
		statusLabel,
	)

	// Create main content with scroll
	scrollContainer := container.NewScroll(output)
	scrollContainer.Resize(fyne.NewSize(800, 400))

	content := container.NewBorder(
		controls, nil, nil, nil,
		scrollContainer,
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(800, 600))
	window.ShowAndRun()
}
