package dropbox

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/state"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"
	"github.com/joho/godotenv"
)

const (
	apiTimeout      = 30 * time.Second
	longpollTimeout = 30 // seconds
	maxBatchSize    = 500
)

func init() {
	// Try loading .env from different possible locations
	locations := []string{
		".env",                    // Current directory
		"../../.env",              // Project root when running from cmd/gui
		"../../../.env",           // In case we're one level deeper
	}

	var loaded bool
	for _, loc := range locations {
		if err := godotenv.Load(loc); err == nil {
			loaded = true
			break
		}
	}

	if !loaded {
		log.Printf("Warning: Could not load .env file from any location")
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
	log.Printf("DROPBOX_ACCESS_TOKEN: %s", token)
	if token == "" {
		return fmt.Errorf("Dropbox access token not set - a")
	}

	config := dropbox.Config{Token: token}
	dbx := files.New(config)

	// Make a test API call to list root folder
	startTime := time.Now()

	// Use a channel to implement timeout
	type result struct {
		res *files.ListFolderResult
		err error
	}
	ch := make(chan result, 1)

	go func() {
		res, err := dbx.ListFolder(files.NewListFolderArg(""))
		ch <- result{res, err}
	}()

	// Wait for result or timeout
	select {
	case r := <-ch:
		if r.err != nil {
			if time.Since(startTime) > apiTimeout {
				return fmt.Errorf("timeout connecting to Dropbox API after %v", apiTimeout)
			}
			return fmt.Errorf("failed to connect to Dropbox API: %v - b", r.err)
		}
		log.Printf("ListFolder API call completed in %v", time.Since(startTime))
	case <-time.After(apiTimeout):
		return fmt.Errorf("timeout connecting to Dropbox API after %v", apiTimeout)
	}

	return nil
}

// DropboxClient wraps the Dropbox client with additional functionality
type DropboxClient struct {
	Client files.Client
	Config dropbox.Config
	db     *sql.DB
	mu     sync.RWMutex

	// Cache state
	cursor    string
	lastCheck time.Time
	statePath string
}

// NewDropboxClient creates a new Dropbox client
func NewDropboxClient(token string, db *sql.DB) (*DropboxClient, error) {
	config := dropbox.Config{
		Token: token,
	}

	return &DropboxClient{
		Client:    files.New(config),
		Config:    config,
		db:        db,
		statePath: "data/dropbox_state.json",
	}, nil
}

func (c *DropboxClient) initDB() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS file_changes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL,
			modified_at TIMESTAMP NOT NULL,
			hash TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(path, modified_at)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_file_changes_modified_at ON file_changes(modified_at)`,
		`CREATE TABLE IF NOT EXISTS sync_state (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cursor TEXT NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := c.db.Exec(query); err != nil {
			return fmt.Errorf("error executing query %q: %v", query, err)
		}
	}

	return nil
}

func (c *DropboxClient) loadState() error {
	var cursor string
	err := c.db.QueryRow("SELECT cursor FROM sync_state ORDER BY timestamp DESC LIMIT 1").Scan(&cursor)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.cursor = cursor
	c.mu.Unlock()
	return nil
}

func (c *DropboxClient) saveState(cursor string) error {
	_, err := c.db.Exec("INSERT INTO sync_state (cursor) VALUES (?)", cursor)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.cursor = cursor
	c.mu.Unlock()
	return nil
}

