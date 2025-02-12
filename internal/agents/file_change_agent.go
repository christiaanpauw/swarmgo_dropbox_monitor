package agents

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/core"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// FileChangeAgent manages file change detection
type FileChangeAgent struct {
	*lifecycle.BaseComponent
	client       *dropbox.DropboxClient
	stateManager *core.StateManager
}

// NewFileChangeAgent creates a new file change agent
func NewFileChangeAgent(client *dropbox.DropboxClient, stateManager *core.StateManager) *FileChangeAgent {
	return &FileChangeAgent{
		BaseComponent: lifecycle.NewBaseComponent("FileChangeAgent"),
		client:       client,
		stateManager: stateManager,
	}
}

// Start implements lifecycle.Component
func (a *FileChangeAgent) Start(ctx context.Context) error {
	if err := a.DefaultStart(ctx); err != nil {
		return err
	}

	log.Printf(" Starting FileChangeAgent...")
	return nil
}

// Stop implements lifecycle.Component
func (a *FileChangeAgent) Stop(ctx context.Context) error {
	log.Printf(" Stopping FileChangeAgent...")
	return a.DefaultStop(ctx)
}

// Health implements lifecycle.Component
func (a *FileChangeAgent) Health(ctx context.Context) error {
	if err := a.DefaultHealth(ctx); err != nil {
		return err
	}

	if a.client == nil {
		return fmt.Errorf("dropbox client is nil")
	}

	if a.stateManager == nil {
		return fmt.Errorf("state manager is nil")
	}

	return nil
}

// GetChanges gets the latest changes from Dropbox
func (a *FileChangeAgent) GetChanges(ctx context.Context) ([]*models.FileChange, error) {
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

// DetectChanges detects changes in Dropbox files
func (a *FileChangeAgent) DetectChanges(ctx context.Context) ([]*models.FileChange, error) {
	return a.GetChanges(ctx)
}

// Execute implements the Agent interface
func (a *FileChangeAgent) Execute(ctx context.Context) error {
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
func (a *FileChangeAgent) processChange(ctx context.Context, change *models.FileChange) error {
	if change.IsDeleted {
		// Handle deleted file
		return nil
	}

	// Get file content is handled by the agent manager
	return nil
}

// GetFileContent gets file content from Dropbox
func (a *FileChangeAgent) GetFileContent(ctx context.Context, path string) ([]byte, error) {
	return a.client.GetFileContent(ctx, path)
}
