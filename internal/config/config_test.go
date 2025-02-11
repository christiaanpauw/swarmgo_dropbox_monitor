package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment
	origDBPath := os.Getenv("DROPBOX_MONITOR_DB")
	origToken := os.Getenv("DROPBOX_ACCESS_TOKEN")
	origPollInterval := os.Getenv("POLL_INTERVAL")
	origMaxRetries := os.Getenv("MAX_RETRIES")
	origRetryDelay := os.Getenv("RETRY_DELAY")

	// Restore environment after test
	defer func() {
		os.Setenv("DROPBOX_MONITOR_DB", origDBPath)
		os.Setenv("DROPBOX_ACCESS_TOKEN", origToken)
		os.Setenv("POLL_INTERVAL", origPollInterval)
		os.Setenv("MAX_RETRIES", origMaxRetries)
		os.Setenv("RETRY_DELAY", origRetryDelay)
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
		wantErr  bool
	}{
		{
			name: "all values set",
			envVars: map[string]string{
				"DROPBOX_MONITOR_DB":    "/path/to/db",
				"DROPBOX_ACCESS_TOKEN":  "test-token",
				"POLL_INTERVAL":         "5m",
				"MAX_RETRIES":          "3",
				"RETRY_DELAY":          "1s",
			},
			expected: &Config{
				DatabasePath:       "/path/to/db",
				DropboxAccessToken: "test-token",
				PollInterval:      5 * time.Minute,
				MaxRetries:        3,
				RetryDelay:        time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing token",
			envVars: map[string]string{
				"DROPBOX_MONITOR_DB": "/path/to/db",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "default values",
			envVars: map[string]string{
				"DROPBOX_ACCESS_TOKEN": "test-token",
			},
			expected: &Config{
				DatabasePath:       "data/dropbox_monitor.db",
				DropboxAccessToken: "test-token",
				PollInterval:      5 * time.Minute,
				MaxRetries:        3,
				RetryDelay:        30 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv("DROPBOX_MONITOR_DB")
			os.Unsetenv("DROPBOX_ACCESS_TOKEN")
			os.Unsetenv("POLL_INTERVAL")
			os.Unsetenv("MAX_RETRIES")
			os.Unsetenv("RETRY_DELAY")

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Load config
			cfg, err := LoadConfig()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.DatabasePath, cfg.DatabasePath)
				assert.Equal(t, tt.expected.DropboxAccessToken, cfg.DropboxAccessToken)
				assert.Equal(t, tt.expected.PollInterval, cfg.PollInterval)
				assert.Equal(t, tt.expected.MaxRetries, cfg.MaxRetries)
				assert.Equal(t, tt.expected.RetryDelay, cfg.RetryDelay)
			}
		})
	}
}
