package dropbox

import (
	"context"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockDropboxClient is a mock implementation of the Client interface
type MockDropboxClient struct {
	mock.Mock
}

// ListFolder mocks the ListFolder method
func (m *MockDropboxClient) ListFolder(ctx context.Context, path string) ([]*models.FileMetadata, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

// GetFileContent mocks the GetFileContent method
func (m *MockDropboxClient) GetFileContent(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// GetChangesLast24Hours mocks the GetChangesLast24Hours method
func (m *MockDropboxClient) GetChangesLast24Hours(ctx context.Context) ([]*models.FileMetadata, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

// GetChangesLast10Minutes mocks the GetChangesLast10Minutes method
func (m *MockDropboxClient) GetChangesLast10Minutes(ctx context.Context) ([]*models.FileMetadata, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

// GetChanges mocks the GetChanges method
func (m *MockDropboxClient) GetChanges(ctx context.Context) ([]*models.FileMetadata, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

// GetFileChanges mocks the GetFileChanges method
func (m *MockDropboxClient) GetFileChanges(ctx context.Context) ([]models.FileChange, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.FileChange), args.Error(1)
}
