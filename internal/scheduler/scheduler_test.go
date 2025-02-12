package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agents"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/interfaces"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDropboxClient is a mock implementation of interfaces.DropboxClient
type MockDropboxClient struct {
	mock.Mock
}

func (m *MockDropboxClient) ListFolder(ctx context.Context, path string) ([]*models.FileMetadata, error) {
	args := m.Called(ctx, path)
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

func (m *MockDropboxClient) GetFileContent(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockDropboxClient) GetChanges(ctx context.Context) ([]*models.FileMetadata, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

func (m *MockDropboxClient) GetChangesLast24Hours(ctx context.Context) ([]*models.FileMetadata, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

func (m *MockDropboxClient) GetChangesLast10Minutes(ctx context.Context) ([]*models.FileMetadata, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

func (m *MockDropboxClient) GetFileChanges(ctx context.Context) ([]models.FileChange, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.FileChange), args.Error(1)
}

// MockReportingAgent is a mock implementation of agents.ReportingAgent
type MockReportingAgent struct {
	mock.Mock
	*lifecycle.BaseComponent
}

func NewMockReportingAgent() *MockReportingAgent {
	m := &MockReportingAgent{
		BaseComponent: lifecycle.NewBaseComponent("MockReportingAgent"),
	}
	m.SetState(lifecycle.StateInitialized)
	return m
}

func (m *MockReportingAgent) GenerateReport(ctx context.Context, changes []models.FileChange) error {
	args := m.Called(ctx, changes)
	return args.Error(0)
}

func (m *MockReportingAgent) NotifyChanges(ctx context.Context, changes []models.FileChange) error {
	args := m.Called(ctx, changes)
	return args.Error(0)
}

func (m *MockReportingAgent) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockReportingAgent) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockReportingAgent) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockReportingAgent) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewScheduler(t *testing.T) {
	tests := []struct {
		name          string
		client        interfaces.DropboxClient
		reportingAgent agents.ReportingAgent
		interval      time.Duration
		expectError   bool
	}{
		{
			name:          "valid configuration",
			client:        new(MockDropboxClient),
			reportingAgent: NewMockReportingAgent(),
			interval:      5 * time.Minute,
			expectError:   false,
		},
		{
			name:          "nil client",
			client:        nil,
			reportingAgent: NewMockReportingAgent(),
			interval:      5 * time.Minute,
			expectError:   true,
		},
		{
			name:          "nil reporting agent",
			client:        new(MockDropboxClient),
			reportingAgent: nil,
			interval:      5 * time.Minute,
			expectError:   true,
		},
		{
			name:          "zero interval",
			client:        new(MockDropboxClient),
			reportingAgent: NewMockReportingAgent(),
			interval:      0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler, err := NewScheduler(tt.client, tt.reportingAgent, tt.interval)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, scheduler)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, scheduler)
				assert.Equal(t, tt.interval, scheduler.interval)
				assert.NotNil(t, scheduler.stopCh)
				assert.Equal(t, lifecycle.StateInitialized, scheduler.State())
			}
		})
	}
}

func TestScheduler_Execute(t *testing.T) {
	tests := []struct {
		name          string
		changes       []*models.FileMetadata
		clientErr     error
		reportingErr  error
		expectedError bool
	}{
		{
			name: "successful execution with changes",
			changes: []*models.FileMetadata{
				{Path: "/test1.txt", Size: 100, Modified: time.Now()},
				{Path: "/test2.txt", Size: 200, Modified: time.Now()},
			},
			clientErr:     nil,
			reportingErr:  nil,
			expectedError: false,
		},
		{
			name:          "successful execution with no changes",
			changes:       []*models.FileMetadata{},
			clientErr:     nil,
			reportingErr:  nil,
			expectedError: false,
		},
		{
			name:          "client error",
			changes:       nil,
			clientErr:     assert.AnError,
			reportingErr:  nil,
			expectedError: true,
		},
		{
			name: "reporting error",
			changes: []*models.FileMetadata{
				{Path: "/test1.txt", Size: 100, Modified: time.Now()},
			},
			clientErr:     nil,
			reportingErr:  assert.AnError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(MockDropboxClient)
			reportingAgent := NewMockReportingAgent()
			scheduler, _ := NewScheduler(client, reportingAgent, time.Minute)

			client.On("GetChanges", mock.Anything).Return(tt.changes, tt.clientErr)
			if tt.changes != nil && len(tt.changes) > 0 && tt.clientErr == nil {
				expectedChanges := make([]models.FileChange, len(tt.changes))
				for i, change := range tt.changes {
					expectedChanges[i] = models.FileChange{
						Path:      change.Path,
						Size:      change.Size,
						Modified:  change.Modified,
						IsDeleted: false,
					}
				}
				reportingAgent.On("GenerateReport", mock.Anything, expectedChanges).Return(tt.reportingErr)
			}

			err := scheduler.execute(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			client.AssertExpectations(t)
			reportingAgent.AssertExpectations(t)
		})
	}
}

func TestScheduler_Lifecycle(t *testing.T) {
	ctx := context.Background()
	client := new(MockDropboxClient)
	reportingAgent := NewMockReportingAgent()
	scheduler, err := NewScheduler(client, reportingAgent, time.Minute)
	assert.NoError(t, err)

	// Test Initialize
	err = scheduler.Initialize(ctx)
	assert.NoError(t, err)
	assert.Equal(t, lifecycle.StateInitialized, scheduler.State())

	// Test Start
	err = scheduler.Start(ctx)
	assert.NoError(t, err)
	assert.Equal(t, lifecycle.StateRunning, scheduler.State())

	// Test Health
	reportingAgent.On("Health", mock.Anything).Return(nil).Once()
	err = scheduler.Health(ctx)
	assert.NoError(t, err)
	reportingAgent.AssertExpectations(t)

	// Test Stop
	err = scheduler.Stop(ctx)
	assert.NoError(t, err)
	assert.Equal(t, lifecycle.StateStopped, scheduler.State())
}

func TestScheduler_Health_Error(t *testing.T) {
	ctx := context.Background()
	client := new(MockDropboxClient)
	reportingAgent := NewMockReportingAgent()
	scheduler, err := NewScheduler(client, reportingAgent, time.Minute)
	assert.NoError(t, err)

	reportingAgent.On("Health", mock.Anything).Return(assert.AnError).Once()
	err = scheduler.Health(ctx)
	assert.Error(t, err)
	reportingAgent.AssertExpectations(t)
}
