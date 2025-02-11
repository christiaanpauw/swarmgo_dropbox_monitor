package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Database configuration
	DatabasePath string `json:"database_path" env:"DROPBOX_MONITOR_DB" default:"data/dropbox_monitor.db"`

	// Dropbox configuration
	DropboxAccessToken string `json:"dropbox_access_token" env:"DROPBOX_ACCESS_TOKEN" required:"true"`

	// Agent configuration
	PollInterval time.Duration `json:"poll_interval" env:"POLL_INTERVAL" default:"5m"`
	MaxRetries   int          `json:"max_retries" env:"MAX_RETRIES" default:"3"`
	RetryDelay   time.Duration `json:"retry_delay" env:"RETRY_DELAY" default:"30s"`

	// Notification configuration
	NotificationEnabled bool   `json:"notification_enabled" env:"NOTIFICATION_ENABLED" default:"true"`
	EmailSMTPHost      string `json:"email_smtp_host" env:"EMAIL_SMTP_HOST" default:"localhost"`
	EmailSMTPPort      int    `json:"email_smtp_port" env:"EMAIL_SMTP_PORT" default:"25"`
	EmailFrom          string `json:"email_from" env:"EMAIL_FROM" default:"dropbox-monitor@localhost"`
	EmailTo            string `json:"email_to" env:"EMAIL_TO" default:"admin@localhost"`

	// Health check configuration
	HealthCheckInterval time.Duration `json:"health_check_interval" env:"HEALTH_CHECK_INTERVAL" default:"1m"`
}

// LoadConfig loads configuration from environment variables and validates it
func LoadConfig() (*Config, error) {
	cfg := &Config{
		// Database defaults
		DatabasePath: getEnvOrDefault("DROPBOX_MONITOR_DB", "data/dropbox_monitor.db"),

		// Dropbox configuration
		DropboxAccessToken: os.Getenv("DROPBOX_ACCESS_TOKEN"),

		// Agent configuration
		PollInterval: getDurationOrDefault("POLL_INTERVAL", 5*time.Minute),
		MaxRetries:   getIntOrDefault("MAX_RETRIES", 3),
		RetryDelay:   getDurationOrDefault("RETRY_DELAY", 30*time.Second),

		// Notification configuration
		NotificationEnabled: getBoolOrDefault("NOTIFICATION_ENABLED", true),
		EmailSMTPHost:      getEnvOrDefault("EMAIL_SMTP_HOST", "localhost"),
		EmailSMTPPort:      getIntOrDefault("EMAIL_SMTP_PORT", 25),
		EmailFrom:          getEnvOrDefault("EMAIL_FROM", "dropbox-monitor@localhost"),
		EmailTo:            getEnvOrDefault("EMAIL_TO", "admin@localhost"),

		// Health check configuration
		HealthCheckInterval: getDurationOrDefault("HEALTH_CHECK_INTERVAL", time.Minute),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate performs validation on all configuration values
func (c *Config) Validate() error {
	// Validate required fields
	if c.DropboxAccessToken == "" {
		return fmt.Errorf("DROPBOX_ACCESS_TOKEN is required")
	}

	// Validate database configuration
	if err := c.validateDatabaseConfig(); err != nil {
		return fmt.Errorf("database configuration error: %w", err)
	}

	// Validate agent configuration
	if err := c.validateAgentConfig(); err != nil {
		return fmt.Errorf("agent configuration error: %w", err)
	}

	// Validate notification configuration
	if err := c.validateNotificationConfig(); err != nil {
		return fmt.Errorf("notification configuration error: %w", err)
	}

	// Validate health check configuration
	if err := c.validateHealthCheckConfig(); err != nil {
		return fmt.Errorf("health check configuration error: %w", err)
	}

	return nil
}

// validateDatabaseConfig validates database-specific configuration
func (c *Config) validateDatabaseConfig() error {
	if c.DatabasePath == "" {
		return fmt.Errorf("database path is required")
	}

	// Ensure database directory exists
	dbDir := filepath.Dir(c.DatabasePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	return nil
}

// validateAgentConfig validates agent-specific configuration
func (c *Config) validateAgentConfig() error {
	if c.PollInterval < time.Second {
		return fmt.Errorf("poll interval must be at least 1 second")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative")
	}

	if c.RetryDelay < time.Second {
		return fmt.Errorf("retry delay must be at least 1 second")
	}

	return nil
}

// validateNotificationConfig validates notification-specific configuration
func (c *Config) validateNotificationConfig() error {
	if !c.NotificationEnabled {
		return nil
	}

	if c.EmailSMTPHost == "" {
		return fmt.Errorf("SMTP host is required when notifications are enabled")
	}

	if c.EmailSMTPPort <= 0 || c.EmailSMTPPort > 65535 {
		return fmt.Errorf("invalid SMTP port: must be between 1 and 65535")
	}

	if c.EmailFrom == "" {
		return fmt.Errorf("email from address is required when notifications are enabled")
	}

	if c.EmailTo == "" {
		return fmt.Errorf("email to address is required when notifications are enabled")
	}

	return nil
}

// validateHealthCheckConfig validates health check-specific configuration
func (c *Config) validateHealthCheckConfig() error {
	if c.HealthCheckInterval < time.Second {
		return fmt.Errorf("health check interval must be at least 1 second")
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
