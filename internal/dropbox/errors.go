package dropbox

import (
	"errors"
	"fmt"
	"net/http"

	cerrors "github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/errors"
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
	// ErrorTypeFileSizeLimit represents a file size limit error
	ErrorTypeFileSizeLimit ErrorType = "file_size_limit"
)

// Error represents a Dropbox API error
type Error struct {
	cerr *cerrors.Error
	Type ErrorType
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.cerr.Error()
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.cerr
}

// NewError creates a new Error with the given type and message
func NewError(typ ErrorType, msg string, cause error) *Error {
	category := typeToCategory(typ)
	err := cerrors.New(category, msg)
	if cause != nil {
		err = cerrors.Wrap(cause, category, msg)
	}
	
	return &Error{
		cerr: err.WithCode(fmt.Sprintf("DROPBOX_%s", typ)).
			WithDetails(map[string]interface{}{
				"error_type": string(typ),
			}),
		Type: typ,
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

// NewFileSizeLimitError creates a new file size limit error
func NewFileSizeLimitError(msg string, cause error) *Error {
	return NewError(ErrorTypeFileSizeLimit, msg, cause)
}

// IsRetryable returns true if the error is retryable
func IsRetryable(err error) bool {
	var dbErr *Error
	if !errors.As(err, &dbErr) {
		return false
	}

	switch dbErr.Type {
	case ErrorTypeNetwork, ErrorTypeRateLimit, ErrorTypeServer:
		return true
	case ErrorTypeAuth, ErrorTypeInvalidInput, ErrorTypeCircuitOpen, ErrorTypeFileSizeLimit:
		return false
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
	case status >= 400:
		return ErrorTypeInvalidInput
	default:
		return ErrorTypeUnknown
	}
}

// typeToCategory maps ErrorType to cerrors.Category
func typeToCategory(typ ErrorType) cerrors.Category {
	switch typ {
	case ErrorTypeAuth:
		return cerrors.CategoryPermissionDenied
	case ErrorTypeRateLimit:
		return cerrors.CategoryUnavailable
	case ErrorTypeNetwork:
		return cerrors.CategoryUnavailable
	case ErrorTypeServer:
		return cerrors.CategoryUnavailable
	case ErrorTypeInvalidInput:
		return cerrors.CategoryInvalidArgument
	case ErrorTypeCircuitOpen:
		return cerrors.CategoryUnavailable
	case ErrorTypeFileSizeLimit:
		return cerrors.CategoryInvalidArgument
	default:
		return cerrors.CategoryUnknown
	}
}
