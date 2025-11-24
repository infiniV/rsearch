package schema

import "testing"

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"camelCase simple", "userName", "user_name"},
		{"camelCase multiple", "productCodeValue", "product_code_value"},
		{"PascalCase", "ProductCode", "product_code"},
		{"already snake_case", "user_name", "user_name"},
		{"single word", "user", "user"},
		{"with numbers", "user123Name", "user123_name"},
		{"acronym at end", "userID", "user_id"},
		{"acronym at start", "IDValue", "id_value"},
		{"consecutive capitals", "HTTPServer", "http_server"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"snake_case simple", "user_name", "userName"},
		{"snake_case multiple", "product_code_value", "productCodeValue"},
		{"already camelCase", "userName", "userName"},
		{"single word", "user", "user"},
		{"with numbers", "user_123_name", "user123Name"},
		{"leading underscore", "_user_name", "userName"},
		{"trailing underscore", "user_name_", "userName"},
		{"multiple underscores", "user__name", "userName"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"snake_case simple", "user_name", "UserName"},
		{"snake_case multiple", "product_code_value", "ProductCodeValue"},
		{"already PascalCase", "UserName", "UserName"},
		{"camelCase", "userName", "UserName"},
		{"single word", "user", "User"},
		{"with numbers", "user_123_name", "User123Name"},
		{"leading underscore", "_user_name", "UserName"},
		{"trailing underscore", "user_name_", "UserName"},
		{"multiple underscores", "user__name", "UserName"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
