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
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/container"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", ".env", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Create container
	c, err := container.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Error creating container: %v", err)
	}

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
		cancel()
	}()

	// Start container
	if err := c.Start(ctx); err != nil {
		log.Fatalf("Error starting container: %v", err)
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Create new context for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	// Stop container
	if err := c.Stop(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
