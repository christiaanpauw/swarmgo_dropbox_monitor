package agents

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/db"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// DatabaseAgent interface for database operations
type DatabaseAgent interface {
	StoreFileContent(ctx context.Context, content *models.FileContent) error
	GetRecentChanges(ctx context.Context, since time.Time) ([]*models.FileChange, error)
	Close() error
}

// databaseAgent implements the DatabaseAgent interface
type databaseAgent struct {
	database *db.DB
}

// NewDatabaseAgent creates a new database agent
func NewDatabaseAgent() (DatabaseAgent, error) {
	// Get database path from environment variable or use default
	dbPath := os.Getenv("DROPBOX_MONITOR_DB")
	if dbPath == "" {
		// Create data directory if it doesn't exist
		if err := os.MkdirAll("data", 0755); err != nil {
			return nil, fmt.Errorf("create data directory: %w", err)
		}
		dbPath = filepath.Join("data", "dropbox_monitor.db")
	}

	database, err := db.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}

	return &databaseAgent{
		database: database,
	}, nil
}

// StoreFileContent stores file content in the database
func (a *databaseAgent) StoreFileContent(ctx context.Context, content *models.FileContent) error {
	// Convert models.FileContent to db.FileContent
	dbContent := &db.FileContent{
		Content:     content.Summary, // Use summary as content
		ContentType: content.ContentType,
	}

	if err := a.database.SaveFileContent(ctx, dbContent); err != nil {
		return fmt.Errorf("store file content: %w", err)
	}

	// Also store a file change record
	dbChange := &db.FileChange{
		FilePath:    content.Path,
		ModifiedAt:  time.Now(),
		ContentHash: content.ContentHash,
		Size:        content.Size,
	}

	if err := a.database.SaveFileChange(ctx, dbChange); err != nil {
		return fmt.Errorf("store file change: %w", err)
	}

	return nil
}

// GetRecentChanges returns recent file changes
func (a *databaseAgent) GetRecentChanges(ctx context.Context, since time.Time) ([]*models.FileChange, error) {
	dbChanges, err := a.database.GetRecentFileChanges(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("get recent changes: %w", err)
	}

	// Convert db.FileChange to models.FileChange
	var changes []*models.FileChange
	for _, dbChange := range dbChanges {
		changes = append(changes, &models.FileChange{
			Path:      dbChange.FilePath,
			ModTime:   dbChange.ModifiedAt,
			IsDeleted: !dbChange.IsDownloadable,
		})
	}

	return changes, nil
}

// Close closes the database connection
func (a *databaseAgent) Close() error {
	if err := a.database.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}
	return nil
}
