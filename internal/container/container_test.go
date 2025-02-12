package container

import (
	"context"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/config"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockReportingAgent mocks the ReportingAgent interface
type MockReportingAgent struct {
	mock.Mock
	*lifecycle.BaseComponent
}

func NewMockReportingAgent() *MockReportingAgent {
	return &MockReportingAgent{
		BaseComponent: lifecycle.NewBaseComponent("MockReportingAgent"),
	}
}

func (m *MockReportingAgent) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.SetState(lifecycle.StateInitialized)
	}
	return args.Error(0)
}

func (m *MockReportingAgent) Start(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.SetState(lifecycle.StateRunning)
	}
	return args.Error(0)
}

func (m *MockReportingAgent) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.SetState(lifecycle.StateStopped)
	}
	return args.Error(0)
}

func (m *MockReportingAgent) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockReportingAgent) GenerateReport(ctx context.Context, changes []models.FileChange) error {
	args := m.Called(ctx, changes)
	return args.Error(0)
}

func (m *MockReportingAgent) NotifyChanges(ctx context.Context, changes []models.FileChange) error {
	args := m.Called(ctx, changes)
	return args.Error(0)
}

// MockFileChangeAgent mocks the FileChangeAgent interface
type MockFileChangeAgent struct {
	mock.Mock
	*lifecycle.BaseComponent
}

func NewMockFileChangeAgent() *MockFileChangeAgent {
	return &MockFileChangeAgent{
		BaseComponent: lifecycle.NewBaseComponent("MockFileChangeAgent"),
	}
}

func (m *MockFileChangeAgent) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.SetState(lifecycle.StateInitialized)
	}
	return args.Error(0)
}

func (m *MockFileChangeAgent) Start(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.SetState(lifecycle.StateRunning)
	}
	return args.Error(0)
}

func (m *MockFileChangeAgent) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.SetState(lifecycle.StateStopped)
	}
	return args.Error(0)
}

func (m *MockFileChangeAgent) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockFileChangeAgent) GetChanges(ctx context.Context) ([]models.FileChange, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.FileChange), args.Error(1)
}

func (m *MockFileChangeAgent) GetFileContent(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockFileChangeAgent) SetPollInterval(interval time.Duration) {
	m.Called(interval)
}

// MockDatabaseAgent mocks the DatabaseAgent interface
type MockDatabaseAgent struct {
	mock.Mock
	*lifecycle.BaseComponent
}

func NewMockDatabaseAgent() *MockDatabaseAgent {
	return &MockDatabaseAgent{
		BaseComponent: lifecycle.NewBaseComponent("MockDatabaseAgent"),
	}
}

func (m *MockDatabaseAgent) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.SetState(lifecycle.StateInitialized)
	}
	return args.Error(0)
}

func (m *MockDatabaseAgent) Start(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.SetState(lifecycle.StateRunning)
	}
	return args.Error(0)
}

func (m *MockDatabaseAgent) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.SetState(lifecycle.StateStopped)
	}
	return args.Error(0)
}

func (m *MockDatabaseAgent) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDatabaseAgent) StoreChange(ctx context.Context, change models.FileMetadata) error {
	args := m.Called(ctx, change)
	return args.Error(0)
}

func (m *MockDatabaseAgent) GetLatestChanges(ctx context.Context, limit int) ([]models.FileMetadata, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]models.FileMetadata), args.Error(1)
}

func (m *MockDatabaseAgent) GetChanges(ctx context.Context, startTime, endTime string) ([]models.FileMetadata, error) {
	args := m.Called(ctx, startTime, endTime)
	return args.Get(0).([]models.FileMetadata), args.Error(1)
}

func (m *MockDatabaseAgent) StoreFileContent(ctx context.Context, content *models.FileContent) error {
	args := m.Called(ctx, content)
	return args.Error(0)
}

