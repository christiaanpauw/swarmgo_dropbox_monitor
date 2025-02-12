package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration settings
type Config struct {
	DropboxToken    string        `yaml:"dropbox_token"`
	PollInterval    time.Duration `yaml:"poll_interval"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	EmailConfig     *EmailConfig  `yaml:"email_config"`
	Database        DatabaseConfig `yaml:"database"`
	Retry          RetryConfig   `yaml:"retry"`
	Notify         NotifyConfig  `yaml:"notify"`
	HealthCheck    HealthCheckConfig `yaml:"health_check"`
	State          StateConfig    `yaml:"state"`
	Web            WebConfig      `yaml:"web"`
	Monitoring     MonitoringConfig `yaml:"monitoring"`
}

// DropboxConfig holds Dropbox-specific configuration
type DropboxConfig struct {
	Token       string        `yaml:"token"`
	PollInterval time.Duration `yaml:"poll_interval"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// WebConfig holds web server configuration
type WebConfig struct {
	Address string `yaml:"address"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// StateConfig holds state management configuration
type StateConfig struct {
	Path string `yaml:"path"`
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts int           `yaml:"max_attempts"`
	Delay       time.Duration `yaml:"delay"`
}

// NotifyConfig holds notification configuration
type NotifyConfig struct {
	Enabled   bool     `yaml:"enabled"`
	SMTPHost  string   `yaml:"smtp_host"`
	SMTPPort  int      `yaml:"smtp_port"`
	FromEmail string   `yaml:"from_email"`
	ToEmails  []string `yaml:"to_emails"`
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	Interval time.Duration `yaml:"interval"`
}

// EmailConfig represents email notification configuration
type EmailConfig struct {
	SMTPHost     string   `yaml:"smtp_host"`
	SMTPPort     int      `yaml:"smtp_port"`
	SMTPUsername string   `yaml:"smtp_username"`
	SMTPPassword string   `yaml:"smtp_password"`
	FromAddress  string   `yaml:"from_address"`
	ToAddresses  []string `yaml:"to_addresses"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate Dropbox configuration
	if c.DropboxToken == "" {
		return fmt.Errorf("dropbox configuration error: access token is required")
	}
	if c.PollInterval <= 0 {
		return fmt.Errorf("dropbox configuration error: poll interval must be positive")
	}

	// Validate retry configuration
	if c.Retry.MaxAttempts <= 0 {
		return fmt.Errorf("retry configuration error: max attempts must be positive")
	}
	if c.Retry.Delay <= 0 {
		return fmt.Errorf("retry configuration error: delay must be positive")
	}

	// Validate health check configuration
	if c.HealthCheck.Interval <= 0 {
		return fmt.Errorf("health check configuration error: interval must be positive")
	}

	// Validate notification configuration
	if c.Notify.Enabled {
		if c.Notify.SMTPHost == "" {
			return fmt.Errorf("notification configuration error: SMTP host is required when notifications are enabled")
		}
		if c.Notify.SMTPPort <= 0 || c.Notify.SMTPPort > 65535 {
			return fmt.Errorf("notification configuration error: invalid SMTP port")
		}
	}

	// Validate state configuration
	if c.State.Path == "" {
		c.State.Path = filepath.Join(os.TempDir(), "dropbox_monitor_state.json")
	} else {
		// Ensure state directory exists
		stateDir := filepath.Dir(c.State.Path)
		if err := os.MkdirAll(stateDir, 0755); err != nil {
			return fmt.Errorf("failed to create state directory: %w", err)
		}
	}

	// Validate database configuration
	if c.Database.Path == "" {
		c.Database.Path = filepath.Join(os.TempDir(), "dropbox_monitor.db")
	}

	// Validate email configuration
	if c.EmailConfig != nil {
		if c.EmailConfig.SMTPHost == "" {
			return fmt.Errorf("email configuration error: SMTP host is required")
		}
		if c.EmailConfig.SMTPPort <= 0 || c.EmailConfig.SMTPPort > 65535 {
			return fmt.Errorf("email configuration error: invalid SMTP port")
		}
	}

	return nil
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetEnvOrDefault gets an environment variable value or returns a default
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetIntOrDefault gets an integer environment variable value or returns a default
func GetIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetBoolOrDefault gets a boolean environment variable value or returns a default
func GetBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetDurationOrDefault gets a duration environment variable value or returns a default
func GetDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		PollInterval: 5 * time.Minute,
		Retry: RetryConfig{
			MaxAttempts: 3,
			Delay:      time.Second * 5,
		},
		HealthCheck: HealthCheckConfig{
			Interval: time.Minute,
		},
		EmailConfig: &EmailConfig{
			SMTPPort: 587,
		},
	}
}
