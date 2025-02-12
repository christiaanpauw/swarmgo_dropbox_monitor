package models

import (
	"sort"
	"time"
)

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

// ActivityPattern represents a pattern of activity
type ActivityPattern struct {
	MainDirectories []string      `json:"main_directories"`
	FileTypes      []string      `json:"file_types"`
	BusyPeriods    []time.Time   `json:"busy_periods"`
	TotalChanges   int           `json:"total_changes"`
	FileContents   []FileContent `json:"file_contents,omitempty"`
}

// Report represents a complete change report
type Report struct {
	Type           ReportType         `json:"type"`
	Period         string             `json:"period"`
	Since          time.Time          `json:"since"`
	Until          time.Time          `json:"until"`
	Changes        []FileChange       `json:"changes"`
	ActivityStats  *ActivityPattern   `json:"activity_stats,omitempty"`
	ExtensionCount map[string]int     `json:"extension_count"`
	DirectoryCount map[string]int     `json:"directory_count"`
	GeneratedAt    time.Time          `json:"generated_at"`
	TotalChanges   int                `json:"total_changes"`
	Metadata       map[string]string  `json:"metadata"`
}

// NewReport creates a new report instance
func NewReport(reportType ReportType) *Report {
	now := time.Now()
	return &Report{
		Type:           reportType,
		Period:         "custom",
		Since:          now.Add(-24 * time.Hour), // Default to last 24 hours
		Until:          now,
		Changes:        make([]FileChange, 0),
		ExtensionCount: make(map[string]int),
		DirectoryCount: make(map[string]int),
		GeneratedAt:    now,
		Metadata:       make(map[string]string),
	}
}

// AddChange adds a file change to the report and updates counts
func (r *Report) AddChange(change FileChange) {
	r.Changes = append(r.Changes, change)
	r.ExtensionCount[change.Extension]++
	r.DirectoryCount[change.Directory]++
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

// SetTimeRange sets the time range for the report
func (r *Report) SetTimeRange(since, until time.Time) {
	r.Since = since
	r.Until = until
	r.Period = since.Format("2006-01-02") + " to " + until.Format("2006-01-02")
}

// SetActivityStats sets the activity stats for the report
func (r *Report) SetActivityStats(stats *ActivityPattern) {
	r.ActivityStats = stats
}

// Helper function to get top n items from a map
func getTopItems(items map[string]int, n int) []string {
	type kv struct {
		Key   string
		Value int
	}

	var sorted []kv
	for k, v := range items {
		sorted = append(sorted, kv{k, v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	result := make([]string, 0, n)
	for i := 0; i < n && i < len(sorted); i++ {
		result = append(result, sorted[i].Key)
	}

	return result
}
