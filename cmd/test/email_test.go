package test

import (
	"os"
	"testing"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/joho/godotenv"
)

func TestEmailSending(t *testing.T) {
	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Create a new notifier
	notifier := notify.NewNotifier()

	// Test cases
	tests := []struct {
		name       string
		recipients []string
		subject    string
		body       string
	}{
		{
			name:       "Basic Email Test",
			recipients: nil, // will use TO_EMAILS from .env
			subject:    "Test Email from Dropbox Monitor",
			body:       "This is a test email to verify the email functionality is working correctly.",
		},
		{
			name:       "Custom Recipients Test",
			recipients: []string{os.Getenv("TO_EMAILS")}, // explicitly set recipient
			subject:    "Test Email with Custom Recipients",
			body:       "This is a test email sent to specific recipients.",
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := notifier.SendEmail(tt.recipients, tt.subject, tt.body)
			if err != nil {
				t.Errorf("SendEmail() error = %v", err)
			}
		})
	}
}
