package report

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agents"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
)

// ActivityPattern represents a pattern of activity in the Dropbox
type ActivityPattern struct {
	MainDirectories []string    // Most active directories
	FileTypes       []string    // Most changed file types
	BusyPeriods     []time.Time // Times with most activity
	TotalChanges    int
	FileContents    []agents.FileContent // Added field for file contents
}

// NarrativeReport generates a human-friendly narrative of Dropbox activity
type NarrativeReport struct {
	Period          string // "day", "hour", "10min"
	Since           time.Time
	Until           time.Time
	Changes         []string
	ActivityPattern ActivityPattern
}

// NewNarrativeReport creates a new narrative report for the specified time period
func NewNarrativeReport(period string, changes []string, since time.Time) *NarrativeReport {
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

	// Get content analyzer
	contentAnalyzer := agents.NewContentAnalyzer(os.Getenv("DROPBOX_ACCESS_TOKEN"))
	
	// Analyze content of changed files
	for _, change := range r.Changes {
		content, err := contentAnalyzer.AnalyzeFile(change)
		if err != nil {
			log.Printf("Error analyzing file %s: %v", change, err)
			continue
		}
		r.ActivityPattern.FileContents = append(r.ActivityPattern.FileContents, content)
	}

	for _, change := range r.Changes {
		// Count directories
		dir := filepath.Dir(change)
		dirCount[dir]++

		// Count file types
		ext := strings.ToLower(filepath.Ext(change))
		if ext == "" {
			ext = "no extension"
		}
		extCount[ext]++
	}

	// Get top directories and file types
	r.ActivityPattern.MainDirectories = getTopKeys(dirCount, 3)
	r.ActivityPattern.FileTypes = getTopKeys(extCount, 3)
	r.ActivityPattern.TotalChanges = len(r.Changes)
}

// getTopKeys returns the top n keys by value from a map
func getTopKeys(m map[string]int, n int) []string {
	if len(m) == 0 {
		return nil
	}

	// Convert map to slice of pairs
	type pair struct {
		key   string
		value int
	}
	pairs := make([]pair, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, pair{k, v})
	}

	// Sort by value (descending)
	for i := 0; i < len(pairs)-1; i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].value > pairs[i].value {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}

	// Get top n keys
	result := make([]string, 0, n)
	for i := 0; i < n && i < len(pairs); i++ {
		result = append(result, pairs[i].key)
	}
	return result
}

// GenerateNarrative creates a human-friendly narrative of the changes
func (r *NarrativeReport) GenerateNarrative() string {
	var sb strings.Builder

	// Introduction
	sb.WriteString(fmt.Sprintf("📊 *Dropbox Activity Report* (%s)\n\n", r.formatTimeRange()))

	// No changes case
	if r.ActivityPattern.TotalChanges == 0 {
		sb.WriteString("🌟 *Summary*\n")
		sb.WriteString("It's been quiet! No changes detected in your Dropbox during this period.\n")
		return sb.String()
	}

	// Activity Summary
	sb.WriteString("🌟 *Summary*\n")
	sb.WriteString(r.generateSummary())
	sb.WriteString("\n")

	// Detailed Analysis
	sb.WriteString("📈 *Activity Analysis*\n")
	sb.WriteString(r.generateAnalysis())
	sb.WriteString("\n")

	// Recommendations or Insights
	sb.WriteString("💡 *Insights*\n")
	sb.WriteString(r.generateInsights())
	sb.WriteString("\n")

	// Footer
	sb.WriteString(fmt.Sprintf("\n📅 Report generated at %s\n", time.Now().Format("2006-01-02 15:04:05")))

	return sb.String()
}

func (r *NarrativeReport) formatTimeRange() string {
	switch r.Period {
	case "day":
		return "Past 24 Hours"
	case "hour":
		return "Past Hour"
	case "10min":
		return "Past 10 Minutes"
	default:
		return fmt.Sprintf("Since %s", r.Since.Format("15:04:05"))
	}
}

