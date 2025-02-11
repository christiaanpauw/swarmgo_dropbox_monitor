package report

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

// ActivityPattern represents a pattern of activity in the Dropbox
type ActivityPattern struct {
	MainDirectories []string    // Most active directories
	FileTypes       []string    // Most changed file types
	BusyPeriods     []time.Time // Times with most activity
	TotalChanges    int
	FileContents    []models.FileContent // Added field for file contents
}

// NarrativeReport generates a human-friendly narrative of Dropbox activity
type NarrativeReport struct {
	Period          string // "day", "hour", "10min"
	Since           time.Time
	Until           time.Time
	Changes         []models.FileChange
	ActivityPattern ActivityPattern
}

// NewNarrativeReport creates a new narrative report for the specified time period
func NewNarrativeReport(period string, changes []models.FileChange, since time.Time) *NarrativeReport {
	r := &NarrativeReport{
		Period:  period,
		Since:   since,
		Until:   time.Now(),
		Changes: changes,
	}

	r.analyzeActivity()
	return r
}

// analyzeActivity analyzes the changes to find patterns
func (r *NarrativeReport) analyzeActivity() {
	dirCount := make(map[string]int)
	extCount := make(map[string]int)

	for _, change := range r.Changes {
		// Count directories
		dir := filepath.Dir(change.Path)
		dirCount[dir]++

		// Count extensions
		ext := filepath.Ext(change.Path)
		if ext != "" {
			ext = strings.TrimPrefix(ext, ".")
			extCount[ext]++
		}
	}

	// Get top directories and file types
	r.ActivityPattern.MainDirectories = r.getTopKeys(dirCount, 5)
	r.ActivityPattern.FileTypes = r.getTopKeys(extCount, 5)
	r.ActivityPattern.TotalChanges = len(r.Changes)
}

// getTopKeys returns the top n keys by value from a map
func (r *NarrativeReport) getTopKeys(m map[string]int, n int) []string {
	type kv struct {
		Key   string
		Value int
	}

	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}

	// Sort by value in descending order
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	// Get top n keys
	var result []string
	for i := 0; i < len(ss) && i < n; i++ {
		result = append(result, ss[i].Key)
	}

	return result
}

// GenerateNarrative creates a human-friendly narrative of the changes
func (r *NarrativeReport) GenerateNarrative() string {
	var narrative strings.Builder

	// Add header
	narrative.WriteString(fmt.Sprintf("Dropbox Activity Report - %s\n\n", r.formatTimeRange()))

	// Add summary
	narrative.WriteString(r.generateSummary())
	narrative.WriteString("\n\n")

	// Add analysis
	narrative.WriteString(r.generateAnalysis())
	narrative.WriteString("\n\n")

	// Add insights
	narrative.WriteString(r.generateInsights())

	return narrative.String()
}

// formatTimeRange formats the time range for display
func (r *NarrativeReport) formatTimeRange() string {
	return fmt.Sprintf("%s to %s",
		r.Since.Format("2006-01-02 15:04:05"),
		r.Until.Format("2006-01-02 15:04:05"))
}

// generateSummary generates a summary of the activity
func (r *NarrativeReport) generateSummary() string {
	var summary strings.Builder

	summary.WriteString("Summary:\n")
	summary.WriteString(fmt.Sprintf("- Total Changes: %d\n", r.ActivityPattern.TotalChanges))

	if len(r.ActivityPattern.MainDirectories) > 0 {
		summary.WriteString("- Most Active Directories:\n")
		for _, dir := range r.ActivityPattern.MainDirectories {
			summary.WriteString(fmt.Sprintf("  * %s\n", dir))
		}
	}

	if len(r.ActivityPattern.FileTypes) > 0 {
		summary.WriteString("- File Types Changed:\n")
		for _, ext := range r.ActivityPattern.FileTypes {
			summary.WriteString(fmt.Sprintf("  * %s\n", ext))
		}
	}

	return summary.String()
}

// generateAnalysis generates an analysis of the activity patterns
func (r *NarrativeReport) generateAnalysis() string {
	var analysis strings.Builder

	analysis.WriteString("Analysis:\n")

	// Analyze directory patterns
	if len(r.ActivityPattern.MainDirectories) > 0 {
		analysis.WriteString("Directory Activity:\n")
		for _, dir := range r.ActivityPattern.MainDirectories {
			analysis.WriteString(fmt.Sprintf("- %s: High activity in this directory\n", dir))
		}
	}

	// Analyze file type patterns
	if len(r.ActivityPattern.FileTypes) > 0 {
		analysis.WriteString("\nFile Type Patterns:\n")
		for _, ext := range r.ActivityPattern.FileTypes {
			analysis.WriteString(fmt.Sprintf("- %s files were frequently modified\n", ext))
		}
	}

	return analysis.String()
}

// generateInsights generates insights from the activity
func (r *NarrativeReport) generateInsights() string {
	var insights strings.Builder

	insights.WriteString("Insights:\n")

	// Add insights based on directory activity
	if len(r.ActivityPattern.MainDirectories) > 0 {
		insights.WriteString("\nDirectory Insights:\n")
		for _, dir := range r.ActivityPattern.MainDirectories {
			insights.WriteString(fmt.Sprintf("- High activity in '%s' suggests focused work in this area\n", dir))
		}
	}

	// Add insights based on file types
	if len(r.ActivityPattern.FileTypes) > 0 {
		insights.WriteString("\nFile Type Insights:\n")
		for _, ext := range r.ActivityPattern.FileTypes {
			insights.WriteString(fmt.Sprintf("- Frequent changes to %s files indicate active development/documentation\n", ext))
		}
	}

	// Add overall activity insight
	if r.ActivityPattern.TotalChanges > 0 {
		insights.WriteString(fmt.Sprintf("\nOverall Activity:\n- %d changes during this period indicate ", r.ActivityPattern.TotalChanges))
		if r.ActivityPattern.TotalChanges > 10 {
			insights.WriteString("high activity level\n")
		} else {
			insights.WriteString("moderate to low activity level\n")
		}
	}

	return insights.String()
}

// SendNarrativeReport generates and sends a narrative report via email
func SendNarrativeReport(changes []models.FileChange, period string, since time.Time) error {
	report := NewNarrativeReport(period, changes, since)
	narrative := report.GenerateNarrative()

	// Initialize email notifier
	notifier := notify.NewNotifier()

	// Send the report
	err := notifier.Send(context.Background(), "Dropbox Activity Report", narrative)
	if err != nil {
		return fmt.Errorf("failed to send narrative report: %w", err)
	}

	return nil
}
