package reporting

import (
	"context"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/reporting/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNotifier implements notify.Notifier for testing
type mockNotifier struct {
	sentMessages int
	lastSubject  string
	lastMessage  string
	shouldError  bool
}

func (m *mockNotifier) Send(ctx context.Context, subject, message string) error {
	if m.shouldError {
		return assert.AnError
	}
	m.sentMessages++
	m.lastSubject = subject
	m.lastMessage = message
	return nil
}

func TestReporter_GenerateReport(t *testing.T) {
	notifier := &mockNotifier{}
	reporter := NewReporter(notifier)

	ctx := context.Background()
	changes := []models.FileChange{
		{
			Path:      "/docs/file1.txt",
			Extension: ".txt",
			Directory: "/docs",
			ModTime:   time.Now(),
			Size:      1024,
		},
		{
			Path:      "/images/photo.jpg",
			Extension: ".jpg",
			Directory: "/images",
			ModTime:   time.Now(),
			Size:      2048,
		},
	}

	report, err := reporter.GenerateReport(ctx, changes)
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	if report.TotalChanges != 2 {
		t.Errorf("Expected 2 changes, got %d", report.TotalChanges)
	}

	if count := report.ExtensionCount[".txt"]; count != 1 {
		t.Errorf("Expected 1 .txt file, got %d", count)
	}

	if count := report.ExtensionCount[".jpg"]; count != 1 {
		t.Errorf("Expected 1 .jpg file, got %d", count)
	}
}

func TestReporter_Lifecycle(t *testing.T) {
	notifier := &mockNotifier{}
	reporter := NewReporter(notifier)
	ctx := context.Background()

	// Test Start
	if err := reporter.Start(ctx); err != nil {
		t.Errorf("Start failed: %v", err)
	}
	if reporter.State() != lifecycle.StateRunning {
		t.Errorf("Expected state Running, got %v", reporter.State())
	}

	// Test Health
	if err := reporter.Health(ctx); err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// Test Stop
	if err := reporter.Stop(ctx); err != nil {
		t.Errorf("Stop failed: %v", err)
	}
	if reporter.State() != lifecycle.StateStopped {
		t.Errorf("Expected state Stopped, got %v", reporter.State())
	}
}

func TestReport_GetTopItems(t *testing.T) {
	report := models.NewReport(models.FileListReport)
	
	// Add some test changes
	changes := []models.FileChange{
		{Path: "/docs/file1.txt", Extension: ".txt", Directory: "/docs"},
		{Path: "/docs/file2.txt", Extension: ".txt", Directory: "/docs"},
		{Path: "/images/photo1.jpg", Extension: ".jpg", Directory: "/images"},
		{Path: "/images/photo2.jpg", Extension: ".jpg", Directory: "/images"},
		{Path: "/images/photo3.jpg", Extension: ".jpg", Directory: "/images"},
	}

	for _, change := range changes {
		report.AddChange(change)
	}

	// Test GetTopExtensions
	topExt := report.GetTopExtensions(2)
	if len(topExt) != 2 {
		t.Errorf("Expected 2 top extensions, got %d", len(topExt))
	}
	if topExt[0] != ".jpg" { // .jpg should be first with 3 occurrences
		t.Errorf("Expected .jpg as top extension, got %s", topExt[0])
	}

	// Test GetTopDirectories
	topDirs := report.GetTopDirectories(2)
	if len(topDirs) != 2 {
		t.Errorf("Expected 2 top directories, got %d", len(topDirs))
	}
	if topDirs[0] != "/images" { // /images should be first with 3 files
		t.Errorf("Expected /images as top directory, got %s", topDirs[0])
	}
}

func TestGenerateReport(t *testing.T) {
	tests := []struct {
		name        string
		changes     []models.FileChange
		notifier    *mockNotifier
		wantSubject string
		wantErr     bool
	}{
		{
			name: "successful report generation",
			changes: []models.FileChange{
				{
					Path:      "/test/file1.txt",
					Extension: ".txt",
					Directory: "/test",
					ModTime:   time.Now(),
					Size:      1024,
				},
			},
			notifier:    &mockNotifier{},
			wantSubject: "Dropbox File Changes Report",
			wantErr:     false,
		},
		{
			name:        "empty changes list",
			changes:     []models.FileChange{},
			notifier:    &mockNotifier{},
			wantSubject: "Dropbox File Changes Report",
			wantErr:     false,
		},
		{
			name: "notifier error",
			changes: []models.FileChange{
				{
					Path:      "/test/file1.txt",
					Extension: ".txt",
					Directory: "/test",
					ModTime:   time.Now(),
					Size:      1024,
				},
			},
			notifier:    &mockNotifier{shouldError: true},
			wantSubject: "Dropbox File Changes Report",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewReporter(tt.notifier)
			err := reporter.Start(context.Background())
			require.NoError(t, err)

			report := models.NewReport(models.FileListReport)
			for _, change := range tt.changes {
				report.AddChange(change)
			}

			err = reporter.SendReport(context.Background(), report)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendReport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.Equal(t, tt.wantSubject, tt.notifier.lastSubject)
				assert.NotEmpty(t, tt.notifier.lastMessage)
			}
		})
	}
}

func TestActivityPattern_Analysis(t *testing.T) {
	pattern := models.NewActivityPattern()

	// Add test time ranges
	now := time.Now()
	ranges := []models.TimeRange{
		{
			Start:    now.Add(-2 * time.Hour),
			End:      now.Add(-1 * time.Hour),
			Changes:  5,
			Patterns: []string{"docs update"},
		},
		{
			Start:    now.Add(-1 * time.Hour),
			End:      now,
			Changes:  10,
			Patterns: []string{"image upload"},
		},
	}

	for _, tr := range ranges {
		pattern.AddTimeRange(tr)
	}

	// Test GetMostActiveTimeRange
	mostActive := pattern.GetMostActiveTimeRange()
	if mostActive == nil {
		t.Fatal("Expected most active time range, got nil")
	}
	if mostActive.Changes != 10 {
		t.Errorf("Expected 10 changes in most active range, got %d", mostActive.Changes)
	}

	// Add test hotspots
	hotspots := []models.DirectoryHotspot{
		{
			Path:        "/docs",
			ChangeCount: 5,
			LastActive:  now,
		},
		{
			Path:        "/images",
			ChangeCount: 10,
			LastActive:  now,
		},
	}

	for _, hs := range hotspots {
		pattern.AddHotspot(hs)
	}

	// Test GetTopHotspots
	topHotspots := pattern.GetTopHotspots(1)
	if len(topHotspots) != 1 {
		t.Errorf("Expected 1 top hotspot, got %d", len(topHotspots))
	}
	if topHotspots[0].Path != "/images" {
		t.Errorf("Expected /images as top hotspot, got %s", topHotspots[0].Path)
	}
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
			wantSubject: "Dropbox File Changes Report",
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
			err := r.Start(context.Background())
			require.NoError(t, err)

			report := &models.Report{
				Changes:     []models.FileChange{},
				GeneratedAt: time.Now(),
			}

			err = r.SendReport(context.Background(), report)
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
