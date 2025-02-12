package generators

import (
	"context"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// Generator defines the interface for report generators
type Generator interface {
	Generate(ctx context.Context, report *models.Report) error
}
