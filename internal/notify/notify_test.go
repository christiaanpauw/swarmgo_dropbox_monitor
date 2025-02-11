package notify

import (
	"context"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewNotifier(t *testing.T) {
	// Save original environment variables
	originals := map[string]string{
		"SMTP_SERVER":   os.Getenv("SMTP_SERVER"),
		"SMTP_PORT":     os.Getenv("SMTP_PORT"),
		"SMTP_USERNAME": os.Getenv("SMTP_USERNAME"),
		"SMTP_PASSWORD": os.Getenv("SMTP_PASSWORD"),
		"FROM_EMAIL":    os.Getenv("FROM_EMAIL"),
		"TO_EMAILS":     os.Getenv("TO_EMAILS"),
	}

	// Create a temporary test environment file
	tmpEnv := `
SMTP_SERVER=smtp.test.com
SMTP_PORT=587
SMTP_USERNAME=test@test.com
SMTP_PASSWORD=password
FROM_EMAIL=from@test.com
TO_EMAILS=to1@test.com,to2@test.com
`
	tmpfile, err := os.CreateTemp("", "test.env")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(tmpEnv)); err != nil {
		t.Fatalf("Could not write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Could not close temp file: %v", err)
	}

	// Restore original environment variables after test
	defer func() {
		for k, v := range originals {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		wantPort string
		wantEmails int
	}{
		{
			name: "Valid configuration",
			envVars: map[string]string{
				"SMTP_SERVER":   "smtp.test.com",
				"SMTP_PORT":     "587",
				"SMTP_USERNAME": "test@test.com",
				"SMTP_PASSWORD": "password",
				"FROM_EMAIL":    "from@test.com",
				"TO_EMAILS":     "to1@test.com,to2@test.com",
			},
			wantPort: "587",
			wantEmails: 2,
		},
		{
			name:     "Empty configuration",
			envVars:  map[string]string{},
			wantPort: "",
			wantEmails: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Clear all environment variables first
			for k := range originals {
				os.Unsetenv(k)
			}
			
			// Set up environment for test
			for k, v := range tc.envVars {
				os.Setenv(k, v)
			}

			notifier := NewNotifier()

			// Check SMTP port
			if notifier.smtpPort != tc.wantPort {
				t.Errorf("NewNotifier() smtpPort = %v, want %v", notifier.smtpPort, tc.wantPort)
			}

			// Check email recipients
			if len(notifier.toEmails) != tc.wantEmails {
				t.Errorf("NewNotifier() toEmails count = %v, want %v", len(notifier.toEmails), tc.wantEmails)
			}
		})
	}
}

func TestLoadEnvFile(t *testing.T) {
	// Create a temporary .env file
	content := `
SMTP_SERVER=smtp.test.com
SMTP_PORT=587
SMTP_USERNAME=test@test.com
SMTP_PASSWORD="secret password"
FROM_EMAIL='from@test.com'
TO_EMAILS=to1@test.com,to2@test.com
# Comment line
INVALID_LINE
`
	tmpfile, err := os.CreateTemp("", "test.env")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Could not write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Could not close temp file: %v", err)
	}

	// Test loading the file
	err = loadEnvFile(tmpfile.Name())
	if err != nil {
		t.Errorf("loadEnvFile() error = %v", err)
	}

	// Check if variables were loaded correctly
	tests := []struct {
		name     string
		envVar   string
		expected string
	}{
		{"SMTP Server", "SMTP_SERVER", "smtp.test.com"},
		{"SMTP Port", "SMTP_PORT", "587"},
		{"SMTP Username", "SMTP_USERNAME", "test@test.com"},
		{"SMTP Password", "SMTP_PASSWORD", "secret password"},
		{"From Email", "FROM_EMAIL", "from@test.com"},
		{"To Emails", "TO_EMAILS", "to1@test.com,to2@test.com"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := os.Getenv(tc.envVar); got != tc.expected {
				t.Errorf("loadEnvFile() %s = %v, want %v", tc.envVar, got, tc.expected)
			}
		})
	}
}

