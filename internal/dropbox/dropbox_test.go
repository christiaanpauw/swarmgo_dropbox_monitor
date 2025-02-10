package dropbox

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	// Load environment variables from .env file
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	// Set environment variables for testing
	if err := os.Setenv("DROPBOX_MONITOR_DB", "test.db"); err != nil {
		fmt.Printf("Error setting environment variable: %v\n", err)
	}
}

func setupTestDB() *sql.DB {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "dropbox_monitor_test")
	if err != nil {
		log.Fatalf("Error creating temp directory: %v", err)
	}

	// Set up the test database path
	dbPath := filepath.Join(tempDir, "test.db")
	dbConn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	return dbConn
}

// mockHTTPClient is a mock HTTP client for testing
type mockHTTPClient struct {
	files []FileMetadata
}

func (m *mockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.URL.Path {
	case "/2/files/list_folder":
		entries := make([]FileMetadata, len(m.files))
		copy(entries, m.files)
		response := struct {
			Entries []FileMetadata `json:"entries"`
			HasMore bool          `json:"has_more"`
			Cursor  string        `json:"cursor"`
		}{
			Entries: entries,
			HasMore: false,
			Cursor:  "mock_cursor",
		}
		body, _ := json.Marshal(response)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}, nil
	case "/2/files/download":
		for _, file := range m.files {
			if file.PathDisplay == req.Header.Get("Dropbox-API-Arg") {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("mock content")),
				}, nil
			}
		}
		return nil, fmt.Errorf("file not found")
	default:
		return nil, fmt.Errorf("unknown endpoint")
	}
}

func TestTestConnection(t *testing.T) {
	now := time.Now()
	// Create a mock Dropbox client that returns success
	client, err := NewDropboxClient("mock_token", setupTestDB())
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}
	client.httpClient = &http.Client{
		Transport: &mockHTTPClient{
			files: []FileMetadata{
				{
					Name:           "test.txt",
					PathDisplay:    "/test.txt",
					ContentHash:    "hash123",
					Size:          1024,
					ClientModified: now.Format(time.RFC3339),
					ServerModified: now.Format(time.RFC3339),
					IsDownloadable: true,
					ID:            "dbx123",
					Rev:           "rev123",
				},
			},
		},
	}

	// Test the connection
	_, err = client.ListFiles("/")
	if err != nil {
		t.Errorf("Error testing connection: %v", err)
	}
}

func TestPopulateFirstNFiles(t *testing.T) {
	now := time.Now()
	// Create a mock Dropbox client that returns a list of files
	client, err := NewDropboxClient("mock_token", setupTestDB())
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}
	client.httpClient = &http.Client{
		Transport: &mockHTTPClient{
			files: []FileMetadata{
				{
					Name:           "test1.txt",
					PathDisplay:    "/test1.txt",
					ContentHash:    "hash123",
					Size:          1024,
					ClientModified: now.Format(time.RFC3339),
					ServerModified: now.Format(time.RFC3339),
					IsDownloadable: true,
					ID:            "dbx123",
					Rev:           "rev123",
				},
				{
					Name:           "test2.txt",
					PathDisplay:    "/test2.txt",
					ContentHash:    "hash456",
					Size:          2048,
					ClientModified: now.Format(time.RFC3339),
					ServerModified: now.Format(time.RFC3339),
					IsDownloadable: true,
					ID:            "dbx456",
					Rev:           "rev456",
				},
			},
		},
	}

	// Test populating first N files
	err = client.PopulateFirstNFiles(10)
	if err != nil {
		t.Errorf("Error populating first N files: %v", err)
	}
}
