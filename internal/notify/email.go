package notify

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/config"
)

// EmailNotifier implements the Notifier interface for email notifications
type EmailNotifier struct {
	config *config.EmailConfig
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier(cfg *config.EmailConfig) Notifier {
	return &EmailNotifier{
		config: cfg,
	}
}

// SendNotification sends an email notification
func (n *EmailNotifier) SendNotification(ctx context.Context, message string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	if n.config == nil {
		return fmt.Errorf("email config is nil")
	}

	// Validate required fields
	if n.config.SMTPHost == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if len(n.config.ToAddresses) == 0 {
		return fmt.Errorf("at least one recipient email address is required")
	}
	if n.config.FromAddress == "" {
		return fmt.Errorf("from email address is required")
	}

	auth := smtp.PlainAuth("", n.config.SMTPUsername, n.config.SMTPPassword, n.config.SMTPHost)

	// Compose email
	from := n.config.FromAddress
	to := n.config.ToAddresses
	subject := "Dropbox Monitor Notification"
	body := message

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", from, strings.Join(to, ", "), subject, body)

	// Send email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", n.config.SMTPHost, n.config.SMTPPort),
		auth,
		from,
		to,
		[]byte(msg),
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
