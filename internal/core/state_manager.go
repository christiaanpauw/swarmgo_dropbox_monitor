package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
)

// StateManager manages application state persistence
type StateManager struct {
	*lifecycle.BaseComponent
	mu        sync.RWMutex
	statePath string
	state     map[string]interface{}
}

// NewStateManager creates a new state manager
func NewStateManager(statePath string) *StateManager {
	return &StateManager{
		BaseComponent: lifecycle.NewBaseComponent("StateManager"),
		statePath:     statePath,
		state:        make(map[string]interface{}),
	}
}

// Start implements lifecycle.Component
func (sm *StateManager) Start(ctx context.Context) error {
	if err := sm.DefaultStart(ctx); err != nil {
		return err
	}

	// Ensure state directory exists
	if err := os.MkdirAll(filepath.Dir(sm.statePath), 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Load initial state
	if err := sm.loadState(); err != nil {
		return fmt.Errorf("failed to load initial state: %w", err)
	}

	return nil
}

// Stop implements lifecycle.Component
func (sm *StateManager) Stop(ctx context.Context) error {
	if err := sm.saveState(); err != nil {
		return fmt.Errorf("failed to save state during shutdown: %w", err)
	}
	return sm.DefaultStop(ctx)
}

// Health implements lifecycle.Component
func (sm *StateManager) Health(ctx context.Context) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.state == nil {
		return fmt.Errorf("state is nil")
	}
	return sm.DefaultHealth(ctx)
}

// GetString retrieves a string value from state
func (sm *StateManager) GetString(key string) string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if val, ok := sm.state[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// SetString stores a string value in state
func (sm *StateManager) SetString(key, value string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.state[key] = value
	return sm.saveState()
}

// loadState loads state from disk
func (sm *StateManager) loadState() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, err := os.ReadFile(sm.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			sm.state = make(map[string]interface{})
			return nil
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}

	if err := json.Unmarshal(data, &sm.state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return nil
}

// saveState saves state to disk
func (sm *StateManager) saveState() error {
	data, err := json.MarshalIndent(sm.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(sm.statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}
