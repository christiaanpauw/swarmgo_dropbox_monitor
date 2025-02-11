package dropbox

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType represents the type of error that occurred
type ErrorType string

const (
	// ErrorTypeUnknown represents an unknown error
	ErrorTypeUnknown ErrorType = "unknown"
	// ErrorTypeAuth represents an authentication error
	ErrorTypeAuth ErrorType = "auth"
	// ErrorTypeRateLimit represents a rate limit error
	ErrorTypeRateLimit ErrorType = "rate_limit"
	// ErrorTypeNetwork represents a network error
	ErrorTypeNetwork ErrorType = "network"
	// ErrorTypeServer represents a server error
	ErrorTypeServer ErrorType = "server"
	// ErrorTypeInvalidInput represents an invalid input error
	ErrorTypeInvalidInput ErrorType = "invalid_input"
	// ErrorTypeCircuitOpen represents a circuit breaker open error
	ErrorTypeCircuitOpen ErrorType = "circuit_open"
)

// Error represents a Dropbox API error
type Error struct {
	Type    ErrorType
	Message string
	Cause   error
}

// Error returns the error message
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Cause
}

// NewError creates a new Error with the given type and message
func NewError(typ ErrorType, msg string, cause error) *Error {
	return &Error{
		Type:    typ,
		Message: msg,
		Cause:   cause,
	}
}

// NewAuthError creates a new authentication error
func NewAuthError(msg string, cause error) *Error {
	return NewError(ErrorTypeAuth, msg, cause)
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(msg string, cause error) *Error {
	return NewError(ErrorTypeRateLimit, msg, cause)
}

// NewNetworkError creates a new network error
func NewNetworkError(msg string, cause error) *Error {
	return NewError(ErrorTypeNetwork, msg, cause)
}

// NewServerError creates a new server error
func NewServerError(msg string, cause error) *Error {
	return NewError(ErrorTypeServer, msg, cause)
}

// NewInvalidInputError creates a new invalid input error
func NewInvalidInputError(msg string, cause error) *Error {
	return NewError(ErrorTypeInvalidInput, msg, cause)
}

// NewCircuitOpenError creates a new circuit breaker open error
func NewCircuitOpenError(msg string, cause error) *Error {
	return NewError(ErrorTypeCircuitOpen, msg, cause)
}

// ErrorAs attempts to convert an error to a Dropbox Error
func ErrorAs(err error, target **Error) bool {
	return errors.As(err, target)
}

// IsRetryable returns true if the error is retryable
func IsRetryable(err error) bool {
	var dbErr *Error
	if !ErrorAs(err, &dbErr) {
		return false
	}

	switch dbErr.Type {
	case ErrorTypeRateLimit, ErrorTypeNetwork, ErrorTypeServer, ErrorTypeCircuitOpen:
		return true
	default:
		return false
	}
}

// statusToErrorType maps HTTP status codes to error types
func statusToErrorType(status int) ErrorType {
	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return ErrorTypeAuth
	case status == http.StatusTooManyRequests:
		return ErrorTypeRateLimit
	case status >= 500:
		return ErrorTypeServer
	default:
		return ErrorTypeUnknown
	}
}
