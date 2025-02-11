package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/config"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/di"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	godotenv.Load()

	// Parse command line flags
	checkNow := flag.Bool("check-now", false, "Run the Dropbox check immediately")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create DI container
	container, err := di.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}
	defer container.Close()

	// Get agent manager from container
	agentManager := container.GetAgentManager()

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		cancel()
	}()

	if *checkNow {
		// Run ad-hoc check
		fmt.Println("Running immediate check...")

		if err := agentManager.Execute(ctx); err != nil {
			log.Fatalf("Error executing agents: %v", err)
		}

		fmt.Println("Check completed successfully!")
		return
	}

	// Start continuous monitoring
	fmt.Println("Starting continuous monitoring...")
	if err := agentManager.Start(ctx); err != nil {
		log.Fatalf("Error starting agent manager: %v", err)
	}

	// Wait for context cancellation
	<-ctx.Done()
	fmt.Println("Shutdown complete")
}
