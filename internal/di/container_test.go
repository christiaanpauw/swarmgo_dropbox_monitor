package di

import (
	"context"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/config"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
)

func TestNewContainer(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		DatabasePath:       ":memory:",
		DropboxAccessToken: "test-token",
		PollInterval:      5 * time.Minute,
		MaxRetries:        3,
		RetryDelay:        30 * time.Second,
		EmailSMTPHost:     "localhost",
		EmailSMTPPort:     25,
		EmailFrom:         "test@localhost",
	}

	// Test container creation
	container, err := NewContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// Test initial state
	if container.State() != lifecycle.StateInitialized {
		t.Errorf("Expected container state to be Initialized, got %s", container.State())
	}

	// Test that all dependencies are initialized
	if container.GetAgentManager() == nil {
		t.Error("AgentManager was not initialized")
	}

	if container.GetNotifier() == nil {
		t.Error("Notifier was not initialized")
	}

	if container.GetContentAnalyzer() == nil {
		t.Error("ContentAnalyzer was not initialized")
	}

	// Test container lifecycle
	ctx := context.Background()

	// Test startup
	if err := container.Start(ctx); err != nil {
		t.Errorf("Failed to start container: %v", err)
	}

	if container.State() != lifecycle.StateRunning {
		t.Errorf("Expected container state to be Running, got %s", container.State())
	}

	// Test health check
	if err := container.Health(ctx); err != nil {
		t.Errorf("Container health check failed: %v", err)
	}

	// Test shutdown
	if err := container.Stop(ctx); err != nil {
		t.Errorf("Failed to stop container: %v", err)
	}

	if container.State() != lifecycle.StateStopped {
		t.Errorf("Expected container state to be Stopped, got %s", container.State())
	}
}

func TestNewContainerWithNilConfig(t *testing.T) {
	container, err := NewContainer(nil)
	if err == nil {
		t.Error("Expected error when creating container with nil config")
	}
	if container != nil {
		t.Error("Expected nil container when creating with nil config")
	}
}

func TestContainerLifecycleErrors(t *testing.T) {
	cfg := &config.Config{
		DatabasePath:       ":memory:",
		DropboxAccessToken: "test-token",
		PollInterval:      5 * time.Minute,
		MaxRetries:        3,
		RetryDelay:        30 * time.Second,
	}

	container, err := NewContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	ctx := context.Background()

	// Test stopping before starting
	if err := container.Stop(ctx); err == nil {
		t.Error("Expected error when stopping container before starting")
	}

	// Test health check before starting
	if err := container.Health(ctx); err == nil {
		t.Error("Expected error when checking health before starting")
	}

	// Start container
	if err := container.Start(ctx); err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Test starting again
	if err := container.Start(ctx); err == nil {
		t.Error("Expected error when starting container twice")
	}

	// Stop container
	if err := container.Stop(ctx); err != nil {
		t.Fatalf("Failed to stop container: %v", err)
	}

	// Test stopping again
	if err := container.Stop(ctx); err == nil {
		t.Error("Expected error when stopping container twice")
	}
}
