package agents

import (
	"context"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNotifier for testing
type mockNotifier struct {
	sentMessages int
	lastMessage  string
	shouldError  bool
}

func (m *mockNotifier) SendNotification(ctx context.Context, message string) error {
	if m.shouldError {
		return assert.AnError
	}
	m.sentMessages++
	m.lastMessage = message
	return nil
}

func TestReportingAgent_Lifecycle(t *testing.T) {
	notifier := &mockNotifier{}
	agent, err := NewReportingAgent(notifier)
	require.NoError(t, err)
	require.NotNil(t, agent)

	// Test initial state
	assert.Equal(t, lifecycle.StateInitialized, agent.State())

	// Test Start
	err = agent.Start(context.Background())
	require.NoError(t, err)
	assert.Equal(t, lifecycle.StateRunning, agent.State())

	// Test Health
	err = agent.Health(context.Background())
	require.NoError(t, err)

	// Test Stop
	err = agent.Stop(context.Background())
	require.NoError(t, err)
	assert.Equal(t, lifecycle.StateStopped, agent.State())
}

func TestReportingAgent_GenerateReport(t *testing.T) {
	tests := []struct {
		name        string
		changes     []models.FileChange
		shouldError bool
		wantErr     bool
		shouldStart bool
	}{
		{
			name: "successful report generation",
			changes: []models.FileChange{
				{
					Path:      "/test/file1.txt",
					Extension: ".txt",
					Directory: "/test",
					ModTime:   time.Now(),
					Size:      1024,
				},
			},
			shouldError: false,
			wantErr:     false,
			shouldStart: true,
		},
		{
			name:        "empty changes list",
			changes:     []models.FileChange{},
			shouldError: false,
			wantErr:     false,
			shouldStart: true,
		},
		{
			name: "notifier error",
			changes: []models.FileChange{
				{
					Path:      "/test/file1.txt",
					Extension: ".txt",
					Directory: "/test",
					ModTime:   time.Now(),
					Size:      1024,
				},
			},
			shouldError: true,
			wantErr:     true,
			shouldStart: true,
		},
		{
			name: "agent not started",
			changes: []models.FileChange{
				{
					Path:      "/test/file1.txt",
					Extension: ".txt",
					Directory: "/test",
					ModTime:   time.Now(),
					Size:      1024,
				},
			},
			shouldError: false,
			wantErr:     true,
			shouldStart: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new notifier for each test case
			notifier := &mockNotifier{shouldError: tt.shouldError}

			agent, err := NewReportingAgent(notifier)
			require.NoError(t, err)
			require.NotNil(t, agent)

			if tt.shouldStart {
				err := agent.Start(context.Background())
				require.NoError(t, err)
			}

			err = agent.GenerateReport(context.Background(), tt.changes)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if len(tt.changes) > 0 {
				assert.Equal(t, 3, notifier.sentMessages) // One message per report type
			}
		})
	}
}
