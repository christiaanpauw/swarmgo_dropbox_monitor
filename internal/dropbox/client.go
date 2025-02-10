package dropbox

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// FileMetadata represents the metadata for a Dropbox file
type FileMetadata struct {
	Tag            string `json:".tag"`
	Name           string `json:"name"`
	PathLower      string `json:"path_lower"`
	PathDisplay    string `json:"path_display"`
	ID             string `json:"id"`
	ClientModified string `json:"client_modified"`
	ServerModified string `json:"server_modified"`
	Rev            string `json:"rev"`
	Size           int64  `json:"size"`
	IsDownloadable bool   `json:"is_downloadable"`
	ContentHash    string `json:"content_hash"`
	SharingInfo    struct {
		ReadOnly             bool        `json:"read_only"`
		ParentSharedFolderID string     `json:"parent_shared_folder_id"`
		ModifiedBy          interface{} `json:"modified_by"` // Changed to interface{} to handle both string and struct
	} `json:"sharing_info"`
}

// RetryConfig defines the configuration for API call retries
type RetryConfig struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:  3,
		InitialWait: 1 * time.Second,
		MaxWait:     10 * time.Second,
	}
}

// DropboxClient handles interactions with the Dropbox API
type DropboxClient struct {
	accessToken string
	db         *sql.DB
	httpClient *http.Client
	retryConfig RetryConfig
}

// NewDropboxClient creates a new Dropbox client
func NewDropboxClient(token string, db *sql.DB) (*DropboxClient, error) {
	if token == "" {
		return nil, fmt.Errorf("Dropbox access token is required")
	}

	return &DropboxClient{
		accessToken: token,
		db:         db,
		httpClient: &http.Client{},
		retryConfig: DefaultRetryConfig(),
	}, nil
}

// doRequestWithRetry performs an HTTP request with retry logic
func (c *DropboxClient) doRequestWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error
	wait := c.retryConfig.InitialWait

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(wait)
			// Exponential backoff with jitter
			wait = time.Duration(float64(wait) * 1.5)
			if wait > c.retryConfig.MaxWait {
				wait = c.retryConfig.MaxWait
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: %v", attempt+1, err)
			continue
		}

		// Check if we should retry based on status code
		if resp.StatusCode == http.StatusTooManyRequests || 
		   resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("attempt %d: received status code %d", attempt+1, resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("all retry attempts failed: %v", lastErr)
}

// GetFileMetadata retrieves metadata for a file from Dropbox
func (c *DropboxClient) GetFileMetadata(path string) (*FileMetadata, error) {
	url := "https://api.dropboxapi.com/2/files/get_metadata"
	data := map[string]interface{}{
		"path": path,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error response from Dropbox (status %d): %s", resp.StatusCode, body)
	}

	var metadata FileMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &metadata, nil
}

