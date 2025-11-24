package schema

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	// columnNameRegex validates column names: alphanumeric and underscores only
	columnNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

	// Valid naming conventions
	validNamingConventions = map[string]bool{
		"snake_case": true,
		"camelCase":  true,
		"PascalCase": true,
		"none":       true,
		"":           true, // Empty is treated as "none"
	}
)

// ValidateSchema validates a schema for correctness
func ValidateSchema(s *Schema) error {
	if s == nil {
		return errors.New("schema is nil")
	}

	// Validate schema name
	if s.Name == "" {
		return errors.New("schema name cannot be empty")
	}

	// Validate fields exist
	if len(s.Fields) == 0 {
		return errors.New("schema must have at least one field")
	}

	// Track aliases and field names to detect duplicates
	seenAliases := make(map[string]string) // alias -> field name
	fieldNames := make(map[string]bool)

	// Collect field names first
	for fieldName := range s.Fields {
		if fieldName == "" {
			return errors.New("field name cannot be empty")
		}
		fieldNames[fieldName] = true
		fieldNames[strings.ToLower(fieldName)] = true // For case-insensitive check
	}

	// Validate each field
	for fieldName, field := range s.Fields {
		// Validate field type
		if !IsValidFieldType(field.Type) {
			return fmt.Errorf("invalid field type %q for field %q", field.Type, fieldName)
		}

		// Validate column name if specified
		if field.Column != "" {
			if !columnNameRegex.MatchString(field.Column) {
				return fmt.Errorf("invalid column name %q for field %q: must contain only alphanumeric characters and underscores", field.Column, fieldName)
			}
		}

		// Validate aliases
		for _, alias := range field.Aliases {
			if alias == "" {
				return fmt.Errorf("empty alias found for field %q", fieldName)
			}

			// Check if alias matches any field name
			normalizedAlias := strings.ToLower(alias)
			if fieldNames[normalizedAlias] {
				return fmt.Errorf("alias %q for field %q conflicts with an existing field name", alias, fieldName)
			}

			// Check for duplicate aliases
			if existingField, exists := seenAliases[normalizedAlias]; exists {
				return fmt.Errorf("duplicate alias %q found in fields %q and %q", alias, existingField, fieldName)
			}
			seenAliases[normalizedAlias] = fieldName
		}
	}

	// Validate naming convention
	if !validNamingConventions[s.Options.NamingConvention] {
		return fmt.Errorf("invalid naming convention %q: must be one of: snake_case, camelCase, PascalCase, none", s.Options.NamingConvention)
	}

	// Validate default field exists
	if s.Options.DefaultField != "" {
		if _, exists := s.Fields[s.Options.DefaultField]; !exists {
			return fmt.Errorf("default field %q does not exist in schema", s.Options.DefaultField)
		}
	}

	return nil
}
