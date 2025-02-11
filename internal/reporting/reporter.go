package reporting

import (
	"context"
	"fmt"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting/generators"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting/models"
)

// Reporter interface defines methods for generating and sending reports
type Reporter interface {
	lifecycle.Component
	GenerateReport(ctx context.Context, changes []models.FileChange) (*models.Report, error)
	SendReport(ctx context.Context, report *models.Report) error
}

// reporter implements the Reporter interface
type reporter struct {
	*lifecycle.BaseComponent
	notifier notify.Notifier
}

// NewReporter creates a new Reporter instance
func NewReporter(notifier notify.Notifier) Reporter {
	r := &reporter{
		BaseComponent: lifecycle.NewBaseComponent("Reporter"),
		notifier:     notifier,
	}
	r.SetState(lifecycle.StateInitialized)
	return r
}

// GenerateReport generates a report from the given file changes
func (r *reporter) GenerateReport(ctx context.Context, changes []models.FileChange) (*models.Report, error) {
	report := models.NewReport(models.FileListReport)
	for _, change := range changes {
		report.AddChange(change)
	}
	return report, nil
}

// SendReport sends the report using the configured notifier
func (r *reporter) SendReport(ctx context.Context, report *models.Report) error {
	if r.State() != lifecycle.StateRunning {
		return fmt.Errorf("reporter is not running")
	}

	// Generate HTML report
	html, err := generators.GenerateHTML(report, nil) // TODO: Add activity pattern
	if err != nil {
		return fmt.Errorf("failed to generate HTML report: %w", err)
	}

	// Send report via notifier
	subject := "Dropbox File Changes Report"
	if err := r.notifier.Send(ctx, subject, html); err != nil {
		return fmt.Errorf("failed to send report: %w", err)
	}

	return nil
}

// Start implements lifecycle.Component
func (r *reporter) Start(ctx context.Context) error {
	return r.DefaultStart(ctx)
}

// Stop implements lifecycle.Component
func (r *reporter) Stop(ctx context.Context) error {
	return r.DefaultStop(ctx)
}

// Health implements lifecycle.Component
func (r *reporter) Health(ctx context.Context) error {
	return r.DefaultHealth(ctx)
}
