package models

import "time"

// FileMetadata represents metadata about a file
type FileMetadata struct {
	Path           string    `json:"path"`
	Name           string    `json:"name"`
	Size           int64     `json:"size"`
	Modified       time.Time `json:"modified"`
	IsDeleted      bool      `json:"is_deleted"`
	PathLower      string    `json:"path_lower"`
	ServerModified time.Time `json:"server_modified"`
}

// FileContent represents analyzed content of a file
type FileContent struct {
	Path     string   `json:"path"`
	Keywords []string `json:"keywords,omitempty"`
	Topics   []string `json:"topics,omitempty"`
	Summary  string   `json:"summary,omitempty"`
}

// FileChange represents a processed file change with additional metadata
type FileChange struct {
	Path      string    `json:"path"`
	Extension string    `json:"extension"`
	Directory string    `json:"directory"`
	ModTime   time.Time `json:"mod_time"`
	IsDeleted bool      `json:"is_deleted"`
}
