package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DBType int

const (
	SQLite DBType = iota
)

type DB struct {
	DB     *sql.DB // Expose the underlying connection
	DBType DBType
}

type Vector []float32

func (v Vector) Value() (interface{}, error) {
	return json.Marshal(v)
}

func (v *Vector) Scan(value interface{}) error {
	switch value := value.(type) {
	case []byte:
		return json.Unmarshal(value, v)
	case string:
		return json.Unmarshal([]byte(value), v)
	case nil:
		*v = nil
		return nil
	default:
		return fmt.Errorf("unsupported type for Vector: %T", value)
	}
}

func NewDB(connStr string) (*DB, error) {
	log.Println("Starting database initialization...")
	return initSQLiteDB(connStr)
}

func initSQLiteDB(connStr string) (*DB, error) {
	log.Println("Initializing SQLite database...")
	
	// Extract database path from connection string
	dbPath := connStr
	if len(dbPath) > 5 && dbPath[:5] == "file:" {
		dbPath = dbPath[5:]
	}

	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("error creating data directory: %v", err)
	}

	// Try to remove any existing WAL files that might be corrupted
	walPath := dbPath + "-wal"
	shmPath := dbPath + "-shm"
	os.Remove(walPath)
	os.Remove(shmPath)

	// Open database with WAL journal mode and normal synchronous mode for better performance
	connStr = fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL", connStr)
	conn, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening SQLite database: %v", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error connecting to SQLite database: %v", err)
	}

	// Initialize schema
	if err := initSQLiteSchema(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error initializing SQLite schema: %v", err)
	}

	log.Printf("Successfully initialized SQLite database at: %s", dbPath)
	return &DB{DB: conn, DBType: SQLite}, nil
}

func initSQLiteSchema(conn *sql.DB) error {
	// Start a transaction for table creation
	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback() // Rollback if we return with error

	// First create tables
	tables := []string{
		`CREATE TABLE IF NOT EXISTS file_changes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_path TEXT NOT NULL,
			modified_at DATETIME NOT NULL,
			file_type TEXT,
			portfolio TEXT,
			project TEXT,
			document_type TEXT,
			author TEXT,
			content_hash TEXT,
			embedding TEXT,
			dropbox_id TEXT,
			dropbox_rev TEXT,
			client_modified DATETIME,
			server_modified DATETIME,
			size INTEGER,
			is_downloadable BOOLEAN,
			modified_by_id TEXT,
			modified_by_name TEXT,
			shared_folder_id TEXT,
			lock_holder_name TEXT,
			lock_holder_id TEXT,
			lock_created_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS file_contents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_change_id INTEGER NOT NULL,
			content TEXT,
			content_type TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (file_change_id) REFERENCES file_changes(id)
		)`,
		`CREATE TABLE IF NOT EXISTS daily_summaries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			summary_date DATE NOT NULL,
			total_files INTEGER NOT NULL,
			summary TEXT,
			portfolio_stats TEXT,
			project_stats TEXT,
			author_stats TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sync_state (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cursor TEXT NOT NULL,
			last_sync DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	// Execute table creation queries
	for _, query := range tables {
		if _, err := tx.Exec(query); err != nil {
			return fmt.Errorf("error executing query %q: %v", query, err)
		}
	}

	// Commit the transaction to ensure tables are created
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	// Verify that the tables exist before creating indexes
	var exists int
	err = conn.QueryRow("SELECT 1 FROM sqlite_master WHERE type='table' AND name='file_changes'").Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking if table exists: %v", err)
	}

	if exists != 1 {
		return fmt.Errorf("file_changes table was not created successfully")
	}

	// Start a new transaction for indexes
	tx, err = conn.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction for indexes: %v", err)
	}
	defer tx.Rollback()

	// Create indexes in a separate transaction
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_file_changes_file_path ON file_changes(file_path)`,
		`CREATE INDEX IF NOT EXISTS idx_file_changes_modified_at ON file_changes(modified_at)`,
		`CREATE INDEX IF NOT EXISTS idx_file_changes_content_hash ON file_changes(content_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_file_changes_dropbox_id ON file_changes(dropbox_id)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_summaries_date ON daily_summaries(summary_date)`,
	}

	// Execute index creation queries
	for _, query := range indexes {
		if _, err := tx.Exec(query); err != nil {
			return fmt.Errorf("error executing query %q: %v", query, err)
		}
	}

	// Commit the index transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing index transaction: %v", err)
	}

	return nil
}

