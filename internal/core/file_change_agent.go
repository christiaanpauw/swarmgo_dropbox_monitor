package core

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agent"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/interfaces"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// FileChangeAgentImpl monitors Dropbox for file changes
type FileChangeAgentImpl struct {
	*lifecycle.BaseComponent
	dropboxClient interfaces.DropboxClient
	stateManager  interfaces.StateManager
	pollInterval  time.Duration
	monitorPath   string
	stopCh        chan struct{}
	mu           sync.RWMutex
}

// NewFileChangeAgent creates a new file change agent
func NewFileChangeAgent(client interfaces.DropboxClient, stateManager interfaces.StateManager, monitorPath string) agent.FileChangeAgent {
	agent := &FileChangeAgentImpl{
		BaseComponent: lifecycle.NewBaseComponent("FileChangeAgent"),
		dropboxClient: client,
		stateManager:  stateManager,
		pollInterval:  5 * time.Minute, // Default poll interval
		monitorPath:   monitorPath,
		stopCh:        make(chan struct{}),
	}
	agent.SetState(lifecycle.StateInitialized)
	return agent
}

// Start starts the file change monitoring
func (a *FileChangeAgentImpl) Start(ctx context.Context) error {
	if err := a.DefaultStart(ctx); err != nil {
		return err
	}

	log.Printf("🔍 Starting FileChangeAgent...")

	// Start monitoring in a goroutine
	go a.monitorChanges(ctx)

	return nil
}

// Stop stops the file change monitoring
func (a *FileChangeAgentImpl) Stop(ctx context.Context) error {
	if err := a.DefaultStop(ctx); err != nil {
		return err
	}

	log.Printf("🛑 Stopping FileChangeAgent...")
	close(a.stopCh)

	return nil
}

// Health checks the health of the file change agent
func (a *FileChangeAgentImpl) Health(ctx context.Context) error {
	if err := a.DefaultHealth(ctx); err != nil {
		return err
	}

	// Try to list files to verify Dropbox connection
	_, err := a.dropboxClient.ListFolder(ctx, a.monitorPath)
	if err != nil {
		return fmt.Errorf("failed to list folder: %w", err)
	}

	return nil
}

// GetChanges returns a list of file changes
func (a *FileChangeAgentImpl) GetChanges(ctx context.Context) ([]models.FileChange, error) {
	// Get the current cursor from state
	cursor := a.stateManager.GetString("cursor")

	// List files from Dropbox
	files, err := a.dropboxClient.ListFolder(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list folder: %w", err)
	}

	// Compare with previous state
	changes := a.detectChanges(files, cursor)

	// Update state with new cursor
	if len(changes) > 0 {
		if err := a.stateManager.SetString("cursor", cursor); err != nil {
			return nil, fmt.Errorf("failed to update cursor: %w", err)
		}
	}

	return changes, nil
}

// GetFileContent returns the content of a file
func (a *FileChangeAgentImpl) GetFileContent(ctx context.Context, path string) ([]byte, error) {
	return a.dropboxClient.GetFileContent(ctx, path)
}

// SetPollInterval sets the polling interval
func (a *FileChangeAgentImpl) SetPollInterval(interval time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pollInterval = interval
}

// monitorChanges polls Dropbox for changes
func (a *FileChangeAgentImpl) monitorChanges(ctx context.Context) {
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		case <-ticker.C:
			if err := a.checkForChanges(ctx); err != nil {
				log.Printf("Error checking for changes: %v", err)
			}
		}
	}
}

// checkForChanges checks for changes in Dropbox
func (a *FileChangeAgentImpl) checkForChanges(ctx context.Context) error {
	changes, err := a.GetChanges(ctx)
	if err != nil {
		return fmt.Errorf("failed to get changes: %w", err)
	}

	if len(changes) > 0 {
		if err := a.processChanges(ctx, changes); err != nil {
			return fmt.Errorf("failed to process changes: %w", err)
		}
	}

	return nil
}

// detectChanges compares current files with previous state
func (a *FileChangeAgentImpl) detectChanges(files []*models.FileMetadata, cursor string) []models.FileChange {
	return models.BatchConvertMetadataToChanges(files)
}

// processChanges handles detected changes
func (a *FileChangeAgentImpl) processChanges(ctx context.Context, changes []models.FileChange) error {
	for _, change := range changes {
		log.Printf("Processing change: %+v", change)
	}
	return nil
}
