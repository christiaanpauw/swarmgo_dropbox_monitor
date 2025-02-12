package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// Clock interface for better testing
type Clock interface {
	Now() time.Time
	Sleep(d time.Duration)
}

// realClock implements Clock using actual time
type realClock struct{}

func (rc *realClock) Now() time.Time {
	return time.Now()
}

func (rc *realClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

// Default API URLs
var (
	listFolderURL = "https://api.dropboxapi.com/2/files/list_folder"
	downloadURL   = "https://content.dropboxapi.com/2/files/download"
)

// CircuitBreakerConfig holds configuration for the circuit breaker
type CircuitBreakerConfig struct {
	MaxFailures      int           // Number of failures before opening circuit
	ResetTimeout     time.Duration // Time to wait before attempting to reset circuit
	HalfOpenMaxTries int           // Number of requests to allow in half-open state
}

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
}

// ClientConfig holds all client configuration
type ClientConfig struct {
	RetryConfig          RetryConfig
	CircuitBreakerConfig CircuitBreakerConfig
	Transport            *http.Transport
}

// DefaultClientConfig returns a default configuration
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		RetryConfig: RetryConfig{
			MaxRetries:  3,
			InitialWait: 1 * time.Second,
			MaxWait:     30 * time.Second,
		},
		CircuitBreakerConfig: CircuitBreakerConfig{
			MaxFailures:      5,
			ResetTimeout:     1 * time.Minute,
			HalfOpenMaxTries: 2,
		},
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   10,
			MaxConnsPerHost:       100,
		},
	}
}

// circuitBreaker implements the circuit breaker pattern
type circuitBreaker struct {
	config        CircuitBreakerConfig
	state         string // "closed", "open", or "half-open"
	failures      int
	lastFailure   time.Time
	clock         Clock
	mu            sync.Mutex
	halfOpenTries int
}

func newCircuitBreaker(config CircuitBreakerConfig) *circuitBreaker {
	return &circuitBreaker{
		config: config,
		state:  "closed",
		clock:  &realClock{},
	}
}

// For testing purposes
func newCircuitBreakerWithClock(config CircuitBreakerConfig, clock Clock) *circuitBreaker {
	return &circuitBreaker{
		config: config,
		state:  "closed",
		clock:  clock,
	}
}

// isOpen returns true if the circuit breaker is in the open state
func (cb *circuitBreaker) isOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == "open" {
		// Check if enough time has passed to transition to half-open
		if cb.clock.Now().Sub(cb.lastFailure) > cb.config.ResetTimeout {
			cb.state = "half-open"
			cb.failures = 0
			cb.halfOpenTries = 0
			return false
		}
		return true
	}
	return false
}

// recordSuccess records a successful request
func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == "half-open" {
		cb.state = "closed"
		cb.halfOpenTries = 0
	} else {
		cb.failures = 0
	}
}

// recordFailure records a failed request
func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = cb.clock.Now()

	if cb.state == "half-open" {
		cb.halfOpenTries++
		if cb.halfOpenTries >= cb.config.HalfOpenMaxTries {
			cb.state = "open"
		}
	} else if cb.state == "closed" && cb.failures >= cb.config.MaxFailures {
		cb.state = "open"
	}
}

// Client defines the interface for Dropbox operations
type Client interface {
	ListFolder(ctx context.Context, path string) ([]*models.FileMetadata, error)
	GetFileContent(ctx context.Context, path string) ([]byte, error)
	GetChangesLast24Hours(ctx context.Context) ([]*models.FileMetadata, error)
	GetChangesLast10Minutes(ctx context.Context) ([]*models.FileMetadata, error)
	GetChanges(ctx context.Context) ([]*models.FileMetadata, error)
	GetFileChanges(ctx context.Context) ([]models.FileChange, error)
}

// DropboxClient handles interactions with the Dropbox API
type DropboxClient struct {
	accessToken    string
	httpClient     *http.Client
	config         ClientConfig
	circuitBreaker *circuitBreaker
	metrics        *clientMetrics
}

// clientMetrics tracks client operation metrics
type clientMetrics struct {
	retryCount    int64
	requestCount  int64
	errorCount    int64
	lastError     error
	lastErrorTime time.Time
	mu            sync.RWMutex
}

func (m *clientMetrics) recordRetry() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.retryCount++
}

