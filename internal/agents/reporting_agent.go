package agents

import (
	"context"
	"fmt"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting"
)

// reportingAgent is responsible for generating reports about file changes
type reportingAgent struct {
	reporter reporting.Reporter
}

// NewReportingAgent creates a new reporting agent
func NewReportingAgent(notifier notify.Notifier) *reportingAgent {
	return &reportingAgent{
		reporter: reporting.NewReporter(notifier),
	}
}

// GenerateReport generates a report for the given file changes
func (ra *reportingAgent) GenerateReport(ctx context.Context, changes []models.FileMetadata) (*models.Report, error) {
	return ra.reporter.GenerateReport(ctx, changes)
}

// GenerateHTMLReport generates an HTML report for the given file changes
func (ra *reportingAgent) GenerateHTMLReport(ctx context.Context, changes []models.FileMetadata) (string, error) {
	report, err := ra.GenerateReport(ctx, changes)
	if err != nil {
		return "", fmt.Errorf("failed to generate report: %w", err)
	}

	html, err := ra.reporter.GenerateHTML(ctx, report)
	if err != nil {
		return "", fmt.Errorf("failed to generate HTML: %w", err)
	}

	return html, nil
}
