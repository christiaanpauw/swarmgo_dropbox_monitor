package notify

import (
	"context"
	"os"
	"testing"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/config"
	"github.com/joho/godotenv"
)

func TestEmailNotifierIntegration(t *testing.T) {
	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Create email config from environment
	emailConfig := &config.EmailConfig{
		SMTPHost:     os.Getenv("SMTP_SERVER"),
		SMTPPort:     587, // Using standard STARTTLS port
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromAddress:  os.Getenv("FROM_EMAIL"),
		ToAddresses:  []string{os.Getenv("TO_EMAILS")},
	}

	// Create notifier
	notifier := NewEmailNotifier(emailConfig)

	// Test cases
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{
			name:    "Basic Email Test",
			message: "This is a test email to verify the email functionality is working correctly.",
			wantErr: false,
		},
		{
			name:    "HTML Content Test",
			message: "<h1>Test Email</h1><p>This is a test email with HTML content.</p>",
			wantErr: false,
		},
	}

	// Run test cases
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := notifier.SendNotification(ctx, tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("EmailNotifier.SendNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
