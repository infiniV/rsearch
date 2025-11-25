package validation

import (
	"testing"
)

func TestSanitizeQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean query unchanged",
			input:    "status:active AND created:[2023-01-01 TO 2023-12-31]",
			expected: "status:active AND created:[2023-01-01 TO 2023-12-31]",
		},
		{
			name:     "remove null bytes",
			input:    "status:active\x00",
			expected: "status:active",
		},
		{
			name:     "remove control characters",
			input:    "status:active\x01\x02\x03",
			expected: "status:active",
		},
		{
			name:     "preserve tabs and newlines",
			input:    "status:active\t\n\r",
			expected: "status:active",
		},
		{
			name:     "remove SQL comments --",
			input:    "status:active -- malicious comment",
			expected: "status:active",
		},
		{
			name:     "remove SQL comments /* */",
			input:    "status:active /* comment */ AND type:user",
			expected: "status:active  AND type:user",
		},
		{
			name:     "remove multiple consecutive comments",
			input:    "status:active /* first */ /* second */",
			expected: "status:active",
		},
		{
			name:     "trim excessive whitespace",
			input:    "  status:active  ",
			expected: "status:active",
		},
		{
			name:     "empty query",
			input:    "",
			expected: "",
		},
		{
			name:     "query with only whitespace",
			input:    "   \t\n  ",
			expected: "",
		},
		{
			name:     "preserve valid special characters",
			input:    "name:john* AND email:*@example.com",
			expected: "name:john* AND email:*@example.com",
		},
		{
			name:     "preserve quotes",
			input:    `title:"hello world"`,
			expected: `title:"hello world"`,
		},
		{
			name:     "preserve parentheses",
			input:    "(status:active OR status:pending)",
			expected: "(status:active OR status:pending)",
		},
		{
			name:     "preserve brackets and braces",
			input:    "age:[18 TO 65] AND price:{0 TO 1000}",
			expected: "age:[18 TO 65] AND price:{0 TO 1000}",
		},
		{
			name:     "remove multiple null bytes",
			input:    "status\x00:active\x00",
			expected: "status:active",
		},
		{
			name:     "complex sanitization",
			input:    "\x01status:active\x00 -- comment\t/* block */ AND\x02 type:user",
			expected: "status:active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeQuery(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeQuery() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizeFieldName(t *testing.T) {
	tests := []struct {
		name                string
		input               string
		allowedSpecialChars string
		expected            string
	}{
		{
			name:                "clean field name unchanged",
			input:               "user_name",
			allowedSpecialChars: ".-_",
			expected:            "user_name",
		},
		{
			name:                "remove invalid characters",
			input:               "user@name#field",
			allowedSpecialChars: ".-_",
			expected:            "usernamefield",
		},
		{
			name:                "preserve allowed special chars",
			input:               "user.name_field-id",
			allowedSpecialChars: ".-_",
			expected:            "user.name_field-id",
		},
		{
			name:                "remove null bytes",
			input:               "user\x00name",
			allowedSpecialChars: ".-_",
			expected:            "username",
		},
		{
			name:                "remove control characters",
			input:               "user\x01name\x02",
			allowedSpecialChars: ".-_",
			expected:            "username",
		},
		{
			name:                "remove spaces",
			input:               "user name",
			allowedSpecialChars: ".-_",
			expected:            "username",
		},
		{
			name:                "preserve alphanumeric",
			input:               "user123Name",
			allowedSpecialChars: ".-_",
			expected:            "user123Name",
		},
		{
			name:                "empty field name",
			input:               "",
			allowedSpecialChars: ".-_",
			expected:            "",
		},
		{
			name:                "only invalid characters",
			input:               "@#$%",
			allowedSpecialChars: ".-_",
			expected:            "",
		},
		{
			name:                "unicode characters removed",
			input:               "user_имя",
			allowedSpecialChars: ".-_",
			expected:            "user_",
		},
		{
			name:                "SQL injection attempt",
			input:               "user'; DROP TABLE",
			allowedSpecialChars: ".-_",
			expected:            "userDROPTABLE",
		},
		{
			name:                "trim whitespace",
			input:               "  user_name  ",
			allowedSpecialChars: ".-_",
			expected:            "user_name",
		},
		{
			name:                "different allowed chars",
			input:               "user@name.field_id",
			allowedSpecialChars: "@.",
			expected:            "user@name.fieldid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFieldName(tt.input, tt.allowedSpecialChars)
			if result != tt.expected {
				t.Errorf("SanitizeFieldName() = %q, want %q", result, tt.expected)
			}
		})
	}
}
