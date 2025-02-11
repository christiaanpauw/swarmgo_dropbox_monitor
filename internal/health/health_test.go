package health

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHealthChecker_RegisterComponent(t *testing.T) {
	checker := NewHealthChecker(time.Second)
	check := func(ctx context.Context) error { return nil }

	// Register component
	checker.RegisterComponent("test", check)

	// Verify component was registered
	status := checker.GetStatus()
	assert.Contains(t, status, "test")
	assert.Equal(t, StatusStarting, status["test"])
}

func TestHealthChecker_GetStatus(t *testing.T) {
	checker := NewHealthChecker(time.Second)

	// Create test checks
	healthyCheck := func(ctx context.Context) error { return nil }
	unhealthyCheck := func(ctx context.Context) error { return errors.New("test error") }

	// Register components
	checker.RegisterComponent("healthy", healthyCheck)
	checker.RegisterComponent("unhealthy", unhealthyCheck)

	// Run health checks
	ctx := context.Background()
	checker.checkAll(ctx)

	// Verify status
	status := checker.GetStatus()
	assert.Equal(t, StatusHealthy, status["healthy"])
	assert.Equal(t, StatusUnhealthy, status["unhealthy"])

	// Verify errors
	errs := checker.GetErrors()
	assert.NoError(t, errs["healthy"])
	assert.Error(t, errs["unhealthy"])
	assert.Equal(t, "test error", errs["unhealthy"].Error())
}

func TestHealthChecker_Start(t *testing.T) {
	checker := NewHealthChecker(100 * time.Millisecond)
	checkCount := 0

	// Create test check that counts calls
	check := func(ctx context.Context) error {
		checkCount++
		return nil
	}

	// Register component
	checker.RegisterComponent("test", check)

	// Start health checker
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	// Start the checker
	checker.Start(ctx)

	// Wait for context to be done
	<-ctx.Done()

	// Stop the checker
	checker.Stop()

	// Verify check was called multiple times
	assert.Greater(t, checkCount, 1)
}