func (db *DB) SaveFileChange(ctx context.Context, fc *FileChange) error {
	// Check if file with same path and content hash already exists
	existing, err := db.GetExistingFileChange(ctx, fc.FilePath, fc.ContentHash)
	if err != nil {
		return fmt.Errorf("error checking for existing file: %v", err)
	}
	if existing != nil {
		// File already exists with same content hash, no need to save
		fc.ID = existing.ID // Set the ID so it can be used for file content
		return nil
	}

	// Convert embedding to JSON for SQLite storage
	embeddingJSON, err := json.Marshal(fc.Embedding)
	if err != nil {
		return fmt.Errorf("error marshaling embedding: %v", err)
	}

	query := `
		INSERT INTO file_changes (
			file_path, modified_at, file_type, portfolio, project, document_type, 
			author, content_hash, embedding, dropbox_id, dropbox_rev, client_modified, 
			server_modified, size, is_downloadable, modified_by_id, modified_by_name, 
			shared_folder_id, lock_holder_name, lock_holder_id, lock_created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at`

	err = db.DB.QueryRowContext(ctx, query,
		fc.FilePath,
		fc.ModifiedAt,
		fc.FileType,
		fc.Portfolio,
		fc.Project,
		fc.DocumentType,
		fc.Author,
		fc.ContentHash,
		string(embeddingJSON),
		fc.DropboxID,
		fc.DropboxRev,
		fc.ClientModified,
		fc.ServerModified,
		fc.Size,
		fc.IsDownloadable,
		fc.ModifiedByID,
		fc.ModifiedByName,
		fc.SharedFolderID,
		fc.LockHolderName,
		fc.LockHolderID,
		fc.LockCreatedAt,
	).Scan(&fc.ID, &fc.CreatedAt)

	if err != nil {
		return fmt.Errorf("error saving file change: %v", err)
	}

	return nil
}

func (db *DB) GetExistingFileChange(ctx context.Context, filePath string, contentHash string) (*FileChange, error) {
	query := `
		SELECT 
			id, file_path, modified_at, file_type, portfolio, project, 
			document_type, author, content_hash, embedding, dropbox_id, 
			dropbox_rev, client_modified, server_modified, size, 
			is_downloadable, modified_by_id, modified_by_name, 
			shared_folder_id, lock_holder_name, lock_holder_id, 
			lock_created_at, created_at
		FROM file_changes
		WHERE file_path = ? AND content_hash = ?
		ORDER BY modified_at DESC
		LIMIT 1`

	var fc FileChange
	var embeddingJSON string
	var clientModified, serverModified, lockCreatedAt sql.NullTime
	err := db.DB.QueryRowContext(ctx, query, filePath, contentHash).Scan(
		&fc.ID,
		&fc.FilePath,
		&fc.ModifiedAt,
		&fc.FileType,
		&fc.Portfolio,
		&fc.Project,
		&fc.DocumentType,
		&fc.Author,
		&fc.ContentHash,
		&embeddingJSON,
		&fc.DropboxID,
		&fc.DropboxRev,
		&clientModified,
		&serverModified,
		&fc.Size,
		&fc.IsDownloadable,
		&fc.ModifiedByID,
		&fc.ModifiedByName,
		&fc.SharedFolderID,
		&fc.LockHolderName,
		&fc.LockHolderID,
		&lockCreatedAt,
		&fc.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error querying file change: %v", err)
	}

	// Parse embedding JSON if present
	if embeddingJSON != "" {
		if err := json.Unmarshal([]byte(embeddingJSON), &fc.Embedding); err != nil {
			return nil, fmt.Errorf("error unmarshaling embedding: %v", err)
		}
	}

	if clientModified.Valid {
		fc.ClientModified = clientModified.Time
	}
	if serverModified.Valid {
		fc.ServerModified = serverModified.Time
	}
	if lockCreatedAt.Valid {
		fc.LockCreatedAt = lockCreatedAt.Time
	}

	return &fc, nil
}

func (db *DB) SaveFileContent(ctx context.Context, fc *FileContent) error {
	// Check if content already exists for this file change
	var exists bool
	err := db.DB.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM file_contents 
			WHERE file_change_id = ?
		)`, fc.FileChangeID).Scan(&exists)
	
	if err != nil {
		return fmt.Errorf("error checking existing content: %v", err)
	}

	if exists {
		// Content already exists for this file change
		return nil
	}

	query := `
		INSERT INTO file_contents (file_change_id, content, content_type)
		VALUES (?, ?, ?)
		RETURNING id, created_at`

	err = db.DB.QueryRowContext(ctx, query,
		fc.FileChangeID,
		fc.Content,
		fc.ContentType,
	).Scan(&fc.ID, &fc.CreatedAt)

	if err != nil {
		return fmt.Errorf("error saving file content: %v", err)
	}

	return nil
}

func (db *DB) SaveDailySummary(ctx context.Context, ds *DailySummary) error {
	portfolioStats, err := json.Marshal(ds.PortfolioStats)
	if err != nil {
		return err
	}

	projectStats, err := json.Marshal(ds.ProjectStats)
	if err != nil {
		return err
	}

	authorStats, err := json.Marshal(ds.AuthorStats)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO daily_summaries (
			summary_date, total_files, summary,
			portfolio_stats, project_stats, author_stats
		) VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, created_at`

	err = db.DB.QueryRowContext(ctx, query,
		ds.SummaryDate,
		ds.TotalFiles,
		ds.Summary,
		string(portfolioStats),
		string(projectStats),
		string(authorStats),
	).Scan(&ds.ID, &ds.CreatedAt)

	if err != nil {
		return fmt.Errorf("error saving daily summary: %v", err)
	}

	return nil
}

