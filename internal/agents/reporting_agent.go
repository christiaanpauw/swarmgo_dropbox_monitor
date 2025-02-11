package agents

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting"
)

// ReportingAgent interface for generating and sending reports
type ReportingAgent interface {
	GenerateReport(ctx context.Context, changes []*models.FileChange) error
}

// reportingAgent implements the ReportingAgent interface
type reportingAgent struct {
	reporter reporting.Reporter
	notifier notify.Notifier
}

// NewReportingAgent creates a new reporting agent
func NewReportingAgent(notifier notify.Notifier) ReportingAgent {
	return &reportingAgent{
		reporter: reporting.NewReporter(notifier),
		notifier: notifier,
	}
}

// GenerateReport generates and sends a report for file changes
func (a *reportingAgent) GenerateReport(ctx context.Context, changes []*models.FileChange) error {
	if len(changes) == 0 {
		return nil
	}

	// Convert FileChange to FileMetadata for backward compatibility
	var metadata []models.FileMetadata
	for _, change := range changes {
		metadata = append(metadata, models.FileMetadata{
			Path:      change.Path,
			Name:      filepath.Base(change.Path),
			Modified:  change.ModTime,
			IsDeleted: change.IsDeleted,
		})
	}

	// Generate report
	report, err := a.reporter.GenerateReport(ctx, metadata)
	if err != nil {
		return fmt.Errorf("generate report: %w", err)
	}

	// Format report content
	var reportContent string
	if report.Changes != nil {
		reportContent = fmt.Sprintf("Changes detected:\n\n")
		for _, change := range report.Changes {
			status := "Modified"
			if change.IsDeleted {
				status = "Deleted"
			}
			reportContent += fmt.Sprintf("- %s: %s (%s)\n", status, change.Path, change.ModTime.Format("2006-01-02 15:04:05"))
		}
	} else {
		reportContent = "No changes detected"
	}

	// Send report via email
	subject := "Dropbox Monitor Report"
	if err := a.notifier.Send(ctx, subject, reportContent); err != nil {
		return fmt.Errorf("send report: %w", err)
	}

	return nil
}
