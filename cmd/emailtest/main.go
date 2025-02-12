package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/config"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/container"
	"github.com/joho/godotenv"
)

func main() {
	var (
		envFile     = flag.String("env", ".env", "Path to .env file")
		message     = flag.String("message", "Test email from Dropbox Monitor", "Email message to send")
		useEnvFile  = flag.Bool("use-env", true, "Use environment file for configuration")
		smtpHost    = flag.String("smtp-host", "", "SMTP server host (overrides env)")
		smtpPort    = flag.Int("smtp-port", 587, "SMTP server port (overrides env)")
		smtpUser    = flag.String("smtp-user", "", "SMTP username (overrides env)")
		smtpPass    = flag.String("smtp-pass", "", "SMTP password (overrides env)")
		fromEmail   = flag.String("from", "", "From email address (overrides env)")
		toEmails    = flag.String("to", "", "Comma-separated list of recipient email addresses (overrides env)")
	)

	flag.Parse()

	// Load environment variables if requested
	if *useEnvFile {
		if err := godotenv.Load(*envFile); err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	// Create email config, preferring command line args over env vars
	emailConfig := &config.EmailConfig{
		SMTPHost:     getConfigValue(*smtpHost, "SMTP_SERVER"),
		SMTPPort:     *smtpPort,
		SMTPUsername: getConfigValue(*smtpUser, "SMTP_USERNAME"),
		SMTPPassword: getConfigValue(*smtpPass, "SMTP_PASSWORD"),
		FromAddress:  getConfigValue(*fromEmail, "FROM_EMAIL"),
		ToAddresses:  getToAddresses(*toEmails),
	}

	// Validate configuration
	if err := validateConfig(emailConfig); err != nil {
		fmt.Fprintln(os.Stderr, "Configuration error:", err)
		flag.Usage()
		os.Exit(1)
	}

	// Create container with email configuration
	cfg := &config.Config{
		EmailConfig: emailConfig,
	}
	c, err := container.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	// Get notifier from container
	notifier := c.GetNotifier()

	// Send test email
	ctx := context.Background()
	if err := notifier.SendNotification(ctx, *message); err != nil {
		log.Fatalf("Failed to send test email: %v", err)
	}

	fmt.Println("Test email sent successfully!")
}

func getConfigValue(flagValue, envKey string) string {
	if flagValue != "" {
		return flagValue
	}
	return os.Getenv(envKey)
}

func getToAddresses(flagValue string) []string {
	if flagValue != "" {
		return []string{flagValue}
	}
	return []string{os.Getenv("TO_EMAILS")}
}

func validateConfig(cfg *config.EmailConfig) error {
	if cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if cfg.SMTPUsername == "" {
		return fmt.Errorf("SMTP username is required")
	}
	if cfg.SMTPPassword == "" {
		return fmt.Errorf("SMTP password is required")
	}
	if cfg.FromAddress == "" {
		return fmt.Errorf("From email address is required")
	}
	if len(cfg.ToAddresses) == 0 || cfg.ToAddresses[0] == "" {
		return fmt.Errorf("At least one recipient email address is required")
	}
	return nil
}
