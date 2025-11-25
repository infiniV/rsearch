package validation

import (
	"strings"
	"testing"

	"github.com/infiniv/rsearch/internal/config"
)

func TestValidateQuery(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		maxLength   int
		blockSQL    bool
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid simple query",
			query:     "status:active",
			maxLength: 1000,
			blockSQL:  true,
			wantErr:   false,
		},
		{
			name:      "valid complex query",
			query:     "(status:active OR status:pending) AND created:[2023-01-01 TO 2023-12-31]",
			maxLength: 1000,
			blockSQL:  true,
			wantErr:   false,
		},
		{
			name:        "query exceeds max length",
			query:       strings.Repeat("a", 1001),
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "exceeds maximum length",
		},
		{
			name:        "SQL keyword SELECT",
			query:       "SELECT * FROM users",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains SQL keyword",
		},
		{
			name:        "SQL keyword INSERT",
			query:       "status:active INSERT INTO",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains SQL keyword",
		},
		{
			name:        "SQL keyword UPDATE",
			query:       "UPDATE users SET",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains SQL keyword",
		},
		{
			name:        "SQL keyword DELETE",
			query:       "DELETE FROM users",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains SQL keyword",
		},
		{
			name:        "SQL keyword DROP",
			query:       "DROP TABLE users",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains SQL keyword",
		},
		{
			name:        "SQL keyword UNION",
			query:       "status:active UNION SELECT",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains SQL keyword",
		},
		{
			name:        "SQL keyword TRUNCATE",
			query:       "TRUNCATE TABLE",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains SQL keyword",
		},
		{
			name:        "SQL keyword ALTER",
			query:       "ALTER TABLE users",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains SQL keyword",
		},
		{
			name:        "SQL keyword EXEC",
			query:       "EXEC sp_droplogin",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains SQL keyword",
		},
		{
			name:        "SQL keyword EXECUTE",
			query:       "EXECUTE procedure",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains SQL keyword",
		},
		{
			name:        "SQL comment --",
			query:       "status:active -- comment",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains SQL comment",
		},
		{
			name:        "SQL comment /* */",
			query:       "status:active /* comment */",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains SQL comment",
		},
		{
			name:        "SQL injection attempt with semicolon",
			query:       "status:active; DROP TABLE users",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains SQL keyword",
		},
		{
			name:        "null byte in query",
			query:       "status:active\x00",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains null byte",
		},
		{
			name:        "control characters in query",
			query:       "status:active\x01\x02",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains control character",
		},
		{
			name:      "SQL keywords allowed when blocking disabled",
			query:     "SELECT * FROM users",
			maxLength: 1000,
			blockSQL:  false,
			wantErr:   false,
		},
		{
			name:      "valid query with wildcards",
			query:     "name:john* AND email:*@example.com",
			maxLength: 1000,
			blockSQL:  true,
			wantErr:   false,
		},
		{
			name:      "valid query with regex",
			query:     "email:/.*@example\\.com/",
			maxLength: 1000,
			blockSQL:  true,
			wantErr:   false,
		},
		{
			name:      "valid query with fuzzy",
			query:     "name:john~2",
			maxLength: 1000,
			blockSQL:  true,
			wantErr:   false,
		},
		{
			name:      "valid query with boost",
			query:     "status:active^2",
			maxLength: 1000,
			blockSQL:  true,
			wantErr:   false,
		},
		{
			name:      "valid query with ranges",
			query:     "age:[18 TO 65] AND price:{0 TO 1000}",
			maxLength: 1000,
			blockSQL:  true,
			wantErr:   false,
		},
		{
			name:        "case insensitive SQL keyword detection",
			query:       "SeLeCt * FrOm users",
			maxLength:   1000,
			blockSQL:    true,
			wantErr:     true,
			errContains: "contains SQL keyword",
		},
		{
			name:      "empty query",
			query:     "",
			maxLength: 1000,
			blockSQL:  true,
			wantErr:   false,
		},
		{
			name:      "query with special characters",
			query:     "title:\"hello world\" AND status:active",
			maxLength: 1000,
			blockSQL:  true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.SecurityConfig{
				BlockSqlKeywords: tt.blockSQL,
			}
			limits := &config.LimitsConfig{
				MaxQueryLength: tt.maxLength,
			}

			validator := NewValidator(cfg, limits)
			err := validator.ValidateQuery(tt.query)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateQuery() expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateQuery() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateQuery() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateSchemaName(t *testing.T) {
	tests := []struct {
		name               string
		schemaName         string
		allowedSpecialChars string
		maxFieldNameLength  int
		wantErr            bool
		errContains        string
	}{
		{
			name:                "valid schema name",
			schemaName:          "users_schema",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             false,
		},
		{
			name:                "valid schema name with dash",
			schemaName:          "user-schema",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             false,
		},
		{
			name:                "valid schema name with dot",
			schemaName:          "user.schema",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             false,
		},
		{
			name:                "empty schema name",
			schemaName:          "",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             true,
			errContains:         "cannot be empty",
		},
		{
			name:                "schema name too long",
			schemaName:          strings.Repeat("a", 256),
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             true,
			errContains:         "exceeds maximum length",
		},
		{
			name:                "schema name with invalid character",
			schemaName:          "user$schema",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             true,
			errContains:         "contains invalid character",
		},
		{
			name:                "schema name with space",
			schemaName:          "user schema",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             true,
			errContains:         "contains invalid character",
		},
		{
			name:                "schema name with SQL injection attempt",
			schemaName:          "users'; DROP TABLE",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             true,
			errContains:         "contains invalid character",
		},
		{
			name:                "schema name with null byte",
			schemaName:          "users\x00",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             true,
			errContains:         "contains null byte",
		},
		{
			name:                "schema name starting with number",
			schemaName:          "123users",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             false,
		},
		{
			name:                "schema name with unicode",
			schemaName:          "users_схема",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             true,
			errContains:         "contains invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.SecurityConfig{
				AllowedSpecialChars: tt.allowedSpecialChars,
			}
			limits := &config.LimitsConfig{
				MaxFieldNameLength: tt.maxFieldNameLength,
			}

			validator := NewValidator(cfg, limits)
			err := validator.ValidateSchemaName(tt.schemaName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateSchemaName() expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateSchemaName() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateSchemaName() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateFieldName(t *testing.T) {
	tests := []struct {
		name               string
		fieldName          string
		allowedSpecialChars string
		maxFieldNameLength  int
		wantErr            bool
		errContains        string
	}{
		{
			name:                "valid field name",
			fieldName:           "user_name",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             false,
		},
		{
			name:                "valid field name camelCase",
			fieldName:           "userName",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             false,
		},
		{
			name:                "valid field name with dot",
			fieldName:           "user.name",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             false,
		},
		{
			name:                "empty field name",
			fieldName:           "",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             true,
			errContains:         "cannot be empty",
		},
		{
			name:                "field name too long",
			fieldName:           strings.Repeat("a", 256),
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             true,
			errContains:         "exceeds maximum length",
		},
		{
			name:                "field name with invalid character",
			fieldName:           "user@name",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             true,
			errContains:         "contains invalid character",
		},
		{
			name:                "field name with space",
			fieldName:           "user name",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             true,
			errContains:         "contains invalid character",
		},
		{
			name:                "field name with null byte",
			fieldName:           "username\x00",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             true,
			errContains:         "contains null byte",
		},
		{
			name:                "field name starting with number",
			fieldName:           "2fa_enabled",
			allowedSpecialChars: ".-_",
			maxFieldNameLength:  255,
			wantErr:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.SecurityConfig{
				AllowedSpecialChars: tt.allowedSpecialChars,
			}
			limits := &config.LimitsConfig{
				MaxFieldNameLength: tt.maxFieldNameLength,
			}

			validator := NewValidator(cfg, limits)
			err := validator.ValidateFieldName(tt.fieldName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateFieldName() expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateFieldName() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateFieldName() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestNewValidator(t *testing.T) {
	cfg := &config.SecurityConfig{
		AllowedSpecialChars: ".-_",
		BlockSqlKeywords:    true,
	}
	limits := &config.LimitsConfig{
		MaxQueryLength:     1000,
		MaxFieldNameLength: 255,
	}

	validator := NewValidator(cfg, limits)
	if validator == nil {
		t.Error("NewValidator() returned nil")
	}
}
