package reporting

import (
	"context"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNotifier implements notify.Notifier for testing
type mockNotifier struct {
	sentMessages int
	lastMessage  string
	shouldError  bool
}

func (m *mockNotifier) SendNotification(ctx context.Context, message string) error {
	if m.shouldError {
		return assert.AnError
	}
	m.sentMessages++
	m.lastMessage = message
	return nil
}

func createTestChanges() []models.FileChange {
	now := time.Now()
	return []models.FileChange{
		{
			Path:      "/docs/file1.txt",
			Extension: ".txt",
			Directory: "/docs",
			ModTime:   now,
			Modified:  now,
			Size:      1024,
		},
		{
			Path:      "/images/photo.jpg",
			Extension: ".jpg",
			Directory: "/images",
			ModTime:   now,
			Modified:  now,
			Size:      2048,
		},
		{
			Path:      "/docs/deleted.pdf",
			Extension: ".pdf",
			Directory: "/docs",
			ModTime:   now,
			Modified:  now,
			Size:      512,
			IsDeleted: true,
		},
	}
}

func TestReporter_GenerateReport(t *testing.T) {
	notifier := &mockNotifier{}
	reporter, err := NewReporter(notifier)
	require.NoError(t, err)
	require.NotNil(t, reporter)

	ctx := context.Background()
	changes := createTestChanges()

	// Test report generation
	report, err := reporter.GenerateReport(ctx, changes, models.FileListReport)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Verify report content
	content, ok := report.Metadata["content"]
	assert.True(t, ok)
	assert.Contains(t, content, "/docs/file1.txt")
	assert.Contains(t, content, "/images/photo.jpg")
	assert.Contains(t, content, "/docs/deleted.pdf")
	assert.Contains(t, content, "Total Changes: 3")
}

func TestReporter_SendReport(t *testing.T) {
	notifier := &mockNotifier{}
	reporter, err := NewReporter(notifier)
	require.NoError(t, err)
	require.NotNil(t, reporter)

	ctx := context.Background()
	changes := createTestChanges()

	// Test successful report sending
	report, err := reporter.GenerateReport(ctx, changes, models.FileListReport)
	require.NoError(t, err)
	err = reporter.SendReport(ctx, report)
	require.NoError(t, err)
	assert.Equal(t, 1, notifier.sentMessages)
	assert.Contains(t, notifier.lastMessage, "Total Changes: 3")

	// Test error case
	notifier.shouldError = true
	err = reporter.SendReport(ctx, report)
	require.Error(t, err)
}

func TestReporter_Lifecycle(t *testing.T) {
	notifier := &mockNotifier{}
	reporter, err := NewReporter(notifier)
	require.NoError(t, err)
	require.NotNil(t, reporter)

	// Test Start
	ctx := context.Background()
	err = reporter.Start(ctx)
	require.NoError(t, err)
	assert.Equal(t, lifecycle.StateRunning, reporter.State())

	// Test Health
	err = reporter.Health(ctx)
	require.NoError(t, err)

	// Test Stop
	err = reporter.Stop(ctx)
	require.NoError(t, err)
	assert.Equal(t, lifecycle.StateStopped, reporter.State())
}

func TestNewReporter(t *testing.T) {
	// Test with valid notifier
	notifier := &mockNotifier{}
	reporter, err := NewReporter(notifier)
	require.NoError(t, err)
	require.NotNil(t, reporter)

	// Test with nil notifier
	reporter, err = NewReporter(nil)
	require.Error(t, err)
	require.Nil(t, reporter)
}
