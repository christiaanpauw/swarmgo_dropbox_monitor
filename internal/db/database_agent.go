package db

import (
	"context"
	"fmt"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/agent"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// DatabaseAgentImpl implements the DatabaseAgent interface
type DatabaseAgentImpl struct {
	*lifecycle.BaseComponent
	db *DB
}

// NewDatabaseAgent creates a new database agent
func NewDatabaseAgent(db *DB) (agent.DatabaseAgent, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	agent := &DatabaseAgentImpl{
		BaseComponent: lifecycle.NewBaseComponent("DatabaseAgent"),
		db:           db,
	}
	agent.SetState(lifecycle.StateInitialized)
	return agent, nil
}

// Start implements lifecycle.Component
func (a *DatabaseAgentImpl) Start(ctx context.Context) error {
	return a.DefaultStart(ctx)
}

// Stop implements lifecycle.Component
func (a *DatabaseAgentImpl) Stop(ctx context.Context) error {
	return a.DefaultStop(ctx)
}

// Health implements lifecycle.Component
func (a *DatabaseAgentImpl) Health(ctx context.Context) error {
	return a.DefaultHealth(ctx)
}

// StoreChange stores a file change in the database
func (a *DatabaseAgentImpl) StoreChange(ctx context.Context, change models.FileMetadata) error {
	// TODO: Implement database storage
	return nil
}

// GetLatestChanges retrieves the latest changes from the database
func (a *DatabaseAgentImpl) GetLatestChanges(ctx context.Context, limit int) ([]models.FileMetadata, error) {
	// TODO: Implement database retrieval
	return nil, nil
}

// GetChanges retrieves changes within a time range
func (a *DatabaseAgentImpl) GetChanges(ctx context.Context, startTime, endTime string) ([]models.FileMetadata, error) {
	// TODO: Implement database retrieval
	return nil, nil
}
