package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Database configuration
	DatabasePath string `json:"database_path"`

	// Dropbox configuration
	DropboxAccessToken string `json:"dropbox_access_token"`

	// Agent configuration
	PollInterval time.Duration `json:"poll_interval"`
	MaxRetries   int          `json:"max_retries"`
	RetryDelay   time.Duration `json:"retry_delay"`

	// Notification configuration
	NotificationEnabled bool   `json:"notification_enabled"`
	EmailSMTPHost      string `json:"email_smtp_host"`
	EmailSMTPPort      int    `json:"email_smtp_port"`
	EmailFrom          string `json:"email_from"`
	EmailTo            string `json:"email_to"`

	// Health check configuration
	HealthCheckInterval time.Duration `json:"health_check_interval"`
}

// LoadConfig loads configuration from environment variables
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

	// Validate required fields
	if cfg.DropboxAccessToken == "" {
		return nil, fmt.Errorf("DROPBOX_ACCESS_TOKEN is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d"); err == nil {
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
