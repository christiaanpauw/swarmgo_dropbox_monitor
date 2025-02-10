package reporting

import (
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/db"
)

func TestGenerateReport(t *testing.T) {
	baseTime := time.Date(2025, 2, 10, 9, 21, 16, 0, time.UTC)

	testCases := []struct {
		name  string
		files []db.FileChange
	}{
		{
			name: "ten_files",
			files: []db.FileChange{
				{
					FilePath:       "/project1/docs/spec.md",
					ModifiedAt:     baseTime,
					FileType:       "md",
					Portfolio:      "Project1",
					Project:        "Documentation",
					ModifiedByName: "John Smith",
					Size:          1024,
				},
				{
					FilePath:       "/project1/src/main.go",
					ModifiedAt:     baseTime.Add(-1 * time.Hour),
					FileType:       "go",
					Portfolio:      "Project1",
					Project:        "Backend",
					ModifiedByName: "Jane Doe",
					Size:          2048,
				},
				{
					FilePath:       "/project2/ui/index.html",
					ModifiedAt:     baseTime.Add(-2 * time.Hour),
					FileType:       "html",
					Portfolio:      "Project2",
					Project:        "Frontend",
					ModifiedByName: "Alice Johnson",
					Size:          512,
				},
				{
					FilePath:       "/project2/ui/styles.css",
					ModifiedAt:     baseTime.Add(-3 * time.Hour),
					FileType:       "css",
					Portfolio:      "Project2",
					Project:        "Frontend",
					ModifiedByName: "Bob Wilson",
					Size:          256,
				},
				{
					FilePath:       "/project1/tests/main_test.go",
					ModifiedAt:     baseTime.Add(-4 * time.Hour),
					FileType:       "go",
					Portfolio:      "Project1",
					Project:        "Testing",
					ModifiedByName: "John Smith",
					Size:          1536,
				},
				{
					FilePath:       "/project2/docs/api.yaml",
					ModifiedAt:     baseTime.Add(-5 * time.Hour),
					FileType:       "yaml",
					Portfolio:      "Project2",
					Project:        "Documentation",
					LockHolderName: "Charlie Brown",
					Size:          768,
				},
				{
					FilePath:       "/project1/config/settings.json",
					ModifiedAt:     baseTime.Add(-6 * time.Hour),
					FileType:       "json",
					Portfolio:      "Project1",
					Project:        "Configuration",
					ModifiedByName: "Jane Doe",
					Size:          384,
				},
				{
					FilePath:       "/project2/src/utils.py",
					ModifiedAt:     baseTime.Add(-7 * time.Hour),
					FileType:       "py",
					Portfolio:      "Project2",
					Project:        "Backend",
					ModifiedByName: "Alice Johnson",
					Size:          896,
				},
				{
					FilePath:       "/project1/assets/logo.png",
					ModifiedAt:     baseTime.Add(-8 * time.Hour),
					FileType:       "png",
					Portfolio:      "Project1",
					Project:        "Assets",
					ModifiedByName: "Bob Wilson",
					Size:          4096,
				},
				{
					FilePath:       "/project2/README.md",
					ModifiedAt:     baseTime.Add(-9 * time.Hour),
					FileType:       "md",
					Portfolio:      "Project2",
					Project:        "Documentation",
					ModifiedByName: "Charlie Brown",
					Size:          128,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			report := GenerateReport(tc.files)
			t.Logf("Generated Report:\n%s", report)
		})
	}
}
