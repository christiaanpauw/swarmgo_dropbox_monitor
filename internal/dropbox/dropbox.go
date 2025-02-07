package dropbox

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/state"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file - a")
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

// TestConnection verifies Dropbox authentication on startup
func TestConnection() error {
	token := getDropboxAccessToken()
	fmt.Println("\nDROPBOX_ACCESS_TOKEN:", token, "\n")
	if token == "" {
		return fmt.Errorf("Dropbox access token not set - a")
	}

	config := dropbox.Config{Token: token}
	dbx := files.New(config)

	// 🔹 Make a test API call to list root folder
	_, err := dbx.ListFolder(files.NewListFolderArg(""))
	if err != nil {
		return fmt.Errorf("failed to connect to Dropbox API: %v - b", err)
	}

	return nil
}

// CheckForChanges connects to Dropbox and checks for file changes since the last check
func CheckForChanges() ([]string, error) {
	token := getDropboxAccessToken()
	if token == "" {
		return nil, fmt.Errorf("Dropbox access token not set - b")
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

func ListFolders() {
	config := dropbox.Config{Token: getDropboxAccessToken()}
	dbx := files.New(config)

	arg := files.NewListFolderArg("")
	res, err := dbx.ListFolder(arg)
	if err != nil {
		log.Printf("Error listing folders: %v", err)
		return
	}

	for _, entry := range res.Entries {
		if folder, ok := entry.(*files.FolderMetadata); ok {
			fmt.Printf("Folder: %s\n", folder.Name)
		}
	}
}

func InspectAccessToken() {
	config := dropbox.Config{Token: getDropboxAccessToken()}
	dbx := users.New(config)

	account, err := dbx.GetCurrentAccount()
	if err != nil {
		log.Printf("Failed to inspect access token: %v", err)
		return
	}

	fmt.Printf("Account ID: %s\n", account.AccountId)
	fmt.Printf("Email: %s\n", account.Email)
	fmt.Printf("Name: %s %s\n", account.Name.GivenName, account.Name.Surname)
}

func ListLastChangedDates() {
	config := dropbox.Config{Token: getDropboxAccessToken()}
	dbx := files.New(config)

	arg := files.NewListFolderArg("")
	res, err := dbx.ListFolder(arg)
	if err != nil {
		log.Printf("Error listing folders: %v", err)
		return
	}

	for _, entry := range res.Entries {
		if folder, ok := entry.(*files.FolderMetadata); ok {
			fmt.Printf("Folder: %s\n", folder.Name)

			// Check for changes in each folder
			folderArg := files.NewListFolderArg(folder.PathLower)
			folderRes, err := dbx.ListFolder(folderArg)
			if err != nil {
				log.Printf("Error checking folder %s: %v", folder.Name, err)
				continue
			}

			for _, fileEntry := range folderRes.Entries {
				if fileMetadata, ok := fileEntry.(*files.FileMetadata); ok {
					fmt.Printf("  File: %s, Last Modified: %s\n", fileMetadata.Name, fileMetadata.ClientModified)
				}
			}
		}
	}
}

func GetFolders() []string {
	config := dropbox.Config{Token: getDropboxAccessToken()}
	dbx := files.New(config)

	arg := files.NewListFolderArg("")
	res, err := dbx.ListFolder(arg)
	if err != nil {
		log.Printf("Error listing folders: %v", err)
		return nil
	}

	var folderNames []string
	for _, entry := range res.Entries {
		if folder, ok := entry.(*files.FolderMetadata); ok {
			folderNames = append(folderNames, folder.Name)
		}
	}
	return folderNames
}

type FolderInfo struct {
	Name         string
	LastModified string
}

func GetLastChangedFolders() []FolderInfo {
	config := dropbox.Config{Token: getDropboxAccessToken()}
	dbx := files.New(config)

	arg := files.NewListFolderArg("")
	arg.Recursive = true // fetch all entries recursively
	res, err := dbx.ListFolder(arg)
	if err != nil {
		log.Printf("Error listing folders: %v", err)
		return nil
	}

	var folderInfos []FolderInfo
	for _, entry := range res.Entries {
		if fileMetadata, ok := entry.(*files.FileMetadata); ok {
			folderInfos = append(folderInfos, FolderInfo{
				Name:         fileMetadata.Name,
				LastModified: fileMetadata.ClientModified.String(),
			})
		}
	}
	return folderInfos
}

func GetChangesLast24Hours() []FolderInfo {
	config := dropbox.Config{Token: getDropboxAccessToken()}
	dbx := files.New(config)

	arg := files.NewListFolderArg("")
	arg.Recursive = true
	res, err := dbx.ListFolder(arg)
	if err != nil {
		log.Printf("Error listing folders: %v", err)
		return nil
	}

	var folderInfos []FolderInfo
	cutoffTime := time.Now().Add(-24 * time.Hour)
	for _, entry := range res.Entries {
		if fileMetadata, ok := entry.(*files.FileMetadata); ok {
			if fileMetadata.ClientModified.After(cutoffTime) {
				folderInfos = append(folderInfos, FolderInfo{
					Name:         fileMetadata.Name,
					LastModified: fileMetadata.ClientModified.String(),
				})
			}
		}
	}
	return folderInfos
}
