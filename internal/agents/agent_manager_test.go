package agents

import (
	"context"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type mockFileChangeAgent struct {
	mock.Mock
}

func (m *mockFileChangeAgent) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockFileChangeAgent) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockFileChangeAgent) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockFileChangeAgent) State() lifecycle.ComponentState {
	args := m.Called()
	return args.Get(0).(lifecycle.ComponentState)
}

func (m *mockFileChangeAgent) GetChanges(ctx context.Context) ([]models.FileChange, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.FileChange), args.Error(1)
}

func (m *mockFileChangeAgent) SetPollInterval(interval time.Duration) {
	m.Called(interval)
}

func (m *mockFileChangeAgent) GetFileContent(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)
	return args.Get(0).([]byte), args.Error(1)
}

type mockDatabaseAgent struct {
	mock.Mock
}

func (m *mockDatabaseAgent) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockDatabaseAgent) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockDatabaseAgent) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockDatabaseAgent) State() lifecycle.ComponentState {
	args := m.Called()
	return args.Get(0).(lifecycle.ComponentState)
}

func (m *mockDatabaseAgent) StoreChange(ctx context.Context, change models.FileMetadata) error {
	args := m.Called(ctx, change)
	return args.Error(0)
}

func (m *mockDatabaseAgent) GetLatestChanges(ctx context.Context, limit int) ([]models.FileMetadata, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]models.FileMetadata), args.Error(1)
}

