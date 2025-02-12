package agents

import (
	"context"
	"fmt"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting"
)

// ReportingAgent interface for generating and sending reports
type ReportingAgent interface {
	lifecycle.Component
	Initialize(ctx context.Context) error
	GenerateReport(ctx context.Context, changes []models.FileChange) error
	NotifyChanges(ctx context.Context, changes []models.FileChange) error
}

// reportingAgent implements the ReportingAgent interface
type reportingAgent struct {
	*lifecycle.BaseComponent
	notifier notify.Notifier
	reporter reporting.Reporter
}

// NewReportingAgent creates a new reporting agent
func NewReportingAgent(notifier notify.Notifier) (ReportingAgent, error) {
	if notifier == nil {
		return nil, fmt.Errorf("notifier cannot be nil")
	}

	reporter, err := reporting.NewReporter(notifier)
	if err != nil {
		return nil, fmt.Errorf("failed to create reporter: %w", err)
	}

	agent := &reportingAgent{
		BaseComponent: lifecycle.NewBaseComponent("ReportingAgent"),
		notifier:      notifier,
		reporter:      reporter,
	}
	agent.SetState(lifecycle.StateInitialized)
	return agent, nil
}

// GenerateReport generates and sends a report for file changes
func (a *reportingAgent) GenerateReport(ctx context.Context, changes []models.FileChange) error {
	if a.State() != lifecycle.StateRunning {
		return fmt.Errorf("reporting agent is not running")
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	if len(changes) == 0 {
		return nil // No changes to report
	}

	// Generate all report types
	reportTypes := []models.ReportType{
		models.FileListReport,
		models.HTMLReport,
		models.NarrativeReport,
	}

	for _, reportType := range reportTypes {
		report, err := a.reporter.GenerateReport(ctx, changes, reportType)
		if err != nil {
			return fmt.Errorf("failed to generate %s report: %w", reportType, err)
		}

		// Send the generated report
		if err := a.reporter.SendReport(ctx, report); err != nil {
			return fmt.Errorf("failed to send %s report: %w", reportType, err)
		}
	}

	return nil
}

// NotifyChanges notifies about file changes
func (a *reportingAgent) NotifyChanges(ctx context.Context, changes []models.FileChange) error {
	return a.GenerateReport(ctx, changes)
}

// Initialize implements lifecycle.Component
func (a *reportingAgent) Initialize(ctx context.Context) error {
	currentState := a.State()
	if currentState != lifecycle.StateUninitialized && currentState != lifecycle.StateInitialized {
		return fmt.Errorf("reporting agent is in invalid state for initialization: %s", currentState)
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	// Validate dependencies
	if a.notifier == nil {
		return fmt.Errorf("notifier is not set")
	}
	if a.reporter == nil {
		return fmt.Errorf("reporter is not set")
	}

	// Initialize the underlying reporter if needed
	if a.reporter.State() == lifecycle.StateUninitialized {
		if err := a.reporter.Start(ctx); err != nil {
			return fmt.Errorf("failed to initialize reporter: %w", err)
		}
	}

	a.SetState(lifecycle.StateInitialized)
	return nil
}

// Start implements lifecycle.Component
func (a *reportingAgent) Start(ctx context.Context) error {
	if err := a.DefaultStart(ctx); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

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

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

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

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	if err := a.reporter.Health(ctx); err != nil {
		return fmt.Errorf("reporter health check failed: %w", err)
	}

	return nil
}
