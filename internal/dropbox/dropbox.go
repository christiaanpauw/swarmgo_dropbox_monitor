package dropbox

import (
	"fmt"
	"log"
	"os"
        "github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/state"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
)

// TestConnection verifies Dropbox authentication on startup
func TestConnection() error {
	token := os.Getenv("DROPBOX_ACCESS_TOKEN")
	if token == "" {
		return fmt.Errorf("Dropbox access token not set")
	}

	config := dropbox.Config{Token: token}
	dbx := files.New(config)

	// 🔹 Make a test API call to list root folder
	_, err := dbx.ListFolder(files.NewListFolderArg(""))
	if err != nil {
		return fmt.Errorf("failed to connect to Dropbox API: %v", err)
	}

	return nil
}

// CheckForChanges connects to Dropbox and checks for file changes since the last check
func CheckForChanges() ([]string, error) {
	token := os.Getenv("DROPBOX_ACCESS_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("Dropbox access token not set")
	}

	config := dropbox.Config{Token: token}
	dbx := files.New(config)

	// Load last known cursor
	cursor, err := state.Load()
	if err != nil {
		log.Printf("Error loading state: %v", err)
	}

	var changes []string

	if cursor == "" {
		// First-time full folder scan
		log.Println("No previous state found. Fetching full Dropbox folder.")
		arg := files.NewListFolderArg("")
		res, err := dbx.ListFolder(arg)
		if err != nil {
			return nil, fmt.Errorf("error listing folder: %v", err)
		}

		for _, entry := range res.Entries {
			switch f := entry.(type) {
			case *files.FileMetadata:
				changes = append(changes, fmt.Sprintf("New File: %s (created at %s)", f.Name, f.ServerModified))
			case *files.FolderMetadata:
				changes = append(changes, fmt.Sprintf("New Folder: %s", f.Name))
			}
		}

		// Save cursor for future incremental updates
		err = state.Save(res.Cursor)
		if err != nil {
			log.Printf("Failed to save state: %v", err)
		}
	} else {
		// Fetch only changes since the last cursor
		log.Println("Fetching Dropbox changes since last check...")
		arg := files.NewListFolderContinueArg(cursor)
		res, err := dbx.ListFolderContinue(arg)
		if err != nil {
			return nil, fmt.Errorf("error listing folder changes: %v", err)
		}

		for _, entry := range res.Entries {
			switch f := entry.(type) {
			case *files.FileMetadata:
				changes = append(changes, fmt.Sprintf("Modified File: %s (at %s)", f.Name, f.ServerModified))
			case *files.DeletedMetadata:
				changes = append(changes, fmt.Sprintf("Deleted: %s", f.Name))
			}
		}

		// Save new cursor state
		err = state.Save(res.Cursor)
		if err != nil {
			log.Printf("Failed to save state: %v", err)
		}
	}

	if len(changes) == 0 {
		log.Println("No new changes detected.")
		return nil, nil
	}

	return changes, nil
}

