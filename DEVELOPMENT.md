# SwarmGo Dropbox Monitor - Technical Documentation

This document provides technical details about the SwarmGo Dropbox Monitor application's architecture, code organization, and implementation details.

## Code Organization

The codebase is organized into several key directories:

```
swarmgo_dropbox_monitor/
├── cmd/                    # Command-line entry points
│   ├── cli/               # CLI interface
│   ├── web/               # Web interface
│   └── gui/               # GUI interface
├── internal/              # Internal packages
│   ├── agents/            # Agent implementations
│   │   ├── content_analyzer.go
│   │   ├── database.go
│   │   ├── file_change.go
│   │   ├── reporting.go
│   │   └── agent_manager.go
│   ├── core/              # Core application logic
│   ├── dropbox/           # Dropbox API integration
│   ├── interfaces/        # Common interfaces for components
│   │   ├── dropbox.go     # Dropbox client interface
│   │   └── state.go       # State management interface
│   ├── lifecycle/         # Component lifecycle management
│   ├── container/         # Dependency injection container
│   ├── models/            # Data models and types
│   └── scheduler/         # Scheduling logic
├── templates/             # HTML templates for web interface
└── data/                  # SQLite database and application data
```

## Core Components

### 1. Interfaces (`internal/interfaces/`)

The application uses clearly defined interfaces to ensure loose coupling between components:

#### DropboxClient Interface
```go
type DropboxClient interface {
    ListFolder(ctx context.Context, path string) ([]*models.FileMetadata, error)
    GetFileContent(ctx context.Context, path string) ([]byte, error)
    GetChanges(ctx context.Context) ([]*models.FileMetadata, error)
    GetChangesLast24Hours(ctx context.Context) ([]*models.FileMetadata, error)
    GetChangesLast10Minutes(ctx context.Context) ([]*models.FileMetadata, error)
}
```

### 2. Agent System (`internal/agents/`)

The application uses an agent-based architecture powered by github.com/prathyushnallamothu/swarmgo:

#### FileChangeAgent
```go
type FileChangeAgent struct {
    dropboxClient *dropbox.DropboxClient
    timeWindow   time.Duration
}

// Identifies and tracks file changes in Dropbox
func (a *FileChangeAgent) Execute(ctx context.Context) error {
    // Get changes from Dropbox
    // Pass changes to other agents
}
```

#### DatabaseAgent
```go
type DatabaseAgent struct {
    db *sql.DB
}

// Stores file changes in the database
func (a *DatabaseAgent) Execute(ctx context.Context) error {
    // Store changes in database
    // Update sync state
}
```

#### ContentAnalyzerAgent
```go
type ContentAnalyzerAgent struct {
    dropboxClient *dropbox.DropboxClient
    db           *sql.DB
}

// Analyzes file contents for insights
func (a *ContentAnalyzerAgent) Execute(ctx context.Context) error {
    // Download and analyze file contents
    // Extract keywords and topics
}
```

#### ReportingAgent
```go
type ReportingAgent struct {
    db *sql.DB
}

// Generates reports based on stored data
func (a *ReportingAgent) Execute(ctx context.Context) error {
    // Generate reports
    // Send email notifications
}
```

### 3. Dropbox Client (`internal/dropbox/`)

The Dropbox client handles all interactions with the Dropbox API:

- **FileMetadata**: Struct representing Dropbox file metadata
- **DropboxClient**: Main client for Dropbox API interactions
- **Key Features**:
  - Content hash tracking for change detection
  - Retry mechanism with exponential backoff
  - Efficient file and folder listing
  - Change detection based on timestamps

### 4. Database Layer (`internal/db/`)

SQLite-based storage for file metadata and changes:

- **Schema**:
  - `file_changes`: Tracks file modifications
  - `file_contents`: Stores file content (if needed)
  - `daily_summaries`: Aggregated daily statistics
  - `sync_state`: Cursor-based synchronization state

### 5. Report Generation (`internal/report/`)

Handles the creation and formatting of change reports:

- Supports multiple report formats
- Email report generation
- Statistical summaries
- Customizable templates

### 6. Scheduler (`internal/scheduler/`)

Manages periodic tasks and monitoring:

- Configurable check intervals
- Daily report generation
- Email notification system
- Error handling and retry logic

