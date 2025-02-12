package notify

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestEmailNotifier(t *testing.T) {
	// Save original environment variables
	originals := map[string]string{
		"SMTP_SERVER":   os.Getenv("SMTP_SERVER"),
		"SMTP_PORT":     os.Getenv("SMTP_PORT"),
		"SMTP_USERNAME": os.Getenv("SMTP_USERNAME"),
		"SMTP_PASSWORD": os.Getenv("SMTP_PASSWORD"),
		"FROM_EMAIL":    os.Getenv("FROM_EMAIL"),
		"TO_EMAILS":     os.Getenv("TO_EMAILS"),
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

	server := newMockSMTPServer(t)
	defer server.close()

	port, err := strconv.Atoi(strings.Split(server.address(), ":")[1])
	if err != nil {
		t.Fatalf("Failed to parse port number: %v", err)
	}

	tests := []struct {
		name    string
		config  config.EmailConfig
		wantErr bool
	}{
		{
			name: "Valid configuration",
			config: config.EmailConfig{
				SMTPHost:     "127.0.0.1",
				SMTPPort:     port,
				SMTPUsername: "test@test.com",
				SMTPPassword: "password",
				FromAddress:  "from@test.com",
				ToAddresses:  []string{"to1@test.com", "to2@test.com"},
			},
			wantErr: false,
		},
		{
			name: "Invalid configuration - empty fields",
			config: config.EmailConfig{
				SMTPHost:    "",
				ToAddresses: []string{}, // Empty slice to test error handling
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			notifier := NewEmailNotifier(&tc.config)
			err := notifier.SendNotification(context.Background(), "Test Message")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// mockSMTPServer implements a mock SMTP server for testing
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

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					t.Errorf("Accept error: %v", err)
				}
				return
			}
			go server.handleConnection(conn)
		}
	}()

	return server
}

func (s *mockSMTPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Send greeting
	conn.Write([]byte("220 mock.smtp.server\r\n"))

	// Simple SMTP conversation
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		cmd := string(buf[:n])
		switch {
		case strings.HasPrefix(cmd, "EHLO"):
			conn.Write([]byte("250-mock.smtp.server\r\n250 AUTH LOGIN PLAIN\r\n"))
		case strings.HasPrefix(cmd, "AUTH PLAIN"):
			conn.Write([]byte("235 Authentication successful\r\n"))
		case strings.HasPrefix(cmd, "MAIL FROM"):
			conn.Write([]byte("250 Ok\r\n"))
		case strings.HasPrefix(cmd, "RCPT TO"):
			conn.Write([]byte("250 Ok\r\n"))
		case strings.HasPrefix(cmd, "DATA"):
			conn.Write([]byte("354 End data with <CR><LF>.<CR><LF>\r\n"))
		case strings.Contains(cmd, "\r\n.\r\n"):
			conn.Write([]byte("250 Ok: message queued\r\n"))
		case strings.HasPrefix(cmd, "QUIT"):
			conn.Write([]byte("221 Bye\r\n"))
			return
		default:
			// For any other command, acknowledge it
			conn.Write([]byte("250 Ok\r\n"))
		}
	}
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
	server := newMockSMTPServer(t)
	defer server.close()

	port, err := strconv.Atoi(strings.Split(server.address(), ":")[1])
	if err != nil {
		t.Fatalf("Failed to parse port number: %v", err)
	}

	cfg := config.EmailConfig{
		SMTPHost:     "127.0.0.1",
		SMTPPort:     port,
		SMTPUsername: "test@test.com",
		SMTPPassword: "password",
		FromAddress:  "from@test.com",
		ToAddresses:  []string{"to@test.com"},
	}

	notifier := NewEmailNotifier(&cfg)

	ctx := context.Background()
	err = notifier.SendNotification(ctx, "Test Message")
	assert.NoError(t, err)
}
