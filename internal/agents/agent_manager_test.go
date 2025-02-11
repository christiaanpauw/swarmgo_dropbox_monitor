package agents

import (
	"context"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/analysis"
	coremodels "github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting/models"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type mockFileChangeAgent struct {
	mock.Mock
}

func (m *mockFileChangeAgent) GetChanges(ctx context.Context) ([]*coremodels.FileChange, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*coremodels.FileChange), args.Error(1)
}

func (m *mockFileChangeAgent) DetectChanges(ctx context.Context) ([]*coremodels.FileChange, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*coremodels.FileChange), args.Error(1)
}

func (m *mockFileChangeAgent) GetFileContent(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)
	return args.Get(0).([]byte), args.Error(1)
}

type mockDatabaseAgent struct {
	mock.Mock
}

func (m *mockDatabaseAgent) StoreFileContent(ctx context.Context, content *coremodels.FileContent) error {
	args := m.Called(ctx, content)
	return args.Error(0)
}

func (m *mockDatabaseAgent) GetRecentChanges(ctx context.Context, since time.Time) ([]*coremodels.FileChange, error) {
	args := m.Called(ctx, since)
	return args.Get(0).([]*coremodels.FileChange), args.Error(1)
}

func (m *mockDatabaseAgent) Close() error {
	args := m.Called()
	return args.Error(0)
}

type mockReportingAgent struct {
	mock.Mock
}

func (m *mockReportingAgent) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockReportingAgent) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockReportingAgent) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockReportingAgent) State() lifecycle.State {
	args := m.Called()
	return args.Get(0).(lifecycle.State)
}

func (m *mockReportingAgent) GenerateReport(ctx context.Context, changes []models.FileChange) error {
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
	fileChangeAgent.On("GetChanges", mock.Anything).Return([]*coremodels.FileChange{}, nil).Once()
	databaseAgent.On("Close").Return(nil).Once()
	reportingAgent.On("Start", mock.Anything).Return(nil).Once()
	reportingAgent.On("State").Return(lifecycle.StateRunning).Once()  // Only called once during Execute
	reportingAgent.On("Stop", mock.Anything).Return(nil).Once()

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
	reportingAgent.AssertExpectations(t)
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
	testTime := time.Now()
	changes := []*coremodels.FileChange{
		{
			Path:      "/test1.txt",
			Extension: ".txt",
			Directory: "/",
			ModTime:   testTime,
		},
	}

	// Expected reporting changes
	reportChanges := []models.FileChange{
		{
			Path:      "/test1.txt",
			Extension: ".txt",
			Directory: "/",
			ModTime:   testTime,
		},
	}

	// Set up expectations
	fileChangeAgent.On("GetChanges", mock.Anything).Return(changes, nil).Once()
	fileChangeAgent.On("GetFileContent", mock.Anything, "/test1.txt").Return([]byte("test content 1"), nil).Once()
	databaseAgent.On("StoreFileContent", mock.Anything, mock.MatchedBy(func(content *coremodels.FileContent) bool {
		return content.Path == "/test1.txt"
	})).Return(nil).Once()
	reportingAgent.On("GenerateReport", mock.Anything, reportChanges).Return(nil).Once()
	reportingAgent.On("State").Return(lifecycle.StateRunning).Once()

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

	tests := []struct {
		name    string
		deps    AgentManagerDeps
		wantErr bool
	}{
		{
			name: "all dependencies present",
			deps: AgentManagerDeps{
				FileChangeAgent:  fileChangeAgent,
				ContentAnalyzer:  contentAnalyzer,
				DatabaseAgent:    databaseAgent,
				ReportingAgent:   reportingAgent,
				Notifier:        notifier,
			},
			wantErr: false,
		},
		{
			name: "missing file change agent",
			deps: AgentManagerDeps{
				ContentAnalyzer:  contentAnalyzer,
				DatabaseAgent:    databaseAgent,
				ReportingAgent:   reportingAgent,
				Notifier:        notifier,
			},
			wantErr: true,
		},
		{
			name: "missing content analyzer",
			deps: AgentManagerDeps{
				FileChangeAgent:  fileChangeAgent,
				DatabaseAgent:    databaseAgent,
				ReportingAgent:   reportingAgent,
				Notifier:        notifier,
			},
			wantErr: true,
		},
		{
			name: "missing database agent",
			deps: AgentManagerDeps{
				FileChangeAgent:  fileChangeAgent,
				ContentAnalyzer:  contentAnalyzer,
				ReportingAgent:   reportingAgent,
				Notifier:        notifier,
			},
			wantErr: true,
		},
		{
			name: "missing reporting agent",
			deps: AgentManagerDeps{
				FileChangeAgent:  fileChangeAgent,
				ContentAnalyzer:  contentAnalyzer,
				DatabaseAgent:    databaseAgent,
				Notifier:        notifier,
			},
			wantErr: true,
		},
		{
			name: "missing notifier",
			deps: AgentManagerDeps{
				FileChangeAgent:  fileChangeAgent,
				ContentAnalyzer:  contentAnalyzer,
				DatabaseAgent:    databaseAgent,
				ReportingAgent:   reportingAgent,
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
