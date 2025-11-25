package errors

import (
	"errors"
	"testing"

	"github.com/infiniv/rsearch/pkg/rsearch"
)

func TestParseError(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		position int
		wantCode string
		wantMsg  string
	}{
		{
			name:     "basic parse error",
			msg:      "unexpected token",
			position: 10,
			wantCode: rsearch.ErrorCodeParseError,
			wantMsg:  "unexpected token",
		},
		{
			name:     "parse error with zero position",
			msg:      "invalid syntax",
			position: 0,
			wantCode: rsearch.ErrorCodeParseError,
			wantMsg:  "invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewParseError(tt.msg, tt.position)
			if err.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", err.Code, tt.wantCode)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("Message = %v, want %v", err.Message, tt.wantMsg)
			}
			if err.Position != tt.position {
				t.Errorf("Position = %v, want %v", err.Position, tt.position)
			}
			if err.Error() == "" {
				t.Error("Error() should return non-empty string")
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		field    string
		wantCode string
		wantMsg  string
	}{
		{
			name:     "basic validation error",
			msg:      "field is required",
			field:    "name",
			wantCode: rsearch.ErrorCodeInvalidSchema,
			wantMsg:  "field is required",
		},
		{
			name:     "validation error without field",
			msg:      "invalid input",
			field:    "",
			wantCode: rsearch.ErrorCodeInvalidSchema,
			wantMsg:  "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.msg, tt.field)
			if err.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", err.Code, tt.wantCode)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("Message = %v, want %v", err.Message, tt.wantMsg)
			}
			if err.Field != tt.field {
				t.Errorf("Field = %v, want %v", err.Field, tt.field)
			}
			if err.Error() == "" {
				t.Error("Error() should return non-empty string")
			}
		})
	}
}

func TestSchemaError(t *testing.T) {
	tests := []struct {
		name       string
		msg        string
		schemaName string
		code       string
		wantCode   string
		wantMsg    string
	}{
		{
			name:       "schema not found",
			msg:        "schema does not exist",
			schemaName: "products",
			code:       rsearch.ErrorCodeSchemaNotFound,
			wantCode:   rsearch.ErrorCodeSchemaNotFound,
			wantMsg:    "schema does not exist",
		},
		{
			name:       "schema already exists",
			msg:        "schema already registered",
			schemaName: "users",
			code:       rsearch.ErrorCodeSchemaExists,
			wantCode:   rsearch.ErrorCodeSchemaExists,
			wantMsg:    "schema already registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewSchemaError(tt.msg, tt.schemaName, tt.code)
			if err.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", err.Code, tt.wantCode)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("Message = %v, want %v", err.Message, tt.wantMsg)
			}
			if err.SchemaName != tt.schemaName {
				t.Errorf("SchemaName = %v, want %v", err.SchemaName, tt.schemaName)
			}
			if err.Error() == "" {
				t.Error("Error() should return non-empty string")
			}
		})
	}
}

func TestTranslationError(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		field    string
		wantMsg  string
	}{
		{
			name:    "basic translation error",
			msg:     "unsupported operator",
			field:   "status",
			wantMsg: "unsupported operator",
		},
		{
			name:    "translation error without field",
			msg:     "translation failed",
			field:   "",
			wantMsg: "translation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewTranslationError(tt.msg, tt.field)
			if err.Message != tt.wantMsg {
				t.Errorf("Message = %v, want %v", err.Message, tt.wantMsg)
			}
			if err.Field != tt.field {
				t.Errorf("Field = %v, want %v", err.Field, tt.field)
			}
			if err.Error() == "" {
				t.Error("Error() should return non-empty string")
			}
		})
	}
}

func TestRateLimitError(t *testing.T) {
	tests := []struct {
		name       string
		msg        string
		retryAfter int
		wantCode   string
		wantMsg    string
	}{
		{
			name:       "basic rate limit error",
			msg:        "too many requests",
			retryAfter: 60,
			wantCode:   rsearch.ErrorCodeRateLimited,
			wantMsg:    "too many requests",
		},
		{
			name:       "rate limit with zero retry",
			msg:        "rate limit exceeded",
			retryAfter: 0,
			wantCode:   rsearch.ErrorCodeRateLimited,
			wantMsg:    "rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewRateLimitError(tt.msg, tt.retryAfter)
			if err.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", err.Code, tt.wantCode)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("Message = %v, want %v", err.Message, tt.wantMsg)
			}
			if err.RetryAfter != tt.retryAfter {
				t.Errorf("RetryAfter = %v, want %v", err.RetryAfter, tt.retryAfter)
			}
			if err.Error() == "" {
				t.Error("Error() should return non-empty string")
			}
		})
	}
}

