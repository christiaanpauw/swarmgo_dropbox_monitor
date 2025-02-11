package generators

import (
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting/models"
	"github.com/stretchr/testify/assert"
)

func TestGenerateFileList(t *testing.T) {
	report := &models.Report{
		Changes: []models.FileChange{
			{
				Path:      "/test/file1.txt",
				Extension: ".txt",
				Directory: "/test",
				ModTime:   time.Now(),
				Size:      1024,
			},
			{
				Path:      "/test/file2.jpg",
				Extension: ".jpg",
				Directory: "/test",
				ModTime:   time.Now(),
				Size:      2048,
			},
		},
		ExtensionCount: map[string]int{".txt": 1, ".jpg": 1},
		DirectoryCount: map[string]int{"/test": 2},
		GeneratedAt:    time.Now(),
		TotalChanges:   2,
	}

	output, err := GenerateFileList(report)
	assert.NoError(t, err)
	assert.Contains(t, output, "Dropbox Change Report")
	assert.Contains(t, output, "/test/file1.txt")
	assert.Contains(t, output, "/test/file2.jpg")
	assert.Contains(t, output, "Total Changes: 2")
	assert.Contains(t, output, "Most Active Extensions")
	assert.Contains(t, output, "Most Active Directories")
}

func TestGenerateNarrative(t *testing.T) {
	report := &models.Report{
		Changes: []models.FileChange{
			{
				Path:      "/test/file1.txt",
				Extension: ".txt",
				Directory: "/test",
				ModTime:   time.Now(),
				Size:      1024,
			},
			{
				Path:      "/test/file2.jpg",
				Extension: ".jpg",
				Directory: "/test",
				ModTime:   time.Now(),
				Size:      2048,
			},
		},
		ExtensionCount: map[string]int{".txt": 1, ".jpg": 1},
		DirectoryCount: map[string]int{"/test": 2},
		GeneratedAt:    time.Now(),
		TotalChanges:   2,
	}

	pattern := models.NewActivityPattern()
	now := time.Now()
	pattern.AddHotspot(models.DirectoryHotspot{
		Path:        "/test",
		ChangeCount: 2,
		LastActive:  now,
	})

	output, err := GenerateNarrative(report, pattern)
	assert.NoError(t, err)
	assert.Contains(t, output, "Activity Report")
	assert.Contains(t, output, "2 changes detected")
	assert.Contains(t, output, "Active Locations")
	assert.Contains(t, output, "File Type Analysis")
}

func TestGenerateHTML(t *testing.T) {
	report := &models.Report{
		Changes: []models.FileChange{
			{
				Path:      "/test/file1.txt",
				Extension: ".txt",
				Directory: "/test",
				ModTime:   time.Now(),
				Size:      1024,
			},
			{
				Path:      "/test/file2.jpg",
				Extension: ".jpg",
				Directory: "/test",
				ModTime:   time.Now(),
				Size:      2048,
			},
		},
		ExtensionCount: map[string]int{".txt": 1, ".jpg": 1},
		DirectoryCount: map[string]int{"/test": 2},
		GeneratedAt:    time.Now(),
		TotalChanges:   2,
	}

	pattern := models.NewActivityPattern()
	now := time.Now()
	pattern.AddHotspot(models.DirectoryHotspot{
		Path:        "/test",
		ChangeCount: 2,
		LastActive:  now,
	})

	output, err := GenerateHTML(report, pattern)
	assert.NoError(t, err)
	assert.Contains(t, output, "<html>")
	assert.Contains(t, output, "</html>")
	assert.Contains(t, output, "Dropbox Change Report")
	assert.Contains(t, output, "/test/file1.txt")
	assert.Contains(t, output, "/test/file2.jpg")
	assert.Contains(t, output, "Total Changes")
}
