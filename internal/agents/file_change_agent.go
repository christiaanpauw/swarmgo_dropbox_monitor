package agents

import (
	"context"
	"log"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agent"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/core"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/interfaces"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// fileChangeAgentImpl implements FileChangeAgent
type fileChangeAgentImpl struct {
	agent.FileChangeAgent // Embed the core agent
}

// NewFileChangeAgent creates a new file change agent
func NewFileChangeAgent(client interfaces.DropboxClient, stateManager interfaces.StateManager, monitorPath string) agent.FileChangeAgent {
	baseAgent := core.NewFileChangeAgent(client, stateManager, monitorPath)
	return &fileChangeAgentImpl{
		FileChangeAgent: baseAgent,
	}
}

// Start starts the file change monitoring
func (a *fileChangeAgentImpl) Start(ctx context.Context) error {
	log.Printf(" Starting FileChangeAgent...")
	return a.FileChangeAgent.Start(ctx)
}

// Stop stops the file change monitoring
func (a *fileChangeAgentImpl) Stop(ctx context.Context) error {
	return a.FileChangeAgent.Stop(ctx)
}

// Health checks the health of the file change agent
func (a *fileChangeAgentImpl) Health(ctx context.Context) error {
	return a.FileChangeAgent.Health(ctx)
}

// GetChanges returns a list of file changes
func (a *fileChangeAgentImpl) GetChanges(ctx context.Context) ([]models.FileChange, error) {
	return a.FileChangeAgent.GetChanges(ctx)
}

// GetFileContent returns the content of a file
func (a *fileChangeAgentImpl) GetFileContent(ctx context.Context, path string) ([]byte, error) {
	return a.FileChangeAgent.GetFileContent(ctx, path)
}

// SetPollInterval sets the polling interval
func (a *fileChangeAgentImpl) SetPollInterval(interval time.Duration) {
	a.FileChangeAgent.SetPollInterval(interval)
}
