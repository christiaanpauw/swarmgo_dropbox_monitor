package container

import (
	"context"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/interfaces"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/stretchr/testify/mock"
)

// mockDropboxClient implements interfaces.DropboxClient for testing
type mockDropboxClient struct {
	mock.Mock
}

func (m *mockDropboxClient) ListFolder(ctx context.Context, path string) ([]*models.FileMetadata, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

func (m *mockDropboxClient) GetFileContent(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockDropboxClient) GetChangesLast24Hours(ctx context.Context) ([]*models.FileMetadata, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

func (m *mockDropboxClient) GetChangesLast10Minutes(ctx context.Context) ([]*models.FileMetadata, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

func (m *mockDropboxClient) GetChanges(ctx context.Context) ([]*models.FileMetadata, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FileMetadata), args.Error(1)
}

func (m *mockDropboxClient) GetFileChanges(ctx context.Context) ([]models.FileChange, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.FileChange), args.Error(1)
}

// Ensure mockDropboxClient implements interfaces.DropboxClient
var _ interfaces.DropboxClient = (*mockDropboxClient)(nil)