func (m *clientMetrics) recordRequest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestCount++
}

func (m *clientMetrics) recordError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCount++
	m.lastError = err
	m.lastErrorTime = time.Now()
}

// NewDropboxClient creates a new Dropbox client
func NewDropboxClient(token string) (*DropboxClient, error) {
	if token == "" {
		return nil, NewInvalidInputError("token cannot be empty", nil)
	}

	config := DefaultClientConfig()
	return NewDropboxClientWithConfig(token, config)
}

// NewDropboxClientWithConfig creates a new Dropbox client with custom configuration
func NewDropboxClientWithConfig(token string, config ClientConfig) (*DropboxClient, error) {
	if token == "" {
		return nil, NewInvalidInputError("token cannot be empty", nil)
	}

	return &DropboxClient{
		accessToken: token,
		httpClient: &http.Client{
			Transport: config.Transport,
		},
		config:         config,
		circuitBreaker: newCircuitBreaker(config.CircuitBreakerConfig),
		metrics:        &clientMetrics{},
	}, nil
}

// GetMetrics returns current client metrics
func (c *DropboxClient) GetMetrics() (retryCount, requestCount, errorCount int64) {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()
	return c.metrics.retryCount, c.metrics.requestCount, c.metrics.errorCount
}

// doRequestWithRetry performs an HTTP request with retry logic and circuit breaker
func (c *DropboxClient) doRequestWithRetry(req *http.Request) (*http.Response, error) {
	if c.circuitBreaker.isOpen() {
		return nil, NewCircuitOpenError("circuit breaker is open", nil)
	}

	c.metrics.recordRequest()
	var lastErr error
	wait := c.config.RetryConfig.InitialWait

	for attempt := 0; attempt <= c.config.RetryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			c.metrics.recordRetry()
			time.Sleep(wait)
			// Exponential backoff with jitter
			wait = time.Duration(float64(wait) * 1.5)
			if wait > c.config.RetryConfig.MaxWait {
				wait = c.config.RetryConfig.MaxWait
			}
		}

		// Clone the request to avoid reusing the same request multiple times
		reqClone := req.Clone(req.Context())
		resp, err := c.httpClient.Do(reqClone)
		if err != nil {
			lastErr = NewNetworkError(fmt.Sprintf("attempt %d: request failed", attempt+1), err)
			c.metrics.recordError(lastErr)
			c.circuitBreaker.recordFailure()
			continue
		}

		// Handle response based on status code
		switch {
		case resp.StatusCode == http.StatusOK:
			c.circuitBreaker.recordSuccess()
			return resp, nil
		case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
			resp.Body.Close()
			err := NewAuthError(fmt.Sprintf("authentication failed: status %d", resp.StatusCode), nil)
			c.metrics.recordError(err)
			return nil, err
		case resp.StatusCode == http.StatusTooManyRequests:
			resp.Body.Close()
			lastErr = NewRateLimitError(fmt.Sprintf("rate limited on attempt %d", attempt+1), nil)
			c.metrics.recordError(lastErr)
			c.circuitBreaker.recordFailure()
			if attempt == c.config.RetryConfig.MaxRetries {
				return nil, lastErr
			}
			continue
		case resp.StatusCode >= 500:
			resp.Body.Close()
			lastErr = NewServerError(fmt.Sprintf("server error on attempt %d: status %d", attempt+1, resp.StatusCode), nil)
			c.metrics.recordError(lastErr)
			c.circuitBreaker.recordFailure()
			if attempt == c.config.RetryConfig.MaxRetries {
				return nil, lastErr
			}
			continue
		default:
			resp.Body.Close()
			err := NewError(ErrorTypeUnknown, fmt.Sprintf("unexpected status: %d", resp.StatusCode), nil)
			c.metrics.recordError(err)
			return nil, err
		}
	}

	return nil, lastErr
}

// dropboxFileMetadata represents the raw metadata from Dropbox API
type dropboxFileMetadata struct {
	Tag            string `json:".tag"`
	Name           string `json:"name"`
	PathLower      string `json:"path_lower"`
	PathDisplay    string `json:"path_display"`
	ID             string `json:"id"`
	ClientModified string `json:"client_modified"`
	ServerModified string `json:"server_modified"`
	Rev            string `json:"rev"`
	Size           int64  `json:"size"`
	IsDownloadable bool   `json:"is_downloadable"`
	ContentHash    string `json:"content_hash"`
	SharingInfo    struct {
		ReadOnly             bool        `json:"read_only"`
		ParentSharedFolderID string      `json:"parent_shared_folder_id"`
		ModifiedBy           interface{} `json:"modified_by"`
	} `json:"sharing_info"`
}

