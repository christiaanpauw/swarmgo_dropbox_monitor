package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agents"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/analysis"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	godotenv.Load()

	// Parse command line flags
	checkNow := flag.Bool("check-now", false, "Run the Dropbox check immediately")
	token := flag.String("token", "", "Dropbox access token")
	flag.Parse()

	if *token == "" {
		tokenValue := os.Getenv("DROPBOX_ACCESS_TOKEN")
		if tokenValue == "" {
			log.Fatal("DROPBOX_ACCESS_TOKEN environment variable is required")
		}
		*token = tokenValue
	}

	fmt.Println("\nStarting Dropbox Monitor...")

	// Initialize database
	dbPath := os.Getenv("DROPBOX_MONITOR_DB")
	if dbPath == "" {
		if err := os.MkdirAll("data", 0755); err != nil {
			log.Fatalf("Error creating data directory: %v", err)
		}
		dbPath = filepath.Join("data", "dropbox_monitor.db")
	}

	// Create dependencies
	fileChangeAgent, err := agents.NewFileChangeAgent(*token)
	if err != nil {
		log.Fatalf("Failed to create file change agent: %v", err)
	}

	databaseAgent, err := agents.NewDatabaseAgent()
	if err != nil {
		log.Fatalf("Failed to create database agent: %v", err)
	}

	contentAnalyzer := analysis.NewContentAnalyzer()
	notifier := notify.NewNotifier()
	reportingAgent := agents.NewReportingAgent(notifier)

	// Create agent manager config
	cfg := agents.AgentManagerConfig{
		PollInterval: 5 * time.Minute,
		MaxRetries:   3,
		RetryDelay:   time.Second,
	}

	// Create agent manager dependencies
	deps := agents.AgentManagerDeps{
		FileChangeAgent:  fileChangeAgent,
		ContentAnalyzer:  contentAnalyzer,
		DatabaseAgent:    databaseAgent,
		ReportingAgent:   reportingAgent,
		Notifier:        notifier,
	}

	// Create and start agent manager
	agentManager := agents.NewAgentManager(cfg, deps)

	if *checkNow {
		// Run ad-hoc check
		fmt.Println("Running immediate check...")

		ctx := context.Background()
		changes, err := fileChangeAgent.GetChanges(ctx)
		if err != nil {
			log.Fatalf("Error getting changes: %v", err)
		}
		fmt.Printf("Found %d changes\n", len(changes))

		// Process changes through the agents
		if err := agentManager.Execute(ctx); err != nil {
			log.Fatalf("Error executing agents: %v", err)
		}

		fmt.Println("Check completed successfully!")

		// Close database connection
		if err := databaseAgent.Close(); err != nil {
			log.Printf("Warning: error closing database: %v", err)
		}

	} else {
		ctx := context.Background()
		if err := agentManager.Start(ctx); err != nil {
			log.Fatalf("Failed to start agent manager: %v", err)
		}

		fmt.Println("Dropbox Monitor is running. Press Ctrl+C to stop.")

		// Wait for interrupt signal
		select {}
	}
}