// GetChanges gets changes since a specific time
func (c *DropboxClient) GetChanges(since time.Time) ([]string, error) {
	log.Printf("Getting changes since %v...", since)
	
	c.mu.Lock()
	cursor := c.cursor
	c.mu.Unlock()

	// Initialize if needed
	if cursor == "" {
		log.Printf("No cursor found, starting initial sync...")
		if err := c.initialSync(); err != nil {
			return nil, fmt.Errorf("error in initial sync: %v", err)
		}
		log.Printf("Initial sync completed successfully")
		
		// Get the cursor after initialization
		c.mu.Lock()
		cursor = c.cursor
		c.mu.Unlock()
	}

	// Get changes using cursor
	changes := make([]string, 0)
	hasMore := true

	for hasMore {
		log.Printf("Fetching changes with cursor: %s", cursor)
		res, err := c.ListFolderContinue(files.NewListFolderContinueArg(cursor))
		if err != nil {
			return nil, fmt.Errorf("error listing folder: %v", err)
		}

		for _, entry := range res.Entries {
			if metadata, ok := entry.(*files.FileMetadata); ok {
				if metadata.ServerModified.After(since) {
					changes = append(changes, metadata.PathLower)
				}
			}
		}

		hasMore = res.HasMore
		cursor = res.Cursor
	}

	// Save the cursor
	if err := c.saveCursor(cursor); err != nil {
		log.Printf("Warning: Could not save cursor: %v", err)
	}

	log.Printf("Found %d changes since %v", len(changes), since)
	return changes, nil
}

func (c *DropboxClient) saveCursor(cursor string) error {
	_, err := c.db.Exec("INSERT INTO sync_state (cursor) VALUES (?)", cursor)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.cursor = cursor
	c.mu.Unlock()
	return nil
}

func (c *DropboxClient) initialSync() error {
	log.Printf("Starting initial sync...")
	startTime := time.Now()

	// Use a channel to implement timeout
	type result struct {
		res *files.ListFolderResult
		err error
	}
	ch := make(chan result, 1)

	go func() {
		arg := files.NewListFolderArg("")
		arg.Recursive = true
		arg.IncludeDeleted = false
		res, err := c.ListFolder(arg)
		ch <- result{res, err}
	}()

	// Wait for result or timeout
	var res *files.ListFolderResult
	select {
	case r := <-ch:
		if r.err != nil {
			if time.Since(startTime) > apiTimeout {
				return fmt.Errorf("timeout listing folder after %v", apiTimeout)
			}
			return fmt.Errorf("error listing folder: %v", r.err)
		}
		res = r.res
		log.Printf("ListFolder API call completed in %v", time.Since(startTime))
	case <-time.After(apiTimeout):
		return fmt.Errorf("timeout listing folder after %v", apiTimeout)
	}

	// Process entries in batch
	for _, entry := range res.Entries {
		switch metadata := entry.(type) {
		case *files.FileMetadata:
			log.Printf("File: %s, Last Modified: %s", metadata.Name, metadata.ClientModified)
		case *files.FolderMetadata:
			log.Printf("Folder: %s", metadata.Name)
		}
	}

	// Save cursor for future incremental updates
	err := c.saveCursor(res.Cursor)
	if err != nil {
		log.Printf("Failed to save state: %v", err)
	}

	return nil
}

// ListFolder wraps the Dropbox ListFolder API with retries
func (c *DropboxClient) ListFolder(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
	log.Printf("Starting ListFolder request...")
	var result *files.ListFolderResult
	err := c.retryOperation(func() error {
		var err error
		result, err = c.Client.ListFolder(arg)
		if err != nil {
			log.Printf("ListFolder error: %v", err)
		} else {
			log.Printf("ListFolder successful, got %d entries", len(result.Entries))
		}
		return err
	})
	return result, err
}

// ListFolderContinue wraps the Dropbox ListFolderContinue API with retries
func (c *DropboxClient) ListFolderContinue(arg *files.ListFolderContinueArg) (*files.ListFolderResult, error) {
	log.Printf("Starting ListFolderContinue request...")
	var result *files.ListFolderResult
	err := c.retryOperation(func() error {
		var err error
		result, err = c.Client.ListFolderContinue(arg)
		if err != nil {
			log.Printf("ListFolderContinue error: %v", err)
		} else {
			log.Printf("ListFolderContinue successful, got %d entries", len(result.Entries))
		}
		return err
	})
	return result, err
}

// GetCurrentAccount gets the current account info
func (c *DropboxClient) GetCurrentAccount() (*users.FullAccount, error) {
	usersClient := users.New(c.Config)
	account, err := usersClient.GetCurrentAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting current account: %v", err)
	}
	return account, nil
}

// GetChangesLast10Minutes gets changes from the last 10 minutes
func (c *DropboxClient) GetChangesLast10Minutes() ([]string, error) {
	since := time.Now().Add(-10 * time.Minute)
	return c.GetChanges(since)
}

