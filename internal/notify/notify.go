package notify

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

// Notifier defines the interface for sending notifications
type Notifier interface {
	Send(ctx context.Context, subject, message string) error
}

// EmailNotifier handles email notifications
type EmailNotifier struct {
	smtpServer   string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	toEmails     []string
	skipTLS      bool // for testing purposes
}

// NewNotifier creates a new EmailNotifier instance
func NewNotifier() *EmailNotifier {
	n := &EmailNotifier{
		smtpServer:   os.Getenv("SMTP_SERVER"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpUsername: os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		fromEmail:    os.Getenv("FROM_EMAIL"),
		skipTLS:      os.Getenv("SKIP_TLS") == "true",
	}

	// Get recipient emails
	if toEmails := os.Getenv("TO_EMAILS"); toEmails != "" {
		n.toEmails = strings.Split(toEmails, ",")
		// Trim spaces from email addresses
		for i, email := range n.toEmails {
			n.toEmails[i] = strings.TrimSpace(email)
		}
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

// Send implements the Notifier interface
func (n *EmailNotifier) Send(ctx context.Context, subject, message string) error {
	if len(n.toEmails) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	auth := smtp.PlainAuth("", n.smtpUsername, n.smtpPassword, n.smtpServer)
	addr := fmt.Sprintf("%s:%s", n.smtpServer, n.smtpPort)

	if n.skipTLS {
		return smtp.SendMail(addr, auth, n.fromEmail, n.toEmails, []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, message)))
	}

	// Create TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         n.smtpServer,
	}

	// Here is the key difference, connect to the SMTP Server with TLS
	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		log.Printf("Error connecting to SMTP server: %v", err)
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, n.smtpServer)
	if err != nil {
		log.Printf("Error creating SMTP client: %v", err)
		return err
	}
	defer c.Close()

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

	for _, addr := range n.toEmails {
		if err = c.Rcpt(addr); err != nil {
			log.Printf("Error setting recipient: %v", err)
			return err
		}
	}

	// Send the email body
	w, err := c.Data()
	if err != nil {
		log.Printf("Error creating data writer: %v", err)
		return err
	}

	msg := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, message)
	_, err = w.Write([]byte(msg))
	if err != nil {
		log.Printf("Error writing message: %v", err)
		return err
	}

	err = w.Close()
	if err != nil {
		log.Printf("Error closing data writer: %v", err)
		return err
	}

	return c.Quit()
}

// SendEmail sends an email notification
func (n *EmailNotifier) SendEmail(recipients []string, subject, body string) error {
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
