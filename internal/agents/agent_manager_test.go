package agents

import (
	"context"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/analysis"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type mockFileChangeAgent struct {
	mock.Mock
}

func (m *mockFileChangeAgent) GetChanges(ctx context.Context) ([]*models.FileChange, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.FileChange), args.Error(1)
}

func (m *mockFileChangeAgent) DetectChanges(ctx context.Context) ([]*models.FileChange, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.FileChange), args.Error(1)
}

func (m *mockFileChangeAgent) GetFileContent(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)
	return args.Get(0).([]byte), args.Error(1)
}

type mockDatabaseAgent struct {
	mock.Mock
}

func (m *mockDatabaseAgent) StoreFileContent(ctx context.Context, content *models.FileContent) error {
	args := m.Called(ctx, content)
	return args.Error(0)
}

func (m *mockDatabaseAgent) GetRecentChanges(ctx context.Context, since time.Time) ([]*models.FileChange, error) {
	args := m.Called(ctx, since)
	return args.Get(0).([]*models.FileChange), args.Error(1)
}

func (m *mockDatabaseAgent) Close() error {
	args := m.Called()
	return args.Error(0)
}

type mockReportingAgent struct {
	mock.Mock
}

func (m *mockReportingAgent) GenerateReport(ctx context.Context, changes []*models.FileChange) error {
	args := m.Called(ctx, changes)
	return args.Error(0)
}

func TestAgentManager_Start(t *testing.T) {
	// Create mocks
	fileChangeAgent := &mockFileChangeAgent{}
	databaseAgent := &mockDatabaseAgent{}
	contentAnalyzer := analysis.NewContentAnalyzer()
	notifier := notify.NewNotifier()
	reportingAgent := &mockReportingAgent{}

	// Create config
	cfg := AgentManagerConfig{
		PollInterval: time.Millisecond * 100, // Short interval for testing
		MaxRetries:   3,
		RetryDelay:   time.Millisecond * 10,
	}

	// Create dependencies
	deps := AgentManagerDeps{
		FileChangeAgent:  fileChangeAgent,
		ContentAnalyzer:  contentAnalyzer,
		DatabaseAgent:    databaseAgent,
		ReportingAgent:   reportingAgent,
		Notifier:        notifier,
	}

	// Create agent manager
	manager := NewAgentManager(cfg, deps)

	// Set up expectations
	fileChangeAgent.On("GetChanges", mock.Anything).Return([]*models.FileChange{}, nil).Once()
	databaseAgent.On("Close").Return(nil).Once()

	// Start the manager
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := manager.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, manager.IsRunning())

	// Wait for at least one poll
	time.Sleep(150 * time.Millisecond)

	// Stop the manager
	err = manager.Stop(ctx)
	assert.NoError(t, err)
	assert.False(t, manager.IsRunning())

	// Verify expectations
	fileChangeAgent.AssertExpectations(t)
	databaseAgent.AssertExpectations(t)
}

func TestAgentManager_Execute(t *testing.T) {
	// Create mocks
	fileChangeAgent := &mockFileChangeAgent{}
	databaseAgent := &mockDatabaseAgent{}
	contentAnalyzer := analysis.NewContentAnalyzer()
	notifier := notify.NewNotifier()
	reportingAgent := &mockReportingAgent{}

	// Create config
	cfg := AgentManagerConfig{
		PollInterval: time.Second,
		MaxRetries:   3,
		RetryDelay:   time.Millisecond * 10,
	}

	// Create dependencies
	deps := AgentManagerDeps{
		FileChangeAgent:  fileChangeAgent,
		ContentAnalyzer:  contentAnalyzer,
		DatabaseAgent:    databaseAgent,
		ReportingAgent:   reportingAgent,
		Notifier:        notifier,
	}

	// Create agent manager
	manager := NewAgentManager(cfg, deps)

	// Set up test data
	changes := []*models.FileChange{
		{
			Path:     "/test1.txt",
			ModTime:  time.Now(),
			IsDeleted: false,
		},
		{
			Path:     "/test2.txt",
			ModTime:  time.Now(),
			IsDeleted: true,
		},
	}

	// Set up expectations
	fileChangeAgent.On("GetChanges", mock.Anything).Return(changes, nil).Once()
	fileChangeAgent.On("GetFileContent", mock.Anything, "/test1.txt").Return([]byte("test content"), nil).Once()
	databaseAgent.On("StoreFileContent", mock.Anything, mock.AnythingOfType("*models.FileContent")).Return(nil).Once()
	reportingAgent.On("GenerateReport", mock.Anything, changes).Return(nil).Once()

	// Execute the manager
	ctx := context.Background()
	err := manager.Execute(ctx)
	assert.NoError(t, err)

	// Verify expectations
	fileChangeAgent.AssertExpectations(t)
	databaseAgent.AssertExpectations(t)
	reportingAgent.AssertExpectations(t)
}

func TestAgentManager_ValidateDependencies(t *testing.T) {
	// Create config
	cfg := AgentManagerConfig{
		PollInterval: time.Second,
		MaxRetries:   3,
		RetryDelay:   time.Millisecond * 10,
	}

	tests := []struct {
		name    string
		deps    AgentManagerDeps
		wantErr bool
	}{
		{
			name: "all dependencies present",
			deps: AgentManagerDeps{
				FileChangeAgent:  &mockFileChangeAgent{},
				ContentAnalyzer:  analysis.NewContentAnalyzer(),
				DatabaseAgent:   &mockDatabaseAgent{},
				ReportingAgent:  &mockReportingAgent{},
				Notifier:       notify.NewNotifier(),
			},
			wantErr: false,
		},
		{
			name: "missing file change agent",
			deps: AgentManagerDeps{
				ContentAnalyzer:  analysis.NewContentAnalyzer(),
				DatabaseAgent:   &mockDatabaseAgent{},
				ReportingAgent:  &mockReportingAgent{},
				Notifier:       notify.NewNotifier(),
			},
			wantErr: true,
		},
		{
			name: "missing database agent",
			deps: AgentManagerDeps{
				FileChangeAgent:  &mockFileChangeAgent{},
				ContentAnalyzer:  analysis.NewContentAnalyzer(),
				ReportingAgent:  &mockReportingAgent{},
				Notifier:       notify.NewNotifier(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewAgentManager(cfg, tt.deps)
			err := manager.validateDependencies()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
