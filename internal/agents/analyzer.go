package agents

import (
	"context"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// ContentAnalyzer analyzes file content
type ContentAnalyzer interface {
	AnalyzeContent(ctx context.Context, path string, content []byte) (*models.FileContent, error)
}
