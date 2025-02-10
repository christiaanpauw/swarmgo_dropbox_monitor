package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileContentStorage(t *testing.T) {
	// Create a temporary database for testing
	tmpDir, err := os.MkdirTemp("", "dropbox_monitor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewDB("file:" + dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test file change
	fileChange := &FileChange{
		FilePath:        "/test/document.txt",
		ModifiedAt:      time.Now(),
		FileType:        "text",
		Portfolio:       "TestPortfolio",
		Project:         "TestProject",
		DocumentType:    "Document",
		Author:          "Test Author",
		ContentHash:     "hash123",
		DropboxID:       "dbx123",
		DropboxRev:      "rev123",
		ClientModified:  time.Now(),
		ServerModified:  time.Now(),
		Size:            1024,
		IsDownloadable:  true,
		ModifiedByID:    "user123",
		ModifiedByName:  "Test User",
		SharedFolderID:  "folder123",
		LockHolderName:  "",
		LockHolderID:    "",
		LockCreatedAt:   time.Time{},
		CreatedAt:       time.Now(),
	}

	// Save file change
	err = db.SaveFileChange(ctx, fileChange)
	if err != nil {
		t.Fatalf("Failed to save file change: %v", err)
	}

	// Test file content
	fileContent := &FileContent{
		FileChangeID: 1, // Should be 1 as it's the first record
		Content:     "This is a test document",
		ContentType: "text/plain",
	}

	// Save file content
	err = db.SaveFileContent(ctx, fileContent)
	if err != nil {
		t.Fatalf("Failed to save file content: %v", err)
	}

	// Verify the content was saved
	var count int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM file_contents").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count file contents: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 file content record, got %d", count)
	}

	// Verify the content matches
	var savedContent string
	err = db.DB.QueryRow("SELECT content FROM file_contents WHERE file_change_id = 1").Scan(&savedContent)
	if err != nil {
		t.Fatalf("Failed to retrieve file content: %v", err)
	}
	if savedContent != "This is a test document" {
		t.Errorf("Content mismatch. Expected 'This is a test document', got '%s'", savedContent)
	}
}
