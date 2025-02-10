package agents

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type databaseAgent struct {
	db *sql.DB
}

func NewDatabaseAgent(connStr string) DatabaseAgent {
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		return nil
	}

	// Create tables if they don't exist
	err = createTables(db)
	if err != nil {
		fmt.Printf("Error creating tables: %v\n", err)
		return nil
	}

	return &databaseAgent{
		db: db,
	}
}

func createTables(db *sql.DB) error {
	// Create file_changes table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS file_changes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL,
			mod_time TIMESTAMP NOT NULL,
			metadata TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating file_changes table: %v", err)
	}

	// Create file_analysis table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS file_analysis (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL,
			summary TEXT,
			keywords TEXT,
			categories TEXT,
			sensitivity TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating file_analysis table: %v", err)
	}

	return nil
}

func (da *databaseAgent) StoreChange(ctx context.Context, change FileChange) error {
	query := `
		INSERT INTO file_changes (path, mod_time, metadata)
		VALUES ($1, $2, $3)
		ON CONFLICT (path) DO UPDATE
		SET mod_time = EXCLUDED.mod_time,
			metadata = EXCLUDED.metadata
	`

	_, err := da.db.ExecContext(ctx, query, change.Path, change.ModTime, change.Metadata)
	if err != nil {
		return fmt.Errorf("error storing file change: %v", err)
	}

	return nil
}

func (da *databaseAgent) StoreAnalysis(ctx context.Context, path string, content *FileContent) error {
	query := `
		INSERT INTO file_analysis (path, summary, keywords, categories, sensitivity)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (path) DO UPDATE
		SET summary = EXCLUDED.summary,
			keywords = EXCLUDED.keywords,
			categories = EXCLUDED.categories,
			sensitivity = EXCLUDED.sensitivity
	`

	_, err := da.db.ExecContext(ctx, query,
		path,
		content.Summary,
		content.Keywords,
		content.Categories,
		content.Sensitivity,
	)
	if err != nil {
		return fmt.Errorf("error storing content analysis: %v", err)
	}

	return nil
}

func (da *databaseAgent) Close() error {
	if da.db != nil {
		return da.db.Close()
	}
	return nil
}
