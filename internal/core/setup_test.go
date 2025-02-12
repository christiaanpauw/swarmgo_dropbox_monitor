package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/db"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
)

type mockDB struct {
	*db.DB
}

func (m *mockDB) Close() error {
	return nil
}

func newMockDB() *mockDB {
	db, err := db.NewDB(":memory:")
	if err != nil {
		panic(err)
	}
	return &mockDB{DB: db}
}

type mockDropboxClient struct {
	*dropbox.DropboxClient
}

func (m *mockDropboxClient) Close() error {
	return nil
}

func newMockDropboxClient() *mockDropboxClient {
	return &mockDropboxClient{
		DropboxClient: &dropbox.DropboxClient{},
	}
}

func TestNewMonitor(t *testing.T) {
	tests := []struct {
		name         string
		dbConnStr    string
		dropboxToken string
		wantErr      bool
	}{
		{
			name:         "valid configuration",
			dbConnStr:    "test.db",
			dropboxToken: "test-token",
			wantErr:      false,
		},
		{
			name:         "missing dropbox token",
			dbConnStr:    "test.db",
			dropboxToken: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor, err := NewMonitor(tt.dbConnStr, tt.dropboxToken)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, monitor)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, monitor)
				assert.NotNil(t, monitor.DB)
				assert.NotNil(t, monitor.DropboxClient)
			}
		})
	}
}

func TestMonitor_Close(t *testing.T) {
	mockDb := newMockDB()
	mockClient := newMockDropboxClient()

	monitor := &Monitor{
		DB:            mockDb.DB,
		DropboxClient: mockClient.DropboxClient,
	}

	err := monitor.Close()
	require.NoError(t, err)
}
