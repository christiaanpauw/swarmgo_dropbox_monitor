package models

import (
	"path/filepath"
	"strings"
	"time"
)

// FileMetadata represents metadata about a file
type FileMetadata struct {
	Path           string    `json:"path"`
	Name           string    `json:"name"`
	Size           int64     `json:"size"`
	Modified       time.Time `json:"modified"`
	IsDeleted      bool      `json:"is_deleted"`
	PathLower      string    `json:"path_lower"`
	ServerModified time.Time `json:"server_modified"`
	Extension      string    `json:"extension"`      // File extension
	Directory      string    `json:"directory"`      // Parent directory
	ModTime        time.Time `json:"mod_time"`      // Last modification time
}

// FileContent represents analyzed content of a file
type FileContent struct {
	Path         string   `json:"path"`
	ContentType  string   `json:"content_type"`
	Size         int64    `json:"size"`
	IsBinary     bool     `json:"is_binary"`
	ContentHash  string   `json:"content_hash"`
	Keywords     []string `json:"keywords,omitempty"`
	Topics       []string `json:"topics,omitempty"`
	Summary      string   `json:"summary,omitempty"`
}

// FileChange represents a processed file change with additional metadata
type FileChange struct {
	Path      string    `json:"path"`
	Extension string    `json:"extension"`
	Directory string    `json:"directory"`
	ModTime   time.Time `json:"mod_time"`
	Modified  time.Time `json:"modified"`
	IsDeleted bool      `json:"is_deleted"`
	Size      int64     `json:"size"`
}

// NewFileMetadata creates a new FileMetadata with computed fields
func NewFileMetadata(path string, size int64, modified time.Time, isDeleted bool) *FileMetadata {
	return &FileMetadata{
		Path:      path,
		Name:      filepath.Base(path),
		Size:      size,
		Modified:  modified,
		IsDeleted: isDeleted,
		PathLower: strings.ToLower(path),
		Extension: strings.ToLower(filepath.Ext(path)),
		Directory: filepath.Dir(path),
		ModTime:   modified,
	}
}

// ToFileChange converts a FileMetadata to a FileChange
func (fm *FileMetadata) ToFileChange() FileChange {
	return FileChange{
		Path:      fm.Path,
		Extension: fm.Extension,
		Directory: fm.Directory,
		ModTime:   fm.ModTime,
		Modified:  fm.Modified,
		IsDeleted: fm.IsDeleted,
		Size:      fm.Size,
	}
}

// FromFileMetadata creates a new FileChange from a FileMetadata
func NewFileChangeFromMetadata(metadata *FileMetadata) *FileChange {
	if metadata == nil {
		return nil
	}
	change := metadata.ToFileChange()
	return &change
}

// BatchConvertMetadataToChanges converts a slice of FileMetadata to FileChanges
func BatchConvertMetadataToChanges(metadata []*FileMetadata) []FileChange {
	changes := make([]FileChange, 0, len(metadata))
	for _, m := range metadata {
		if m != nil {
			changes = append(changes, m.ToFileChange())
		}
	}
	return changes
}
