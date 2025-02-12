package generators

import (
	"context"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestChanges() []models.FileChange {
	now := time.Date(2025, 2, 12, 10, 6, 0, 0, time.UTC)
	return []models.FileChange{
		{
			Path:      "/test/file1.txt",
			Extension: ".txt",
			Directory: "/test",
			ModTime:   now,
			Modified:  now,
			Size:      1024 * 1024, // 1 MB
		},
		{
			Path:      "/test/file2.jpg",
			Extension: ".jpg",
			Directory: "/test",
			ModTime:   now,
			Modified:  now,
			Size:      2 * 1024 * 1024, // 2 MB
		},
		{
			Path:      "/test/subdir/file3.txt",
			Extension: ".txt",
			Directory: "/test/subdir",
			ModTime:   now,
			Modified:  now,
			Size:      512 * 1024, // 0.5 MB
			IsDeleted: true,
		},
	}
}

func TestFileListGenerator(t *testing.T) {
	generator := NewFileListGenerator()
	require.NotNil(t, generator)

	changes := createTestChanges()
	report := models.NewReport(models.FileListReport)
	for _, change := range changes {
		report.AddChange(change)
	}

	err := generator.Generate(context.Background(), report)
	require.NoError(t, err)

	content, ok := report.Metadata["content"]
	require.True(t, ok, "content should be present in metadata")
	require.NotEmpty(t, content, "content should not be empty")

	// Check report content
	assert.Contains(t, content, "Dropbox Change Report")
	assert.Contains(t, content, "Total Changes: 3")
	assert.Contains(t, content, "/test/file1.txt")
	assert.Contains(t, content, "/test/file2.jpg")
	assert.Contains(t, content, "/test/subdir/file3.txt")
	assert.Contains(t, content, ".txt: 2 files")
	assert.Contains(t, content, "/test: 2 changes")
	assert.Contains(t, content, "/test/subdir: 1 changes")
	assert.Contains(t, content, "Total Size: 3.50 MB")
	assert.Contains(t, content, "Deleted Files: 1")
	assert.Contains(t, content, "Modified Files: 2")
}

func TestHTMLGenerator(t *testing.T) {
	generator := NewHTMLGenerator()
	require.NotNil(t, generator)

	changes := createTestChanges()
	report := models.NewReport(models.HTMLReport)
	for _, change := range changes {
		report.AddChange(change)
	}

	err := generator.Generate(context.Background(), report)
	require.NoError(t, err)

	content, ok := report.Metadata["content"]
	require.True(t, ok, "content should be present in metadata")
	require.NotEmpty(t, content, "content should not be empty")

	// Check HTML content
	assert.Contains(t, content, "<!DOCTYPE html>")
	assert.Contains(t, content, "Dropbox Change Report")
	assert.Contains(t, content, "Total Changes: 3")
	assert.Contains(t, content, "/test/file1.txt")
	assert.Contains(t, content, "/test/file2.jpg")
	assert.Contains(t, content, "/test/subdir/file3.txt")
	assert.Contains(t, content, ".txt: 2 files")
	assert.Contains(t, content, "/test: 2 changes")
	assert.Contains(t, content, "/test/subdir: 1 changes")
	assert.Contains(t, content, "Total Size: 3.50 MB")
	assert.Contains(t, content, "Deleted Files: 1")
	assert.Contains(t, content, "Modified Files: 2")
}

func TestNarrativeGenerator(t *testing.T) {
	generator := NewNarrativeGenerator()
	require.NotNil(t, generator)

	changes := createTestChanges()
	report := models.NewReport(models.NarrativeReport)
	for _, change := range changes {
		report.AddChange(change)
	}

	err := generator.Generate(context.Background(), report)
	require.NoError(t, err)

	content, ok := report.Metadata["content"]
	require.True(t, ok, "content should be present in metadata")
	require.NotEmpty(t, content, "content should not be empty")

	// Check narrative content
	assert.Contains(t, content, "Dropbox Activity Report")
	assert.Contains(t, content, "3 file changes")
	assert.Contains(t, content, "1 files were deleted")
	assert.Contains(t, content, ".txt (2 files)")
	assert.Contains(t, content, ".jpg (1 files)")
	assert.Contains(t, content, "3.50 MB")
}