func (db *DB) GetRecentFileChanges(ctx context.Context, since time.Time) ([]FileChange, error) {
	query := `
		SELECT 
			id, file_path, modified_at, file_type, portfolio, project, 
			document_type, author, content_hash, embedding, dropbox_id, 
			dropbox_rev, client_modified, server_modified, size, 
			is_downloadable, modified_by_id, modified_by_name, 
			shared_folder_id, lock_holder_name, lock_holder_id, 
			lock_created_at, created_at
		FROM file_changes
		WHERE modified_at > ?
		ORDER BY modified_at DESC`

	rows, err := db.DB.QueryContext(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("error querying file changes: %v", err)
	}
	defer rows.Close()

	var files []FileChange
	for rows.Next() {
		var fc FileChange
		var embeddingJSON string
		var clientModified, serverModified, lockCreatedAt sql.NullTime
		err := rows.Scan(
			&fc.ID,
			&fc.FilePath,
			&fc.ModifiedAt,
			&fc.FileType,
			&fc.Portfolio,
			&fc.Project,
			&fc.DocumentType,
			&fc.Author,
			&fc.ContentHash,
			&embeddingJSON,
			&fc.DropboxID,
			&fc.DropboxRev,
			&clientModified,
			&serverModified,
			&fc.Size,
			&fc.IsDownloadable,
			&fc.ModifiedByID,
			&fc.ModifiedByName,
			&fc.SharedFolderID,
			&fc.LockHolderName,
			&fc.LockHolderID,
			&lockCreatedAt,
			&fc.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning file change: %v", err)
		}

		// Parse embedding JSON if present
		if embeddingJSON != "" {
			if err := json.Unmarshal([]byte(embeddingJSON), &fc.Embedding); err != nil {
				return nil, fmt.Errorf("error unmarshaling embedding: %v", err)
			}
		}

		if clientModified.Valid {
			fc.ClientModified = clientModified.Time
		}
		if serverModified.Valid {
			fc.ServerModified = serverModified.Time
		}
		if lockCreatedAt.Valid {
			fc.LockCreatedAt = lockCreatedAt.Time
		}

		files = append(files, fc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return files, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

type FileChange struct {
	ID              int64     `json:"id"`
	FilePath        string    `json:"file_path"`
	ModifiedAt      time.Time `json:"modified_at"`
	FileType        string    `json:"file_type"`
	Portfolio       string    `json:"portfolio"`
	Project         string    `json:"project"`
	DocumentType    string    `json:"document_type"`
	Author          string    `json:"author"`
	ContentHash     string    `json:"content_hash"`
	Embedding       Vector    `json:"embedding"`
	DropboxID       string    `json:"dropbox_id"`
	DropboxRev      string    `json:"dropbox_rev"`
	ClientModified  time.Time `json:"client_modified"`
	ServerModified  time.Time `json:"server_modified"`
	Size            int64     `json:"size"`
	IsDownloadable  bool      `json:"is_downloadable"`
	ModifiedByID    string    `json:"modified_by_id"`
	ModifiedByName  string    `json:"modified_by_name"`
	SharedFolderID  string    `json:"shared_folder_id"`
	LockHolderName  string    `json:"lock_holder_name"`
	LockHolderID    string    `json:"lock_holder_id"`
	LockCreatedAt   time.Time `json:"lock_created_at"`
	CreatedAt       time.Time `json:"created_at"`
}

type FileContent struct {
	ID           int64
	FileChangeID int64
	Content      string
	ContentType  string
	CreatedAt    time.Time
}

type DailySummary struct {
	ID             int64
	SummaryDate    time.Time
	TotalFiles     int
	Summary        string
	PortfolioStats map[string]interface{}
	ProjectStats   map[string]interface{}
	AuthorStats    map[string]interface{}
	CreatedAt      time.Time
}
