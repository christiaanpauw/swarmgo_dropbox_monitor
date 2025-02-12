package database

import (
	"errors"
	"fmt"

	cerrors "github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/errors"
)

// ErrorType represents the type of database error that occurred
type ErrorType string

const (
	// ErrorTypeConnection represents a connection error
	ErrorTypeConnection ErrorType = "connection"
	// ErrorTypeQuery represents a query error
	ErrorTypeQuery ErrorType = "query"
	// ErrorTypeTransaction represents a transaction error
	ErrorTypeTransaction ErrorType = "transaction"
	// ErrorTypeConstraint represents a constraint violation error
	ErrorTypeConstraint ErrorType = "constraint"
	// ErrorTypeNotFound represents a not found error
	ErrorTypeNotFound ErrorType = "not_found"
)

// Error represents a database error
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

// NewError creates a new database Error with the given type and message
func NewError(typ ErrorType, msg string, cause error) *Error {
	category := typeToCategory(typ)
	err := cerrors.New(category, msg)
	if cause != nil {
		err = cerrors.Wrap(cause, category, msg)
	}
	
	return &Error{
		cerr: err.WithCode(fmt.Sprintf("DB_%s", typ)).
			WithDetails(map[string]interface{}{
				"error_type": string(typ),
			}),
		Type: typ,
	}
}

// NewConnectionError creates a new connection error
func NewConnectionError(msg string, cause error) *Error {
	return NewError(ErrorTypeConnection, msg, cause)
}

// NewQueryError creates a new query error
func NewQueryError(msg string, cause error) *Error {
	return NewError(ErrorTypeQuery, msg, cause)
}

// NewTransactionError creates a new transaction error
func NewTransactionError(msg string, cause error) *Error {
	return NewError(ErrorTypeTransaction, msg, cause)
}

// NewConstraintError creates a new constraint violation error
func NewConstraintError(msg string, cause error) *Error {
	return NewError(ErrorTypeConstraint, msg, cause)
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(msg string, cause error) *Error {
	return NewError(ErrorTypeNotFound, msg, cause)
}

// IsRetryable returns true if the error is retryable
func IsRetryable(err error) bool {
	var dbErr *Error
	if !errors.As(err, &dbErr) {
		return false
	}

	switch dbErr.Type {
	case ErrorTypeConnection, ErrorTypeTransaction:
		return true
	default:
		return false
	}
}

// typeToCategory maps ErrorType to cerrors.Category
func typeToCategory(typ ErrorType) cerrors.Category {
	switch typ {
	case ErrorTypeConnection:
		return cerrors.CategoryUnavailable
	case ErrorTypeQuery:
		return cerrors.CategoryInvalidArgument
	case ErrorTypeTransaction:
		return cerrors.CategoryUnavailable
	case ErrorTypeConstraint:
		return cerrors.CategoryInvalidArgument
	case ErrorTypeNotFound:
		return cerrors.CategoryNotFound
	default:
		return cerrors.CategoryUnknown
	}
}
