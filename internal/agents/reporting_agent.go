package agents

import (
	"context"
	"fmt"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting/models"
)

// ReportingAgent interface for generating and sending reports
type ReportingAgent interface {
	lifecycle.Component
	GenerateReport(ctx context.Context, changes []models.FileChange) error
}

// reportingAgent implements the ReportingAgent interface
type reportingAgent struct {
	*lifecycle.BaseComponent
	reporter reporting.Reporter
	notifier notify.Notifier
}

// NewReportingAgent creates a new reporting agent
func NewReportingAgent(notifier notify.Notifier) ReportingAgent {
	agent := &reportingAgent{
		BaseComponent: lifecycle.NewBaseComponent("ReportingAgent"),
		reporter:     reporting.NewReporter(notifier),
		notifier:     notifier,
	}
	agent.SetState(lifecycle.StateInitialized)
	return agent
}

// GenerateReport generates and sends a report for file changes
func (a *reportingAgent) GenerateReport(ctx context.Context, changes []models.FileChange) error {
	if a.State() != lifecycle.StateRunning {
		return fmt.Errorf("reporting agent is not running")
	}

	if len(changes) == 0 {
		return nil
	}

	// Generate report using the reporter
	report, err := a.reporter.GenerateReport(ctx, changes)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Send the report
	if err := a.reporter.SendReport(ctx, report); err != nil {
		return fmt.Errorf("failed to send report: %w", err)
	}

	return nil
}

// Start implements lifecycle.Component
func (a *reportingAgent) Start(ctx context.Context) error {
	if err := a.DefaultStart(ctx); err != nil {
		return err
	}

	// Start the reporter
	if err := a.reporter.Start(ctx); err != nil {
		return fmt.Errorf("failed to start reporter: %w", err)
	}

	return nil
}

// Stop implements lifecycle.Component
func (a *reportingAgent) Stop(ctx context.Context) error {
	if err := a.DefaultStop(ctx); err != nil {
		return err
	}

	// Stop the reporter
	if err := a.reporter.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop reporter: %w", err)
	}

	return nil
}

// Health implements lifecycle.Component
func (a *reportingAgent) Health(ctx context.Context) error {
	if err := a.DefaultHealth(ctx); err != nil {
		return err
	}

	// Check reporter health
	if err := a.reporter.Health(ctx); err != nil {
		return fmt.Errorf("reporter health check failed: %w", err)
	}

	return nil
}
