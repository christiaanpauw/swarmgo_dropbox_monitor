package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create temporary directory for database
	tmpDir, err := os.MkdirTemp("", "dropbox_monitor_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Save original environment
	origEnv := map[string]string{
		"DROPBOX_MONITOR_DB":     os.Getenv("DROPBOX_MONITOR_DB"),
		"DROPBOX_ACCESS_TOKEN":   os.Getenv("DROPBOX_ACCESS_TOKEN"),
		"POLL_INTERVAL":          os.Getenv("POLL_INTERVAL"),
		"MAX_RETRIES":           os.Getenv("MAX_RETRIES"),
		"RETRY_DELAY":           os.Getenv("RETRY_DELAY"),
		"NOTIFICATION_ENABLED":   os.Getenv("NOTIFICATION_ENABLED"),
		"EMAIL_SMTP_HOST":       os.Getenv("EMAIL_SMTP_HOST"),
		"EMAIL_SMTP_PORT":       os.Getenv("EMAIL_SMTP_PORT"),
		"EMAIL_FROM":            os.Getenv("EMAIL_FROM"),
		"EMAIL_TO":              os.Getenv("EMAIL_TO"),
		"HEALTH_CHECK_INTERVAL": os.Getenv("HEALTH_CHECK_INTERVAL"),
	}

	// Restore environment after test
	defer func() {
		for k, v := range origEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
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
				"DROPBOX_MONITOR_DB":     filepath.Join(tmpDir, "test.db"),
				"DROPBOX_ACCESS_TOKEN":   "test-token",
				"POLL_INTERVAL":         "1m",
				"MAX_RETRIES":          "5",
				"RETRY_DELAY":          "10s",
				"NOTIFICATION_ENABLED":  "true",
				"EMAIL_SMTP_HOST":      "smtp.test.com",
				"EMAIL_SMTP_PORT":      "587",
				"EMAIL_FROM":           "test@test.com",
				"EMAIL_TO":             "admin@test.com",
				"HEALTH_CHECK_INTERVAL": "30s",
			},
			expected: &Config{
				DatabasePath:        filepath.Join(tmpDir, "test.db"),
				DropboxAccessToken:  "test-token",
				PollInterval:       time.Minute,
				MaxRetries:         5,
				RetryDelay:         10 * time.Second,
				NotificationEnabled: true,
				EmailSMTPHost:      "smtp.test.com",
				EmailSMTPPort:      587,
				EmailFrom:          "test@test.com",
				EmailTo:            "admin@test.com",
				HealthCheckInterval: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "default values",
			envVars: map[string]string{
				"DROPBOX_MONITOR_DB":   filepath.Join(tmpDir, "test.db"),
				"DROPBOX_ACCESS_TOKEN": "test-token",
			},
			expected: &Config{
				DatabasePath:        filepath.Join(tmpDir, "test.db"),
				DropboxAccessToken:  "test-token",
				PollInterval:       5 * time.Minute,
				MaxRetries:         3,
				RetryDelay:         30 * time.Second,
				NotificationEnabled: true,
				EmailSMTPHost:      "localhost",
				EmailSMTPPort:      25,
				EmailFrom:          "dropbox-monitor@localhost",
				EmailTo:            "admin@localhost",
				HealthCheckInterval: time.Minute,
			},
			wantErr: false,
		},
		{
			name: "missing dropbox token",
			envVars: map[string]string{
				"DROPBOX_MONITOR_DB": filepath.Join(tmpDir, "test.db"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables
			for k := range origEnv {
				os.Unsetenv(k)
			}

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Load config
			cfg, err := LoadConfig()

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, cfg)

			// Verify loaded values
			assert.Equal(t, tt.expected.DatabasePath, cfg.DatabasePath)
			assert.Equal(t, tt.expected.DropboxAccessToken, cfg.DropboxAccessToken)
			assert.Equal(t, tt.expected.PollInterval, cfg.PollInterval)
			assert.Equal(t, tt.expected.MaxRetries, cfg.MaxRetries)
			assert.Equal(t, tt.expected.RetryDelay, cfg.RetryDelay)
			assert.Equal(t, tt.expected.NotificationEnabled, cfg.NotificationEnabled)
			assert.Equal(t, tt.expected.EmailSMTPHost, cfg.EmailSMTPHost)
			assert.Equal(t, tt.expected.EmailSMTPPort, cfg.EmailSMTPPort)
			assert.Equal(t, tt.expected.EmailFrom, cfg.EmailFrom)
			assert.Equal(t, tt.expected.EmailTo, cfg.EmailTo)
			assert.Equal(t, tt.expected.HealthCheckInterval, cfg.HealthCheckInterval)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	// Create temporary directory for database
	tmpDir, err := os.MkdirTemp("", "dropbox_monitor_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				DatabasePath:        filepath.Join(tmpDir, "test.db"),
				DropboxAccessToken:  "test-token",
				PollInterval:       time.Minute,
				MaxRetries:         3,
				RetryDelay:         time.Second,
				NotificationEnabled: true,
				EmailSMTPHost:      "smtp.test.com",
				EmailSMTPPort:      587,
				EmailFrom:          "test@test.com",
				EmailTo:            "admin@test.com",
				HealthCheckInterval: time.Minute,
			},
			wantErr: false,
		},
		{
			name: "missing dropbox token",
			cfg: Config{
				DatabasePath:        filepath.Join(tmpDir, "test.db"),
				PollInterval:       time.Minute,
				MaxRetries:         3,
				RetryDelay:         time.Second,
				NotificationEnabled: false,
				HealthCheckInterval: time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid poll interval",
			cfg: Config{
				DatabasePath:        filepath.Join(tmpDir, "test.db"),
				DropboxAccessToken:  "test-token",
				PollInterval:       time.Millisecond,
				MaxRetries:         3,
				RetryDelay:         time.Second,
				NotificationEnabled: false,
				HealthCheckInterval: time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid retry delay",
			cfg: Config{
				DatabasePath:        filepath.Join(tmpDir, "test.db"),
				DropboxAccessToken:  "test-token",
				PollInterval:       time.Minute,
				MaxRetries:         3,
				RetryDelay:         time.Millisecond,
				NotificationEnabled: false,
				HealthCheckInterval: time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid health check interval",
			cfg: Config{
				DatabasePath:        filepath.Join(tmpDir, "test.db"),
				DropboxAccessToken:  "test-token",
				PollInterval:       time.Minute,
				MaxRetries:         3,
				RetryDelay:         time.Second,
				NotificationEnabled: false,
				HealthCheckInterval: time.Millisecond,
			},
			wantErr: true,
		},
		{
			name: "notifications enabled but missing smtp host",
			cfg: Config{
				DatabasePath:        filepath.Join(tmpDir, "test.db"),
				DropboxAccessToken:  "test-token",
				PollInterval:       time.Minute,
				MaxRetries:         3,
				RetryDelay:         time.Second,
				NotificationEnabled: true,
				EmailSMTPPort:      587,
				EmailFrom:          "test@test.com",
				EmailTo:            "admin@test.com",
				HealthCheckInterval: time.Minute,
			},
			wantErr: true,
		},
		{
			name: "notifications enabled but invalid smtp port",
			cfg: Config{
				DatabasePath:        filepath.Join(tmpDir, "test.db"),
				DropboxAccessToken:  "test-token",
				PollInterval:       time.Minute,
				MaxRetries:         3,
				RetryDelay:         time.Second,
				NotificationEnabled: true,
				EmailSMTPHost:      "smtp.test.com",
				EmailSMTPPort:      0,
				EmailFrom:          "test@test.com",
				EmailTo:            "admin@test.com",
				HealthCheckInterval: time.Minute,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	// Save original environment
	origValue := os.Getenv("TEST_KEY")
	defer os.Setenv("TEST_KEY", origValue)

	// Test with environment variable set
	os.Setenv("TEST_KEY", "test-value")
	assert.Equal(t, "test-value", getEnvOrDefault("TEST_KEY", "default"))

	// Test with environment variable not set
	os.Unsetenv("TEST_KEY")
	assert.Equal(t, "default", getEnvOrDefault("TEST_KEY", "default"))
}

func TestGetIntOrDefault(t *testing.T) {
	// Save original environment
	origValue := os.Getenv("TEST_INT")
	defer os.Setenv("TEST_INT", origValue)

	// Test with valid integer
	os.Setenv("TEST_INT", "42")
	assert.Equal(t, 42, getIntOrDefault("TEST_INT", 0))

	// Test with invalid integer
	os.Setenv("TEST_INT", "invalid")
	assert.Equal(t, 0, getIntOrDefault("TEST_INT", 0))

	// Test with environment variable not set
	os.Unsetenv("TEST_INT")
	assert.Equal(t, 0, getIntOrDefault("TEST_INT", 0))
}

func TestGetBoolOrDefault(t *testing.T) {
	// Save original environment
	origValue := os.Getenv("TEST_BOOL")
	defer os.Setenv("TEST_BOOL", origValue)

	// Test with "true"
	os.Setenv("TEST_BOOL", "true")
	assert.True(t, getBoolOrDefault("TEST_BOOL", false))

	// Test with "1"
	os.Setenv("TEST_BOOL", "1")
	assert.True(t, getBoolOrDefault("TEST_BOOL", false))

	// Test with "yes"
	os.Setenv("TEST_BOOL", "yes")
	assert.True(t, getBoolOrDefault("TEST_BOOL", false))

	// Test with "false"
	os.Setenv("TEST_BOOL", "false")
	assert.False(t, getBoolOrDefault("TEST_BOOL", true))

	// Test with environment variable not set
	os.Unsetenv("TEST_BOOL")
	assert.True(t, getBoolOrDefault("TEST_BOOL", true))
}

func TestGetDurationOrDefault(t *testing.T) {
	// Save original environment
	origValue := os.Getenv("TEST_DURATION")
	defer os.Setenv("TEST_DURATION", origValue)

	// Test with valid duration
	os.Setenv("TEST_DURATION", "1h")
	assert.Equal(t, time.Hour, getDurationOrDefault("TEST_DURATION", time.Minute))

	// Test with invalid duration
	os.Setenv("TEST_DURATION", "invalid")
	assert.Equal(t, time.Minute, getDurationOrDefault("TEST_DURATION", time.Minute))

	// Test with environment variable not set
	os.Unsetenv("TEST_DURATION")
	assert.Equal(t, time.Minute, getDurationOrDefault("TEST_DURATION", time.Minute))
}