func (r *NarrativeReport) generateSummary() string {
	var sb strings.Builder

	intensity := "moderate"
	if r.ActivityPattern.TotalChanges > 100 {
		intensity = "high"
	} else if r.ActivityPattern.TotalChanges < 10 {
		intensity = "light"
	}

	sb.WriteString(fmt.Sprintf("There has been %s activity in your Dropbox with %d changes ", 
		intensity, r.ActivityPattern.TotalChanges))
	
	switch r.Period {
	case "day":
		sb.WriteString("over the past 24 hours")
	case "hour":
		sb.WriteString("in the last hour")
	case "10min":
		sb.WriteString("in the past 10 minutes")
	}
	sb.WriteString(".\n")

	return sb.String()
}

func (r *NarrativeReport) generateAnalysis() string {
	var sb strings.Builder

	// Directory activity
	if len(r.ActivityPattern.MainDirectories) > 0 {
		sb.WriteString("Most active areas:\n")
		for _, dir := range r.ActivityPattern.MainDirectories {
			sb.WriteString(fmt.Sprintf("- %s\n", dir))
		}
		sb.WriteString("\n")
	}

	// File types
	if len(r.ActivityPattern.FileTypes) > 0 {
		sb.WriteString("Main file types changed:\n")
		for _, ext := range r.ActivityPattern.FileTypes {
			sb.WriteString(fmt.Sprintf("- %s files\n", strings.TrimPrefix(ext, ".")))
		}
		sb.WriteString("\n")
	}

	// Content analysis
	if len(r.ActivityPattern.FileContents) > 0 {
		sb.WriteString("📄 Content Changes:\n")
		for _, content := range r.ActivityPattern.FileContents {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", filepath.Base(content.Path), content.Summary))
			if len(content.Keywords) > 0 {
				sb.WriteString(fmt.Sprintf("  Keywords: %s\n", strings.Join(content.Keywords, ", ")))
			}
			if len(content.Topics) > 0 && content.Topics[0] != "general" {
				sb.WriteString(fmt.Sprintf("  Topics: %s\n", strings.Join(content.Topics, ", ")))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (r *NarrativeReport) generateInsights() string {
	var sb strings.Builder

	// Generate insights based on the patterns
	if len(r.ActivityPattern.MainDirectories) > 0 {
		mainDir := r.ActivityPattern.MainDirectories[0]
		sb.WriteString(fmt.Sprintf("👉 Most changes are concentrated in '%s'. ", mainDir))
		
		if strings.Contains(strings.ToLower(mainDir), "project") ||
			strings.Contains(strings.ToLower(mainDir), "doc") {
			sb.WriteString("This suggests active project work or documentation updates.\n")
		} else {
			sb.WriteString("You might want to review these changes.\n")
		}
	}

	// Content-based insights
	if len(r.ActivityPattern.FileContents) > 0 {
		var documentCount, codeCount, dataCount int
		for _, content := range r.ActivityPattern.FileContents {
			switch content.ContentType {
			case "document":
				documentCount++
			case "code":
				codeCount++
			case "data":
				dataCount++
			}
		}

		if documentCount > 0 {
			sb.WriteString(fmt.Sprintf("📚 %d document(s) were modified, suggesting active documentation work.\n", documentCount))
		}
		if codeCount > 0 {
			sb.WriteString(fmt.Sprintf("💻 %d code file(s) were changed, indicating development activity.\n", codeCount))
		}
		if dataCount > 0 {
			sb.WriteString(fmt.Sprintf("📊 %d data file(s) were updated, suggesting data analysis or configuration changes.\n", dataCount))
		}
	}

	return sb.String()
}

// SendNarrativeReport generates and sends a narrative report via email
func SendNarrativeReport(changes []string, period string, since time.Time) error {
	report := NewNarrativeReport(period, changes, since)
	narrative := report.GenerateNarrative()

	notifier := notify.NewNotifier()
	subject := fmt.Sprintf("Dropbox Activity Report - %s", report.formatTimeRange())

	if err := notifier.SendEmail(nil, subject, narrative); err != nil {
		return fmt.Errorf("failed to send narrative report: %v", err)
	}

	log.Printf("Narrative report sent successfully")
	return nil
}
