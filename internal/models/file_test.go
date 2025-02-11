package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFileMetadataJSON(t *testing.T) {
	now := time.Now()
	metadata := FileMetadata{
		Path:           "/test/file.txt",
		Name:           "file.txt",
		Size:           1024,
		Modified:       now,
		IsDeleted:      false,
		PathLower:      "/test/file.txt",
		ServerModified: now,
	}

	// Test marshaling
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal FileMetadata: %v", err)
	}

	// Test unmarshaling
	var unmarshaled FileMetadata
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal FileMetadata: %v", err)
	}

	// Compare fields
	if metadata.Path != unmarshaled.Path {
		t.Errorf("Path mismatch: got %s, want %s", unmarshaled.Path, metadata.Path)
	}
	if metadata.Name != unmarshaled.Name {
		t.Errorf("Name mismatch: got %s, want %s", unmarshaled.Name, metadata.Name)
	}
	if metadata.Size != unmarshaled.Size {
		t.Errorf("Size mismatch: got %d, want %d", unmarshaled.Size, metadata.Size)
	}
	if !metadata.Modified.Equal(unmarshaled.Modified) {
		t.Errorf("Modified time mismatch: got %v, want %v", unmarshaled.Modified, metadata.Modified)
	}
	if metadata.IsDeleted != unmarshaled.IsDeleted {
		t.Errorf("IsDeleted mismatch: got %v, want %v", unmarshaled.IsDeleted, metadata.IsDeleted)
	}
	if metadata.PathLower != unmarshaled.PathLower {
		t.Errorf("PathLower mismatch: got %s, want %s", unmarshaled.PathLower, metadata.PathLower)
	}
	if !metadata.ServerModified.Equal(unmarshaled.ServerModified) {
		t.Errorf("ServerModified time mismatch: got %v, want %v", unmarshaled.ServerModified, metadata.ServerModified)
	}
}

func TestFileContentJSON(t *testing.T) {
	content := FileContent{
		Path:     "/test/file.txt",
		Keywords: []string{"key1", "key2"},
		Topics:   []string{"topic1", "topic2"},
		Summary:  "Test summary",
	}

	// Test marshaling
	data, err := json.Marshal(content)
	if err != nil {
		t.Fatalf("Failed to marshal FileContent: %v", err)
	}

	// Test unmarshaling
	var unmarshaled FileContent
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal FileContent: %v", err)
	}

	// Compare fields
	if content.Path != unmarshaled.Path {
		t.Errorf("Path mismatch: got %s, want %s", unmarshaled.Path, content.Path)
	}
	if len(content.Keywords) != len(unmarshaled.Keywords) {
		t.Errorf("Keywords length mismatch: got %d, want %d", len(unmarshaled.Keywords), len(content.Keywords))
	}
	for i, kw := range content.Keywords {
		if unmarshaled.Keywords[i] != kw {
			t.Errorf("Keyword mismatch at index %d: got %s, want %s", i, unmarshaled.Keywords[i], kw)
		}
	}
	if len(content.Topics) != len(unmarshaled.Topics) {
		t.Errorf("Topics length mismatch: got %d, want %d", len(unmarshaled.Topics), len(content.Topics))
	}
	for i, topic := range content.Topics {
		if unmarshaled.Topics[i] != topic {
			t.Errorf("Topic mismatch at index %d: got %s, want %s", i, unmarshaled.Topics[i], topic)
		}
	}
	if content.Summary != unmarshaled.Summary {
		t.Errorf("Summary mismatch: got %s, want %s", unmarshaled.Summary, content.Summary)
	}
}

func TestFileChangeJSON(t *testing.T) {
	now := time.Now()
	change := FileChange{
		Path:      "/test/file.txt",
		Extension: ".txt",
		Directory: "/test",
		ModTime:   now,
		IsDeleted: false,
	}

	// Test marshaling
	data, err := json.Marshal(change)
	if err != nil {
		t.Fatalf("Failed to marshal FileChange: %v", err)
	}

	// Test unmarshaling
	var unmarshaled FileChange
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal FileChange: %v", err)
	}

	// Compare fields
	if change.Path != unmarshaled.Path {
		t.Errorf("Path mismatch: got %s, want %s", unmarshaled.Path, change.Path)
	}
	if change.Extension != unmarshaled.Extension {
		t.Errorf("Extension mismatch: got %s, want %s", unmarshaled.Extension, change.Extension)
	}
	if change.Directory != unmarshaled.Directory {
		t.Errorf("Directory mismatch: got %s, want %s", unmarshaled.Directory, change.Directory)
	}
	if !change.ModTime.Equal(unmarshaled.ModTime) {
		t.Errorf("ModTime mismatch: got %v, want %v", unmarshaled.ModTime, change.ModTime)
	}
	if change.IsDeleted != unmarshaled.IsDeleted {
		t.Errorf("IsDeleted mismatch: got %v, want %v", unmarshaled.IsDeleted, change.IsDeleted)
	}
}
