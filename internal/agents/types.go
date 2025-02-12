package agents

import (
	"context"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// StateManager defines the interface for state management
type StateManager interface {
	GetString(key string) string
	SetString(key, value string) error
}

// FileChangeHandler is a function that handles file changes
type FileChangeHandler func(context.Context, []models.FileChange) error

// FileChangeProcessor processes file changes
type FileChangeProcessor interface {
	ProcessFileChanges(context.Context, []models.FileChange) error
}

// FileChangeProcessorFunc is a function that implements FileChangeProcessor
type FileChangeProcessorFunc func(context.Context, []models.FileChange) error

// ProcessFileChanges implements FileChangeProcessor
func (f FileChangeProcessorFunc) ProcessFileChanges(ctx context.Context, changes []models.FileChange) error {
	return f(ctx, changes)
}

// FileChangeConfig holds configuration for file change monitoring
type FileChangeConfig struct {
	PollInterval time.Duration
	MaxRetries   int
	RetryDelay   time.Duration
}
