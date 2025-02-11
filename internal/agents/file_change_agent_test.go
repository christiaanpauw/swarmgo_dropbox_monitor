package agents

import (
	"context"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestFileChangeAgent_DetectChanges(t *testing.T) {
	now := time.Now()
	oldTime := now.Add(-24 * time.Hour)

	testCases := []struct {
		name       string
		files      []*models.FileMetadata
		err        error
		wantErr    bool
		wantCount  int
	}{
		{
			name: "Recent changes found",
			files: []*models.FileMetadata{
				{
					Path:           "/test1.txt",
					Name:          "test1.txt",
					Size:          100,
					Modified:      now,
					PathLower:     "/test1.txt",
					ServerModified: now,
				},
				{
					Path:           "/test2.txt",
					Name:          "test2.txt",
					Size:          200,
					Modified:      oldTime,
					PathLower:     "/test2.txt",
					ServerModified: oldTime,
				},
			},
			err:        nil,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:       "Dropbox error",
			files:      nil,
			err:        &DropboxError{Message: "API error"},
			wantErr:    true,
			wantCount:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mockDropboxClient{}

			// Set up mock expectations
			mockClient.On("ListFolder", mock.Anything, "").Return(tc.files, tc.err)

			// Create agent with mock client
			agent := &fileChangeAgent{
				client: mockClient,
			}

			// Test DetectChanges
			changes, err := agent.DetectChanges(context.Background())

			// Check error
			if (err != nil) != tc.wantErr {
				t.Errorf("DetectChanges() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// Check number of changes
			if !tc.wantErr {
				assert.Equal(t, tc.wantCount, len(changes))
			}

			// Verify mock expectations
			mockClient.AssertExpectations(t)
		})
	}
}

// DropboxError is a mock error type for testing
type DropboxError struct {
	Message string
}

func (e *DropboxError) Error() string {
	return e.Message
}