// CheckForChanges connects to Dropbox and checks for file changes since the last check
func CheckForChanges() ([]string, error) {
	token := getDropboxAccessToken()
	if token == "" {
		return nil, fmt.Errorf("Dropbox access token not set - b")
	}

	db, err := sql.Open("sqlite3", "./dropbox.db")
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	c, err := NewDropboxClient(token, db)
	if err != nil {
		return nil, err
	}

	// Load last known cursor
	cursor, err := state.Load()
	if err != nil {
		log.Printf("Error loading state: %v", err)
	}

	var changes []string

	if cursor == "" {
		// First-time full folder scan
		log.Println("No previous state found. Fetching full Dropbox folder.")
		startTime := time.Now()

		// Use a channel to implement timeout
		type result struct {
			res *files.ListFolderResult
			err error
		}
		ch := make(chan result, 1)

		go func() {
			res, err := c.ListFolder(files.NewListFolderArg(""))
			ch <- result{res, err}
		}()

		// Wait for result or timeout
		var res *files.ListFolderResult
		select {
		case r := <-ch:
			if r.err != nil {
				if time.Since(startTime) > apiTimeout {
					return nil, fmt.Errorf("timeout listing folder after %v", apiTimeout)
				}
				return nil, fmt.Errorf("error listing folder: %v", r.err)
			}
			res = r.res
			log.Printf("ListFolder API call completed in %v", time.Since(startTime))
		case <-time.After(apiTimeout):
			return nil, fmt.Errorf("timeout listing folder after %v", apiTimeout)
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
		startTime := time.Now()

		// Use a channel to implement timeout
		type result struct {
			res *files.ListFolderResult
			err error
		}
		ch := make(chan result, 1)

		go func() {
			res, err := c.ListFolderContinue(files.NewListFolderContinueArg(cursor))
			ch <- result{res, err}
		}()

		// Wait for result or timeout
		var res *files.ListFolderResult
		select {
		case r := <-ch:
			if r.err != nil {
				if time.Since(startTime) > apiTimeout {
					return nil, fmt.Errorf("timeout listing folder changes after %v", apiTimeout)
				}
				return nil, fmt.Errorf("error listing folder changes: %v", r.err)
			}
			res = r.res
			log.Printf("ListFolderContinue API call completed in %v", time.Since(startTime))
		case <-time.After(apiTimeout):
			return nil, fmt.Errorf("timeout listing folder changes after %v", apiTimeout)
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

// ListFolders lists all folders in the Dropbox account
func ListFolders() {
	token := getDropboxAccessToken()
	if token == "" {
		log.Println("Dropbox access token not set")
		return
	}

	db, err := sql.Open("sqlite3", "./dropbox.db")
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return
	}
	defer db.Close()

	c, err := NewDropboxClient(token, db)
	if err != nil {
		log.Printf("Error creating Dropbox client: %v", err)
		return
	}

	startTime := time.Now()

	// Use a channel to implement timeout
	type result struct {
		res *files.ListFolderResult
		err error
	}
	ch := make(chan result, 1)

	go func() {
		res, err := c.ListFolder(files.NewListFolderArg(""))
		ch <- result{res, err}
	}()

	// Wait for result or timeout
	var res *files.ListFolderResult
	select {
	case r := <-ch:
		if r.err != nil {
			if time.Since(startTime) > apiTimeout {
				log.Printf("timeout listing folders after %v", apiTimeout)
				return
			}
			log.Printf("Error listing folders: %v", r.err)
			return
		}
		res = r.res
		log.Printf("ListFolder API call completed in %v", time.Since(startTime))
	case <-time.After(apiTimeout):
		log.Printf("timeout listing folders after %v", apiTimeout)
		return
	}

	for _, entry := range res.Entries {
		if folder, ok := entry.(*files.FolderMetadata); ok {
			fmt.Printf("Folder: %s\n", folder.Name)
		}
	}
}

// InspectAccessToken inspects the Dropbox access token
func InspectAccessToken() {
	token := getDropboxAccessToken()
	if token == "" {
		log.Println("Dropbox access token not set")
		return
	}

	db, err := sql.Open("sqlite3", "./dropbox.db")
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return
	}
	defer db.Close()

	c, err := NewDropboxClient(token, db)
	if err != nil {
		log.Printf("Error creating Dropbox client: %v", err)
		return
	}

	startTime := time.Now()

	// Use a channel to implement timeout
	type result struct {
		res *users.FullAccount
		err error
	}
	ch := make(chan result, 1)

	go func() {
		res, err := c.GetCurrentAccount()
		ch <- result{res, err}
	}()

	// Wait for result or timeout
	var res *users.FullAccount
	select {
	case r := <-ch:
		if r.err != nil {
			if time.Since(startTime) > apiTimeout {
				log.Printf("timeout inspecting access token after %v", apiTimeout)
				return
			}
			log.Printf("Failed to inspect access token: %v", r.err)
			return
		}
		res = r.res
		log.Printf("GetCurrentAccount API call completed in %v", time.Since(startTime))
	case <-time.After(apiTimeout):
		log.Printf("timeout inspecting access token after %v", apiTimeout)
		return
	}

	fmt.Printf("Account ID: %s\n", res.AccountId)
	fmt.Printf("Email: %s\n", res.Email)
	fmt.Printf("Name: %s %s\n", res.Name.GivenName, res.Name.Surname)
}

// ListLastChangedDates lists the last changed dates of all files
func ListLastChangedDates() {
	token := getDropboxAccessToken()
	if token == "" {
		log.Println("Dropbox access token not set")
		return
	}

	db, err := sql.Open("sqlite3", "./dropbox.db")
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return
	}
	defer db.Close()

	c, err := NewDropboxClient(token, db)
	if err != nil {
		log.Printf("Error creating Dropbox client: %v", err)
		return
	}

	startTime := time.Now()

	// Use a channel to implement timeout
	type result struct {
		res *files.ListFolderResult
		err error
	}
	ch := make(chan result, 1)

	go func() {
		res, err := c.ListFolder(files.NewListFolderArg(""))
		ch <- result{res, err}
	}()

	// Wait for result or timeout
	var res *files.ListFolderResult
	select {
	case r := <-ch:
		if r.err != nil {
			if time.Since(startTime) > apiTimeout {
				log.Printf("timeout listing folders after %v", apiTimeout)
				return
			}
			log.Printf("Error listing folders: %v", r.err)
			return
		}
		res = r.res
		log.Printf("ListFolder API call completed in %v", time.Since(startTime))
	case <-time.After(apiTimeout):
		log.Printf("timeout listing folders after %v", apiTimeout)
		return
	}

	for _, entry := range res.Entries {
		if folder, ok := entry.(*files.FolderMetadata); ok {
			fmt.Printf("Folder: %s\n", folder.Name)

			// Check for changes in each folder
			folderArg := files.NewListFolderArg(folder.PathLower)
			startTime := time.Now()

			// Use a channel to implement timeout
			type result struct {
				res *files.ListFolderResult
				err error
			}
			ch := make(chan result, 1)

			go func() {
				res, err := c.ListFolder(folderArg)
				ch <- result{res, err}
			}()

			// Wait for result or timeout
			var folderRes *files.ListFolderResult
			select {
			case r := <-ch:
				if r.err != nil {
					if time.Since(startTime) > apiTimeout {
						log.Printf("timeout checking folder %s after %v", folder.Name, apiTimeout)
						continue
					}
					log.Printf("Error checking folder %s: %v", folder.Name, r.err)
					continue
				}
				folderRes = r.res
				log.Printf("ListFolder API call completed in %v", time.Since(startTime))
			case <-time.After(apiTimeout):
				log.Printf("timeout checking folder %s after %v", folder.Name, apiTimeout)
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

// GetFolders returns a list of all folders in the Dropbox account
func GetFolders() []string {
	token := getDropboxAccessToken()
	if token == "" {
		log.Println("Dropbox access token not set")
		return nil
	}

	db, err := sql.Open("sqlite3", "./dropbox.db")
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return nil
	}
	defer db.Close()

	c, err := NewDropboxClient(token, db)
	if err != nil {
		log.Printf("Error creating Dropbox client: %v", err)
		return nil
	}

	startTime := time.Now()

	// Use a channel to implement timeout
	type result struct {
		res *files.ListFolderResult
		err error
	}
	ch := make(chan result, 1)

	go func() {
		res, err := c.ListFolder(files.NewListFolderArg(""))
		ch <- result{res, err}
	}()

	// Wait for result or timeout
	var res *files.ListFolderResult
	select {
	case r := <-ch:
		if r.err != nil {
			if time.Since(startTime) > apiTimeout {
				log.Printf("timeout listing folders after %v", apiTimeout)
				return nil
			}
			log.Printf("Error listing folders: %v", r.err)
			return nil
		}
		res = r.res
		log.Printf("ListFolder API call completed in %v", time.Since(startTime))
	case <-time.After(apiTimeout):
		log.Printf("timeout listing folders after %v", apiTimeout)
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

// GetLastChangedFolders returns a list of folders with their last changed dates
func GetLastChangedFolders() []FolderInfo {
	token := getDropboxAccessToken()
	if token == "" {
		log.Println("Dropbox access token not set")
		return nil
	}

	db, err := sql.Open("sqlite3", "./dropbox.db")
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return nil
	}
	defer db.Close()

	c, err := NewDropboxClient(token, db)
	if err != nil {
		log.Printf("Error creating Dropbox client: %v", err)
		return nil
	}

	startTime := time.Now()

	// Use a channel to implement timeout
	type result struct {
		res *files.ListFolderResult
		err error
	}
	ch := make(chan result, 1)

	go func() {
		res, err := c.ListFolder(files.NewListFolderArg(""))
		ch <- result{res, err}
	}()

	// Wait for result or timeout
	var res *files.ListFolderResult
	select {
	case r := <-ch:
		if r.err != nil {
			if time.Since(startTime) > apiTimeout {
				log.Printf("timeout listing folders after %v", apiTimeout)
				return nil
			}
			log.Printf("Error listing folders: %v", r.err)
			return nil
		}
		res = r.res
		log.Printf("ListFolder API call completed in %v", time.Since(startTime))
	case <-time.After(apiTimeout):
		log.Printf("timeout listing folders after %v", apiTimeout)
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

// GetChangesLast24Hours returns a list of files that have changed in the last 24 hours
func GetChangesLast24Hours() ([]string, error) {
	token := getDropboxAccessToken()
	if token == "" {
		return nil, fmt.Errorf("Dropbox access token not set")
	}

	db, err := sql.Open("sqlite3", "./dropbox.db")
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	c, err := NewDropboxClient(token, db)
	if err != nil {
		return nil, err
	}

	return c.GetChanges(time.Now().Add(-24 * time.Hour))
}

// GetChangesLast10Minutes returns a list of files that have changed in the last 10 minutes
func GetChangesLast10Minutes() ([]string, error) {
	token := getDropboxAccessToken()
	if token == "" {
		return nil, fmt.Errorf("Dropbox access token not set")
	}

	db, err := sql.Open("sqlite3", "./dropbox.db")
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	c, err := NewDropboxClient(token, db)
	if err != nil {
		return nil, err
	}

	return c.GetChanges(time.Now().Add(-10 * time.Minute))
}

// retryOperation executes an operation with retries
func (c *DropboxClient) retryOperation(operation func() error) error {
	maxRetries := 3
	backoff := time.Second

	var err error
	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		if i < maxRetries-1 {
			log.Printf("Operation failed, retrying in %v... (attempt %d/%d): %v", backoff, i+1, maxRetries, err)
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}
	}
	return fmt.Errorf("operation failed after %d attempts: %v", maxRetries, err)
}

// ListFolderLongpoll wraps the Dropbox ListFolderLongpoll API with retries
func (c *DropboxClient) ListFolderLongpoll(arg *files.ListFolderLongpollArg) (*files.ListFolderLongpollResult, error) {
	var result *files.ListFolderLongpollResult
	err := c.retryOperation(func() error {
		var err error
		result, err = c.Client.ListFolderLongpoll(arg)
		return err
	})
	return result, err
}
