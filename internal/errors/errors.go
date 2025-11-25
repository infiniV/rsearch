package errors

import (
	"fmt"

	"github.com/infiniv/rsearch/pkg/rsearch"
)

// ParseError represents an error that occurred during query parsing
type ParseError struct {
	Code     string
	Message  string
	Position int
	Cause    error
}

// NewParseError creates a new ParseError
func NewParseError(msg string, position int) *ParseError {
	return &ParseError{
		Code:     rsearch.ErrorCodeParseError,
		Message:  msg,
		Position: position,
	}
}

// Error implements the error interface
func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error: %s (position: %d)", e.Message, e.Position)
}

// Unwrap implements the unwrap interface for error wrapping
func (e *ParseError) Unwrap() error {
	return e.Cause
}

// Wrap wraps an underlying error
func (e *ParseError) Wrap(cause error) *ParseError {
	e.Cause = cause
	return e
}

// ToErrorDetail converts the error to an ErrorDetail for API responses
func (e *ParseError) ToErrorDetail() rsearch.ErrorDetail {
	return rsearch.ErrorDetail{
		Code:    e.Code,
		Message: e.Message,
		Details: []rsearch.ErrorInfo{
			{
				Position: e.Position,
				Message:  e.Message,
			},
		},
	}
}

// ValidationError represents an error that occurred during validation
type ValidationError struct {
	Code    string
	Message string
	Field   string
	Cause   error
}

// NewValidationError creates a new ValidationError
func NewValidationError(msg string, field string) *ValidationError {
	return &ValidationError{
		Code:    rsearch.ErrorCodeInvalidSchema,
		Message: msg,
		Field:   field,
	}
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error: %s (field: %s)", e.Message, e.Field)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// Unwrap implements the unwrap interface for error wrapping
func (e *ValidationError) Unwrap() error {
	return e.Cause
}

// Wrap wraps an underlying error
func (e *ValidationError) Wrap(cause error) *ValidationError {
	e.Cause = cause
	return e
}

// ToErrorDetail converts the error to an ErrorDetail for API responses
func (e *ValidationError) ToErrorDetail() rsearch.ErrorDetail {
	details := []rsearch.ErrorInfo{
		{
			Message: fmt.Sprintf("field: %s", e.Field),
		},
	}
	return rsearch.ErrorDetail{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// SchemaError represents an error related to schema operations
type SchemaError struct {
	Code       string
	Message    string
	SchemaName string
	Cause      error
}

// NewSchemaError creates a new SchemaError
func NewSchemaError(msg string, schemaName string, code string) *SchemaError {
	return &SchemaError{
		Code:       code,
		Message:    msg,
		SchemaName: schemaName,
	}
}

// Error implements the error interface
func (e *SchemaError) Error() string {
	return fmt.Sprintf("schema error: %s (schema: %s)", e.Message, e.SchemaName)
}

// Unwrap implements the unwrap interface for error wrapping
func (e *SchemaError) Unwrap() error {
	return e.Cause
}

// Wrap wraps an underlying error
func (e *SchemaError) Wrap(cause error) *SchemaError {
	e.Cause = cause
	return e
}

// ToErrorDetail converts the error to an ErrorDetail for API responses
func (e *SchemaError) ToErrorDetail() rsearch.ErrorDetail {
	details := []rsearch.ErrorInfo{
		{
			Message: fmt.Sprintf("schema: %s", e.SchemaName),
		},
	}
	return rsearch.ErrorDetail{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// TranslationError represents an error that occurred during query translation
type TranslationError struct {
	Code    string
	Message string
	Field   string
	Cause   error
}

// NewTranslationError creates a new TranslationError
func NewTranslationError(msg string, field string) *TranslationError {
	return &TranslationError{
		Code:    rsearch.ErrorCodeInternalError,
		Message: msg,
		Field:   field,
	}
}

// Error implements the error interface
func (e *TranslationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("translation error: %s (field: %s)", e.Message, e.Field)
	}
	return fmt.Sprintf("translation error: %s", e.Message)
}

// Unwrap implements the unwrap interface for error wrapping
func (e *TranslationError) Unwrap() error {
	return e.Cause
}

// Wrap wraps an underlying error
func (e *TranslationError) Wrap(cause error) *TranslationError {
	e.Cause = cause
	return e
}

// ToErrorDetail converts the error to an ErrorDetail for API responses
func (e *TranslationError) ToErrorDetail() rsearch.ErrorDetail {
	details := []rsearch.ErrorInfo{}
	if e.Field != "" {
		details = append(details, rsearch.ErrorInfo{
			Message: fmt.Sprintf("field: %s", e.Field),
		})
	}
	return rsearch.ErrorDetail{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// RateLimitError represents a rate limiting error
type RateLimitError struct {
	Code       string
	Message    string
	RetryAfter int
	Cause      error
}

// NewRateLimitError creates a new RateLimitError
func NewRateLimitError(msg string, retryAfter int) *RateLimitError {
	return &RateLimitError{
		Code:       rsearch.ErrorCodeRateLimited,
		Message:    msg,
		RetryAfter: retryAfter,
	}
}

// Error implements the error interface
func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit error: %s (retry after: %ds)", e.Message, e.RetryAfter)
}

// Unwrap implements the unwrap interface for error wrapping
func (e *RateLimitError) Unwrap() error {
	return e.Cause
}

// Wrap wraps an underlying error
func (e *RateLimitError) Wrap(cause error) *RateLimitError {
	e.Cause = cause
	return e
}

// ToErrorDetail converts the error to an ErrorDetail for API responses
func (e *RateLimitError) ToErrorDetail() rsearch.ErrorDetail {
	details := []rsearch.ErrorInfo{
		{
			Message: fmt.Sprintf("retry after %d seconds", e.RetryAfter),
		},
	}
	return rsearch.ErrorDetail{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// AuthError represents an authentication or authorization error
type AuthError struct {
	Code    string
	Message string
	Cause   error
}

// NewAuthError creates a new AuthError
func NewAuthError(msg string, code string) *AuthError {
	return &AuthError{
		Code:    code,
		Message: msg,
	}
}

// Error implements the error interface
func (e *AuthError) Error() string {
	return fmt.Sprintf("auth error: %s", e.Message)
}

// Unwrap implements the unwrap interface for error wrapping
func (e *AuthError) Unwrap() error {
	return e.Cause
}

// Wrap wraps an underlying error
func (e *AuthError) Wrap(cause error) *AuthError {
	e.Cause = cause
	return e
}

// ToErrorDetail converts the error to an ErrorDetail for API responses
func (e *AuthError) ToErrorDetail() rsearch.ErrorDetail {
	return rsearch.ErrorDetail{
		Code:    e.Code,
		Message: e.Message,
	}
}
