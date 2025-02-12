package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStateManager(t *testing.T) {
	// Create temp directory for test state file
	tmpDir, err := os.MkdirTemp("", "state_manager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	statePath := filepath.Join(tmpDir, "state.json")
	sm := NewStateManager(statePath)

	// Test Start
	t.Run("Start", func(t *testing.T) {
		ctx := context.Background()
		if err := sm.Start(ctx); err != nil {
			t.Errorf("Start() error = %v", err)
		}

		// Verify state directory was created
		if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
			t.Error("State directory was not created")
		}
	})

	// Test GetString/SetString
	t.Run("GetString/SetString", func(t *testing.T) {
		key := "test_key"
		value := "test_value"

		// Set value
		if err := sm.SetString(key, value); err != nil {
			t.Errorf("SetString() error = %v", err)
		}

		// Get value
		if got := sm.GetString(key); got != value {
			t.Errorf("GetString() = %v, want %v", got, value)
		}

		// Verify state was persisted
		if _, err := os.Stat(statePath); os.IsNotExist(err) {
			t.Error("State file was not created")
		}
	})

	// Test Health
	t.Run("Health", func(t *testing.T) {
		ctx := context.Background()
		if err := sm.Health(ctx); err != nil {
			t.Errorf("Health() error = %v", err)
		}
	})

	// Test Stop
	t.Run("Stop", func(t *testing.T) {
		ctx := context.Background()
		if err := sm.Stop(ctx); err != nil {
			t.Errorf("Stop() error = %v", err)
		}
	})

	// Test persistence across restarts
	t.Run("Persistence", func(t *testing.T) {
		key := "persist_key"
		value := "persist_value"

		// Create new state manager
		sm1 := NewStateManager(statePath)
		ctx := context.Background()

		// Start and set value
		if err := sm1.Start(ctx); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		if err := sm1.SetString(key, value); err != nil {
			t.Fatalf("SetString() error = %v", err)
		}
		if err := sm1.Stop(ctx); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}

		// Create new state manager and verify value persists
		sm2 := NewStateManager(statePath)
		if err := sm2.Start(ctx); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		if got := sm2.GetString(key); got != value {
			t.Errorf("GetString() = %v, want %v", got, value)
		}
		if err := sm2.Stop(ctx); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
	})
}

func TestStateManagerConcurrency(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state_manager_concurrent_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	statePath := filepath.Join(tmpDir, "state.json")
	sm := NewStateManager(statePath)
	ctx := context.Background()

	if err := sm.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer sm.Stop(ctx)

	// Test concurrent access
	t.Run("Concurrent Access", func(t *testing.T) {
		const (
			numGoroutines = 10
			numOperations = 100
		)

		done := make(chan bool)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				for j := 0; j < numOperations; j++ {
					key := "key"
					value := "value"
					
					// Write
					if err := sm.SetString(key, value); err != nil {
						t.Errorf("SetString() error = %v", err)
					}

					// Read
					if got := sm.GetString(key); got != value {
						t.Errorf("GetString() = %v, want %v", got, value)
					}

					// Small sleep to increase chance of race conditions
					time.Sleep(time.Millisecond)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}
