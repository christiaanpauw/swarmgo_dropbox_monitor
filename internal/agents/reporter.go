package agents

import (
	"context"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// ReportingAgent generates reports about file changes
type ReportingAgent interface {
	GenerateReport(ctx context.Context, changes []models.FileMetadata) (*models.Report, error)
	GenerateHTMLReport(ctx context.Context, changes []models.FileMetadata) (string, error)
}