func TestAuthError(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		code     string
		wantCode string
		wantMsg  string
	}{
		{
			name:     "unauthorized error",
			msg:      "invalid credentials",
			code:     rsearch.ErrorCodeUnauthorized,
			wantCode: rsearch.ErrorCodeUnauthorized,
			wantMsg:  "invalid credentials",
		},
		{
			name:     "forbidden error",
			msg:      "access denied",
			code:     rsearch.ErrorCodeForbidden,
			wantCode: rsearch.ErrorCodeForbidden,
			wantMsg:  "access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAuthError(tt.msg, tt.code)
			if err.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", err.Code, tt.wantCode)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("Message = %v, want %v", err.Message, tt.wantMsg)
			}
			if err.Error() == "" {
				t.Error("Error() should return non-empty string")
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	t.Run("wrap parse error", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := NewParseError("parse failed", 5).Wrap(cause)

		if err.Cause == nil {
			t.Error("Cause should not be nil")
		}

		unwrapped := err.Unwrap()
		if unwrapped != cause {
			t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
		}

		if !errors.Is(err, cause) {
			t.Error("errors.Is should return true for wrapped error")
		}
	})

	t.Run("wrap validation error", func(t *testing.T) {
		cause := errors.New("validation failed")
		err := NewValidationError("invalid field", "name").Wrap(cause)

		if err.Cause == nil {
			t.Error("Cause should not be nil")
		}

		unwrapped := err.Unwrap()
		if unwrapped != cause {
			t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
		}
	})

	t.Run("wrap schema error", func(t *testing.T) {
		cause := errors.New("schema error")
		err := NewSchemaError("schema not found", "products", rsearch.ErrorCodeSchemaNotFound).Wrap(cause)

		if err.Cause == nil {
			t.Error("Cause should not be nil")
		}

		unwrapped := err.Unwrap()
		if unwrapped != cause {
			t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
		}
	})

	t.Run("wrap translation error", func(t *testing.T) {
		cause := errors.New("translation failed")
		err := NewTranslationError("cannot translate", "status").Wrap(cause)

		if err.Cause == nil {
			t.Error("Cause should not be nil")
		}

		unwrapped := err.Unwrap()
		if unwrapped != cause {
			t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
		}
	})
}

func TestErrorsAs(t *testing.T) {
	t.Run("errors.As with ParseError", func(t *testing.T) {
		err := NewParseError("syntax error", 10)
		var parseErr *ParseError
		if !errors.As(err, &parseErr) {
			t.Error("errors.As should return true for ParseError")
		}
		if parseErr.Position != 10 {
			t.Errorf("Position = %v, want 10", parseErr.Position)
		}
	})

	t.Run("errors.As with ValidationError", func(t *testing.T) {
		err := NewValidationError("invalid field", "name")
		var validErr *ValidationError
		if !errors.As(err, &validErr) {
			t.Error("errors.As should return true for ValidationError")
		}
		if validErr.Field != "name" {
			t.Errorf("Field = %v, want name", validErr.Field)
		}
	})

	t.Run("errors.As with SchemaError", func(t *testing.T) {
		err := NewSchemaError("not found", "products", rsearch.ErrorCodeSchemaNotFound)
		var schemaErr *SchemaError
		if !errors.As(err, &schemaErr) {
			t.Error("errors.As should return true for SchemaError")
		}
		if schemaErr.SchemaName != "products" {
			t.Errorf("SchemaName = %v, want products", schemaErr.SchemaName)
		}
	})
}

