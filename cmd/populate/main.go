package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get Dropbox token
	token := os.Getenv("DROPBOX_ACCESS_TOKEN")
	if token == "" {
		log.Fatal("DROPBOX_ACCESS_TOKEN not set in .env")
	}

	// Create Dropbox client
	client, err := dropbox.NewDropboxClient(token)
	if err != nil {
		log.Fatalf("Error creating Dropbox client: %v", err)
	}

	// List first 10 files from root directory
	log.Println("Listing first 10 files from Dropbox...")
	files, err := client.ListFolder(context.Background(), "")
	if err != nil {
		log.Fatalf("Error listing files: %v", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", "data/dropbox_monitor.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Create table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS files (
			path TEXT PRIMARY KEY,
			name TEXT,
			size INTEGER,
			modified DATETIME,
			is_deleted BOOLEAN
		)
	`)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	// Insert first 10 files into database
	count := 0
	for _, file := range files {
		if count >= 10 {
			break
		}

		_, err := db.Exec(`
			INSERT OR REPLACE INTO files (path, name, size, modified, is_deleted)
			VALUES (?, ?, ?, ?, ?)
		`, file.Path, file.Name, file.Size, file.Modified, file.IsDeleted)
		if err != nil {
			log.Printf("Error inserting file %s: %v", file.Path, err)
			continue
		}
		count++
	}

	log.Printf("Successfully populated %d files from Dropbox", count)
}
