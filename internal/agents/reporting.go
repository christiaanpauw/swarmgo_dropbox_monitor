package agents

import (
	"context"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"time"
)

type reportingAgent struct{}

func NewReportingAgent() ReportingAgent {
	return &reportingAgent{}
}

func (ra *reportingAgent) GenerateReport(ctx context.Context, changes []FileChange) (*Report, error) {
	if len(changes) == 0 {
		return &Report{
			GeneratedAt: time.Now(),
			TimeWindow: "unknown",
			FileCount:  0,
			Summary:    "No changes detected in the monitored period.",
			Changes:    changes,
		}, nil
	}

	// Process changes
	filesByType := make(map[string]int)
	topKeywords := make(map[string]int)
	topTopics := make(map[string]int)
	for _, change := range changes {
		// Count file types
		fileType := ra.getFileType(change.Path)
		filesByType[fileType]++
	}

	// Get top items
	topKeywords = ra.getTopItems(topKeywords, 5)
	topTopics = ra.getTopItems(topTopics, 5)

	// Generate summary
	summary := fmt.Sprintf("Detected %d file changes:\n", len(changes))
	for _, change := range changes {
		summary += fmt.Sprintf("- %s (modified at %s)\n", change.Path, change.ModTime)
	}

	return &Report{
		GeneratedAt:  time.Now(),
		TimeWindow:   "unknown",
		FileCount:    len(changes),
		FilesByType:  filesByType,
		TopKeywords:  topKeywords,
		TopTopics:    topTopics,
		RecentFiles:  []map[string]interface{}{},
		Summary:      summary,
		Changes:      changes,
	}, nil
}

func (ra *reportingAgent) getFileType(path string) string {
	ext := strings.ToLower(strings.TrimPrefix(path[strings.LastIndex(path, "."):], "."))
	switch ext {
	case "txt", "md", "rst", "doc", "docx":
		return "document"
	case "jpg", "jpeg", "png", "gif", "bmp":
		return "image"
	case "mp3", "wav", "ogg":
		return "audio"
	case "mp4", "avi", "mov":
		return "video"
	case "pdf":
		return "pdf"
	case "go", "py", "js", "java", "cpp", "c", "h":
		return "code"
	default:
		return "other"
	}
}

func (ra *reportingAgent) getTopItems(items map[string]int, n int) map[string]int {
	// Convert map to slice of pairs
	type pair struct {
		key   string
		value int
	}
	var pairs []pair
	for k, v := range items {
		pairs = append(pairs, pair{k, v})
	}

	// Sort by value
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].value > pairs[j].value
	})

	// Take top N
	result := make(map[string]int)
	for i := 0; i < n && i < len(pairs); i++ {
		result[pairs[i].key] = pairs[i].value
	}

	return result
}

func (ra *reportingAgent) generateHTMLReport(report *Report) string {
	tmpl := template.Must(template.New("report").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Dropbox Monitor Report</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
            line-height: 1.6;
        }
        .section {
            margin-bottom: 20px;
        }
        h1, h2 {
            color: #333;
        }
        table {
            border-collapse: collapse;
            width: 100%;
            margin-bottom: 20px;
        }
        th, td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
        }
        th {
            background-color: #f5f5f5;
        }
        .file-list {
            list-style-type: none;
            padding: 0;
        }
        .file-item {
            padding: 5px 0;
            border-bottom: 1px solid #eee;
        }
    </style>
</head>
<body>
    <h1>Dropbox Monitor Report</h1>
    
    <div class="section">
        <h2>Overview</h2>
        <p>Generated at: {{.GeneratedAt}}</p>
        <p>Time window: {{.TimeWindow}}</p>
        <p>Total files: {{.FileCount}}</p>
    </div>

    <div class="section">
        <h2>File Types</h2>
        <table>
            <tr><th>Type</th><th>Count</th></tr>
            {{range $type, $count := .FilesByType}}
            <tr><td>{{$type}}</td><td>{{$count}}</td></tr>
            {{end}}
        </table>
    </div>

    {{if .TopKeywords}}
    <div class="section">
        <h2>Top Keywords</h2>
        <table>
            <tr><th>Keyword</th><th>Count</th></tr>
            {{range $keyword, $count := .TopKeywords}}
            <tr><td>{{$keyword}}</td><td>{{$count}}</td></tr>
            {{end}}
        </table>
    </div>
    {{end}}

    {{if .TopTopics}}
    <div class="section">
        <h2>Top Topics</h2>
        <table>
            <tr><th>Topic</th><th>Count</th></tr>
            {{range $topic, $count := .TopTopics}}
            <tr><td>{{$topic}}</td><td>{{$count}}</td></tr>
            {{end}}
        </table>
    </div>
    {{end}}

    <div class="section">
        <h2>Recent Files</h2>
        <ul class="file-list">
        {{range .Changes}}
            <li class="file-item">{{.Path}} (modified at {{.ModTime}})</li>
        {{end}}
        </ul>
    </div>
</body>
</html>
`))

	var output strings.Builder
	err := tmpl.Execute(&output, report)
	if err != nil {
		return fmt.Sprintf("Error generating HTML report: %v", err)
	}

	return output.String()
}

func (ra *reportingAgent) generateTextReport(report *Report) string {
	return report.Summary
}
