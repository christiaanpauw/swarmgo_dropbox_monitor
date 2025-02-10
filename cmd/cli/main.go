package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agents"
	"github.com/joho/godotenv"
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
	flag.Parse()

	fmt.Println("\nStarting Dropbox Monitor...")

	// Initialize database
	dbPath := os.Getenv("DROPBOX_MONITOR_DB")
	if dbPath == "" {
		// Create data directory if it doesn't exist
		if err := os.MkdirAll("data", 0755); err != nil {
			log.Fatalf("Error creating data directory: %v", err)
		}
		dbPath = filepath.Join("data", "dropbox_monitor.db")
	}
	dropboxToken := os.Getenv("DROPBOX_ACCESS_TOKEN")
	if dropboxToken == "" {
		log.Fatal("DROPBOX_ACCESS_TOKEN environment variable is required")
	}

	manager, err := agents.NewAgentManager("file:"+dbPath, dropboxToken)
	if err != nil {
		log.Fatalf("Error initializing agent manager: %v", err)
	}

	if *checkNow {
		// Run ad-hoc check
		log.Println("Running ad-hoc Dropbox check...")
		timeWindow := "24h"
		if *quickCheck {
			timeWindow = "10m"
		} else if *lastHour {
			timeWindow = "1h"
		}

		ctx := context.Background()
		report, err := manager.ProcessChanges(ctx, timeWindow)
		if err != nil {
			log.Fatalf("Error processing changes: %v", err)
		}

		// Print report summary
		fmt.Printf("\nReport Summary:\n")
		fmt.Printf("Generated at: %v\n", report.GeneratedAt)
		fmt.Printf("Total Files: %d\n", report.FileCount)
		fmt.Printf("\nSummary:\n%s\n", report.Summary)
		
		if len(report.TopKeywords) > 0 {
			fmt.Printf("\nTop Keywords:\n")
			for kw, count := range report.TopKeywords {
				fmt.Printf("- %s (%d)\n", kw, count)
			}
		}

		if len(report.TopTopics) > 0 {
			fmt.Printf("\nCommon Topics:\n")
			for topic, count := range report.TopTopics {
				fmt.Printf("- %s (%d)\n", topic, count)
			}
		}
	} else {
		log.Fatal("Please specify --check-now to run an immediate check")
	}
}