## Key Features Implementation

### 1. Change Detection

The application uses multiple methods to detect changes:

```go
func (c *DropboxClient) GetChanges(since time.Time) ([]FileMetadata, error) {
    // Lists files and compares:
    // 1. Content hash
    // 2. Modification time
    // 3. File size
    // 4. Revision number
}
```

### 2. Error Handling

Robust error handling with retry mechanism:

```go
type RetryConfig struct {
    MaxRetries  int
    InitialWait time.Duration
    MaxWait     time.Duration
}

func (c *DropboxClient) doRequestWithRetry(req *http.Request) (*http.Response, error) {
    // Implements exponential backoff
    // Handles rate limits
    // Retries on network errors
}
```

### 3. Database Operations

Efficient database operations with prepared statements:

```sql
-- File changes tracking
CREATE TABLE IF NOT EXISTS file_changes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL,
    modified_at DATETIME NOT NULL,
    content_hash TEXT,
    -- Additional metadata fields
);
```

## Testing

### 1. Unit Tests

Each component has comprehensive unit tests:

```go
func TestDropboxClient_GetChanges(t *testing.T) {
    // Test change detection
}

func TestContentAnalyzer_ExtractKeywords(t *testing.T) {
    // Test content analysis
}

func TestReportingAgent_GenerateReport(t *testing.T) {
    // Test report generation
}
```

### 2. Mock Testing

The application uses mock objects for external dependencies:

```go
type mockHTTPClient struct {
    files []FileMetadata
}

func (m *mockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
    // Simulate Dropbox API responses
}
```

### 3. Integration Tests

Tests that verify component interactions:

```go
func TestAgentWorkflow(t *testing.T) {
    // Test complete workflow:
    // 1. FileChangeAgent detects changes
    // 2. DatabaseAgent stores them
    // 3. ContentAnalyzerAgent analyzes
    // 4. ReportingAgent generates report
}
```

### 4. Test Coverage

Ensure good test coverage:
```bash
go test -cover ./...
```

## Interface Implementation

### 1. CLI Interface

The CLI interface (`cmd/cli/`) provides command-line functionality:

- Flags for different time windows
- Interactive and non-interactive modes
- Service mode for continuous monitoring

### 2. Web Interface

The web interface (`cmd/web/`) offers a browser-based dashboard:

- Real-time updates
- File browsing
- Change history visualization
- Configuration management

### 3. GUI Interface

The GUI application (`cmd/gui/`) provides a desktop experience:

- System tray integration
- Native file system integration
- Real-time notifications
- Configuration management

## Development Guidelines

### 1. Code Style

- Follow Go standard formatting (`go fmt`)
- Use meaningful variable and function names
- Document public functions and types
- Keep functions focused and small

### 2. Testing

- Write unit tests for core functionality
- Use table-driven tests where appropriate
- Mock external services (Dropbox API, SMTP)
- Test error conditions and edge cases

### 3. Error Handling

- Use descriptive error messages
- Implement proper error wrapping
- Log errors with appropriate context
- Handle API rate limits gracefully

### 4. Configuration

- Use environment variables for configuration
- Support multiple configuration sources
- Validate configuration at startup
- Document all configuration options

## Adding New Features

1. **Plan the Feature**:
   - Define the feature scope
   - Consider backward compatibility
   - Plan database migrations if needed

2. **Implementation**:
   - Create new package if necessary
   - Follow existing patterns
   - Add appropriate tests
   - Update documentation

3. **Testing**:
   - Run existing test suite
   - Add new test cases
   - Test edge cases
   - Verify performance impact

4. **Documentation**:
   - Update README.md
   - Update DEVELOPMENT.md
   - Add inline documentation
   - Update API documentation

## Troubleshooting

Common issues and solutions:

1. **Dropbox API Rate Limits**:
   - Use exponential backoff
   - Cache responses where possible
   - Monitor API usage

2. **Database Performance**:
   - Use indexes for frequent queries
   - Regular maintenance
   - Monitor growth

3. **Memory Usage**:
   - Stream large files
   - Clean up resources
   - Profile memory usage

4. **Agent Communication**:
   - Monitor agent health
   - Handle timeouts
   - Implement circuit breakers