func TestToErrorDetail(t *testing.T) {
	t.Run("ParseError to ErrorDetail", func(t *testing.T) {
		err := NewParseError("syntax error", 10)
		detail := err.ToErrorDetail()

		if detail.Code != rsearch.ErrorCodeParseError {
			t.Errorf("Code = %v, want %v", detail.Code, rsearch.ErrorCodeParseError)
		}
		if detail.Message != "syntax error" {
			t.Errorf("Message = %v, want syntax error", detail.Message)
		}
		if len(detail.Details) != 1 {
			t.Errorf("Details length = %v, want 1", len(detail.Details))
		}
		if detail.Details[0].Position != 10 {
			t.Errorf("Details[0].Position = %v, want 10", detail.Details[0].Position)
		}
	})

	t.Run("ValidationError to ErrorDetail", func(t *testing.T) {
		err := NewValidationError("required field", "name")
		detail := err.ToErrorDetail()

		if detail.Code != rsearch.ErrorCodeInvalidSchema {
			t.Errorf("Code = %v, want %v", detail.Code, rsearch.ErrorCodeInvalidSchema)
		}
		if detail.Message != "required field" {
			t.Errorf("Message = %v, want required field", detail.Message)
		}
		if len(detail.Details) != 1 {
			t.Errorf("Details length = %v, want 1", len(detail.Details))
		}
		if detail.Details[0].Message != "field: name" {
			t.Errorf("Details[0].Message = %v, want field: name", detail.Details[0].Message)
		}
	})

	t.Run("SchemaError to ErrorDetail", func(t *testing.T) {
		err := NewSchemaError("not found", "products", rsearch.ErrorCodeSchemaNotFound)
		detail := err.ToErrorDetail()

		if detail.Code != rsearch.ErrorCodeSchemaNotFound {
			t.Errorf("Code = %v, want %v", detail.Code, rsearch.ErrorCodeSchemaNotFound)
		}
		if detail.Message != "not found" {
			t.Errorf("Message = %v, want not found", detail.Message)
		}
		if len(detail.Details) != 1 {
			t.Errorf("Details length = %v, want 1", len(detail.Details))
		}
		if detail.Details[0].Message != "schema: products" {
			t.Errorf("Details[0].Message = %v, want schema: products", detail.Details[0].Message)
		}
	})

	t.Run("RateLimitError to ErrorDetail", func(t *testing.T) {
		err := NewRateLimitError("too many requests", 60)
		detail := err.ToErrorDetail()

		if detail.Code != rsearch.ErrorCodeRateLimited {
			t.Errorf("Code = %v, want %v", detail.Code, rsearch.ErrorCodeRateLimited)
		}
		if detail.Message != "too many requests" {
			t.Errorf("Message = %v, want too many requests", detail.Message)
		}
		if len(detail.Details) != 1 {
			t.Errorf("Details length = %v, want 1", len(detail.Details))
		}
		if detail.Details[0].Message != "retry after 60 seconds" {
			t.Errorf("Details[0].Message = %v, want retry after 60 seconds", detail.Details[0].Message)
		}
	})
}

func TestErrorMessageFormatting(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{
			name:    "ParseError message",
			err:     NewParseError("unexpected token", 5),
			wantMsg: "parse error: unexpected token (position: 5)",
		},
		{
			name:    "ValidationError message with field",
			err:     NewValidationError("field required", "name"),
			wantMsg: "validation error: field required (field: name)",
		},
		{
			name:    "ValidationError message without field",
			err:     NewValidationError("invalid input", ""),
			wantMsg: "validation error: invalid input",
		},
		{
			name:    "SchemaError message",
			err:     NewSchemaError("not found", "products", rsearch.ErrorCodeSchemaNotFound),
			wantMsg: "schema error: not found (schema: products)",
		},
		{
			name:    "TranslationError message with field",
			err:     NewTranslationError("unsupported type", "status"),
			wantMsg: "translation error: unsupported type (field: status)",
		},
		{
			name:    "TranslationError message without field",
			err:     NewTranslationError("failed", ""),
			wantMsg: "translation error: failed",
		},
		{
			name:    "RateLimitError message",
			err:     NewRateLimitError("too many requests", 60),
			wantMsg: "rate limit error: too many requests (retry after: 60s)",
		},
		{
			name:    "AuthError message",
			err:     NewAuthError("invalid token", rsearch.ErrorCodeUnauthorized),
			wantMsg: "auth error: invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if msg := tt.err.Error(); msg != tt.wantMsg {
				t.Errorf("Error() = %v, want %v", msg, tt.wantMsg)
			}
		})
	}
}
