package generators

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// HTMLGenerator generates HTML reports
type HTMLGenerator struct{}

// NewHTMLGenerator creates a new HTML generator
func NewHTMLGenerator() *HTMLGenerator {
	return &HTMLGenerator{}
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Dropbox Change Report</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
            line-height: 1.6;
            color: #333;
        }
        .header {
            background-color: #0061ff;
            color: white;
            padding: 20px;
            margin-bottom: 20px;
            border-radius: 5px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .section {
            margin-bottom: 30px;
            padding: 20px;
            background-color: #f8f9fa;
            border-radius: 5px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.05);
        }
        .change-item {
            padding: 10px;
            margin: 5px 0;
            border-left: 4px solid #0061ff;
            background-color: white;
            transition: transform 0.2s;
        }
        .change-item:hover {
            transform: translateX(5px);
        }
        .deleted {
            border-left-color: #dc3545;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        .stat-box {
            background-color: white;
            padding: 15px;
            border-radius: 5px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.05);
        }
        .stat-box h3 {
            margin-top: 0;
            color: #0061ff;
        }
        .file-list {
            max-height: 400px;
            overflow-y: auto;
            padding-right: 10px;
        }
        .file-list::-webkit-scrollbar {
            width: 8px;
        }
        .file-list::-webkit-scrollbar-track {
            background: #f1f1f1;
        }
        .file-list::-webkit-scrollbar-thumb {
            background: #0061ff;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Dropbox Change Report</h1>
        <p>Generated at: {{ .GeneratedAt.Format "2006-01-02 15:04:05" }}</p>
    </div>

    <div class="section">
        <h2>Summary</h2>
        <div class="stats-grid">
            <div class="stat-box">
                <h3>Overview</h3>
                <ul>
                    <li>Total Changes: {{ .TotalChanges }}</li>
                    <li>Total Size: {{ printf "%.2f" (divideFloat .TotalSize 1048576) }} MB</li>
                    <li>Deleted Files: {{ .DeletedCount }}</li>
                    <li>Modified Files: {{ .ModifiedCount }}</li>
                </ul>
            </div>
            <div class="stat-box">
                <h3>Top Extensions</h3>
                <ul>
                    {{range $ext, $count := .ExtensionCount}}
                    <li>{{$ext}}: {{$count}} files</li>
                    {{end}}
                </ul>
            </div>
            <div class="stat-box">
                <h3>Most Active Directories</h3>
                <ul>
                    {{range $dir, $count := .DirectoryCount}}
                    <li>{{$dir}}: {{$count}} changes</li>
                    {{end}}
                </ul>
            </div>
        </div>
    </div>

    <div class="section">
        <h2>File Changes</h2>
        <div class="file-list">
            {{range .Changes}}
            <div class="change-item {{if .IsDeleted}}deleted{{end}}">
                <strong>{{.Path}}</strong><br>
                Size: {{printf "%.2f" (divideFloat .Size 1048576)}} MB<br>
                {{if .IsDeleted}}
                Status: Deleted<br>
                {{else}}
                Modified: {{.Modified.Format "2006-01-02 15:04:05"}}<br>
                {{end}}
            </div>
            {{end}}
        </div>
    </div>
</body>
</html>
`

// HTMLData represents the data needed for HTML report generation
type HTMLData struct {
	*models.Report
	TotalSize     int64
	DeletedCount  int
	ModifiedCount int
}

// Generate generates an HTML report
func (g *HTMLGenerator) Generate(ctx context.Context, report *models.Report) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	if report == nil {
		return fmt.Errorf("report cannot be nil")
	}

	// Calculate additional stats
	var totalSize int64
	var deletedCount, modifiedCount int
	for _, change := range report.Changes {
		// Always add to total size
		totalSize += change.Size

		if change.IsDeleted {
			deletedCount++
		} else {
			modifiedCount++
		}
	}

	data := HTMLData{
		Report:        report,
		TotalSize:     totalSize,
		DeletedCount:  deletedCount,
		ModifiedCount: modifiedCount,
	}

	funcMap := template.FuncMap{
		"divideFloat": func(a int64, b float64) float64 {
			return float64(a) / b
		},
	}

	tmpl, err := template.New("html").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	if report.Metadata == nil {
		report.Metadata = make(map[string]string)
	}
	report.Metadata["content"] = buf.String()
	report.Type = models.HTMLReport

	return nil
}
