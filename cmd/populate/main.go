package main

import (
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

	// Open database connection
	db, err := sql.Open("sqlite3", "data/dropbox_monitor.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Create Dropbox client
	client, err := dropbox.NewDropboxClient(token, db)
	if err != nil {
		log.Fatalf("Error creating Dropbox client: %v", err)
	}

	// Populate first 10 files
	log.Println("Populating first 10 files from Dropbox...")
	if err := client.PopulateFirstNFiles(10); err != nil {
		log.Fatalf("Error populating files: %v", err)
	}
	log.Println("Successfully populated 10 files from Dropbox")
}
