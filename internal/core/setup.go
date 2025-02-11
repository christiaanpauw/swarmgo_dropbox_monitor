package core

import (
	"fmt"
	"log"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agents"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/db"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
)

// Monitor represents the main application monitor
type Monitor struct {
	DB            *db.DB
	DBAgent       agents.DatabaseAgent
	DropboxClient *dropbox.DropboxClient
}

// NewMonitor creates a new monitor with the given database connection string and Dropbox access token
func NewMonitor(dbConnStr, dropboxToken string) (*Monitor, error) {
	if dropboxToken == "" {
		return nil, fmt.Errorf("DROPBOX_ACCESS_TOKEN not set")
	}

	// Initialize DB
	log.Println("Initializing database connection...")
	db, err := db.NewDB(dbConnStr)
	if err != nil {
		return nil, fmt.Errorf("error initializing database: %w", err)
	}

	// Initialize DatabaseAgent
	dbAgent, err := agents.NewDatabaseAgent()
	if err != nil {
		db.Close() // Clean up DB connection if agent fails
		return nil, fmt.Errorf("error creating database agent: %w", err)
	}

	// Initialize Dropbox client
	log.Println("Initializing Dropbox client...")
	dropboxClient, err := dropbox.NewDropboxClient(dropboxToken)
	if err != nil {
		db.Close() // Clean up DB connection if Dropbox client fails
		dbAgent.Close() // Clean up DB agent
		return nil, fmt.Errorf("error creating Dropbox client: %w", err)
	}

	return &Monitor{
		DB:            db,
		DBAgent:       dbAgent,
		DropboxClient: dropboxClient,
	}, nil
}

// Close cleanly shuts down the monitor
func (m *Monitor) Close() error {
	var errs []error

	if err := m.DB.Close(); err != nil {
		errs = append(errs, fmt.Errorf("error closing DB: %w", err))
	}

	if err := m.DBAgent.Close(); err != nil {
		errs = append(errs, fmt.Errorf("error closing DB agent: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}
