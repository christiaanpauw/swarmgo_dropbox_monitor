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
	dbType DBType
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
	return initSQLiteDB()
}

func initSQLiteDB() (*DB, error) {
	log.Println("Initializing SQLite database...")
	
	// Create data directory if it doesn't exist
	dataDir := filepath.Join(".", "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating data directory: %v", err)
	}

	// Open SQLite database
	dbPath := filepath.Join(dataDir, "dropbox_monitor.db")
	log.Printf("Opening SQLite database at: %s", dbPath)

	// Try to remove any existing WAL files that might be corrupted
	walPath := dbPath + "-wal"
	shmPath := dbPath + "-shm"
	os.Remove(walPath)
	os.Remove(shmPath)

	// Open database with WAL journal mode and normal synchronous mode for better performance
	connStr := fmt.Sprintf("file:%s?_journal_mode=WAL&_synchronous=NORMAL", dbPath)
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
	return &DB{DB: conn, dbType: SQLite}, nil
}

func initSQLiteSchema(conn *sql.DB) error {
	queries := []string{
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

	for _, query := range queries {
		if _, err := conn.Exec(query); err != nil {
			return fmt.Errorf("error executing schema query: %v", err)
		}
	}

	return nil
}

func (db *DB) SaveFileChange(ctx context.Context, fc *FileChange) error {
	// Convert embedding to JSON for SQLite storage
	embeddingJSON, err := json.Marshal(fc.Embedding)
	if err != nil {
		return fmt.Errorf("error marshaling embedding: %v", err)
	}

	query := `
		INSERT INTO file_changes (file_path, modified_at, file_type, portfolio, project, document_type, author, content_hash, embedding)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at`

	_, err = db.DB.ExecContext(ctx, query,
		fc.FilePath,
		fc.ModifiedAt,
		fc.FileType,
		fc.Portfolio,
		fc.Project,
		fc.DocumentType,
		fc.Author,
		fc.ContentHash,
		string(embeddingJSON),
	)
	return err
}

func (db *DB) SaveFileContent(ctx context.Context, fc *FileContent) error {
	query := `
		INSERT INTO file_contents (file_change_id, content, content_type)
		VALUES (?, ?, ?)
		RETURNING id, created_at`

	_, err := db.DB.ExecContext(ctx, query,
		fc.FileChangeID,
		fc.Content,
		fc.ContentType,
	)
	return err
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
			summary_date, total_files, narrative_summary,
			portfolio_stats, project_stats, author_stats
		) VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, created_at`

	_, err = db.DB.ExecContext(ctx, query,
		ds.SummaryDate,
		ds.TotalFiles,
		ds.Summary,
		string(portfolioStats),
		string(projectStats),
		string(authorStats),
	)
	return err
}

func (db *DB) GetRecentFileChanges(ctx context.Context, since time.Time) ([]FileChange, error) {
	query := `
		SELECT id, file_path, modified_at, file_type, portfolio, project,
			   document_type, author, content_hash, embedding, created_at
		FROM file_changes
		WHERE modified_at >= ?
		ORDER BY modified_at DESC`

	rows, err := db.DB.QueryContext(ctx, query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []FileChange
	for rows.Next() {
		var fc FileChange
		var embeddingJSON string
		if err := rows.Scan(
			&fc.ID, &fc.FilePath, &fc.ModifiedAt, &fc.FileType, &fc.Portfolio,
			&fc.Project, &fc.DocumentType, &fc.Author, &fc.ContentHash,
			&embeddingJSON, &fc.CreatedAt,
		); err != nil {
			return nil, err
		}
		var embedding Vector
		if err := json.Unmarshal([]byte(embeddingJSON), &embedding); err != nil {
			return nil, err
		}
		fc.Embedding = embedding
		changes = append(changes, fc)
	}
	return changes, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

type FileChange struct {
	ID           int64
	FilePath     string
	ModifiedAt   time.Time
	FileType     string
	Portfolio    string
	Project      string
	DocumentType string
	Author       string
	ContentHash  string
	Embedding    Vector
	CreatedAt    time.Time
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
