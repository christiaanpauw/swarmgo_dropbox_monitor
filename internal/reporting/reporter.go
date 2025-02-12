package reporting

import (
	"context"
	"fmt"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting/generators"
)

// Reporter interface defines methods for generating and sending reports
type Reporter interface {
	lifecycle.Component
	GenerateReport(ctx context.Context, changes []models.FileChange, reportType models.ReportType) (*models.Report, error)
	SendReport(ctx context.Context, report *models.Report) error
}

// reporter implements the Reporter interface
type reporter struct {
	*lifecycle.BaseComponent
	notifier notify.Notifier
	generators map[models.ReportType]generators.Generator
}

// NewReporter creates a new Reporter instance
func NewReporter(notifier notify.Notifier) (Reporter, error) {
	if notifier == nil {
		return nil, fmt.Errorf("notifier cannot be nil")
	}

	r := &reporter{
		BaseComponent: lifecycle.NewBaseComponent("Reporter"),
		notifier:     notifier,
		generators:   make(map[models.ReportType]generators.Generator),
	}
	r.SetState(lifecycle.StateInitialized)

	// Register default generators
	r.generators[models.FileListReport] = generators.NewFileListGenerator()
	r.generators[models.NarrativeReport] = generators.NewNarrativeGenerator()
	r.generators[models.HTMLReport] = generators.NewHTMLGenerator()

	return r, nil
}

// GenerateReport generates a report from the given file changes
func (r *reporter) GenerateReport(ctx context.Context, changes []models.FileChange, reportType models.ReportType) (*models.Report, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	if changes == nil {
		changes = []models.FileChange{} // Use empty slice instead of nil
	}

	generator, ok := r.generators[reportType]
	if !ok {
		return nil, fmt.Errorf("unsupported report type: %s", reportType)
	}

	report := models.NewReport(reportType)
	report.GeneratedAt = time.Now()
	for _, change := range changes {
		report.AddChange(change)
	}

	if err := generator.Generate(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	return report, nil
}

// SendReport sends the report using the configured notifier
func (r *reporter) SendReport(ctx context.Context, report *models.Report) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	if report == nil {
		return fmt.Errorf("report cannot be nil")
	}

	if report.Metadata == nil || report.Metadata["content"] == "" {
		return fmt.Errorf("report has no content")
	}

	// Format report message
	message := fmt.Sprintf("Dropbox Changes Report - %s\n\n%s", 
		report.GeneratedAt.Format("2006-01-02 15:04:05"),
		report.Metadata["content"])

	// Send report via notifier
	if err := r.notifier.SendNotification(ctx, message); err != nil {
		return fmt.Errorf("failed to send report: %w", err)
	}

	return nil
}

// Start implements lifecycle.Component
func (r *reporter) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	r.SetState(lifecycle.StateRunning)
	return nil
}

// Stop implements lifecycle.Component
func (r *reporter) Stop(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	r.SetState(lifecycle.StateStopped)
	return nil
}

// Health implements lifecycle.Component
func (r *reporter) Health(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	if r.State() != lifecycle.StateRunning {
		return fmt.Errorf("reporter is not running")
	}
	return nil
}

// RegisterGenerator registers a custom generator for a report type
func (r *reporter) RegisterGenerator(reportType models.ReportType, generator generators.Generator) error {
	if generator == nil {
		return fmt.Errorf("generator cannot be nil")
	}

	r.generators[reportType] = generator
	return nil
}
