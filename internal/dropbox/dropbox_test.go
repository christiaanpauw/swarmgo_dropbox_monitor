package dropbox

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockClock struct {
	now time.Time
}

func newMockClock() *mockClock {
	return &mockClock{now: time.Now()}
}

func (mc *mockClock) Now() time.Time {
	return mc.now
}

func (mc *mockClock) Sleep(d time.Duration) {
	mc.now = mc.now.Add(d)
}

func TestNewDropboxClient(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		wantErr     bool
		wantErrType ErrorType
	}{
		{
			name:        "Valid token",
			token:       "valid-token",
			wantErr:     false,
			wantErrType: ErrorTypeUnknown,
		},
		{
			name:        "Empty token",
			token:       "",
			wantErr:     true,
			wantErrType: ErrorTypeInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewDropboxClient(tt.token)
			if tt.wantErr {
				require.Error(t, err)
				var dbErr *Error
				require.True(t, ErrorAs(err, &dbErr))
				assert.Equal(t, tt.wantErrType, dbErr.Type)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, client)
		})
	}
}

func setupTestServer(t *testing.T, statusCode int, response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify common headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Set response headers based on endpoint
		if r.URL.Path == "/2/files/list_folder" {
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			w.Header().Set("Content-Type", "application/json")
		} else if r.URL.Path == "/2/files/download" {
			assert.NotEmpty(t, r.Header.Get("Dropbox-API-Arg"))
		}

		// Set response status and body
		w.WriteHeader(statusCode)
		_, err := w.Write([]byte(response))
		require.NoError(t, err)
	}))
}

func setupTestClient(t *testing.T, server *httptest.Server, config ClientConfig) *DropboxClient {
	clock := newMockClock()
	client := &DropboxClient{
		accessToken: "test-token",
		httpClient:  server.Client(),
		config:     config,
		circuitBreaker: &circuitBreaker{
			config: config.CircuitBreakerConfig,
			state:  "closed",
			clock:  clock,
		},
		metrics: &clientMetrics{},
	}
	return client
}

func TestDropboxClient_ListFolder(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		statusCode  int
		response    string
		wantErr     bool
		wantErrType ErrorType
	}{
		{
			name:       "Success",
			path:      "/test",
			statusCode: http.StatusOK,
			response: `{
				"entries": [
					{
						".tag": "file",
						"name": "test.txt",
						"path_display": "/test.txt",
						"id": "id:123",
						"client_modified": "2021-01-01T00:00:00Z",
						"server_modified": "2021-01-01T00:00:00Z",
						"rev": "123",
						"size": 100,
						"is_downloadable": true
					}
				]
			}`,
			wantErr: false,
		},
		{
			name:        "Unauthorized",
			path:        "/test",
			statusCode:  http.StatusUnauthorized,
			response:    `{"error": "unauthorized"}`,
			wantErr:     true,
			wantErrType: ErrorTypeAuth,
		},
		{
			name:        "Rate Limited",
			path:        "/test",
			statusCode:  http.StatusTooManyRequests,
			response:    `{"error": "too_many_requests"}`,
			wantErr:     true,
			wantErrType: ErrorTypeRateLimit,
		},
		{
			name:        "Server Error",
			path:        "/test",
			statusCode:  http.StatusInternalServerError,
			response:    `{"error": "internal_server_error"}`,
			wantErr:     true,
			wantErrType: ErrorTypeServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupTestServer(t, tt.statusCode, tt.response)
			defer server.Close()

			config := DefaultClientConfig()
			config.RetryConfig = RetryConfig{
				MaxRetries:  0, // Disable retries for testing
				InitialWait: 1 * time.Millisecond,
				MaxWait:     10 * time.Millisecond,
			}

			client := setupTestClient(t, server, config)

			// Override the API URL for testing
			origURL := listFolderURL
			listFolderURL = server.URL + "/2/files/list_folder"
			defer func() { listFolderURL = origURL }()

			files, err := client.ListFolder(context.Background(), tt.path)
			if tt.wantErr {
				require.Error(t, err)
				var dbErr *Error
				if assert.True(t, ErrorAs(err, &dbErr)) {
					assert.Equal(t, tt.wantErrType, dbErr.Type)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, files)
			if len(files) > 0 {
				assert.Equal(t, "/test.txt", files[0].Path)
			}
		})
	}
}

