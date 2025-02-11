package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Creation(t *testing.T) {
	// Test New
	err := New(CategoryNotFound, "resource not found")
	assert.Equal(t, CategoryNotFound, err.Category)
	assert.Equal(t, "resource not found", err.Message)
	assert.NotNil(t, err.Details)

	// Test Wrap
	originalErr := errors.New("original error")
	wrapped := Wrap(originalErr, CategoryInvalidArgument, "wrapped error")
	assert.Equal(t, CategoryInvalidArgument, wrapped.Category)
	assert.Equal(t, "wrapped error", wrapped.Message)
	assert.Equal(t, originalErr, wrapped.Original)
}

func TestError_WithMethods(t *testing.T) {
	err := New(CategoryNotFound, "not found")
	
	// Test WithCode
	err = err.WithCode("USER_404")
	assert.Equal(t, "USER_404", err.Code)

	// Test WithDetails
	details := map[string]interface{}{
		"user_id": 123,
		"action":  "get",
	}
	err = err.WithDetails(details)
	assert.Equal(t, 123, err.Details["user_id"])
	assert.Equal(t, "get", err.Details["action"])
}

func TestError_ErrorInterface(t *testing.T) {
	// Test Error() method
	err := New(CategoryNotFound, "custom message")
	assert.Equal(t, "custom message", err.Error())

	// Test with empty message
	err = &Error{Original: errors.New("original message")}
	assert.Equal(t, "original message", err.Error())
}

func TestError_Unwrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, CategoryInvalidArgument, "wrapped error")
	
	// Test Unwrap
	unwrapped := errors.Unwrap(wrapped)
	assert.Equal(t, original, unwrapped)
}

func TestError_Is(t *testing.T) {
	err1 := New(CategoryNotFound, "not found").WithCode("404")
	err2 := New(CategoryNotFound, "different message").WithCode("404")
	err3 := New(CategoryInvalidArgument, "invalid").WithCode("400")

	// Same category and code should match
	assert.True(t, errors.Is(err1, err2))
	
	// Different category should not match
	assert.False(t, errors.Is(err1, err3))
}

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected Category
	}{
		{
			name:     "structured error",
			err:      New(CategoryNotFound, "not found"),
			expected: CategoryNotFound,
		},
		{
			name:     "wrapped standard error",
			err:      Wrap(errors.New("standard error"), CategoryInvalidArgument, "wrapped"),
			expected: CategoryInvalidArgument,
		},
		{
			name:     "standard error - not found",
			err:      ErrNotFound,
			expected: CategoryNotFound,
		},
		{
			name:     "unknown error",
			err:      errors.New("unknown error"),
			expected: CategoryUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := GetCategory(tt.err)
			assert.Equal(t, tt.expected, category)
		})
	}
}

func TestFormatError(t *testing.T) {
	// Test basic error
	err := New(CategoryNotFound, "resource not found")
	assert.Equal(t, "resource not found", FormatError(err))

	// Test error with code
	err = err.WithCode("USER_404")
	assert.Equal(t, "[USER_404] resource not found", FormatError(err))

	// Test error with details
	err = err.WithDetails(map[string]interface{}{"id": 123})
	assert.Contains(t, FormatError(err), "[USER_404] resource not found")
	assert.Contains(t, FormatError(err), "Details: map[id:123]")

	// Test standard error
	standardErr := errors.New("standard error")
	assert.Equal(t, "standard error", FormatError(standardErr))
}
