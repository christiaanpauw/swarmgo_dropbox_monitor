package agent

import (
	"context"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// Agent is the base interface for all agents
type Agent interface {
	lifecycle.Component
}

// FileChangeAgent monitors file changes
type FileChangeAgent interface {
	Agent
	GetChanges(ctx context.Context) ([]models.FileChange, error)
	GetFileContent(ctx context.Context, path string) ([]byte, error)
	SetPollInterval(interval time.Duration)
}

// DatabaseAgent handles database operations
type DatabaseAgent interface {
	Agent
	StoreChange(ctx context.Context, change models.FileMetadata) error
	GetLatestChanges(ctx context.Context, limit int) ([]models.FileMetadata, error)
	GetChanges(ctx context.Context, startTime, endTime string) ([]models.FileMetadata, error)
}

// ReportingAgent handles report generation and notifications
type ReportingAgent interface {
	Agent
	GenerateReport(ctx context.Context, changes []models.FileChange) error
	NotifyChanges(ctx context.Context, changes []models.FileChange) error
}
