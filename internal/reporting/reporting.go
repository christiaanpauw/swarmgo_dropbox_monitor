package reporting

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/db"
)

// GenerateReport creates a report of recent file changes
func GenerateReport(files []db.FileChange) string {
	if len(files) == 0 {
		return "No file changes found in the specified time period."
	}

	// Sort files by modified time
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModifiedAt.After(files[j].ModifiedAt)
	})

	// Group files by directory
	dirCounts := make(map[string]int)
	for _, file := range files {
		dir := filepath.Dir(file.FilePath)
		dirCounts[dir]++
	}

	// Count file types
	typeCounts := make(map[string]int)
	for _, file := range files {
		fileType := file.FileType
		if fileType == "" {
			fileType = "no extension"
		}
		typeCounts[fileType]++
	}

	// Count portfolios
	portfolioCounts := make(map[string]int)
	for _, file := range files {
		if file.Portfolio != "" {
			portfolioCounts[file.Portfolio]++
		}
	}

	// Count projects
	projectCounts := make(map[string]int)
	for _, file := range files {
		if file.Project != "" {
			projectCounts[file.Project]++
		}
	}

	// Count authors
	authorCounts := make(map[string]int)
	for _, file := range files {
		author := file.Author
		if file.ModifiedByName != "" {
			author = file.ModifiedByName
		} else if file.LockHolderName != "" {
			author = file.LockHolderName
		}
		if author != "" && author != "unknown" {
			authorCounts[author]++
		}
	}

	// Build report
	var report strings.Builder
	report.WriteString(fmt.Sprintf("File Change Report - %s\n", time.Now().Format("2006-01-02 15:04:05")))
	report.WriteString(fmt.Sprintf("Total Files: %d\n\n", len(files)))

	// List recent changes
	report.WriteString("Recent File Changes:\n")
	for i, file := range files {
		author := file.Author
		if file.ModifiedByName != "" {
			author = file.ModifiedByName
		} else if file.LockHolderName != "" {
			author = file.LockHolderName
		}

		authorStr := ""
		if author != "" && author != "unknown" {
			authorStr = fmt.Sprintf(" (Author: %s)", author)
		}

		report.WriteString(fmt.Sprintf("%d. %s (Modified: %s)%s\n",
			i+1,
			file.FilePath,
			file.ModifiedAt.Format("2006-01-02 15:04:05"),
			authorStr,
		))
	}
	report.WriteString("\n")

	// List directories
	report.WriteString("Directories:\n")
	for dir, count := range dirCounts {
		report.WriteString(fmt.Sprintf("- %s: %d files\n", dir, count))
	}
	report.WriteString("\n")

	// List file types
	report.WriteString("File Types:\n")
	for fileType, count := range typeCounts {
		report.WriteString(fmt.Sprintf("- %s: %d files\n", fileType, count))
	}
	report.WriteString("\n")

	// List portfolios
	if len(portfolioCounts) > 0 {
		report.WriteString("Portfolios (extracted from paths):\n")
		for portfolio, count := range portfolioCounts {
			report.WriteString(fmt.Sprintf("- %s: %d files\n", portfolio, count))
		}
		report.WriteString("\n")
	}

	// List projects
	if len(projectCounts) > 0 {
		report.WriteString("Projects (extracted from paths):\n")
		for project, count := range projectCounts {
			report.WriteString(fmt.Sprintf("- %s: %d files\n", project, count))
		}
		report.WriteString("\n")
	}

	// List authors
	if len(authorCounts) > 0 {
		report.WriteString("Authors (extracted from paths):\n")
		for author, count := range authorCounts {
			report.WriteString(fmt.Sprintf("- %s: %d files\n", author, count))
		}
		report.WriteString("\n")
	}

	return report.String()
}
