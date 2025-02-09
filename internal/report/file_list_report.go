package report

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
)

// FileChange represents a change to a file
type FileChange struct {
	Path      string
	Extension string
	Directory string
}

// Report represents a complete change report
type Report struct {
	Changes        []FileChange
	ExtensionCount map[string]int
	DirectoryCount map[string]int
	GeneratedAt    time.Time
	TotalChanges   int
}

// Generate formats the Dropbox changes into a readable report
func Generate(changes []string) *Report {
	startTime := time.Now()
	log.Printf("Generating report for %d changes...", len(changes))

	report := &Report{
		Changes:        make([]FileChange, 0, len(changes)),
		ExtensionCount: make(map[string]int),
		DirectoryCount: make(map[string]int),
		GeneratedAt:    time.Now(),
		TotalChanges:   len(changes),
	}

	if len(changes) == 0 {
		log.Println("No changes to report")
		return report
	}

	// Analyze changes
	for _, change := range changes {
		ext := strings.ToLower(filepath.Ext(change))
		dir := filepath.Dir(change)

		fileChange := FileChange{
			Path:      change,
			Extension: ext,
			Directory: dir,
		}

		report.Changes = append(report.Changes, fileChange)
		report.ExtensionCount[ext]++
		report.DirectoryCount[dir]++
	}

	log.Printf("Report generation completed in %v", time.Since(startTime))
	return report
}

// FormatReport formats the report into a readable string
func (r *Report) FormatReport() string {
	var sb strings.Builder

	// Header
	sb.WriteString("📢 *Dropbox Change Report*\n\n")

	// Summary section
	sb.WriteString("📊 *Summary*\n")
	if r.TotalChanges == 0 {
		sb.WriteString("No changes detected in Dropbox.\n")
		return sb.String()
	}
	sb.WriteString(fmt.Sprintf("Total changes: %d\n", r.TotalChanges))

	// File types section
	sb.WriteString("\n📁 *File Types Changed*\n")
	for ext, count := range r.ExtensionCount {
		if ext == "" {
			sb.WriteString(fmt.Sprintf("- No extension: %d\n", count))
		} else {
			sb.WriteString(fmt.Sprintf("- %s: %d\n", ext, count))
		}
	}

	// Directories section
	sb.WriteString("\n📂 *Directories Changed*\n")
	for dir, count := range r.DirectoryCount {
		sb.WriteString(fmt.Sprintf("- %s: %d changes\n", dir, count))
	}

	// Detailed changes
	sb.WriteString("\n📝 *Detailed Changes*\n")
	for _, change := range r.Changes {
		sb.WriteString(fmt.Sprintf("- %s\n", change.Path))
	}

	// Footer
	sb.WriteString(fmt.Sprintf("\n📅 Report generated at %s\n", r.GeneratedAt.Format("2006-01-02 15:04:05")))

	return sb.String()
}

// SendReport sends the report via email
func (r *Report) SendReport() error {
	notifier := notify.NewNotifier()
	
	subject := fmt.Sprintf("Dropbox Change Report - %s", r.GeneratedAt.Format("2006-01-02"))
	if r.TotalChanges > 0 {
		subject = fmt.Sprintf("%s (%d changes)", subject, r.TotalChanges)
	}

	body := r.FormatReport()
	if err := notifier.SendEmail(nil, subject, body); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("Report sent successfully to configured recipients")
	return nil
}