// GetModifier gets the user info for a given account ID
func (c *DropboxClient) GetModifier(accountID string) (string, error) {
	if accountID == "" {
		return "", nil
	}

	url := "https://api.dropboxapi.com/2/users/get_account"
	data := map[string]string{
		"account_id": accountID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error response from Dropbox (status %d): %s", resp.StatusCode, body)
	}

	var result struct {
		Name struct {
			DisplayName string `json:"display_name"`
		} `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	return result.Name.DisplayName, nil
}

// PopulateFirstNFiles gets metadata for the first N files and stores it in the database
func (c *DropboxClient) PopulateFirstNFiles(n int) error {
	// List files endpoint
	url := "https://api.dropboxapi.com/2/files/list_folder"
	data := map[string]interface{}{
		"path": "",
		"recursive": true,
		"limit": n * 2, // Request more items since we'll filter out folders
		"include_media_info": false,
		"include_deleted": false,
		"include_has_explicit_shared_members": false,
		"include_mounted_folders": true,
		"include_non_downloadable_files": true,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error response from Dropbox (status %d): %s", resp.StatusCode, body)
	}

	var result struct {
		Entries []FileMetadata `json:"entries"`
		Cursor  string        `json:"cursor"`
		HasMore bool          `json:"has_more"`
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response from Dropbox: %s\n", string(body))

	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}

	fmt.Printf("Found %d entries\n", len(result.Entries))

	// Store each file's metadata in the database
	filesProcessed := 0
	for _, file := range result.Entries {
		if file.Tag != "file" {
			fmt.Printf("Skipping folder: %s\n", file.PathLower)
			continue
		}

		fmt.Printf("Processing file: %s\n", file.PathLower)

		// Get user info for modifier
		modifierName := ""
		if file.SharingInfo.ModifiedBy != nil {
			modifierName, err = c.GetModifier(file.SharingInfo.ModifiedBy.(string))
			if err != nil {
				// Log error but continue processing
				fmt.Printf("Warning: Error getting modifier info for %s: %v\n", file.PathLower, err)
			}
		}

		fmt.Printf("Inserting file %s with modifier %s\n", file.PathLower, modifierName)

		// Get file extension
		fileExt := ""
		if lastDot := strings.LastIndex(file.Name, "."); lastDot >= 0 {
			fileExt = strings.ToLower(file.Name[lastDot+1:])
		}

		// Parse timestamps
		clientModified, err := time.Parse(time.RFC3339, file.ClientModified)
		if err != nil {
			fmt.Printf("Warning: Error parsing client_modified time for %s: %v\n", file.PathLower, err)
			continue
		}

		serverModified, err := time.Parse(time.RFC3339, file.ServerModified)
		if err != nil {
			fmt.Printf("Warning: Error parsing server_modified time for %s: %v\n", file.PathLower, err)
			continue
		}

		// Insert into database
		_, err = c.db.Exec(`
			INSERT INTO file_changes (
				file_path, modified_at, file_type, dropbox_id, dropbox_rev,
				client_modified, server_modified, size, is_downloadable,
				modified_by_id, modified_by_name, content_hash
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			file.PathLower,
			serverModified,
			fileExt,
			file.ID,
			file.Rev,
			clientModified,
			serverModified,
			file.Size,
			file.IsDownloadable,
			"",
			modifierName,
			file.ContentHash,
		)
		if err != nil {
			return fmt.Errorf("error inserting file %s: %v", file.PathLower, err)
		}
		fmt.Printf("Successfully inserted file %s\n", file.PathLower)

		filesProcessed++
		if filesProcessed >= n {
			break
		}
	}

	return nil
}

// GetChangesLast10Minutes retrieves changes from the last 10 minutes
func (c *DropboxClient) GetChangesLast10Minutes() ([]FileMetadata, error) {
	tenMinutesAgo := time.Now().Add(-10 * time.Minute)
	return c.GetChanges(tenMinutesAgo)
}

// GetChangesLast24Hours retrieves changes from the last 24 hours
func (c *DropboxClient) GetChangesLast24Hours() ([]FileMetadata, error) {
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
	return c.GetChanges(twentyFourHoursAgo)
}

