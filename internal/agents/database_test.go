package agents

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) string {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "dropbox_monitor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create the database path
	dbPath := filepath.Join(tempDir, "test.db")

	return dbPath
}

func cleanupTestDB(t *testing.T, dbPath string) {
	// Remove the test database file and its directory
	dir := filepath.Dir(dbPath)
	if err := os.RemoveAll(dir); err != nil {
		t.Errorf("Failed to cleanup test DB: %v", err)
	}
}

func TestDatabaseAgent_Execute(t *testing.T) {
	dbPath := setupTestDB(t)
	defer cleanupTestDB(t, dbPath)

	agent := NewDatabaseAgent("sqlite3://" + dbPath)
	now := time.Now()

	changes := []map[string]interface{}{
		{
			"Path":    "/test1.txt",
			"Type":    "modified",
			"ModTime": now,
			"Metadata": map[string]interface{}{
				"id":           "id1",
				"name":         "test1.txt",
				"path_display": "/test1.txt",
				"size":         100,
			},
		},
	}

	result := agent.Execute(nil, map[string]interface{}{
		"changes": changes,
	})

	if !result.Success {
		t.Errorf("Execute() failed: %v", result.Error)
	}
}

func TestDatabaseAgent_Connect(t *testing.T) {
	dbPath := setupTestDB(t)
	defer cleanupTestDB(t, dbPath)

	agent := NewDatabaseAgent("sqlite3://" + dbPath)
	if err := agent.Connect(); err != nil {
		t.Errorf("Connect() failed: %v", err)
	}
}

func TestDatabaseAgent_Close(t *testing.T) {
	dbPath := setupTestDB(t)
	defer cleanupTestDB(t, dbPath)

	agent := NewDatabaseAgent("sqlite3://" + dbPath)
	if err := agent.Connect(); err != nil {
		t.Errorf("Connect() failed: %v", err)
	}

	if err := agent.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}
