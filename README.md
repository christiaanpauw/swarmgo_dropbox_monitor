# SwarmGo Dropbox Monitor

SwarmGo Dropbox Monitor is a Go application designed to monitor and interact with Dropbox. It establishes a connection to Dropbox and starts a scheduler for periodic tasks.

## Features

- **Dropbox Authentication**: Tests the connection to Dropbox to ensure authentication is successful.
- **Scheduler**: Starts a scheduler to manage periodic tasks.

## Installation

1. **Clone the repository**:
    ```bash
    git clone https://github.com/christiaanpauw/swarmgo_dropbox_monitor.git
    cd swarmgo_dropbox_monitor
    ```

2. **Install dependencies**:
    Ensure you have Go installed. Then, install the required packages:
    ```bash
    go mod tidy
    ```

3. **Using the .env File for Environment Variables**
   
Create a .env file in the root directory of your project with the following content:


``` env
# Dropbox API token
DROPBOX_ACCESS_TOKEN=your_dropbox_api_token

# SMTP Configuration
SMTP_SERVER=smtp.yourserver.com
SMTP_PORT=587
SMTP_USER=your_smtp_user
SMTP_PASS=your_smtp_password
NOTIFY_EMAIL=your_email@example.com
```
4. **Install the godotenv package** if you haven't already:

```bash
go get github.com/joho/godotenv
```

5 **Ensure your Go files are set up** to load the .env file.

This is already done in the provided code.

## Usage

1. **Run the application**:
    ```bash
    go run main.go
    ```

    The application will:
    - Test the connection to Dropbox.
    - Start the scheduler.
    - Keep running to monitor Dropbox.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License.
