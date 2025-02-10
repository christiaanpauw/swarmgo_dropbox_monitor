-- File changes tracking
CREATE TABLE IF NOT EXISTS file_changes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL,
    modified_at DATETIME NOT NULL,
    file_type TEXT,
    dropbox_id TEXT,
    dropbox_rev TEXT,
    client_modified DATETIME,
    server_modified DATETIME,
    size INTEGER,
    is_downloadable BOOLEAN,
    modified_by_id TEXT,
    modified_by_name TEXT,
    content_hash TEXT,
    embedding TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- File content storage
CREATE TABLE file_contents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_change_id INTEGER,
    content TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_change_id) REFERENCES file_changes(id)
);

-- Daily summaries
CREATE TABLE daily_summaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date DATE NOT NULL,
    total_files INTEGER,
    total_changes INTEGER,
    total_size BIGINT,
    most_active_user TEXT,
    most_modified_file TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Sync state
CREATE TABLE IF NOT EXISTS sync_state (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cursor TEXT NOT NULL,
    last_sync DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_file_changes_modified_at ON file_changes(modified_at);
CREATE INDEX idx_file_changes_dropbox_id ON file_changes(dropbox_id);
CREATE INDEX idx_file_changes_modified_by_id ON file_changes(modified_by_id);
CREATE INDEX idx_daily_summaries_date ON daily_summaries(date);
