package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/notify"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or error loading it: %v", err)
	}

	// Command line flags
	smtpServer := flag.String("smtp-server", os.Getenv("SMTP_SERVER"), "SMTP server address")
	smtpPort := flag.String("smtp-port", os.Getenv("SMTP_PORT"), "SMTP server port")
	username := flag.String("username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	password := flag.String("password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	fromEmail := flag.String("from", os.Getenv("FROM_EMAIL"), "From email address")
	toEmails := flag.String("to", os.Getenv("TO_EMAILS"), "Comma-separated list of recipient email addresses")
	subject := flag.String("subject", "Test Email", "Email subject")
	body := flag.String("body", "This is a test email from the Dropbox Monitor application.", "Email body")

	flag.Parse()

	// Validate required fields
	if *smtpServer == "" || *smtpPort == "" || *username == "" || *password == "" || *fromEmail == "" || *toEmails == "" {
		log.Fatal("All SMTP settings are required. Use -h for help.")
	}

	// Override environment variables with command line flags
	os.Setenv("SMTP_SERVER", *smtpServer)
	os.Setenv("SMTP_PORT", *smtpPort)
	os.Setenv("SMTP_USERNAME", *username)
	os.Setenv("SMTP_PASSWORD", *password)
	os.Setenv("FROM_EMAIL", *fromEmail)
	os.Setenv("TO_EMAILS", *toEmails)

	// Create notifier
	notifier := notify.NewNotifier()

	// Send test email
	recipients := strings.Split(*toEmails, ",")
	fmt.Printf("Sending test email to: %v\n", recipients)
	fmt.Printf("Using SMTP server: %s:%s\n", *smtpServer, *smtpPort)
	fmt.Printf("From: %s\n", *fromEmail)

	if err := notifier.SendEmail(recipients, *subject, *body); err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	fmt.Println("Email sent successfully!")
}
