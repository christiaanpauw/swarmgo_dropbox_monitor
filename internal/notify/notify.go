package notify

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

// Send sends the Dropbox report via email
func Send(report string) error {
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	toEmail := os.Getenv("NOTIFY_EMAIL")

	if smtpServer == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || toEmail == "" {
		return fmt.Errorf("SMTP configuration is missing")
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpServer)
	msg := []byte("To: " + toEmail + "\r\n" +
		"Subject: Dropbox Daily Report\r\n" +
		"\r\n" + report + "\r\n")

	addr := smtpServer + ":" + smtpPort

	err := smtp.SendMail(addr, auth, smtpUser, []string{toEmail}, msg)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return err
	}

	log.Println("📧 Report sent successfully!")
	return nil
}

