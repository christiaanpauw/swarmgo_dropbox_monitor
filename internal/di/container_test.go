package di

import (
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/config"
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

	// Test container cleanup
	if err := container.Close(); err != nil {
		t.Errorf("Failed to close container: %v", err)
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
