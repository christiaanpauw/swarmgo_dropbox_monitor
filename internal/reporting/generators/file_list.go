package generators

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

const fileListTemplate = `Dropbox Change Report - {{ .GeneratedAt.Format "2006-01-02 15:04:05" }}

Total Changes: {{ .TotalChanges }}

File Changes:
{{ range .Changes }}  - {{ if .IsDeleted }}[Deleted] {{ end }}{{ .Path }} ({{ printf "%.2f" (divideFloat .Size 1048576) }} MB)
{{ end }}

Most Active Extensions:
{{ range $ext, $count := .ExtensionCount }}  - {{ $ext }}: {{ $count }} files
{{ end }}

Most Active Directories:
{{ range $dir, $count := .DirectoryCount }}  - {{ $dir }}: {{ $count }} changes
{{ end }}

Activity Summary:
- Total Size: {{ printf "%.2f" (divideFloat .TotalSize 1048576) }} MB
- Deleted Files: {{ .DeletedCount }}
- Modified Files: {{ .ModifiedCount }}
`

// FileListData represents the data needed for file list report generation
type FileListData struct {
	*models.Report
	TotalSize     int64
	DeletedCount  int
	ModifiedCount int
	ExtensionCount map[string]int
	DirectoryCount map[string]int
}

// GenerateFileList generates a text-based file list report
func GenerateFileList(ctx context.Context, report *models.Report) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context cancelled: %w", err)
	}

	if report == nil {
		return "", fmt.Errorf("report cannot be nil")
	}

	// Calculate additional stats
	var totalSize int64
	var deletedCount, modifiedCount int
	extensionCount := make(map[string]int)
	directoryCount := make(map[string]int)
	for _, change := range report.Changes {
		// Always add to total size
		totalSize += change.Size

		if change.IsDeleted {
			deletedCount++
		} else {
			modifiedCount++
		}
		
		// Use the Extension field directly
		if change.Extension != "" {
			extensionCount[change.Extension]++
		}
		
		// Use the Directory field directly
		if change.Directory != "" {
			directoryCount[change.Directory]++
		}
	}

	data := FileListData{
		Report:        report,
		TotalSize:     totalSize,
		DeletedCount:  deletedCount,
		ModifiedCount: modifiedCount,
		ExtensionCount: extensionCount,
		DirectoryCount: directoryCount,
	}

	funcMap := template.FuncMap{
		"divideFloat": func(a int64, b float64) float64 {
			return float64(a) / b
		},
	}

	tmpl, err := template.New("filelist").Funcs(funcMap).Parse(fileListTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// FileListGenerator generates a simple list of file changes
type FileListGenerator struct{}

// NewFileListGenerator creates a new file list generator
func NewFileListGenerator() *FileListGenerator {
	return &FileListGenerator{}
}

// Generate generates a file list report
func (g *FileListGenerator) Generate(ctx context.Context, report *models.Report) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	if report == nil {
		return fmt.Errorf("report cannot be nil")
	}

	content, err := GenerateFileList(ctx, report)
	if err != nil {
		return fmt.Errorf("failed to generate file list: %w", err)
	}

	if report.Metadata == nil {
		report.Metadata = make(map[string]string)
	}
	report.Metadata["content"] = content
	report.Type = models.FileListReport

	return nil
}
