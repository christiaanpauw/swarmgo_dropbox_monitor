package agents

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/db"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// DatabaseAgent interface for database operations
type DatabaseAgent interface {
	lifecycle.Component
	StoreChange(ctx context.Context, change models.FileMetadata) error
	GetLatestChanges(ctx context.Context, limit int) ([]models.FileMetadata, error)
	GetChanges(ctx context.Context, startTime, endTime string) ([]models.FileMetadata, error)
	StoreFileContent(ctx context.Context, content *models.FileContent) error
	Close() error
}

// databaseAgent implements the DatabaseAgent interface
type databaseAgent struct {
	*lifecycle.BaseComponent
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

	agent := &databaseAgent{
		BaseComponent: lifecycle.NewBaseComponent("DatabaseAgent"),
		database:      database,
	}

	return agent, nil
}

// Initialize implements lifecycle.Component
func (a *databaseAgent) Initialize(ctx context.Context) error {
	a.SetState(lifecycle.StateInitialized)
	return nil
}

// Start implements lifecycle.Component
func (a *databaseAgent) Start(ctx context.Context) error {
	return a.DefaultStart(ctx)
}

// Stop implements lifecycle.Component
func (a *databaseAgent) Stop(ctx context.Context) error {
	if err := a.DefaultStop(ctx); err != nil {
		return err
	}
	return a.Close()
}

// Health implements lifecycle.Component
func (a *databaseAgent) Health(ctx context.Context) error {
	if a.database == nil {
		return fmt.Errorf("database connection is nil")
	}
	if err := a.database.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	return nil
}

// StoreChange stores a file change in the database
func (a *databaseAgent) StoreChange(ctx context.Context, change models.FileMetadata) error {
	dbChange := &db.FileChange{
		FilePath:       change.Path,
		ModifiedAt:     change.ModTime,
		IsDownloadable: true,
		CreatedAt:      time.Now(),
		Size:          change.Size,
	}

	if err := a.database.SaveFileChange(ctx, dbChange); err != nil {
		return fmt.Errorf("store file change: %w", err)
	}

	return nil
}

// GetLatestChanges retrieves the latest changes from the database
func (a *databaseAgent) GetLatestChanges(ctx context.Context, limit int) ([]models.FileMetadata, error) {
	dbChanges, err := a.database.GetRecentFileChanges(ctx, time.Now().AddDate(0, 0, -7)) // Get last week's changes
	if err != nil {
		return nil, fmt.Errorf("get latest changes: %w", err)
	}

	// Convert db.FileChange to models.FileMetadata and limit the results
	changes := make([]models.FileMetadata, 0, limit)
	for i, dbChange := range dbChanges {
		if i >= limit {
			break
		}
		changes = append(changes, models.FileMetadata{
			Path:     dbChange.FilePath,
			Size:     dbChange.Size,
			ModTime:  dbChange.ModifiedAt,
		})
	}

	return changes, nil
}

// GetChanges retrieves changes within a time range
func (a *databaseAgent) GetChanges(ctx context.Context, startTime, endTime string) ([]models.FileMetadata, error) {
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return nil, fmt.Errorf("parse start time: %w", err)
	}

	end, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return nil, fmt.Errorf("parse end time: %w", err)
	}

	dbChanges, err := a.database.GetRecentFileChanges(ctx, start)
	if err != nil {
		return nil, fmt.Errorf("get changes: %w", err)
	}

	// Filter changes by time range and convert to FileMetadata
	changes := make([]models.FileMetadata, 0)
	for _, dbChange := range dbChanges {
		if dbChange.ModifiedAt.After(start) && dbChange.ModifiedAt.Before(end) {
			changes = append(changes, models.FileMetadata{
				Path:     dbChange.FilePath,
				Size:     dbChange.Size,
				ModTime:  dbChange.ModifiedAt,
			})
		}
	}

	return changes, nil
}

// StoreFileContent stores file content in the database
func (a *databaseAgent) StoreFileContent(ctx context.Context, content *models.FileContent) error {
	dbContent := &db.FileContent{
		Content:     "",  // We don't store the actual content
		ContentType: content.ContentType,
		CreatedAt:   time.Now(),
	}

	if err := a.database.SaveFileContent(ctx, dbContent); err != nil {
		return fmt.Errorf("store file content: %w", err)
	}

	return nil
}

// Close closes the database connection
func (a *databaseAgent) Close() error {
	if err := a.database.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}
	return nil
}
