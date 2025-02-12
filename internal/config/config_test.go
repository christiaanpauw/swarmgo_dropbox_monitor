package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				DropboxToken: "test-token",
				PollInterval: time.Second * 30,
				ShutdownTimeout: time.Second * 10,
				EmailConfig: &EmailConfig{
					SMTPHost:     "smtp.test.com",
					SMTPPort:     587,
					SMTPUsername: "test@test.com",
					SMTPPassword: "password",
					FromAddress:  "from@test.com",
					ToAddresses:  []string{"to@test.com"},
				},
				Database: DatabaseConfig{
					Path: "test.db",
				},
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      time.Second * 5,
				},
				Notify: NotifyConfig{
					Enabled:   true,
					SMTPHost:  "smtp.test.com",
					SMTPPort:  587,
					FromEmail: "from@test.com",
					ToEmails:  []string{"to@test.com"},
				},
				HealthCheck: HealthCheckConfig{
					Interval: time.Minute,
				},
				State: StateConfig{
					Path: "state.db",
				},
				Web: WebConfig{
					Address: ":8080",
				},
				Monitoring: MonitoringConfig{
					Enabled: true,
				},
			},
			wantErr: false,
		},
		{
			name: "missing dropbox token",
			config: Config{
				DropboxToken: "",
				PollInterval: time.Second * 30,
				Database: DatabaseConfig{
					Path: "test.db",
				},
				Web: WebConfig{
					Address: ":8080",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid poll interval",
			config: Config{
				DropboxToken: "test-token",
				PollInterval: 0,
				Database: DatabaseConfig{
					Path: "test.db",
				},
				Web: WebConfig{
					Address: ":8080",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				DropboxToken:    "test-token",
				PollInterval:    time.Minute,
				ShutdownTimeout: time.Second * 10,
				EmailConfig: &EmailConfig{
					SMTPHost:     "smtp.test.com",
					SMTPPort:     587,
					SMTPUsername: "test@test.com",
					SMTPPassword: "password",
					FromAddress:  "from@test.com",
					ToAddresses:  []string{"to@test.com"},
				},
				Database: DatabaseConfig{
					Path: filepath.Join(tmpDir, "test.db"),
				},
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      time.Second,
				},
				HealthCheck: HealthCheckConfig{
					Interval: time.Minute,
				},
				Notify: NotifyConfig{
					Enabled:   true,
					SMTPHost:  "smtp.test.com",
					SMTPPort:  587,
					FromEmail: "from@test.com",
					ToEmails:  []string{"to@test.com"},
				},
				State: StateConfig{
					Path: "state.db",
				},
				Web: WebConfig{
					Address: ":8080",
				},
				Monitoring: MonitoringConfig{
					Enabled: true,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid poll interval",
			cfg: Config{
				DropboxToken: "test-token",
				PollInterval: time.Millisecond,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      time.Second,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid retry delay",
			cfg: Config{
				DropboxToken: "test-token",
				PollInterval: time.Minute,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      time.Millisecond,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid health check interval",
			cfg: Config{
				DropboxToken: "test-token",
				PollInterval: time.Minute,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      time.Second,
				},
				HealthCheck: HealthCheckConfig{
					Interval: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "notifications enabled but missing smtp host",
			cfg: Config{
				DropboxToken: "test-token",
				PollInterval: time.Minute,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      time.Second,
				},
				Notify: NotifyConfig{
					Enabled: true,
				},
			},
			wantErr: true,
		},
		{
			name: "notifications enabled but invalid smtp port",
			cfg: Config{
				DropboxToken: "test-token",
				PollInterval: time.Minute,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      time.Second,
				},
				Notify: NotifyConfig{
					Enabled:  true,
					SMTPHost: "smtp.test.com",
					SMTPPort: -1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
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
	assert.Equal(t, "test-value", GetEnvOrDefault("TEST_KEY", "default"))
	assert.Equal(t, "default", GetEnvOrDefault("NON_EXISTENT_KEY", "default"))
}

func TestGetIntOrDefault(t *testing.T) {
	// Save original environment
	origValue := os.Getenv("TEST_INT")
	defer os.Setenv("TEST_INT", origValue)

	// Test with valid integer
	os.Setenv("TEST_INT", "42")
	assert.Equal(t, 42, GetIntOrDefault("TEST_INT", 0))

	// Test with invalid integer
	os.Setenv("TEST_INT", "invalid")
	assert.Equal(t, 0, GetIntOrDefault("TEST_INT", 0))

	// Test with environment variable not set
	os.Unsetenv("TEST_INT")
	assert.Equal(t, 0, GetIntOrDefault("TEST_INT", 0))
}

func TestGetBoolOrDefault(t *testing.T) {
	// Save original environment
	origValue := os.Getenv("TEST_BOOL")
	defer os.Setenv("TEST_BOOL", origValue)

	// Test with "true"
	os.Setenv("TEST_BOOL", "true")
	assert.True(t, GetBoolOrDefault("TEST_BOOL", false))

	// Test with "1"
	os.Setenv("TEST_BOOL", "1")
	assert.True(t, GetBoolOrDefault("TEST_BOOL", false))

	// Test with "false"
	os.Setenv("TEST_BOOL", "false")
	assert.False(t, GetBoolOrDefault("TEST_BOOL", true))

	// Test with environment variable not set
	os.Unsetenv("TEST_BOOL")
	assert.True(t, GetBoolOrDefault("TEST_BOOL", true))
}

func TestGetDurationOrDefault(t *testing.T) {
	// Save original environment
	origValue := os.Getenv("TEST_DURATION")
	defer os.Setenv("TEST_DURATION", origValue)

	// Test with valid duration
	os.Setenv("TEST_DURATION", "1h")
	assert.Equal(t, time.Hour, GetDurationOrDefault("TEST_DURATION", time.Minute))

	// Test with invalid duration
	os.Setenv("TEST_DURATION", "invalid")
	assert.Equal(t, time.Minute, GetDurationOrDefault("TEST_DURATION", time.Minute))

	// Test with environment variable not set
	os.Unsetenv("TEST_DURATION")
	assert.Equal(t, time.Minute, GetDurationOrDefault("TEST_DURATION", time.Minute))
}

func TestLoadConfigNew(t *testing.T) {
	testCases := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "all values set",
			config: Config{
				DropboxToken: "test-token",
				PollInterval: 5 * time.Minute,
				Database: DatabaseConfig{
					Path: "test.db",
				},
				Web: WebConfig{
					Address: ":8080",
				},
				Monitoring: MonitoringConfig{
					Enabled: true,
				},
				State: StateConfig{
					Path: "state.json",
				},
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      30 * time.Second,
				},
				Notify: NotifyConfig{
					Enabled:   true,
					SMTPHost: "smtp.test.com",
					SMTPPort: 587,
					FromEmail: "test@test.com",
					ToEmails: []string{"admin@test.com"},
				},
				HealthCheck: HealthCheckConfig{
					Interval: time.Minute,
				},
			},
			wantErr: false,
		},
		{
			name: "default values",
			config: Config{
				DropboxToken: "test-token",
				PollInterval: 5 * time.Minute,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      30 * time.Second,
				},
				HealthCheck: HealthCheckConfig{
					Interval: time.Minute,
				},
			},
			wantErr: false,
		},
		{
			name: "missing dropbox token",
			config: Config{
				PollInterval: 5 * time.Minute,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      30 * time.Second,
				},
				HealthCheck: HealthCheckConfig{
					Interval: time.Minute,
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_ValidateNew(t *testing.T) {
	testCases := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				DropboxToken: "test-token",
				PollInterval: 5 * time.Minute,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      30 * time.Second,
				},
				HealthCheck: HealthCheckConfig{
					Interval: time.Minute,
				},
			},
			wantErr: false,
		},
		{
			name: "missing dropbox token",
			config: Config{
				PollInterval: 5 * time.Minute,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      30 * time.Second,
				},
				HealthCheck: HealthCheckConfig{
					Interval: time.Minute,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid poll interval",
			config: Config{
				DropboxToken: "test-token",
				PollInterval: -1,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      30 * time.Second,
				},
				HealthCheck: HealthCheckConfig{
					Interval: time.Minute,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid retry delay",
			config: Config{
				DropboxToken: "test-token",
				PollInterval: 5 * time.Minute,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      -1,
				},
				HealthCheck: HealthCheckConfig{
					Interval: time.Minute,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid health check interval",
			config: Config{
				DropboxToken: "test-token",
				PollInterval: 5 * time.Minute,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      30 * time.Second,
				},
				HealthCheck: HealthCheckConfig{
					Interval: -1,
				},
			},
			wantErr: true,
		},
		{
			name: "notifications enabled but missing smtp host",
			config: Config{
				DropboxToken: "test-token",
				PollInterval: 5 * time.Minute,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      30 * time.Second,
				},
				HealthCheck: HealthCheckConfig{
					Interval: time.Minute,
				},
				Notify: NotifyConfig{
					Enabled: true,
				},
			},
			wantErr: true,
		},
		{
			name: "notifications enabled but invalid smtp port",
			config: Config{
				DropboxToken: "test-token",
				PollInterval: 5 * time.Minute,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Delay:      30 * time.Second,
				},
				HealthCheck: HealthCheckConfig{
					Interval: time.Minute,
				},
				Notify: NotifyConfig{
					Enabled:  true,
					SMTPHost: "smtp.test.com",
					SMTPPort: -1,
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetEnvOrDefaultNew(t *testing.T) {
	t.Setenv("TEST_KEY", "test-value")
	assert.Equal(t, "test-value", GetEnvOrDefault("TEST_KEY", "default"))
	assert.Equal(t, "default", GetEnvOrDefault("NON_EXISTENT_KEY", "default"))
}

func TestGetIntOrDefaultNew(t *testing.T) {
	t.Setenv("TEST_INT", "123")
	assert.Equal(t, 123, GetIntOrDefault("TEST_INT", 456))
	assert.Equal(t, 456, GetIntOrDefault("NON_EXISTENT_INT", 456))
}

func TestGetBoolOrDefaultNew(t *testing.T) {
	t.Setenv("TEST_BOOL", "true")
	assert.True(t, GetBoolOrDefault("TEST_BOOL", false))
	assert.False(t, GetBoolOrDefault("NON_EXISTENT_BOOL", false))
}

func TestGetDurationOrDefaultNew(t *testing.T) {
	t.Setenv("TEST_DURATION", "1h")
	assert.Equal(t, time.Hour, GetDurationOrDefault("TEST_DURATION", time.Minute))
	assert.Equal(t, time.Minute, GetDurationOrDefault("NON_EXISTENT_DURATION", time.Minute))
}
