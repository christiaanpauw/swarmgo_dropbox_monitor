package interfaces

import (
	"context"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// DropboxClient defines the interface for Dropbox operations
type DropboxClient interface {
	ListFolder(ctx context.Context, path string) ([]*models.FileMetadata, error)
	GetFileContent(ctx context.Context, path string) ([]byte, error)
	GetChangesLast24Hours(ctx context.Context) ([]*models.FileMetadata, error)
	GetChangesLast10Minutes(ctx context.Context) ([]*models.FileMetadata, error)
	GetChanges(ctx context.Context) ([]*models.FileMetadata, error)
}
