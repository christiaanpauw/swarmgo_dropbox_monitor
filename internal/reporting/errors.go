package reporting

import (
	"errors"
	"fmt"

	cerrors "github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/errors"
)

// ErrorType represents the type of reporting error that occurred
type ErrorType string

const (
	// ErrorTypeGeneration represents a report generation error
	ErrorTypeGeneration ErrorType = "generation"
	// ErrorTypeValidation represents a validation error
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeTemplate represents a template error
	ErrorTypeTemplate ErrorType = "template"
	// ErrorTypeStorage represents a storage error
	ErrorTypeStorage ErrorType = "storage"
	// ErrorTypeLifecycle represents a lifecycle error
	ErrorTypeLifecycle ErrorType = "lifecycle"
)

// Error represents a reporting error
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

// NewError creates a new reporting Error with the given type and message
func NewError(typ ErrorType, msg string, cause error) *Error {
	category := typeToCategory(typ)
	err := cerrors.New(category, msg)
	if cause != nil {
		err = cerrors.Wrap(cause, category, msg)
	}
	
	return &Error{
		cerr: err.WithCode(fmt.Sprintf("REPORT_%s", typ)).
			WithDetails(map[string]interface{}{
				"error_type": string(typ),
			}),
		Type: typ,
	}
}

// NewGenerationError creates a new generation error
func NewGenerationError(msg string, cause error) *Error {
	return NewError(ErrorTypeGeneration, msg, cause)
}

// NewValidationError creates a new validation error
func NewValidationError(msg string, cause error) *Error {
	return NewError(ErrorTypeValidation, msg, cause)
}

// NewTemplateError creates a new template error
func NewTemplateError(msg string, cause error) *Error {
	return NewError(ErrorTypeTemplate, msg, cause)
}

// NewStorageError creates a new storage error
func NewStorageError(msg string, cause error) *Error {
	return NewError(ErrorTypeStorage, msg, cause)
}

// NewLifecycleError creates a new lifecycle error
func NewLifecycleError(msg string, cause error) *Error {
	return NewError(ErrorTypeLifecycle, msg, cause)
}

// IsRetryable returns true if the error is retryable
func IsRetryable(err error) bool {
	var repErr *Error
	if !errors.As(err, &repErr) {
		return false
	}

	switch repErr.Type {
	case ErrorTypeGeneration, ErrorTypeStorage:
		return true
	default:
		return false
	}
}

// typeToCategory maps ErrorType to cerrors.Category
func typeToCategory(typ ErrorType) cerrors.Category {
	switch typ {
	case ErrorTypeGeneration:
		return cerrors.CategoryUnavailable
	case ErrorTypeValidation:
		return cerrors.CategoryInvalidArgument
	case ErrorTypeTemplate:
		return cerrors.CategoryInvalidArgument
	case ErrorTypeStorage:
		return cerrors.CategoryUnavailable
	case ErrorTypeLifecycle:
		return cerrors.CategoryInvalidState
	default:
		return cerrors.CategoryUnknown
	}
}
