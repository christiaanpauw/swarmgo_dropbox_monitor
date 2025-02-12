package agents

import (
	"context"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/core"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockDropboxClient is a mock implementation of the Dropbox client
type mockDropboxClient struct {
	mock.Mock
	dropbox.Client
}

func (m *mockDropboxClient) ListFolder(ctx context.Context, path string) ([]*models.FileMetadata, error) {
	args := m.Called(ctx, path)
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

func (m *mockDropboxClient) GetFileContent(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockDropboxClient) GetChangesLast24Hours(ctx context.Context) ([]*models.FileMetadata, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

func (m *mockDropboxClient) GetChangesLast10Minutes(ctx context.Context) ([]*models.FileMetadata, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

func (m *mockDropboxClient) GetChanges(ctx context.Context) ([]*models.FileMetadata, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

// mockStateManager is a mock implementation of the StateManager
type mockStateManager struct {
	mock.Mock
	core.StateManager
}

func (m *mockStateManager) GetString(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *mockStateManager) SetString(key, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func TestFileChangeAgent_GetChanges(t *testing.T) {
	// Create test files
	now := time.Now()
	testFiles := []*models.FileMetadata{
		models.NewFileMetadata("/test1.txt", 1024, now, false),
		models.NewFileMetadata("/test2.txt", 2048, now, false),
	}

	// Create expected changes
	expectedChanges := models.BatchConvertMetadataToChanges(testFiles)

	tests := []struct {
		name      string
		files     []*models.FileMetadata
		wantErr   bool
		expected  []models.FileChange
		err       error
	}{
		{
			name:      "Recent changes found",
			files:     testFiles,
			wantErr:   false,
			expected:  expectedChanges,
			err:       nil,
		},
		{
			name:      "Dropbox error",
			files:     nil,
			wantErr:   true,
			expected:  nil,
			err:       assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockClient := &mockDropboxClient{}
			mockState := &mockStateManager{}

			// Set up mock expectations for the embedded FileChangeAgent
			mockClient.On("ListFolder", mock.Anything, "").Return(tt.files, tt.err).Once()
			mockState.On("GetString", "cursor").Return("").Once()
			if !tt.wantErr {
				mockState.On("SetString", "cursor", mock.Anything).Return(nil).Once()
			}

			// Create agent with mocks
			agent := NewFileChangeAgent(mockClient, mockState, "")

			// Initialize the embedded FileChangeAgent
			agent.(*fileChangeAgentImpl).FileChangeAgent.(*core.FileChangeAgentImpl).SetState(lifecycle.StateInitialized)

			// Run startup sequence
			err := agent.Start(context.Background())
			require.NoError(t, err)

			// Get changes
			changes, err := agent.GetChanges(context.Background())

			// Verify expectations
			mockClient.AssertExpectations(t)
			mockState.AssertExpectations(t)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, changes)
		})
	}
}

func TestFileChangeAgent_Lifecycle(t *testing.T) {
	// Create mocks
	mockClient := &mockDropboxClient{}
	mockState := &mockStateManager{}

	// Set up mock expectations
	mockClient.On("ListFolder", mock.Anything, "").Return([]*models.FileMetadata{}, nil).Times(2)

	// Create agent with mocks
	agent := NewFileChangeAgent(mockClient, mockState, "")

	// Initialize the embedded FileChangeAgent
	agent.(*fileChangeAgentImpl).FileChangeAgent.(*core.FileChangeAgentImpl).SetState(lifecycle.StateInitialized)

	// Run startup sequence
	err := agent.Start(context.Background())
	assert.NoError(t, err)

	// Test Health
	t.Run("Health", func(t *testing.T) {
		err := agent.Health(context.Background())
		assert.NoError(t, err)
	})

	// Run shutdown sequence
	err = agent.Stop(context.Background())
	assert.NoError(t, err)
}

// DropboxError is a mock error type for testing
type DropboxError struct {
	Message string
}

func (e *DropboxError) Error() string {
	return e.Message
}