// GetChanges retrieves changes since the specified time
func (c *DropboxClient) GetChanges(since time.Time) ([]FileMetadata, error) {
	log.Printf("🔍 Checking for changes since: %v", since.Format(time.RFC3339))
	log.Printf("🔑 Using access token: %s...", c.accessToken[:10])

	// Use list_folder directly to get all files
	listURL := "https://api.dropboxapi.com/2/files/list_folder"
	listData := map[string]interface{}{
		"path": "",
		"recursive": true,
		"include_media_info": false,
		"include_deleted": false,
		"include_has_explicit_shared_members": false,
		"include_mounted_folders": true,
		"include_non_downloadable_files": true,
		"limit": 2000,
	}

	jsonData, err := json.Marshal(listData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling list request: %v", err)
	}

	log.Printf("📤 Sending list request with data: %s", string(jsonData))
	
	req, err := http.NewRequest("POST", listURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating list request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Set a timeout for the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making list request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("📥 List response (status %d): %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response from Dropbox (status %d): %s", resp.StatusCode, body)
	}

	var result struct {
		Entries []FileMetadata `json:"entries"`
		Cursor  string        `json:"cursor"`
		HasMore bool          `json:"has_more"`
	}

	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding list response: %v", err)
	}

	log.Printf("📚 Found %d entries in response", len(result.Entries))

	// Filter entries by time
	var changes []FileMetadata
	log.Printf("🕒 Looking for files modified after: %v", since.Format(time.RFC3339))
	
	for _, file := range result.Entries {
		if file.Tag != "file" {
			continue
		}

		log.Printf("🔎 Examining entry: %+v", file)

		serverModified, err := time.Parse(time.RFC3339, file.ServerModified)
		if err != nil {
			log.Printf("⚠️ Error parsing time for file %s: %v", file.PathDisplay, err)
			continue
		}

		log.Printf("📄 Checking file: %s (modified: %s)", file.PathDisplay, serverModified.Format(time.RFC3339))
		if serverModified.After(since) {
			log.Printf("✨ Found change: %s (modified at %s)", file.PathDisplay, serverModified.Format(time.RFC3339))
			changes = append(changes, file)
		}
	}

	// If we have more results, get them using the cursor
	for result.HasMore {
		log.Printf("📑 Getting more results using cursor: %s", result.Cursor)
		continueURL := "https://api.dropboxapi.com/2/files/list_folder/continue"
		continueData := map[string]interface{}{
			"cursor": result.Cursor,
		}

		jsonData, err = json.Marshal(continueData)
		if err != nil {
			return nil, fmt.Errorf("error marshaling continue request: %v", err)
		}

		req, err = http.NewRequest("POST", continueURL, bytes.NewReader(jsonData))
		if err != nil {
			return nil, fmt.Errorf("error creating continue request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.accessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error making continue request: %v", err)
		}
		defer resp.Body.Close()

		body, _ = io.ReadAll(resp.Body)
		log.Printf("📥 Continue response (status %d): %s", resp.StatusCode, string(body))

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error response from Dropbox (status %d): %s", resp.StatusCode, body)
		}

		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
			return nil, fmt.Errorf("error decoding continue response: %v", err)
		}

		for _, file := range result.Entries {
			if file.Tag != "file" {
				continue
			}

			log.Printf("🔎 Examining entry: %+v", file)

			serverModified, err := time.Parse(time.RFC3339, file.ServerModified)
			if err != nil {
				log.Printf("⚠️ Error parsing time for file %s: %v", file.PathDisplay, err)
				continue
			}

			log.Printf("📄 Checking file: %s (modified: %s)", file.PathDisplay, serverModified.Format(time.RFC3339))
			if serverModified.After(since) {
				log.Printf("✨ Found change: %s (modified at %s)", file.PathDisplay, serverModified.Format(time.RFC3339))
				changes = append(changes, file)
			}
		}
	}

	log.Printf("✅ Found %d changes after filtering", len(changes))
	return changes, nil
}

// GetFolders retrieves all folders from Dropbox
func (c *DropboxClient) GetFolders() ([]string, error) {
	url := "https://api.dropboxapi.com/2/files/list_folder"
	data := map[string]interface{}{
		"path": "",
		"recursive": true,
		"include_media_info": false,
		"include_deleted": false,
		"include_has_explicit_shared_members": false,
		"include_mounted_folders": true,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error response from Dropbox (status %d): %s", resp.StatusCode, body)
	}

	var result struct {
		Entries []FileMetadata `json:"entries"`
		Cursor  string        `json:"cursor"`
		HasMore bool          `json:"has_more"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	var folders []string
	for _, entry := range result.Entries {
		if entry.Tag == "folder" {
			folders = append(folders, entry.PathLower)
		}
	}

	return folders, nil
}

// GetLastChangedFolders retrieves folders that have had changes in the last 24 hours
func (c *DropboxClient) GetLastChangedFolders() ([]string, error) {
	changes, err := c.GetChangesLast24Hours()
	if err != nil {
		return nil, err
	}

	folderSet := make(map[string]bool)
	var folders []string

	for _, change := range changes {
		folder := filepath.Dir(change.PathLower)
		if !folderSet[folder] {
			folderSet[folder] = true
			folders = append(folders, folder)
		}
	}

	return folders, nil
}

// ListFiles lists files in a Dropbox folder
func (c *DropboxClient) ListFiles(path string) ([]FileMetadata, error) {
	url := "https://api.dropboxapi.com/2/files/list_folder"
	data := map[string]interface{}{
		"path": path,
		"recursive": true,
		"include_media_info": false,
		"include_deleted": false,
		"include_has_explicit_shared_members": false,
		"include_mounted_folders": true,
		"include_non_downloadable_files": true,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request data: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error response from Dropbox (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Entries []FileMetadata `json:"entries"`
		Cursor string         `json:"cursor"`
		HasMore bool          `json:"has_more"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return result.Entries, nil
}