func (m *mockDatabaseAgent) GetChanges(ctx context.Context, startTime, endTime string) ([]models.FileMetadata, error) {
	args := m.Called(ctx, startTime, endTime)
	return args.Get(0).([]models.FileMetadata), args.Error(1)
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

func (m *mockReportingAgent) State() lifecycle.ComponentState {
	args := m.Called()
	return args.Get(0).(lifecycle.ComponentState)
}

func (m *mockReportingAgent) GenerateReport(ctx context.Context, changes []models.FileChange) error {
	args := m.Called(ctx, changes)
	return args.Error(0)
}

func (m *mockReportingAgent) NotifyChanges(ctx context.Context, changes []models.FileChange) error {
	args := m.Called(ctx, changes)
	return args.Error(0)
}

func TestAgentManager_Start(t *testing.T) {
	// Create mocks
	fileChangeAgent := new(mockFileChangeAgent)
	databaseAgent := new(mockDatabaseAgent)
	reportingAgent := new(mockReportingAgent)

	// Setup expectations for initialization
	fileChangeAgent.On("State").Return(lifecycle.StateInitialized).Times(1)
	databaseAgent.On("State").Return(lifecycle.StateInitialized).Times(1)
	reportingAgent.On("State").Return(lifecycle.StateInitialized).Times(1)

	// Create agent manager
	am := NewAgentManager(AgentManagerDeps{
		FileChangeAgent: fileChangeAgent,
		DatabaseAgent:   databaseAgent,
		ReportingAgent:  reportingAgent,
		Notifier:        nil,
		ContentAnalyzer: nil,
	})

	// Initialize agent manager
	err := am.Initialize(context.Background())
	assert.NoError(t, err)

	// Setup expectations for start
	fileChangeAgent.On("Start", mock.Anything).Return(nil).Times(1)
	databaseAgent.On("Start", mock.Anything).Return(nil).Times(1)
	reportingAgent.On("Start", mock.Anything).Return(nil).Times(1)

	fileChangeAgent.On("State").Return(lifecycle.StateRunning).Times(1)
	databaseAgent.On("State").Return(lifecycle.StateRunning).Times(1)
	reportingAgent.On("State").Return(lifecycle.StateRunning).Times(1)

	// Start agent manager
	err = am.Start(context.Background())
	assert.NoError(t, err)

	// Assert expectations
	fileChangeAgent.AssertExpectations(t)
	databaseAgent.AssertExpectations(t)
	reportingAgent.AssertExpectations(t)
}

func TestAgentManager_Stop(t *testing.T) {
	// Create mocks
	fileChangeAgent := new(mockFileChangeAgent)
	databaseAgent := new(mockDatabaseAgent)
	reportingAgent := new(mockReportingAgent)

	// Setup expectations for initialization and start
	fileChangeAgent.On("State").Return(lifecycle.StateInitialized).Once()
	databaseAgent.On("State").Return(lifecycle.StateInitialized).Once()
	reportingAgent.On("State").Return(lifecycle.StateInitialized).Once()

	// Create agent manager
	am := NewAgentManager(AgentManagerDeps{
		FileChangeAgent: fileChangeAgent,
		DatabaseAgent:   databaseAgent,
		ReportingAgent:  reportingAgent,
		Notifier:        nil,
		ContentAnalyzer: nil,
	})

	// Initialize agent manager
	err := am.Initialize(context.Background())
	assert.NoError(t, err)

	// Setup expectations for start
	fileChangeAgent.On("Start", mock.Anything).Return(nil).Once()
	databaseAgent.On("Start", mock.Anything).Return(nil).Once()
	reportingAgent.On("Start", mock.Anything).Return(nil).Once()

	fileChangeAgent.On("State").Return(lifecycle.StateRunning).Once()
	databaseAgent.On("State").Return(lifecycle.StateRunning).Once()
	reportingAgent.On("State").Return(lifecycle.StateRunning).Once()

	// Start agent manager
	err = am.Start(context.Background())
	assert.NoError(t, err)

	// Setup expectations for stop
	fileChangeAgent.On("Stop", mock.Anything).Return(nil).Once()
	databaseAgent.On("Stop", mock.Anything).Return(nil).Once()
	reportingAgent.On("Stop", mock.Anything).Return(nil).Once()

	fileChangeAgent.On("State").Return(lifecycle.StateRunning).Once()
	databaseAgent.On("State").Return(lifecycle.StateRunning).Once()
	reportingAgent.On("State").Return(lifecycle.StateRunning).Once()

	// Stop agent manager
	err = am.Stop(context.Background())
	assert.NoError(t, err)

	// Assert expectations
	fileChangeAgent.AssertExpectations(t)
	databaseAgent.AssertExpectations(t)
	reportingAgent.AssertExpectations(t)
}

func TestAgentManager_Health(t *testing.T) {
	// Create mocks
	fileChangeAgent := new(mockFileChangeAgent)
	databaseAgent := new(mockDatabaseAgent)
	reportingAgent := new(mockReportingAgent)

	// Setup expectations for initialization and start
	fileChangeAgent.On("State").Return(lifecycle.StateInitialized).Once()
	databaseAgent.On("State").Return(lifecycle.StateInitialized).Once()
	reportingAgent.On("State").Return(lifecycle.StateInitialized).Once()

	// Create agent manager
	am := NewAgentManager(AgentManagerDeps{
		FileChangeAgent: fileChangeAgent,
		DatabaseAgent:   databaseAgent,
		ReportingAgent:  reportingAgent,
		Notifier:        nil,
		ContentAnalyzer: nil,
	})

	// Initialize agent manager
	err := am.Initialize(context.Background())
	assert.NoError(t, err)

	// Setup expectations for start
	fileChangeAgent.On("Start", mock.Anything).Return(nil).Once()
	databaseAgent.On("Start", mock.Anything).Return(nil).Once()
	reportingAgent.On("Start", mock.Anything).Return(nil).Once()

	fileChangeAgent.On("State").Return(lifecycle.StateRunning).Once()
	databaseAgent.On("State").Return(lifecycle.StateRunning).Once()
	reportingAgent.On("State").Return(lifecycle.StateRunning).Once()

	// Start agent manager
	err = am.Start(context.Background())
	assert.NoError(t, err)

	// Setup expectations for health check
	fileChangeAgent.On("Health", mock.Anything).Return(nil).Once()
	databaseAgent.On("Health", mock.Anything).Return(nil).Once()
	reportingAgent.On("Health", mock.Anything).Return(nil).Once()

	fileChangeAgent.On("State").Return(lifecycle.StateRunning).Once()
	databaseAgent.On("State").Return(lifecycle.StateRunning).Once()
	reportingAgent.On("State").Return(lifecycle.StateRunning).Once()

	// Check health
	err = am.Health(context.Background())
	assert.NoError(t, err)

	// Assert expectations
	fileChangeAgent.AssertExpectations(t)
	databaseAgent.AssertExpectations(t)
	reportingAgent.AssertExpectations(t)
}
