# Use the official Go image as a build stage
FROM golang:1.23 AS builder

# Set the working directory
WORKDIR /app

# Copy the entire project first (before downloading dependencies)
COPY . . 

# Download dependencies
RUN go mod tidy

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dropbox-monitor main.go

# Use a minimal image for runtime
FROM alpine:latest

# Install ca-certificates (required for Dropbox API)
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/dropbox-monitor .

# Expose no ports (runs as a background process)
CMD ["./dropbox-monitor"]

