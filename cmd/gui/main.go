package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
)

func main() {
	myApp := app.New()
	window := myApp.NewWindow("Dropbox Monitor")

	// Create Dropbox client
	client, err := dropbox.NewDropboxClient()
	if err != nil {
		log.Fatalf("Error creating Dropbox client: %v", err)
	}

	// Create output text area
	output := widget.NewTextGrid()
	output.SetText("Welcome to Dropbox Monitor!\nClick 'Check Now' to check for file changes.")

	// Create Check Now button
	checkButton := widget.NewButton("Check Now", func() {
		output.SetText("Checking for changes...")
		go func() {
			changes, err := client.GetChangesLast24Hours()
			if err != nil {
				output.SetText(fmt.Sprintf("Error checking changes: %v", err))
				return
			}

			// Prepare the report
			var report string
			if len(changes) > 0 {
				report = fmt.Sprintf("Files changed in the last 24 hours (as of %s):\n\n", time.Now().Format("2006-01-02 15:04:05"))
				report += strings.Join(changes, "\n")
			} else {
				report = "No file changes detected in the last 24 hours."
			}

			// Update GUI
			output.SetText(report)

			// Send notification
			err = notify.Send(report)
			if err != nil {
				log.Printf("Error sending notification: %v", err)
				output.SetText(output.Text() + "\n\nError sending notification: " + err.Error())
			} else {
				output.SetText(output.Text() + "\n\nNotification sent successfully!")
			}
		}()
	})

	// Create layout
	content := container.NewVBox(
		widget.NewLabel("Dropbox Monitor"),
		checkButton,
		output,
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(600, 400))
	window.ShowAndRun()
}
