# SwarmGo Dropbox Monitor

SwarmGo Dropbox Monitor is a Go application designed to monitor changes in your Dropbox account using an agent-based architecture. It tracks file modifications, additions, and updates, storing the metadata in a local database and sending periodic reports via email.

## Features

- **Agent-Based Architecture**: Uses swarmgo for efficient agent coordination
  - FileChangeAgent: Identifies Dropbox file changes
  - DatabaseAgent: Stores changes in PostgreSQL
  - ContentAnalyzerAgent: Analyzes file contents
  - ReportingAgent: Generates reports
- **Real-time Dropbox Monitoring**: Tracks file changes, modifications, and updates in your Dropbox account
- **Metadata Storage**: Stores file metadata in a SQLite database for efficient querying and tracking
- **Change Detection**: Uses Dropbox's content hash to accurately detect file changes
- **Flexible Reporting**: Generate reports for different time windows:
  - Last 10 minutes (quick check)
  - Last hour
  - Last 24 hours
- **Email Notifications**: Sends formatted email reports about file changes
- **Retry Mechanism**: Implements exponential backoff for reliable API communication
- **Multiple Interfaces**:
  - CLI for command-line operations
  - Web interface for visual monitoring
  - GUI application for desktop usage

## Installation

1. **Clone the repository**:
    ```bash
    git clone https://github.com/christiaanpauw/swarmgo_dropbox_monitor.git
    cd swarmgo_dropbox_monitor
    ```

2. **Install dependencies**:
    ```bash
    go mod tidy
    ```

3. **Configure Environment Variables**:
   Create a `.env` file in the root directory with the following content:
   ```env
   # Dropbox Configuration
   DROPBOX_ACCESS_TOKEN=your_dropbox_api_token

   # SMTP Configuration
   SMTP_SERVER=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your.email@gmail.com
   SMTP_PASSWORD=your_app_password
   FROM_EMAIL=your.email@gmail.com
   TO_EMAILS=recipient1@example.com,recipient2@example.com

   # Scheduler Configuration
   SCHEDULER_INTERVAL=15

   # Logging
   LOG_LEVEL=INFO
   ```

## Usage

### CLI Interface

1. **Quick check (last 10 minutes)**:
   ```bash
   go run cmd/cli/main.go --check-now --quick
   ```

2. **Check last hour**:
   ```bash
   go run cmd/cli/main.go --check-now --last-hour
   ```

3. **Check last 24 hours**:
   ```bash
   go run cmd/cli/main.go --check-now --last-24h
   ```

4. **Run as a service** (checks daily at midnight):
   ```bash
   go run cmd/cli/main.go
   ```

### Web Interface
```bash
go run cmd/web/main.go
```
Access the web interface at `http://localhost:8080`

### GUI Application
```bash
go run cmd/gui/main.go
```

## Email Configuration

The application uses SMTP to send email reports. For Gmail:
1. Enable 2-Factor Authentication
2. Generate an App Password
3. Use the App Password in your `.env` file

## Building from Source

Build all binaries:
```bash
go build ./...
```

Build specific interfaces:
```bash
go build -o dropbox-monitor cmd/cli/main.go    # CLI
go build -o dropbox-web cmd/web/main.go        # Web
go build -o dropbox-gui cmd/gui/main.go        # GUI
```

## Testing

The application includes a comprehensive test suite:

- **Unit Tests**: Test individual components and functions
- **Integration Tests**: Test interactions between components
- **Mock Tests**: Use mock HTTP client for Dropbox API testing
- **Agent Tests**: Test agent coordination and communication

Run the tests:
```bash
go test ./...
```

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for technical details about the codebase organization and implementation details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
