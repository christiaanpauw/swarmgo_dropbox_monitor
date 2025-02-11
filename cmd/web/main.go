package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/core"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/report"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

type ChangeResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Changes []string `json:"changes"`
}

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

	// Create cron scheduler
	c := cron.New()
	_, err = c.AddFunc("@midnight", func() {
		ctx := context.Background()
		changes, err := monitor.DropboxClient.GetChangesLast24Hours(ctx)
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
		log.Printf("Warning: Failed to set up scheduler: %v", err)
	} else {
		c.Start()
		log.Println("Scheduler started. Will check Dropbox at midnight.")
	}

	// Set up Gin router
	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLGlob("templates/*")

	// Serve static files
	router.Static("/static", "./static")

	// Routes
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Dropbox Monitor",
		})
	})

	router.GET("/api/changes", func(c *gin.Context) {
		timeWindow := c.Query("window")
		log.Printf("🔍 Received request to check changes for window: %s", timeWindow)
		
		var changes []string
		var err error
		var metadata []*models.FileMetadata
		ctx := context.Background()

		switch timeWindow {
		case "10min":
			log.Printf("Checking last 10 minutes...")
			metadata, err = monitor.DropboxClient.GetChangesLast10Minutes(ctx)
		case "1hour":
			log.Printf("Checking last hour...")
			ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
			defer cancel()
			metadata, err = monitor.DropboxClient.GetChanges(ctx)
		default: // 24hours
			log.Printf("Checking last 24 hours...")
			metadata, err = monitor.DropboxClient.GetChangesLast24Hours(ctx)
		}

		if err != nil {
			log.Printf("❌ Error getting changes: %v", err)
			c.JSON(http.StatusInternalServerError, ChangeResponse{
				Success: false,
				Message: fmt.Sprintf("Error getting changes: %v", err),
				Changes: []string{},
			})
			return
		}

		// Convert metadata to strings and log each change
		for _, m := range metadata {
			change := fmt.Sprintf("%s (modified at %s)", m.PathLower, m.ServerModified)
			log.Printf("✨ Found change: %s", change)
			changes = append(changes, change)
		}

		// Generate report
		log.Printf("📝 Generating report for %d changes...", len(changes))
		reportData := report.Generate(changes)
		reportText := reportData.FormatReport()

		// Send email if requested
		if c.Query("email") == "true" {
			log.Printf("📧 Attempting to send email report...")
			if err := reportData.SendReport(); err != nil {
				log.Printf("❌ Error sending email: %v", err)
				c.JSON(http.StatusInternalServerError, ChangeResponse{
					Success: false,
					Message: fmt.Sprintf("Error sending email report: %v", err),
					Changes: []string{},
				})
				return
			}
			log.Printf("✅ Email report sent successfully")
		}

		// Ensure changes is never null
		if changes == nil {
			changes = []string{}
		}

		log.Printf("✅ Returning %d changes", len(changes))
		c.JSON(http.StatusOK, ChangeResponse{
			Success: true,
			Message: reportText,
			Changes: changes,
		})
	})

	// Start server
	port := os.Getenv("WEB_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting web server on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}