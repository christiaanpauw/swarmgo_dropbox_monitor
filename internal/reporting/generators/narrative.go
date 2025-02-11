package generators

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting/models"
)

const narrativeTemplate = `Activity Report - {{ .GeneratedAt.Format "2006-01-02 15:04:05" }}

Summary:
There were {{ .TotalChanges }} changes detected in your Dropbox folders.

{{ if .MostActiveTime }}Most Active Period:
The highest activity was observed between {{ .MostActiveTime.Start.Format "15:04" }} and {{ .MostActiveTime.End.Format "15:04" }},
with {{ .MostActiveTime.Changes }} changes during this period.
{{ end }}

{{ if .TopHotspots }}Active Locations:
{{ range .TopHotspots }}  - {{ .Path }} ({{ .ChangeCount }} changes)
{{ end }}{{ end }}

{{ if .Patterns }}Notable Patterns:
{{ range .Patterns }}  - {{ .Pattern }} ({{ .Occurrences }} occurrences)
{{ end }}{{ end }}

File Type Analysis:
{{ range $ext, $count := .Report.ExtensionCount }}  - {{ $ext }}: {{ $count }} files
{{ end }}

Recommendations:
{{ range .Recommendations }}  - {{ . }}
{{ end }}
`

// NarrativeData represents the data needed for narrative report generation
type NarrativeData struct {
	*models.Report
	MostActiveTime  *models.TimeRange
	TopHotspots     []models.DirectoryHotspot
	Patterns        []models.FilePattern
	Recommendations []string
}

// GenerateNarrative generates a narrative report from the activity pattern
func GenerateNarrative(report *models.Report, pattern *models.ActivityPattern) (string, error) {
	// Generate recommendations based on patterns
	recommendations := generateRecommendations(report, pattern)

	data := NarrativeData{
		Report:          report,
		MostActiveTime:  pattern.GetMostActiveTimeRange(),
		TopHotspots:     pattern.GetTopHotspots(3),
		Patterns:        pattern.FilePatterns,
		Recommendations: recommendations,
	}

	tmpl, err := template.New("narrative").Parse(narrativeTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// generateRecommendations analyzes patterns and generates recommendations
func generateRecommendations(report *models.Report, pattern *models.ActivityPattern) []string {
	var recommendations []string

	// Check for high activity directories
	if pattern != nil {
		if hotspots := pattern.GetTopHotspots(1); len(hotspots) > 0 {
			if hotspots[0].ChangeCount > 10 {
				recommendations = append(recommendations,
					fmt.Sprintf("Consider organizing files in '%s' as it shows high activity", hotspots[0].Path))
			}
		}

		// Check time-based patterns
		if mostActive := pattern.GetMostActiveTimeRange(); mostActive != nil {
			if mostActive.Changes > 100 {
				recommendations = append(recommendations,
					"Consider spreading out large file operations to avoid system overload")
			}
		}
	}

	// Check file type patterns
	for ext, count := range report.ExtensionCount {
		switch {
		case count > 20 && (ext == ".tmp" || ext == ".temp"):
			recommendations = append(recommendations,
				"Consider cleaning up temporary files as they make up a significant portion of changes")
		case count > 50 && (ext == ".jpg" || ext == ".png"):
			recommendations = append(recommendations,
				"Consider using a dedicated photo management tool for better organization")
		case count > 30 && (ext == ".doc" || ext == ".docx"):
			recommendations = append(recommendations,
				"Consider using document version control for better file management")
		}
	}

	return recommendations
}