func (m *MockDatabaseAgent) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewContainer(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
		},
		{
			name: "valid config",
			cfg: &config.Config{
				DropboxToken: "test-token",
				PollInterval: 5 * time.Minute,
				Monitoring: config.MonitoringConfig{
					Path:    "/test/monitor",
					Enabled: true,
				},
				Retry: config.RetryConfig{
					MaxAttempts: 3,
					Delay:      time.Second,
				},
				HealthCheck: config.HealthCheckConfig{
					Interval: time.Second,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := NewContainer(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, container)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, container)
			}
		})
	}
}

func TestContainer_Lifecycle(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		DropboxToken: "test-token",
		PollInterval: 5 * time.Minute,
		Monitoring: config.MonitoringConfig{
			Path:    "/test/monitor",
			Enabled: true,
		},
		Retry: config.RetryConfig{
			MaxAttempts: 3,
			Delay:      time.Second,
		},
		HealthCheck: config.HealthCheckConfig{
			Interval: time.Second,
		},
	}

	// Create mock agents
	mockClient := &dropbox.MockDropboxClient{}

	mockReportingAgent := NewMockReportingAgent()
	mockReportingAgent.On("Initialize", mock.Anything).Return(nil).Once()
	mockReportingAgent.On("Start", mock.Anything).Return(nil).Once()
	mockReportingAgent.On("Stop", mock.Anything).Return(nil).Once()
	mockReportingAgent.On("Health", mock.Anything).Return(nil).Maybe()
	mockReportingAgent.On("State").Return(lifecycle.StateInitialized).Maybe()

	mockFileChangeAgent := NewMockFileChangeAgent()
	mockFileChangeAgent.On("Initialize", mock.Anything).Return(nil).Once()
	mockFileChangeAgent.On("Start", mock.Anything).Return(nil).Once()
	mockFileChangeAgent.On("Stop", mock.Anything).Return(nil).Once()
	mockFileChangeAgent.On("Health", mock.Anything).Return(nil).Maybe()
	mockFileChangeAgent.On("State").Return(lifecycle.StateInitialized).Maybe()

	mockDatabaseAgent := NewMockDatabaseAgent()
	mockDatabaseAgent.On("Initialize", mock.Anything).Return(nil).Once()
	mockDatabaseAgent.On("Start", mock.Anything).Return(nil).Once()
	mockDatabaseAgent.On("Stop", mock.Anything).Return(nil).Once()
	mockDatabaseAgent.On("Health", mock.Anything).Return(nil).Maybe()
	mockDatabaseAgent.On("State").Return(lifecycle.StateInitialized).Maybe()

	// Create scheduler
	scheduler, err := scheduler.NewScheduler(mockClient, mockReportingAgent, cfg.PollInterval)
	assert.NoError(t, err)

	// Create container with mocks
	container, err := NewContainerWithMocks(cfg, mockClient, mockReportingAgent, mockFileChangeAgent, mockDatabaseAgent, scheduler)
	assert.NoError(t, err)
	assert.NotNil(t, container)

	// Initialize container components
	ctx := context.Background()
	err = mockReportingAgent.Initialize(ctx)
	assert.NoError(t, err)
	err = mockFileChangeAgent.Initialize(ctx)
	assert.NoError(t, err)
	err = mockDatabaseAgent.Initialize(ctx)
	assert.NoError(t, err)

	// Update state expectations for running state
	mockReportingAgent.On("State").Return(lifecycle.StateRunning).Maybe()
	mockFileChangeAgent.On("State").Return(lifecycle.StateRunning).Maybe()
	mockDatabaseAgent.On("State").Return(lifecycle.StateRunning).Maybe()

	// Test container lifecycle
	// Test Start
	err = container.Start(ctx)
	assert.NoError(t, err)
	assert.Equal(t, lifecycle.StateRunning, container.State())

	// Test Health
	err = container.Health(ctx)
	assert.NoError(t, err)

	// Test Stop
	err = container.Stop(ctx)
	assert.NoError(t, err)
	assert.Equal(t, lifecycle.StateStopped, container.State())

	// Verify all expectations were met
	mockClient.AssertExpectations(t)
	mockReportingAgent.AssertExpectations(t)
	mockFileChangeAgent.AssertExpectations(t)
	mockDatabaseAgent.AssertExpectations(t)
}
