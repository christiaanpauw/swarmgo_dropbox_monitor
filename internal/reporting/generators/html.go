package generators

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting/models"
)

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Dropbox Change Report - {{ .GeneratedAt }}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .header {
            background-color: #2196F3;
            color: white;
            padding: 20px;
            border-radius: 5px;
            margin-bottom: 20px;
        }
        .section {
            background-color: white;
            padding: 20px;
            margin-bottom: 20px;
            border-radius: 5px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .stats {
            display: flex;
            justify-content: space-between;
            flex-wrap: wrap;
        }
        .stat-box {
            flex: 1;
            min-width: 200px;
            margin: 10px;
            padding: 15px;
            background-color: #f8f9fa;
            border-radius: 5px;
            text-align: center;
        }
        .changes-list {
            list-style-type: none;
            padding: 0;
        }
        .changes-list li {
            padding: 10px;
            border-bottom: 1px solid #eee;
        }
        .changes-list li:last-child {
            border-bottom: none;
        }
        .recommendations {
            background-color: #fff3cd;
            padding: 15px;
            border-radius: 5px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Dropbox Change Report</h1>
        <p>Generated: {{ .GeneratedAt }}</p>
    </div>

    <div class="section">
        <h2>Summary</h2>
        <div class="stats">
            <div class="stat-box">
                <h3>Total Changes</h3>
                <p>{{ .Report.TotalChanges }}</p>
            </div>
            <div class="stat-box">
                <h3>Directories</h3>
                <p>{{ len .TopDirectories }}</p>
            </div>
            <div class="stat-box">
                <h3>File Types</h3>
                <p>{{ len .TopExtensions }}</p>
            </div>
        </div>
    </div>

    {{ if .Hotspots }}
    <div class="section">
        <h2>Active Locations</h2>
        <ul class="changes-list">
        {{ range .Hotspots }}
            <li>
                <strong>{{ .Path }}</strong> - {{ .ChangeCount }} changes
                {{ if .CommonPatterns }}
                <br>
                <small>Common patterns: {{ .CommonPatterns }}</small>
                {{ end }}
            </li>
        {{ end }}
        </ul>
    </div>
    {{ end }}

    {{ if .MostActive }}
    <div class="section">
        <h2>Most Active Time Range</h2>
        <p>{{ .MostActive.Start.Format "2006-01-02 15:04:05" }} - {{ .MostActive.End.Format "2006-01-02 15:04:05" }}</p>
    </div>
    {{ end }}

    {{ if .Recommendations }}
    <div class="section recommendations">
        <h2>Recommendations</h2>
        <ul class="changes-list">
        {{ range .Recommendations }}
            <li>{{ . }}</li>
        {{ end }}
        </ul>
    </div>
    {{ end }}

    <div class="section">
        <h2>Recent Changes</h2>
        <ul class="changes-list">
        {{ range .Report.Changes }}
            <li>
                <strong>{{ .Path }}</strong>
                <br>
                <small>
                    Directory: {{ .Directory }}<br>
                    Size: {{ .Size }} bytes<br>
                    Modified: {{ .ModTime.Format "2006-01-02 15:04:05" }}
                </small>
            </li>
        {{ end }}
        </ul>
    </div>
</body>
</html>
`

// GenerateHTML generates an HTML report
func GenerateHTML(report *models.Report, pattern *models.ActivityPattern) (string, error) {
	var hotspots []models.DirectoryHotspot
	var mostActive *models.TimeRange
	if pattern != nil {
		hotspots = pattern.GetTopHotspots(5)
		mostActive = pattern.GetMostActiveTimeRange()
	}

	// Get top extensions and directories
	topExtensions := report.GetTopExtensions(5)
	topDirectories := report.GetTopDirectories(5)

	// Generate recommendations
	recommendations := generateRecommendations(report, pattern)

	// Create HTML template data
	data := struct {
		Report          *models.Report
		TopExtensions   []string
		TopDirectories  []string
		Hotspots        []models.DirectoryHotspot
		MostActive      *models.TimeRange
		Recommendations []string
		GeneratedAt     string
	}{
		Report:          report,
		TopExtensions:   topExtensions,
		TopDirectories:  topDirectories,
		Hotspots:        hotspots,
		MostActive:      mostActive,
		Recommendations: recommendations,
		GeneratedAt:     report.GeneratedAt.Format(time.RFC1123),
	}

	// Parse and execute template
	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
