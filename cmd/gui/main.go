package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/config"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/container"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/gui"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
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

	// Start container
	if err := c.Start(ctx); err != nil {
		log.Fatalf("Error starting container: %v", err)
	}

	// Create and start GUI
	guiApp, err := gui.NewApp(c)
	if err != nil {
		log.Fatalf("Error creating GUI app: %v", err)
	}
	go guiApp.Start(ctx)

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, initiating shutdown", sig)
		cancel()
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Shutdown gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := guiApp.Stop(shutdownCtx); err != nil {
		log.Printf("Error stopping GUI application: %v", err)
	}

	if err := c.Stop(shutdownCtx); err != nil {
		log.Printf("Error stopping container: %v", err)
	}
}
