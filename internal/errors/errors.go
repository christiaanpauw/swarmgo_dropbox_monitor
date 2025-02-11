package errors

import (
	"errors"
	"fmt"
)

// Common error types
var (
	ErrNotFound          = errors.New("resource not found")
	ErrInvalidState      = errors.New("invalid state")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrNotImplemented    = errors.New("not implemented")
	ErrUnavailable       = errors.New("service unavailable")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrPermissionDenied  = errors.New("permission denied")
)

// Error categories
type Category string

const (
	CategoryNotFound         Category = "not_found"
	CategoryInvalidState     Category = "invalid_state"
	CategoryInvalidArgument  Category = "invalid_argument"
	CategoryNotImplemented   Category = "not_implemented"
	CategoryUnavailable      Category = "unavailable"
	CategoryAlreadyExists    Category = "already_exists"
	CategoryPermissionDenied Category = "permission_denied"
	CategoryUnknown         Category = "unknown"
)

// Error represents a structured error with additional context
type Error struct {
	// Original is the underlying error
	Original error
	// Category is the error category
	Category Category
	// Code is a specific error code within the category
	Code string
	// Message is a human-readable error message
	Message string
	// Details contains additional error context
	Details map[string]interface{}
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Original.Error()
}

// Unwrap implements the unwrap interface
func (e *Error) Unwrap() error {
	return e.Original
}

// Is implements the Is interface for error comparison
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return errors.Is(e.Original, target)
	}
	return e.Category == t.Category && e.Code == t.Code
}

// New creates a new Error with the given category and message
func New(category Category, message string) *Error {
	return &Error{
		Original: errors.New(message),
		Category: category,
		Message:  message,
		Details: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, category Category, message string) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Original: err,
		Category: category,
		Message:  message,
		Details: make(map[string]interface{}),
	}
}

// WithCode adds an error code to the Error
func (e *Error) WithCode(code string) *Error {
	e.Code = code
	return e
}

// WithDetails adds details to the Error
func (e *Error) WithDetails(details map[string]interface{}) *Error {
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// GetCategory returns the error category for any error
func GetCategory(err error) Category {
	var e *Error
	if errors.As(err, &e) {
		return e.Category
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return CategoryNotFound
	case errors.Is(err, ErrInvalidState):
		return CategoryInvalidState
	case errors.Is(err, ErrInvalidArgument):
		return CategoryInvalidArgument
	case errors.Is(err, ErrNotImplemented):
		return CategoryNotImplemented
	case errors.Is(err, ErrUnavailable):
		return CategoryUnavailable
	case errors.Is(err, ErrAlreadyExists):
		return CategoryAlreadyExists
	case errors.Is(err, ErrPermissionDenied):
		return CategoryPermissionDenied
	default:
		return CategoryUnknown
	}
}

// FormatError formats an error with its full context
func FormatError(err error) string {
	var e *Error
	if !errors.As(err, &e) {
		return err.Error()
	}

	var result string
	if e.Message != "" {
		result = e.Message
	} else {
		result = e.Original.Error()
	}

	if e.Code != "" {
		result = fmt.Sprintf("[%s] %s", e.Code, result)
	}

	if len(e.Details) > 0 {
		result = fmt.Sprintf("%s (Details: %v)", result, e.Details)
	}

	return result
}
