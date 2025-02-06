package dropbox

import (
    "fmt"
    "log"
    "os"
    "strings"
    "testing"
    "github.com/joho/godotenv"
)

func init() {
    err := godotenv.Load("../../.env")
    if err != nil {
        log.Fatalf("Error loading .env file - 1")
    }

    fmt.Println("Environment Variables:")
    for _, e := range os.Environ() {
        fmt.Println(e)
    }
}

func getDropboxAccessToken() string {
    for _, e := range os.Environ() {
        if strings.HasPrefix(e, "DROPBOX_ACCESS_TOKEN=") {
            return strings.TrimPrefix(e, "DROPBOX_ACCESS_TOKEN=")
        }
    }
    return ""
}

func TestTestConnection(t *testing.T) {
    // Log the current working directory
    cwd, err := os.Getwd()
    if err != nil {
        t.Fatalf("Error getting current working directory: %v", err)
    }
    log.Printf("Current working directory: %s", cwd)

    // Check if .env file exists
    if _, err := os.Stat("../../.env"); os.IsNotExist(err) {
        t.Fatalf(".env file does not exist")
    }

    // Load environment variables from .env file
    err = godotenv.Load("../../.env")
    if err != nil {
        t.Fatalf("Error loading .env file - 2")
    }

    // Test case: Environment variable set to a valid value
    token := getDropboxAccessToken()
    if token == "" {
        t.Fatalf("DROPBOX_ACCESS_TOKEN not found in environment variables")
    }
    err = TestConnection()
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
}

func TestConnection() error {
    token := getDropboxAccessToken()
    fmt.Println("DROPBOX_ACCESS_TOKEN:", token)

    if token == "" {
        return fmt.Errorf("Dropbox access token not set - a")
    }

    // Simulate other logic here
    if token == "a" {
        return fmt.Errorf("Simulated error with token 'a'")
    }

    return nil
}
