package container

import (
	"context"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestContainer_New(t *testing.T) {
	cfg := &config.Config{
		DatabasePath:        "test.db",
		DropboxAccessToken: "test-token",
		PollInterval:      5 * time.Minute,
		MaxRetries:        3,
		RetryDelay:        time.Second,
		HealthCheckInterval: time.Second,
	}

	container := New(cfg)
	assert.NotNil(t, container)
	assert.NotNil(t, container.healthChecker)
	assert.Equal(t, cfg, container.config)
}

func TestContainer_Start(t *testing.T) {
	cfg := &config.Config{
		DatabasePath:        "test.db",
		DropboxAccessToken: "test-token",
		PollInterval:      5 * time.Minute,
		MaxRetries:        3,
		RetryDelay:        time.Second,
		HealthCheckInterval: time.Second,
	}

	container := New(cfg)
	assert.NotNil(t, container)

	// Start container
	ctx := context.Background()
	err := container.Start(ctx)
	assert.NoError(t, err)

	// Verify components are initialized
	assert.NotNil(t, container.GetAgentManager())
	assert.NotNil(t, container.GetHealthChecker())

	// Try starting again - should fail
	err = container.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already started")

	// Stop container
	err = container.Stop(ctx)
	assert.NoError(t, err)
}

func TestContainer_Stop(t *testing.T) {
	cfg := &config.Config{
		DatabasePath:        "test.db",
		DropboxAccessToken: "test-token",
		PollInterval:      5 * time.Minute,
		MaxRetries:        3,
		RetryDelay:        time.Second,
		HealthCheckInterval: time.Second,
	}

	container := New(cfg)
	assert.NotNil(t, container)

	// Start container
	ctx := context.Background()
	err := container.Start(ctx)
	assert.NoError(t, err)

	// Stop container
	err = container.Stop(ctx)
	assert.NoError(t, err)
}

func TestContainer_IsHealthy(t *testing.T) {
	cfg := &config.Config{
		DatabasePath:        "test.db",
		DropboxAccessToken: "test-token",
		PollInterval:      5 * time.Minute,
		MaxRetries:        3,
		RetryDelay:        time.Second,
		HealthCheckInterval: 100 * time.Millisecond,
	}

	container := New(cfg)
	assert.NotNil(t, container)

	// Start container
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	err := container.Start(ctx)
	assert.NoError(t, err)

	// Wait for health checks to run
	time.Sleep(150 * time.Millisecond)

	// Check health
	assert.True(t, container.IsHealthy())

	// Stop container
	err = container.Stop(ctx)
	assert.NoError(t, err)
}
