package validation

import (
	"strings"
)

// SanitizeQuery removes dangerous patterns from a query string
// This provides defense in depth even though the translator uses parameterized queries
func SanitizeQuery(query string) string {
	if query == "" {
		return ""
	}

	// Remove null bytes
	query = strings.ReplaceAll(query, "\x00", "")

	// Remove control characters (except tab, newline, carriage return)
	var sanitized strings.Builder
	sanitized.Grow(len(query))
	for _, r := range query {
		if r >= 32 || r == '\t' || r == '\n' || r == '\r' {
			sanitized.WriteRune(r)
		}
	}
	query = sanitized.String()

	// Remove SQL comment patterns
	// Remove line comments (--) - everything after -- is removed
	if idx := strings.Index(query, "--"); idx != -1 {
		query = query[:idx]
	}

	// Remove block comments (/* */) - keep looping until all are removed (handles nested comments)
	for {
		startIdx := strings.Index(query, "/*")
		if startIdx == -1 {
			break
		}
		// Search for */ after the /*
		endIdx := strings.Index(query[startIdx+2:], "*/")
		if endIdx == -1 {
			// Unclosed comment, remove from start to end
			query = query[:startIdx]
			break
		}
		// Remove the comment block (startIdx + 2 + endIdx + 2 for the */)
		query = query[:startIdx] + query[startIdx+2+endIdx+2:]
	}

	// Trim excessive whitespace
	query = strings.TrimSpace(query)

	return query
}

// SanitizeFieldName normalizes a field name by removing invalid characters
// Only allows alphanumeric characters and the specified special characters
func SanitizeFieldName(name string, allowedSpecialChars string) string {
	if name == "" {
		return ""
	}

	// Trim whitespace first
	name = strings.TrimSpace(name)

	// Remove null bytes and control characters
	name = strings.ReplaceAll(name, "\x00", "")

	var sanitized strings.Builder
	sanitized.Grow(len(name))
	for _, r := range name {
		// Keep alphanumeric characters
		if isAlphanumeric(r) {
			sanitized.WriteRune(r)
		} else if strings.ContainsRune(allowedSpecialChars, r) {
			// Keep allowed special characters
			sanitized.WriteRune(r)
		}
		// Skip all other characters (including control chars)
	}

	return sanitized.String()
}
