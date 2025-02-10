package core

import (
	"fmt"
	"log"
	"os"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agents"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/db"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
	"github.com/joho/godotenv"
)

// Monitor represents the main application monitor
type Monitor struct {
	DB            *db.DB
	DBAgent       *agents.DatabaseAgent
	DropboxClient *dropbox.DropboxClient
}

func NewMonitor(dbConnStr string) (*Monitor, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v\nPlease ensure .env file exists and is readable", err)
	}

	// Initialize DB
	log.Println("Initializing database connection...")
	db, err := db.NewDB(dbConnStr)
	if err != nil {
		return nil, fmt.Errorf("error initializing database: %v", err)
	}

	// Initialize Dropbox client
	log.Println("Initializing Dropbox client...")
	accessToken := os.Getenv("DROPBOX_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, fmt.Errorf("DROPBOX_ACCESS_TOKEN not set in environment variables")
	}

	dropboxClient, err := dropbox.NewDropboxClient(accessToken, db.DB)
	if err != nil {
		db.Close() // Clean up DB connection if Dropbox client fails
		return nil, fmt.Errorf("error initializing Dropbox client: %v", err)
	}

	// Initialize DatabaseAgent
	dbAgent := agents.NewDatabaseAgent(dbConnStr)

	return &Monitor{
		DB:            db,
		DBAgent:       dbAgent,
		DropboxClient: dropboxClient,
	}, nil
}

// Close cleanly shuts down the monitor
func (m *Monitor) Close() error {
	return m.DB.Close()
}
