package agents

import (
	"context"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// FileChangeAgent detects changes in files
type FileChangeAgent interface {
	DetectChanges(ctx context.Context, timeWindow string) ([]models.FileMetadata, error)
}
