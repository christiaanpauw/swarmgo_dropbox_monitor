package agents

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// FileChangeAgent interface for detecting file changes
type FileChangeAgent interface {
	GetChanges(ctx context.Context) ([]*models.FileChange, error)
	DetectChanges(ctx context.Context) ([]*models.FileChange, error)
	GetFileContent(ctx context.Context, path string) ([]byte, error)
}

// fileChangeAgent implements the FileChangeAgent interface
type fileChangeAgent struct {
	client dropbox.Client
}

// NewFileChangeAgent creates a new file change agent
func NewFileChangeAgent(token string) (FileChangeAgent, error) {
	client, err := dropbox.NewDropboxClient(token)
	if err != nil {
		return nil, fmt.Errorf("create dropbox client: %w", err)
	}

	return &fileChangeAgent{
		client: client,
	}, nil
}

// GetChanges gets the latest changes from Dropbox
func (a *fileChangeAgent) GetChanges(ctx context.Context) ([]*models.FileChange, error) {
	return a.DetectChanges(ctx)
}

// DetectChanges detects changes in Dropbox files
func (a *fileChangeAgent) DetectChanges(ctx context.Context) ([]*models.FileChange, error) {
	entries, err := a.client.ListFolder(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("list folder: %w", err)
	}

	var changes []*models.FileChange
	for _, entry := range entries {
		changes = append(changes, &models.FileChange{
			Path:      entry.Path,
			Extension: filepath.Ext(entry.Path),
			Directory: filepath.Dir(entry.Path),
			ModTime:   entry.Modified,
			IsDeleted: entry.IsDeleted,
		})
	}

	return changes, nil
}

// Execute implements the Agent interface
func (a *fileChangeAgent) Execute(ctx context.Context) error {
	changes, err := a.GetChanges(ctx)
	if err != nil {
		return fmt.Errorf("get changes: %w", err)
	}

	// Process changes
	for _, change := range changes {
		if err := a.processChange(ctx, change); err != nil {
			return fmt.Errorf("process change %s: %w", change.Path, err)
		}
	}

	return nil
}

// processChange processes a single file change
func (a *fileChangeAgent) processChange(ctx context.Context, change *models.FileChange) error {
	if change.IsDeleted {
		// Handle deleted file
		return nil
	}

	// Get file content is handled by the agent manager
	return nil
}

// GetFileContent gets file content from Dropbox
func (a *fileChangeAgent) GetFileContent(ctx context.Context, path string) ([]byte, error) {
	return a.client.GetFileContent(ctx, path)
}
