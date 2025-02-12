package generators

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

const narrativeTemplate = `Dropbox Activity Report - {{ .Time.Format "2006-01-02 15:04:05" }}

During this period, there were {{ .TotalChanges }} file changes in your Dropbox account.

File Activity:
{{ if gt .DeletedFiles 0 }}- {{ .DeletedFiles }} files were deleted{{ end }}
{{ if gt .ModifiedFiles 0 }}- {{ .ModifiedFiles }} files were modified{{ end }}

Most Active Extensions:
{{ range $ext, $count := .ExtensionCount }}- {{ $ext }} ({{ $count }} files)
{{ end }}

Most Active Directories:
{{ range $dir, $count := .DirectoryCount }}- {{ $dir }}: {{ $count }} changes
{{ end }}

Total Size of Changes: {{ printf "%.2f" .TotalSize }} MB`

type narrativeData struct {
	Time           time.Time
	TotalChanges   int
	DeletedFiles   int
	ModifiedFiles  int
	ExtensionCount map[string]int
	DirectoryCount map[string]int
	TotalSize      float64
}

type narrativeGenerator struct {
	template *template.Template
}

// NewNarrativeGenerator creates a new narrative generator
func NewNarrativeGenerator() Generator {
	tmpl := template.Must(template.New("narrative").Parse(narrativeTemplate))
	return &narrativeGenerator{template: tmpl}
}

// Generate generates a narrative report
func (g *narrativeGenerator) Generate(ctx context.Context, report *models.Report) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	if report == nil {
		return fmt.Errorf("report cannot be nil")
	}

	data := &narrativeData{
		Time:           time.Now(),
		ExtensionCount: make(map[string]int),
		DirectoryCount: make(map[string]int),
	}

	for _, change := range report.Changes {
		data.TotalChanges++
		if change.IsDeleted {
			data.DeletedFiles++
		} else {
			data.ModifiedFiles++
		}
		data.ExtensionCount[change.Extension]++
		data.DirectoryCount[change.Directory]++
		data.TotalSize += float64(change.Size) / (1024 * 1024) // Convert to MB
	}

	var buf bytes.Buffer
	if err := g.template.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute narrative template: %w", err)
	}

	if report.Metadata == nil {
		report.Metadata = make(map[string]string)
	}
	report.Metadata["content"] = buf.String()
	return nil
}