// toFileMetadata converts Dropbox API metadata to our consolidated type
func (c *DropboxClient) toFileMetadata(dbx *dropboxFileMetadata) (*models.FileMetadata, error) {
	if dbx == nil {
		return nil, NewInvalidInputError("nil dropbox metadata", nil)
	}

	modTime, err := time.Parse(time.RFC3339, dbx.ServerModified)
	if err != nil {
		return nil, NewInvalidInputError("invalid server modified time", err)
	}

	return &models.FileMetadata{
		Path:     dbx.PathDisplay,
		Name:     dbx.Name,
		Size:     dbx.Size,
		Modified: modTime,
	}, nil
}

// ListFolder lists files in a Dropbox folder
func (c *DropboxClient) ListFolder(ctx context.Context, path string) ([]*models.FileMetadata, error) {
	if path == "" {
		return nil, NewInvalidInputError("path cannot be empty", nil)
	}

	body := map[string]interface{}{
		"path": path,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, NewInvalidInputError(fmt.Sprintf("failed to marshal request body for path %s", path), err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", listFolderURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, NewInvalidInputError(fmt.Sprintf("failed to create request for path %s", path), err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err // Already wrapped by doRequestWithRetry with proper context
	}
	defer resp.Body.Close()

	var result struct {
		Entries []dropboxFileMetadata `json:"entries"`
		HasMore bool                  `json:"has_more"`
		Cursor  string                `json:"cursor"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, NewServerError(fmt.Sprintf("failed to decode response for path %s", path), err)
	}

	files := make([]*models.FileMetadata, 0, len(result.Entries))
	for i := range result.Entries {
		file, err := c.toFileMetadata(&result.Entries[i])
		if err != nil {
			return nil, NewServerError(fmt.Sprintf("failed to convert metadata for file %s in path %s", result.Entries[i].Name, path), err)
		}
		files = append(files, file)
	}

	return files, nil
}

// GetFileContent downloads a file's content from Dropbox
func (c *DropboxClient) GetFileContent(ctx context.Context, path string) ([]byte, error) {
	if path == "" {
		return nil, NewInvalidInputError("path cannot be empty", nil)
	}

	body := map[string]interface{}{
		"path": path,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, NewInvalidInputError(fmt.Sprintf("failed to marshal request body for path %s", path), err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", downloadURL, nil)
	if err != nil {
		return nil, NewInvalidInputError(fmt.Sprintf("failed to create request for path %s", path), err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Dropbox-API-Arg", string(jsonBody))

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err // Already wrapped by doRequestWithRetry with proper context
	}
	defer resp.Body.Close()

	// Check if the file is too large (>100MB) to prevent memory issues
	if resp.ContentLength > 100*1024*1024 {
		return nil, NewFileSizeLimitError(fmt.Sprintf("file %s exceeds maximum size of 100MB (size: %d bytes)", path, resp.ContentLength), nil)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewNetworkError(fmt.Sprintf("failed to read response body for path %s", path), err)
	}

	return content, nil
}

// GetChangesLast24Hours returns changes from the last 24 hours
func (c *DropboxClient) GetChangesLast24Hours(ctx context.Context) ([]*models.FileMetadata, error) {
	return c.ListFolder(ctx, "")
}

// GetChangesLast10Minutes returns changes from the last 10 minutes
func (c *DropboxClient) GetChangesLast10Minutes(ctx context.Context) ([]*models.FileMetadata, error) {
	return c.ListFolder(ctx, "")
}

// GetChanges returns all changes
func (c *DropboxClient) GetChanges(ctx context.Context) ([]*models.FileMetadata, error) {
	return c.ListFolder(ctx, "")
}

// GetFileChanges retrieves file changes from Dropbox
func (c *DropboxClient) GetFileChanges(ctx context.Context) ([]models.FileChange, error) {
	changes, err := c.GetChanges(ctx)
	if err != nil {
		return nil, err
	}
	return models.BatchConvertMetadataToChanges(changes), nil
}
