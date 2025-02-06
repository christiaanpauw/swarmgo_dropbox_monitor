package dropbox

import (
    "log"
    "os"
    "testing"
    "github.com/joho/godotenv"
)

func init() {
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }
}

func TestTestConnection(t *testing.T) {
    // Load environment variables from .env file
    err := godotenv.Load()
    if err != nil {
        t.Fatalf("Error loading .env file")
    }

    // Backup current environment variable
    originalToken := os.Getenv("DROPBOX_ACCESS_TOKEN")

    // Test case: Environment variable not set
    os.Setenv("DROPBOX_ACCESS_TOKEN", "")
    err = TestConnection()
    if err == nil || err.Error() != "Dropbox access token not set" {
        t.Errorf("Expected error 'Dropbox access token not set', got %v", err)
    }

    // Test case: Environment variable set to a valid value
    os.Setenv("DROPBOX_ACCESS_TOKEN", originalToken)
    err = TestConnection()
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }

    // Restore original environment variable
    os.Setenv("DROPBOX_ACCESS_TOKEN", originalToken)
}
