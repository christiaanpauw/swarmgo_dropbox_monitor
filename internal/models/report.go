package models

import "time"

// ActivityPattern represents a pattern of activity
type ActivityPattern struct {
	MainDirectories []string      `json:"main_directories"`
	FileTypes      []string      `json:"file_types"`
	BusyPeriods    []time.Time   `json:"busy_periods"`
	TotalChanges   int           `json:"total_changes"`
	FileContents   []FileContent `json:"file_contents,omitempty"`
}

// Report represents a generated report about file changes
type Report struct {
	Period         string          `json:"period"`
	Since          time.Time       `json:"since"`
	Until          time.Time       `json:"until"`
	Changes        []FileChange    `json:"changes"`
	ActivityStats  ActivityPattern `json:"activity_stats"`
	Summary        string          `json:"summary"`
	HTMLContent    string          `json:"html_content,omitempty"`
	ExtensionCount map[string]int  `json:"extension_count"`
	DirectoryCount map[string]int  `json:"directory_count"`
	GeneratedAt    time.Time       `json:"generated_at"`
	TotalChanges   int            `json:"total_changes"`
}
