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
