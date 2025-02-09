-- File changes tracking
CREATE TABLE IF NOT EXISTS file_changes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL,
    modified_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(path, modified_at)
);

-- File content storage
CREATE TABLE file_contents (
    id SERIAL PRIMARY KEY,
    file_change_id INTEGER REFERENCES file_changes(id),
    content TEXT,
    content_type TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Daily summaries
CREATE TABLE daily_summaries (
    id SERIAL PRIMARY KEY,
    summary_date DATE NOT NULL,
    total_files INTEGER NOT NULL,
    narrative_summary TEXT NOT NULL,
    portfolio_stats JSONB,
    project_stats JSONB,
    author_stats JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_file_changes_modified_at ON file_changes(modified_at);
CREATE INDEX idx_file_changes_portfolio ON file_changes(portfolio);
CREATE INDEX idx_file_changes_project ON file_changes(project);
CREATE INDEX idx_file_changes_author ON file_changes(author);
CREATE INDEX idx_daily_summaries_date ON daily_summaries(summary_date);
