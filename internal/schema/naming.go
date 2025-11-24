package schema

import (
	"strings"
	"unicode"
)

// ToSnakeCase converts a string to snake_case.
// Handles camelCase, PascalCase, and already snake_case strings.
func ToSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	result.Grow(len(s) + 5) // Allocate slightly more for underscores

	var prevLower, prevUnderscore bool
	for i, r := range s {
		if r == '_' {
			if !prevUnderscore && result.Len() > 0 {
				result.WriteRune('_')
				prevUnderscore = true
			}
			prevLower = false
			continue
		}

		if unicode.IsUpper(r) {
			// Add underscore before uppercase letter if:
			// 1. Not at the start
			// 2. Previous char was lowercase
			// 3. Next char is lowercase (for consecutive capitals like "HTTPServer")
			if result.Len() > 0 && !prevUnderscore {
				if prevLower {
					result.WriteRune('_')
				} else if i+1 < len(s) {
					next := rune(s[i+1])
					if unicode.IsLower(next) {
						result.WriteRune('_')
					}
				}
			}
			result.WriteRune(unicode.ToLower(r))
			prevLower = false
		} else {
			result.WriteRune(r)
			prevLower = unicode.IsLetter(r)
		}
		prevUnderscore = false
	}

	return result.String()
}

// ToCamelCase converts a string to camelCase.
// Handles snake_case and already camelCase strings.
func ToCamelCase(s string) string {
	if s == "" {
		return ""
	}

	// Remove leading and trailing underscores
	s = strings.Trim(s, "_")

	parts := strings.Split(s, "_")
	if len(parts) == 0 {
		return ""
	}

	var result strings.Builder
	result.Grow(len(s))

	// First part stays lowercase
	firstPart := strings.TrimSpace(parts[0])
	if firstPart != "" {
		result.WriteString(strings.ToLower(string(firstPart[0])))
		if len(firstPart) > 1 {
			result.WriteString(firstPart[1:])
		}
	}

	// Capitalize subsequent parts
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		result.WriteString(strings.ToUpper(string(part[0])))
		if len(part) > 1 {
			result.WriteString(part[1:])
		}
	}

	return result.String()
}

// ToPascalCase converts a string to PascalCase.
// Handles snake_case, camelCase, and already PascalCase strings.
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}

	// If already in camelCase or PascalCase without underscores, just capitalize first letter
	if !strings.Contains(s, "_") {
		if len(s) == 1 {
			return strings.ToUpper(s)
		}
		return strings.ToUpper(string(s[0])) + s[1:]
	}

	// Remove leading and trailing underscores
	s = strings.Trim(s, "_")

	parts := strings.Split(s, "_")
	if len(parts) == 0 {
		return ""
	}

	var result strings.Builder
	result.Grow(len(s))

	// Capitalize each part
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		result.WriteString(strings.ToUpper(string(part[0])))
		if len(part) > 1 {
			result.WriteString(part[1:])
		}
	}

	return result.String()
}
