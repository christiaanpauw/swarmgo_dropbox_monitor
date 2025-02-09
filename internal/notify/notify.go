package notify

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

// Notifier handles email notifications
type Notifier struct {
	smtpServer   string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	toEmails     []string
}

// NewNotifier creates a new Notifier instance
func NewNotifier() *Notifier {
	// Try to load .env from different locations
	envPaths := []string{
		".env",
		"../.env",
		"../../.env",
		"/Users/christiaanpauw/go/src/github.com/christiaanpauw/swarmgo_dropbox_monitor/.env",
	}

	var loaded bool
	for _, path := range envPaths {
		if err := loadEnvFile(path); err == nil {
			loaded = true
			break
		}
	}

	if !loaded {
		log.Println("Warning: Could not load .env file")
	}

	// Get email configuration from environment
	n := &Notifier{
		smtpServer:   os.Getenv("SMTP_SERVER"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpUsername: os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		fromEmail:    os.Getenv("FROM_EMAIL"),
	}

	// Get recipient emails
	if toEmails := os.Getenv("TO_EMAILS"); toEmails != "" {
		n.toEmails = strings.Split(toEmails, ",")
	}

	return n
}

// loadEnvFile reads and parses the .env file
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// Remove quotes if present
		value = strings.Trim(value, `"'`)
		os.Setenv(key, value)
	}
	return scanner.Err()
}

// SendEmail sends an email notification
func (n *Notifier) SendEmail(recipients []string, subject, body string) error {
	if len(recipients) == 0 {
		if len(n.toEmails) == 0 {
			return fmt.Errorf("no recipients specified")
		}
		recipients = n.toEmails
	}

	// Set up authentication information
	auth := smtp.PlainAuth("", n.smtpUsername, n.smtpPassword, n.smtpServer)

	// Format the email
	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", n.fromEmail, strings.Join(recipients, ", "), subject, body)

	// Connect to the server
	smtpAddr := fmt.Sprintf("%s:%s", n.smtpServer, n.smtpPort)
	c, err := smtp.Dial(smtpAddr)
	if err != nil {
		log.Printf("Error connecting to SMTP server: %v", err)
		return err
	}
	defer c.Close()

	// Start TLS
	if err = c.StartTLS(&tls.Config{ServerName: n.smtpServer}); err != nil {
		log.Printf("Error starting TLS: %v", err)
		return err
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		log.Printf("Error authenticating: %v", err)
		return err
	}

	// Set the sender and recipient
	if err = c.Mail(n.fromEmail); err != nil {
		log.Printf("Error setting sender: %v", err)
		return err
	}

	for _, addr := range recipients {
		if err = c.Rcpt(addr); err != nil {
			log.Printf("Error setting recipient %s: %v", addr, err)
			return err
		}
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		log.Printf("Error getting data writer: %v", err)
		return err
	}
	defer wc.Close()

	_, err = fmt.Fprintf(wc, msg)
	if err != nil {
		log.Printf("Error writing message: %v", err)
		return err
	}

	return nil
}
