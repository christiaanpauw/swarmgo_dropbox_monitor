package generators

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting/models"
)

const fileListTemplate = `Dropbox Change Report - {{ .GeneratedAt.Format "2006-01-02 15:04:05" }}

Total Changes: {{ .TotalChanges }}

File Changes:
{{ range .Changes }}  - {{ .Path }}
{{ end }}

Most Active Extensions:
{{ range .TopExtensions }}  - {{ . }}
{{ end }}

Most Active Directories:
{{ range .TopDirectories }}  - {{ . }}
{{ end }}
`

// FileListData represents the data needed for file list report generation
type FileListData struct {
	*models.Report
	TopExtensions  []string
	TopDirectories []string
}

// GenerateFileList generates a text-based file list report
func GenerateFileList(report *models.Report) (string, error) {
	data := FileListData{
		Report:         report,
		TopExtensions:  report.GetTopExtensions(5),
		TopDirectories: report.GetTopDirectories(5),
	}

	tmpl, err := template.New("filelist").Parse(fileListTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
