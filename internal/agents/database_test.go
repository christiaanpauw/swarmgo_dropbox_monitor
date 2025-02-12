package agents

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) (DatabaseAgent, string, func()) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "dropbox_monitor_test")
	assert.NoError(t, err)

	// Set environment variable for database path
	dbPath := filepath.Join(tempDir, "test.db")
	oldDBPath := os.Getenv("DROPBOX_MONITOR_DB")
	os.Setenv("DROPBOX_MONITOR_DB", dbPath)

	// Initialize database agent
	agent, err := NewDatabaseAgent()
	assert.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		agent.Close()
		os.RemoveAll(tempDir)
		if oldDBPath != "" {
			os.Setenv("DROPBOX_MONITOR_DB", oldDBPath)
		} else {
			os.Unsetenv("DROPBOX_MONITOR_DB")
		}
	}

	return agent, tempDir, cleanup
}

func TestDatabaseAgent_StoreFileContent(t *testing.T) {
	agent, _, cleanup := setupTestDB(t)
	defer cleanup()

	// Test storing file content
	content := &models.FileContent{
		ContentType: "text/plain",
	}

	err := agent.StoreFileContent(context.Background(), content)
	assert.NoError(t, err)

	// Test storing another file content
	content = &models.FileContent{
		ContentType: "application/json",
	}
	err = agent.StoreFileContent(context.Background(), content)
	assert.NoError(t, err)
}

func TestDatabaseAgent_GetRecentChanges(t *testing.T) {
	agent, _, cleanup := setupTestDB(t)
	defer cleanup()

	// Store some test data
	content1 := &models.FileMetadata{
		Path:    "/test1.txt",
		Size:    100,
		ModTime: time.Now(),
	}
	err := agent.StoreChange(context.Background(), *content1)
	assert.NoError(t, err)

	content2 := &models.FileMetadata{
		Path:    "/test2.txt",
		Size:    200,
		ModTime: time.Now(),
	}
	err = agent.StoreChange(context.Background(), *content2)
	assert.NoError(t, err)

	// Test getting latest changes
	changes, err := agent.GetLatestChanges(context.Background(), 10)
	assert.NoError(t, err)
	assert.Len(t, changes, 2)

	// Verify the changes are returned in the correct order
	assert.Equal(t, content2.Path, changes[0].Path)
	assert.Equal(t, content1.Path, changes[1].Path)
}

func TestDatabaseAgent_GetChanges(t *testing.T) {
	agent, _, cleanup := setupTestDB(t)
	defer cleanup()

	startTime := time.Now().Add(-time.Hour).Format(time.RFC3339)
	endTime := time.Now().Add(time.Hour).Format(time.RFC3339)

	// Store some test data
	content := &models.FileMetadata{
		Path:    "/test.txt",
		Size:    100,
		ModTime: time.Now(),
	}
	err := agent.StoreChange(context.Background(), *content)
	assert.NoError(t, err)

	// Test getting changes within time range
	changes, err := agent.GetChanges(context.Background(), startTime, endTime)
	assert.NoError(t, err)
	assert.Len(t, changes, 1)
	assert.Equal(t, content.Path, changes[0].Path)
}

func TestDatabaseAgent_Health(t *testing.T) {
	agent, _, cleanup := setupTestDB(t)
	defer cleanup()

	// Test health check
	err := agent.Health(context.Background())
	assert.NoError(t, err)

	// Test health check after closing connection
	agent.Close()
	err = agent.Health(context.Background())
	assert.Error(t, err)
}
