package reporting

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNotifier implements notify.Notifier for testing
type mockNotifier struct {
	lastSubject string
	lastMessage string
	shouldError bool
}

func (m *mockNotifier) Send(ctx context.Context, subject, message string) error {
	if m.shouldError {
		return assert.AnError
	}
	m.lastSubject = subject
	m.lastMessage = message
	return nil
}

func TestGenerateReport(t *testing.T) {
	tests := []struct {
		name    string
		changes []models.FileMetadata
		want    *models.Report
		wantErr bool
	}{
		{
			name:    "Empty changes",
			changes: []models.FileMetadata{},
			want: &models.Report{
				Changes:     []models.FileChange{},
				GeneratedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "Single change",
			changes: []models.FileMetadata{
				{
					Path:     "/test/file.txt",
					Modified: time.Now(),
				},
			},
			want: &models.Report{
				Changes: []models.FileChange{
					{
						Path:      "/test/file.txt",
						Extension: "txt",
						Directory: "/test",
					},
				},
				ExtensionCount: map[string]int{"txt": 1},
				DirectoryCount: map[string]int{"/test": 1},
				TotalChanges:   1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReporter(&mockNotifier{})
			got, err := r.GenerateReport(context.Background(), tt.changes)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)

			// Compare everything except GeneratedAt
			assert.Equal(t, len(tt.want.Changes), len(got.Changes))
			if len(tt.want.Changes) > 0 {
				assert.Equal(t, tt.want.Changes[0].Path, got.Changes[0].Path)
				assert.Equal(t, tt.want.Changes[0].Extension, got.Changes[0].Extension)
				assert.Equal(t, tt.want.Changes[0].Directory, got.Changes[0].Directory)
			}
			assert.Equal(t, tt.want.ExtensionCount, got.ExtensionCount)
			assert.Equal(t, tt.want.DirectoryCount, got.DirectoryCount)
			assert.Equal(t, tt.want.TotalChanges, got.TotalChanges)
		})
	}
}

func TestGenerateHTML(t *testing.T) {
	report := &models.Report{
		Changes: []models.FileChange{
			{
				Path:      "/test/file.txt",
				Extension: "txt",
				Directory: "/test",
			},
		},
		ExtensionCount: map[string]int{"txt": 1},
		DirectoryCount: map[string]int{"/test": 1},
		GeneratedAt:    time.Now(),
		TotalChanges:   1,
	}

	r := NewReporter(&mockNotifier{})
	html, err := r.GenerateHTML(context.Background(), report)
	require.NoError(t, err)
	require.NotEmpty(t, html)

	// Check for expected HTML elements
	assert.Contains(t, html, "<html>")
	assert.Contains(t, html, "</html>")
	assert.Contains(t, html, "Dropbox Changes Report")
	assert.Contains(t, html, "/test/file.txt")
	assert.Contains(t, html, "txt: 1 files")
	assert.Contains(t, html, "/test: 1 files")
}

func TestSendReport(t *testing.T) {
	tests := []struct {
		name        string
		notifier    *mockNotifier
		wantErr     bool
		wantSubject string
	}{
		{
			name:        "Successful send",
			notifier:    &mockNotifier{},
			wantErr:     false,
			wantSubject: "Dropbox Changes Report",
		},
		{
			name:        "Send error",
			notifier:    &mockNotifier{shouldError: true},
			wantErr:     true,
			wantSubject: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReporter(tt.notifier)
			report := &models.Report{
				Changes:     []models.FileChange{},
				GeneratedAt: time.Now(),
			}

			err := r.SendReport(context.Background(), report)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantSubject, tt.notifier.lastSubject)
			assert.NotEmpty(t, tt.notifier.lastMessage)
		})
	}
}

func TestNarrativeReporter(t *testing.T) {
	report := &models.Report{
		Changes: []models.FileChange{
			{
				Path:      "/test/file1.txt",
				Extension: "txt",
				Directory: "/test",
			},
			{
				Path:      "/test/file2.txt",
				Extension: "txt",
				Directory: "/test",
			},
			{
				Path:      "/other/file.doc",
				Extension: "doc",
				Directory: "/other",
			},
		},
		ExtensionCount: map[string]int{"txt": 2, "doc": 1},
		DirectoryCount: map[string]int{"/test": 2, "/other": 1},
		GeneratedAt:    time.Now(),
		TotalChanges:   3,
	}

	r := NewNarrativeReporter(&mockNotifier{})

	// Test AnalyzeActivity
	pattern, err := r.AnalyzeActivity(context.Background(), report)
	require.NoError(t, err)
	require.NotNil(t, pattern)
	assert.Equal(t, 3, pattern.TotalChanges)
	assert.Contains(t, pattern.MainDirectories, "/test")
	assert.Contains(t, pattern.FileTypes, "txt")

	// Test GenerateNarrative
	narrative, err := r.GenerateNarrative(context.Background(), report, pattern)
	require.NoError(t, err)
	require.NotEmpty(t, narrative)

	// Check narrative content
	assert.True(t, strings.Contains(narrative, "Activity Report"))
	assert.True(t, strings.Contains(narrative, "Total Changes: 3"))
	assert.True(t, strings.Contains(narrative, "/test"))
	assert.True(t, strings.Contains(narrative, "txt files"))
}
