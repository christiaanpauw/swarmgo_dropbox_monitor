package lifecycle

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockComponent is a test implementation of Component
type mockComponent struct {
	*BaseComponent
	startErr  error
	stopErr   error
	healthErr error
}

func newMockComponent(name string) *mockComponent {
	return &mockComponent{
		BaseComponent: NewBaseComponent(name),
	}
}

func (m *mockComponent) Start(ctx context.Context) error {
	if m.startErr != nil {
		m.SetState(StateFailed)
		return m.startErr
	}
	m.SetState(StateStarting)
	m.SetState(StateRunning)
	return nil
}

func (m *mockComponent) Stop(ctx context.Context) error {
	if m.stopErr != nil {
		m.SetState(StateFailed)
		return m.stopErr
	}
	m.SetState(StateStopping)
	m.SetState(StateStopped)
	return nil
}

func (m *mockComponent) Health(ctx context.Context) error {
	if m.healthErr != nil {
		return m.healthErr
	}
	if m.State() != StateRunning {
		return errors.New("component not running")
	}
	return nil
}

func TestStartupSequence(t *testing.T) {
	tests := []struct {
		name        string
		component   *mockComponent
		timeout     time.Duration
		expectError bool
	}{
		{
			name:        "successful startup",
			component:   newMockComponent("test"),
			timeout:     1 * time.Second,
			expectError: false,
		},
		{
			name: "startup error",
			component: func() *mockComponent {
				m := newMockComponent("test")
				m.startErr = errors.New("start error")
				return m
			}(),
			timeout:     1 * time.Second,
			expectError: true,
		},
		{
			name: "health check timeout",
			component: func() *mockComponent {
				m := newMockComponent("test")
				m.healthErr = errors.New("health error")
				return m
			}(),
			timeout:     100 * time.Millisecond,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := StartupSequence(context.Background(), tt.component, tt.timeout)
			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestShutdownSequence(t *testing.T) {
	tests := []struct {
		name        string
		component   *mockComponent
		timeout     time.Duration
		expectError bool
	}{
		{
			name:        "successful shutdown",
			component:   newMockComponent("test"),
			timeout:     1 * time.Second,
			expectError: false,
		},
		{
			name: "shutdown error",
			component: func() *mockComponent {
				m := newMockComponent("test")
				m.stopErr = errors.New("stop error")
				return m
			}(),
			timeout:     1 * time.Second,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First start the component
			if err := tt.component.Start(context.Background()); err != nil {
				t.Fatalf("failed to start component: %v", err)
			}

			err := ShutdownSequence(context.Background(), tt.component, tt.timeout)
			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestComponentState(t *testing.T) {
	c := NewBaseComponent("test")

	if c.State() != StateUninitialized {
		t.Errorf("expected initial state to be Uninitialized, got %v", c.State())
	}

	c.SetState(StateRunning)
	if c.State() != StateRunning {
		t.Errorf("expected state to be Running, got %v", c.State())
	}

	if c.Name() != "test" {
		t.Errorf("expected name to be 'test', got %v", c.Name())
	}
}
