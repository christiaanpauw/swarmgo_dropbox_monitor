package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

// InitDB creates the database if it doesn't exist and initializes the schema
func InitDB(host, port, user, password, dbname string) error {
	// Connect to postgres database to create our database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, password)
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error connecting to postgres: %v", err)
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbname).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking if database exists: %v", err)
	}

	// Create database if it doesn't exist
	if !exists {
		log.Printf("Creating database %s...", dbname)
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
		if err != nil {
			return fmt.Errorf("error creating database: %v", err)
		}
	}

	// Connect to our new database
	connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error connecting to new database: %v", err)
	}
	defer db.Close()

	// Create pgvector extension if it doesn't exist
	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS vector`)
	if err != nil {
		return fmt.Errorf("error creating vector extension: %v", err)
	}

	// Create tables
	tables := []string{
		`CREATE TABLE IF NOT EXISTS file_changes (
			id SERIAL PRIMARY KEY,
			file_path TEXT NOT NULL,
			modified_at TIMESTAMP NOT NULL,
			file_type TEXT,
			portfolio TEXT,
			project TEXT,
			document_type TEXT,
			author TEXT,
			content_hash TEXT,
			embedding vector(1536),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS file_contents (
			id SERIAL PRIMARY KEY,
			file_change_id INTEGER REFERENCES file_changes(id),
			content TEXT NOT NULL,
			content_type TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS daily_summaries (
			id SERIAL PRIMARY KEY,
			summary_date DATE NOT NULL,
			total_files INTEGER NOT NULL,
			narrative_summary TEXT NOT NULL,
			portfolio_stats JSONB,
			project_stats JSONB,
			author_stats JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	// Create indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_file_changes_modified_at ON file_changes(modified_at)`,
		`CREATE INDEX IF NOT EXISTS idx_file_changes_portfolio ON file_changes(portfolio)`,
		`CREATE INDEX IF NOT EXISTS idx_file_changes_project ON file_changes(project)`,
		`CREATE INDEX IF NOT EXISTS idx_file_changes_author ON file_changes(author)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_summaries_date ON daily_summaries(summary_date)`,
	}

	// Execute all table creation statements
	for _, table := range tables {
		_, err = db.Exec(table)
		if err != nil {
			return fmt.Errorf("error creating table: %v", err)
		}
	}

	// Execute all index creation statements
	for _, index := range indexes {
		_, err = db.Exec(index)
		if err != nil {
			return fmt.Errorf("error creating index: %v", err)
		}
	}

	log.Println("Database initialization completed successfully")
	return nil
}

// ParseConnectionString parses a PostgreSQL connection string into its components
func ParseConnectionString(connStr string) (host, port, user, password, dbname string, err error) {
	// Default values
	host = "localhost"
	port = "5432"

	// Split the connection string into parts
	parts := strings.Split(connStr, " ")
	for _, part := range parts {
		keyVal := strings.Split(part, "=")
		if len(keyVal) != 2 {
			continue
		}
		key := keyVal[0]
		val := keyVal[1]

		switch key {
		case "host":
			host = val
		case "port":
			port = val
		case "user":
			user = val
		case "password":
			password = val
		case "dbname":
			dbname = val
		}
	}

	// Validate required fields
	if user == "" || password == "" || dbname == "" {
		return "", "", "", "", "", fmt.Errorf("missing required database connection parameters")
	}

	return host, port, user, password, dbname, nil
}
