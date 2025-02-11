package agents

import (
	"context"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// DatabaseAgent handles storage of file changes and analysis
type DatabaseAgent interface {
	StoreChange(ctx context.Context, change models.FileMetadata) error
	StoreAnalysis(ctx context.Context, content models.FileContent) error
	Close() error
}
