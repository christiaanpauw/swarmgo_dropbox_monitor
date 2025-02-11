package agents

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseAgent_StoreFileContent(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "dropbox_monitor_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Initialize database agent
	agent, err := NewDatabaseAgent()
	assert.NoError(t, err)

	// Test storing file content
	content := &models.FileContent{
		Path:        "/test.txt",
		ContentType: "text/plain",
		Size:        100,
		ContentHash: "abc123",
		IsBinary:    false,
	}

	err = agent.StoreFileContent(context.Background(), content)
	assert.NoError(t, err)
}

func TestDatabaseAgent_GetRecentChanges(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "dropbox_monitor_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Initialize database agent
	agent, err := NewDatabaseAgent()
	assert.NoError(t, err)

	// Store an old file that shouldn't be included in recent changes
	content3 := &models.FileContent{
		Path:        "/test3.txt",
		ContentType: "text/plain",
		Size:        300,
		ContentHash: "ghi789_v1", 
		IsBinary:    false,
	}
	err = agent.StoreFileContent(context.Background(), content3)
	assert.NoError(t, err)

	// Sleep to ensure time difference
	time.Sleep(time.Second)

	// Record the time for our "since" filter
	since := time.Now()

	// Sleep again to ensure new files are after "since"
	time.Sleep(time.Second)

	// Store some test data
	content1 := &models.FileContent{
		Path:        "/test1.txt",
		ContentType: "text/plain",
		Size:        100,
		ContentHash: "abc123_v1", 
		IsBinary:    false,
	}
	err = agent.StoreFileContent(context.Background(), content1)
	assert.NoError(t, err)

	content2 := &models.FileContent{
		Path:        "/test2.txt",
		ContentType: "text/plain",
		Size:        200,
		ContentHash: "def456_v1", 
		IsBinary:    false,
	}
	err = agent.StoreFileContent(context.Background(), content2)
	assert.NoError(t, err)

	// Test getting recent changes
	changes, err := agent.GetRecentChanges(context.Background(), since)
	assert.NoError(t, err)
	assert.Len(t, changes, 2)
}
