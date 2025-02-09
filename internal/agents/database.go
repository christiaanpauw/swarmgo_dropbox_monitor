package agents

import (
	"context"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/db"
)

// DatabaseAgent handles all database operations
type DatabaseAgent struct {
	db *db.DB
}

// NewDatabaseAgent creates a new database agent
func NewDatabaseAgent(db *db.DB) *DatabaseAgent {
	return &DatabaseAgent{
		db: db,
	}
}

// StoreFileChange stores a file change in the database
func (d *DatabaseAgent) StoreFileChange(filePath string, modifiedAt time.Time, metadata map[string]interface{}) error {
	fileChange := &db.FileChange{
		FilePath:   filePath,
		ModifiedAt: modifiedAt,
	}

	if metadata != nil {
		if portfolio, ok := metadata["portfolio"].(string); ok {
			fileChange.Portfolio = portfolio
		}
		if project, ok := metadata["project"].(string); ok {
			fileChange.Project = project
		}
		if docType, ok := metadata["documentType"].(string); ok {
			fileChange.DocumentType = docType
		}
		if author, ok := metadata["author"].(string); ok {
			fileChange.Author = author
		}
	}

	ctx := context.Background()
	return d.db.SaveFileChange(ctx, fileChange)
}
