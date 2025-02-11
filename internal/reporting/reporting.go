package reporting

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
)

// ReportType defines the type of report to generate
type ReportType string

const (
	ReportTypeBasic     ReportType = "basic"
	ReportTypeNarrative ReportType = "narrative"
)

// Reporter defines the interface for generating reports
type Reporter interface {
	// GenerateReport creates a report from file changes
	GenerateReport(ctx context.Context, changes []models.FileMetadata) (*models.Report, error)
	// GenerateHTML generates an HTML version of the report
	GenerateHTML(ctx context.Context, report *models.Report) (string, error)
	// SendReport sends the report via the configured notification method
	SendReport(ctx context.Context, report *models.Report) error
}

// NarrativeReporter extends Reporter with narrative-specific functionality
type NarrativeReporter interface {
	Reporter
	// AnalyzeActivity analyzes patterns in the changes
	AnalyzeActivity(ctx context.Context, report *models.Report) (*models.ActivityPattern, error)
	// GenerateNarrative creates a human-friendly narrative
	GenerateNarrative(ctx context.Context, report *models.Report, pattern *models.ActivityPattern) (string, error)
}

// reporter implements both Reporter and NarrativeReporter interfaces
type reporter struct {
	notifier notify.Notifier
}

// NewReporter creates a new reporter instance
func NewReporter(notifier notify.Notifier) Reporter {
	return &reporter{notifier: notifier}
}

// NewNarrativeReporter creates a new narrative reporter instance
func NewNarrativeReporter(notifier notify.Notifier) NarrativeReporter {
	return &reporter{notifier: notifier}
}

// GenerateReport creates a report from file changes
func (r *reporter) GenerateReport(ctx context.Context, changes []models.FileMetadata) (*models.Report, error) {
	if len(changes) == 0 {
		return &models.Report{
			Changes:     make([]models.FileChange, 0),
			GeneratedAt: time.Now(),
		}, nil
	}

	report := &models.Report{
		Changes:        make([]models.FileChange, 0, len(changes)),
		ExtensionCount: make(map[string]int),
		DirectoryCount: make(map[string]int),
		GeneratedAt:    time.Now(),
		TotalChanges:   len(changes),
	}

	for _, change := range changes {
		fileChange := models.FileChange{
			Path:      change.Path,
			Extension: strings.TrimPrefix(filepath.Ext(change.Path), "."),
			Directory: filepath.Dir(change.Path),
		}
		report.Changes = append(report.Changes, fileChange)
		report.ExtensionCount[fileChange.Extension]++
		report.DirectoryCount[fileChange.Directory]++
	}

	return report, nil
}

// GenerateHTML generates an HTML version of the report
func (r *reporter) GenerateHTML(ctx context.Context, report *models.Report) (string, error) {
	var html strings.Builder

	html.WriteString("<html><body>")
	html.WriteString("<h1>Dropbox Changes Report</h1>")
	html.WriteString(fmt.Sprintf("<p>Generated at: %s</p>", report.GeneratedAt.Format(time.RFC3339)))
	html.WriteString(fmt.Sprintf("<p>Total Changes: %d</p>", report.TotalChanges))

	html.WriteString("<h2>Changes by Extension</h2><ul>")
	for ext, count := range report.ExtensionCount {
		html.WriteString(fmt.Sprintf("<li>%s: %d files</li>", ext, count))
	}
	html.WriteString("</ul>")

	html.WriteString("<h2>Changes by Directory</h2><ul>")
	for dir, count := range report.DirectoryCount {
		html.WriteString(fmt.Sprintf("<li>%s: %d files</li>", dir, count))
	}
	html.WriteString("</ul>")

	html.WriteString("<h2>Changed Files</h2><ul>")
	for _, change := range report.Changes {
		html.WriteString(fmt.Sprintf("<li>%s</li>", change.Path))
	}
	html.WriteString("</ul>")

	html.WriteString("</body></html>")
	return html.String(), nil
}

// SendReport sends the report via the configured notification method
func (r *reporter) SendReport(ctx context.Context, report *models.Report) error {
	html, err := r.GenerateHTML(ctx, report)
	if err != nil {
		return fmt.Errorf("failed to generate HTML report: %w", err)
	}

	return r.notifier.Send(ctx, "Dropbox Changes Report", html)
}

// AnalyzeActivity analyzes patterns in the changes
func (r *reporter) AnalyzeActivity(ctx context.Context, report *models.Report) (*models.ActivityPattern, error) {
	pattern := &models.ActivityPattern{
		MainDirectories: getTopItems(report.DirectoryCount, 5),
		FileTypes:      getTopItems(report.ExtensionCount, 5),
		TotalChanges:   report.TotalChanges,
	}
	return pattern, nil
}

// GenerateNarrative creates a human-friendly narrative
func (r *reporter) GenerateNarrative(ctx context.Context, report *models.Report, pattern *models.ActivityPattern) (string, error) {
	var narrative strings.Builder

	// Generate summary
	narrative.WriteString(fmt.Sprintf("Activity Report for %s\n\n", report.GeneratedAt.Format("January 2, 2006")))
	narrative.WriteString(fmt.Sprintf("Total Changes: %d\n\n", pattern.TotalChanges))

	// Add directory analysis
	if len(pattern.MainDirectories) > 0 {
		narrative.WriteString("Most Active Directories:\n")
		for _, dir := range pattern.MainDirectories {
			narrative.WriteString(fmt.Sprintf("- %s (%d changes)\n", dir, report.DirectoryCount[dir]))
		}
		narrative.WriteString("\n")
	}

	// Add file type analysis
	if len(pattern.FileTypes) > 0 {
		narrative.WriteString("Most Changed File Types:\n")
		for _, ext := range pattern.FileTypes {
			narrative.WriteString(fmt.Sprintf("- %s files (%d changes)\n", ext, report.ExtensionCount[ext]))
		}
		narrative.WriteString("\n")
	}

	return narrative.String(), nil
}

// Helper functions

func getTopItems(items map[string]int, n int) []string {
	type kv struct {
		Key   string
		Value int
	}

	var kvs []kv
	for k, v := range items {
		kvs = append(kvs, kv{k, v})
	}

	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Value > kvs[j].Value
	})

	result := make([]string, 0, n)
	for i := 0; i < n && i < len(kvs); i++ {
		result = append(result, kvs[i].Key)
	}

	return result
}