type mockNotifier struct {
	sent     bool
	lastSubj string
	lastMsg  string
	err      error
}

func (m *mockNotifier) Send(ctx context.Context, subject, message string) error {
	if m.err != nil {
		return m.err
	}
	m.sent = true
	m.lastSubj = subject
	m.lastMsg = message
	return nil
}

type mockSMTPServer struct {
	t       *testing.T
	ln      net.Listener
	handler func(net.Conn)
}

func newMockSMTPServer(t *testing.T) *mockSMTPServer {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock SMTP server: %v", err)
	}

	server := &mockSMTPServer{
		t:  t,
		ln: ln,
	}

	server.handler = func(conn net.Conn) {
		defer conn.Close()

		// Send greeting
		conn.Write([]byte("220 mock.smtp.server ESMTP\r\n"))

		// Read client commands
		buf := make([]byte, 1024)

		// EHLO/HELO
		conn.Read(buf)
		conn.Write([]byte("250-mock.smtp.server\r\n250-AUTH LOGIN PLAIN\r\n250 8BITMIME\r\n"))

		// AUTH
		conn.Read(buf)
		conn.Write([]byte("235 Authentication successful\r\n"))

		// MAIL FROM
		conn.Read(buf)
		conn.Write([]byte("250 Sender OK\r\n"))

		// RCPT TO
		conn.Read(buf)
		conn.Write([]byte("250 Recipient OK\r\n"))

		// DATA
		conn.Read(buf)
		conn.Write([]byte("354 Start mail input\r\n"))

		// Read email content
		for {
			n, _ := conn.Read(buf)
			if n > 0 && strings.Contains(string(buf[:n]), "\r\n.\r\n") {
				break
			}
		}
		conn.Write([]byte("250 OK\r\n"))

		// QUIT
		conn.Read(buf)
		conn.Write([]byte("221 Bye\r\n"))
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go server.handler(conn)
		}
	}()

	return server
}

func (s *mockSMTPServer) close() {
	if s.ln != nil {
		s.ln.Close()
	}
}

func (s *mockSMTPServer) address() string {
	return s.ln.Addr().String()
}

func TestEmailNotifierSend(t *testing.T) {
	// Start mock SMTP server
	mockServer := newMockSMTPServer(t)
	defer mockServer.close()

	tests := []struct {
		name    string
		subject string
		message string
		wantErr bool
	}{
		{
			name:    "Valid email",
			subject: "Test Subject",
			message: "Test Message",
			wantErr: false,
		},
		{
			name:    "Empty subject",
			subject: "",
			message: "Test Message",
			wantErr: false,
		},
		{
			name:    "Empty message",
			subject: "Test Subject",
			message: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment for test
			os.Setenv("SMTP_SERVER", "127.0.0.1")
			os.Setenv("SMTP_PORT", mockServer.address()[strings.LastIndex(mockServer.address(), ":")+1:])
			os.Setenv("SMTP_USERNAME", "test@example.com")
			os.Setenv("SMTP_PASSWORD", "password")
			os.Setenv("FROM_EMAIL", "from@example.com")
			os.Setenv("TO_EMAILS", "to@example.com")
			os.Setenv("SKIP_TLS", "true")

			notifier := NewNotifier()

			// Add timeout to prevent test from hanging
			done := make(chan error)
			go func() {
				done <- notifier.Send(context.Background(), tt.subject, tt.message)
			}()

			select {
			case err := <-done:
				if (err != nil) != tt.wantErr {
					t.Errorf("EmailNotifier.Send() error = %v, wantErr %v", err, tt.wantErr)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("Test timed out after 2 seconds")
			}

			// Clean up environment
			os.Unsetenv("SMTP_SERVER")
			os.Unsetenv("SMTP_PORT")
			os.Unsetenv("SMTP_USERNAME")
			os.Unsetenv("SMTP_PASSWORD")
			os.Unsetenv("FROM_EMAIL")
			os.Unsetenv("TO_EMAILS")
			os.Unsetenv("SKIP_TLS")
		})
	}
}