func TestDropboxClient_GetFileContent(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		statusCode  int
		response    string
		wantErr     bool
		wantErrType ErrorType
	}{
		{
			name:       "Success",
			path:      "/test.txt",
			statusCode: http.StatusOK,
			response:  "file content",
			wantErr:   false,
		},
		{
			name:        "Unauthorized",
			path:        "/test.txt",
			statusCode:  http.StatusUnauthorized,
			response:    "unauthorized",
			wantErr:     true,
			wantErrType: ErrorTypeAuth,
		},
		{
			name:        "Rate Limited",
			path:        "/test.txt",
			statusCode:  http.StatusTooManyRequests,
			response:    "too many requests",
			wantErr:     true,
			wantErrType: ErrorTypeRateLimit,
		},
		{
			name:        "Server Error",
			path:        "/test.txt",
			statusCode:  http.StatusInternalServerError,
			response:    "internal server error",
			wantErr:     true,
			wantErrType: ErrorTypeServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupTestServer(t, tt.statusCode, tt.response)
			defer server.Close()

			config := DefaultClientConfig()
			config.RetryConfig = RetryConfig{
				MaxRetries:  0, // Disable retries for testing
				InitialWait: 1 * time.Millisecond,
				MaxWait:     10 * time.Millisecond,
			}

			client := setupTestClient(t, server, config)

			// Override the API URL for testing
			origURL := downloadURL
			downloadURL = server.URL + "/2/files/download"
			defer func() { downloadURL = origURL }()

			content, err := client.GetFileContent(context.Background(), tt.path)
			if tt.wantErr {
				require.Error(t, err)
				var dbErr *Error
				if assert.True(t, ErrorAs(err, &dbErr)) {
					assert.Equal(t, tt.wantErrType, dbErr.Type)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, content)
			assert.Equal(t, tt.response, string(content))
		})
	}
}

func TestCircuitBreaker(t *testing.T) {
	clock := newMockClock()
	config := CircuitBreakerConfig{
		MaxFailures:      2,
		ResetTimeout:     100 * time.Millisecond,
		HalfOpenMaxTries: 1,
	}

	cb := &circuitBreaker{
		config: config,
		state:  "closed",
		clock:  clock,
	}

	// Test initial state
	assert.False(t, cb.isOpen())

	// Test opening circuit
	cb.recordFailure()
	assert.False(t, cb.isOpen())
	cb.recordFailure()
	assert.True(t, cb.isOpen())

	// Test half-open state after timeout
	clock.Sleep(config.ResetTimeout + time.Millisecond)
	assert.False(t, cb.isOpen())

	// Test successful recovery
	cb.recordSuccess()
	assert.False(t, cb.isOpen())

	// Test failure in half-open state
	cb.recordFailure()
	assert.False(t, cb.isOpen()) // First failure in half-open doesn't immediately open
	cb.recordFailure()          // Second failure should open the circuit
	assert.True(t, cb.isOpen())
}

func TestClientMetrics(t *testing.T) {
	metrics := &clientMetrics{}

	// Test recording metrics
	metrics.recordRequest()
	metrics.recordRetry()
	metrics.recordError(assert.AnError)

	retries, requests, errors := metrics.retryCount, metrics.requestCount, metrics.errorCount
	assert.Equal(t, int64(1), retries)
	assert.Equal(t, int64(1), requests)
	assert.Equal(t, int64(1), errors)
	assert.Equal(t, assert.AnError, metrics.lastError)
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "Rate limit error",
			err:  NewRateLimitError("too many requests", nil),
			want: true,
		},
		{
			name: "Network error",
			err:  NewNetworkError("connection failed", nil),
			want: true,
		},
		{
			name: "Server error",
			err:  NewServerError("internal error", nil),
			want: true,
		},
		{
			name: "Auth error",
			err:  NewAuthError("unauthorized", nil),
			want: false,
		},
		{
			name: "Invalid input error",
			err:  NewInvalidInputError("bad request", nil),
			want: false,
		},
		{
			name: "Non-dropbox error",
			err:  fmt.Errorf("standard error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryable(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}
