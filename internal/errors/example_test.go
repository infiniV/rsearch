package errors_test

import (
	stderrors "errors"
	"fmt"

	"github.com/infiniv/rsearch/internal/errors"
	"github.com/infiniv/rsearch/pkg/rsearch"
)

// Example demonstrates basic error creation and usage
func ExampleNewParseError() {
	err := errors.NewParseError("unexpected token '}'", 42)
	fmt.Println(err.Error())
	// Output: parse error: unexpected token '}' (position: 42)
}

// Example demonstrates error wrapping
func ExampleParseError_Wrap() {
	underlying := stderrors.New("IO error")
	err := errors.NewParseError("failed to read query", 0).Wrap(underlying)

	// Check if error is the underlying error
	if stderrors.Is(err, underlying) {
		fmt.Println("Error contains underlying IO error")
	}

	// Unwrap to get the underlying error
	if unwrapped := stderrors.Unwrap(err); unwrapped != nil {
		fmt.Println("Unwrapped:", unwrapped.Error())
	}

	// Output:
	// Error contains underlying IO error
	// Unwrapped: IO error
}

// Example demonstrates converting error to API response
func ExampleParseError_ToErrorDetail() {
	err := errors.NewParseError("missing closing bracket", 15)
	detail := err.ToErrorDetail()

	fmt.Printf("Code: %s\n", detail.Code)
	fmt.Printf("Message: %s\n", detail.Message)
	fmt.Printf("Position: %d\n", detail.Details[0].Position)

	// Output:
	// Code: PARSE_ERROR
	// Message: missing closing bracket
	// Position: 15
}

// Example demonstrates validation error with field
func ExampleNewValidationError() {
	err := errors.NewValidationError("field is required", "email")
	fmt.Println(err.Error())
	// Output: validation error: field is required (field: email)
}

// Example demonstrates schema error
func ExampleNewSchemaError() {
	err := errors.NewSchemaError("schema not found", "products", rsearch.ErrorCodeSchemaNotFound)
	fmt.Println(err.Error())
	fmt.Println("Code:", err.Code)

	// Output:
	// schema error: schema not found (schema: products)
	// Code: SCHEMA_NOT_FOUND
}

// Example demonstrates rate limit error
func ExampleNewRateLimitError() {
	err := errors.NewRateLimitError("too many requests", 60)
	fmt.Println(err.Error())
	fmt.Println("Retry after:", err.RetryAfter, "seconds")

	// Output:
	// rate limit error: too many requests (retry after: 60s)
	// Retry after: 60 seconds
}

// Example demonstrates using errors.As for type assertion
func ExampleParseError_errorsAs() {
	var parseErr *errors.ParseError

	err := errors.NewParseError("syntax error", 10)

	if stderrors.As(err, &parseErr) {
		fmt.Printf("Parse error at position %d: %s\n", parseErr.Position, parseErr.Message)
	}

	// Output: Parse error at position 10: syntax error
}

// Example demonstrates translation error
func ExampleNewTranslationError() {
	err := errors.NewTranslationError("unsupported operator", "status")
	fmt.Println(err.Error())
	// Output: translation error: unsupported operator (field: status)
}

// Example demonstrates auth error
func ExampleNewAuthError() {
	err := errors.NewAuthError("invalid credentials", rsearch.ErrorCodeUnauthorized)
	fmt.Println(err.Error())
	fmt.Println("Code:", err.Code)

	// Output:
	// auth error: invalid credentials
	// Code: UNAUTHORIZED
}
