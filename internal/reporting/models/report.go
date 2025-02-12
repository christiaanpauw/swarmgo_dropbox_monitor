package models

import (
	"sort"
	"time"
	"path/filepath"
)

// FileChange represents a single file change in Dropbox
type FileChange struct {
	Path      string    // Full path to the file
	Extension string    // File extension (lowercase)
	Directory string    // Parent directory
	ModTime   time.Time // Modification time
	Size      int64     // File size in bytes
}

// Report represents a complete change report
type Report struct {
	Type           ReportType         // Type of report
	Changes        []FileChange        // List of file changes
	ExtensionCount map[string]int     // Count of changes by extension
	DirectoryCount map[string]int     // Count of changes by directory
	GeneratedAt    time.Time          // Report generation time
	TotalChanges   int                // Total number of changes
	Metadata       map[string]string  // Additional metadata
}

// ReportType defines the type of report
type ReportType string

const (
	// FileListReport is a simple list of file changes
	FileListReport ReportType = "file_list"
	// NarrativeReport includes analysis and narrative description
	NarrativeReport ReportType = "narrative"
	// HTMLReport is formatted in HTML
	HTMLReport ReportType = "html"
)

// NewReport creates a new report instance
func NewReport(reportType ReportType) *Report {
	return &Report{
		Type:           reportType,
		Changes:        make([]FileChange, 0),
		ExtensionCount: make(map[string]int),
		DirectoryCount: make(map[string]int),
		GeneratedAt:    time.Now(),
		Metadata:       make(map[string]string),
	}
}

// AddChange adds a file change to the report and updates counts
func (r *Report) AddChange(change FileChange) {
	r.Changes = append(r.Changes, change)
	r.ExtensionCount[change.Extension]++
	r.DirectoryCount[filepath.Dir(change.Path)]++
	r.TotalChanges++
}

// GetTopExtensions returns the n most common file extensions
func (r *Report) GetTopExtensions(n int) []string {
	return getTopItems(r.ExtensionCount, n)
}

// GetTopDirectories returns the n most active directories
func (r *Report) GetTopDirectories(n int) []string {
	return getTopItems(r.DirectoryCount, n)
}

// Helper function to get top n items from a map
func getTopItems(items map[string]int, n int) []string {
	type kv struct {
		Key   string
		Value int
	}

	// Convert map to slice of kv
	var ss []kv
	for k, v := range items {
		ss = append(ss, kv{k, v})
	}

	// Sort slice by value in descending order
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	// Get top n items
	result := make([]string, 0, n)
	for i := 0; i < n && i < len(ss); i++ {
		result = append(result, ss[i].Key)
	}

	return result
}
