package core

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agents"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/db"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

func TestNewMonitor(t *testing.T) {
	// Create a temporary database file
	tmpDB := "test.db"
	defer os.Remove(tmpDB)

	tests := []struct {
		name          string
		dbConn        string
		dropboxToken  string
		expectedError string
	}{
		{
			name:          "Valid configuration",
			dbConn:        tmpDB,
			dropboxToken:  "test-token",
			expectedError: "",
		},
		{
			name:          "Missing access token",
			dbConn:        tmpDB,
			dropboxToken:  "",
			expectedError: "DROPBOX_ACCESS_TOKEN not set",
		},
		{
			name:          "Invalid DB connection string",
			dbConn:        "/invalid/path/test.db",
			dropboxToken:  "test-token",
			expectedError: "error initializing database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor, err := NewMonitor(tt.dbConn, tt.dropboxToken)
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if monitor != nil {
				monitor.Close()
			}
		})
	}
}

type mockDatabaseAgent struct {
	agents.DatabaseAgent
}

func (m *mockDatabaseAgent) StoreChange(ctx context.Context, change models.FileMetadata) error {
	return nil
}

func (m *mockDatabaseAgent) StoreAnalysis(ctx context.Context, content models.FileContent) error {
	return nil
}

func (m *mockDatabaseAgent) GetAnalysis(ctx context.Context, path string) (string, bool, error) {
	return "", false, nil
}

func (m *mockDatabaseAgent) Close() error {
	return nil
}

type mockDB struct {
	db *db.DB
}

func newMockDB() *db.DB {
	dbPath := "test.db"
	db, _ := db.NewDB(dbPath)
	return db
}

type mockDropboxClient struct {
	*dropbox.DropboxClient
}

func newMockDropboxClient() *dropbox.DropboxClient {
	return &dropbox.DropboxClient{}
}

func TestMonitorClose(t *testing.T) {
	tests := []struct {
		name    string
		monitor *Monitor
		wantErr bool
	}{
		{
			name: "Valid monitor",
			monitor: &Monitor{
				DB:            newMockDB(),
				DBAgent:       &mockDatabaseAgent{},
				DropboxClient: newMockDropboxClient(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.monitor.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Monitor.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}
