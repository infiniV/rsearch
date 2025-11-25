package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/infiniv/rsearch/internal/config"
)

// Validator validates user input for security
type Validator struct {
	allowedSpecialChars string
	blockSqlKeywords    bool
	maxQueryLength      int
	maxFieldNameLength  int
	sqlKeywords         []string
	sqlComments         []string
}

// NewValidator creates a new validator with the given configuration
func NewValidator(cfg *config.SecurityConfig, limits *config.LimitsConfig) *Validator {
	return &Validator{
		allowedSpecialChars: cfg.AllowedSpecialChars,
		blockSqlKeywords:    cfg.BlockSqlKeywords,
		maxQueryLength:      limits.MaxQueryLength,
		maxFieldNameLength:  limits.MaxFieldNameLength,
		sqlKeywords: []string{
			"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "UNION",
			"TRUNCATE", "ALTER", "EXEC", "EXECUTE", "CREATE", "GRANT",
			"REVOKE", "DENY", "DECLARE", "CAST", "CONVERT", "SHUTDOWN",
		},
		sqlComments: []string{"--", "/*", "*/"},
	}
}

// ValidateQuery validates a query string for security issues
func (v *Validator) ValidateQuery(query string) error {
	// Allow empty queries
	if query == "" {
		return nil
	}

	// Check query length
	if len(query) > v.maxQueryLength {
		return fmt.Errorf("query exceeds maximum length of %d characters", v.maxQueryLength)
	}

	// Check for null bytes
	if strings.ContainsRune(query, '\x00') {
		return fmt.Errorf("query contains null byte")
	}

	// Check for control characters (except whitespace)
	for _, r := range query {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return fmt.Errorf("query contains control character at position")
		}
	}

	// Check for SQL keywords if enabled
	if v.blockSqlKeywords {
		upperQuery := strings.ToUpper(query)
		for _, keyword := range v.sqlKeywords {
			// Use word boundary matching to avoid false positives
			// Match keyword as a whole word
			pattern := `\b` + keyword + `\b`
			matched, _ := regexp.MatchString(pattern, upperQuery)
			if matched {
				return fmt.Errorf("query contains SQL keyword: %s", keyword)
			}
		}

		// Check for SQL comments
		for _, comment := range v.sqlComments {
			if strings.Contains(query, comment) {
				return fmt.Errorf("query contains SQL comment pattern: %s", comment)
			}
		}
	}

	return nil
}

// ValidateSchemaName validates a schema name for security and format
func (v *Validator) ValidateSchemaName(name string) error {
	if name == "" {
		return fmt.Errorf("schema name cannot be empty")
	}

	// Check length
	if len(name) > v.maxFieldNameLength {
		return fmt.Errorf("schema name exceeds maximum length of %d characters", v.maxFieldNameLength)
	}

	// Check for null bytes
	if strings.ContainsRune(name, '\x00') {
		return fmt.Errorf("schema name contains null byte")
	}

	// Validate characters: alphanumeric + allowed special chars
	for i, r := range name {
		if !isAlphanumeric(r) && !strings.ContainsRune(v.allowedSpecialChars, r) {
			return fmt.Errorf("schema name contains invalid character '%c' at position %d", r, i)
		}
	}

	return nil
}

// ValidateFieldName validates a field name for security and format
func (v *Validator) ValidateFieldName(name string) error {
	if name == "" {
		return fmt.Errorf("field name cannot be empty")
	}

	// Check length
	if len(name) > v.maxFieldNameLength {
		return fmt.Errorf("field name exceeds maximum length of %d characters", v.maxFieldNameLength)
	}

	// Check for null bytes
	if strings.ContainsRune(name, '\x00') {
		return fmt.Errorf("field name contains null byte")
	}

	// Validate characters: alphanumeric + allowed special chars
	for i, r := range name {
		if !isAlphanumeric(r) && !strings.ContainsRune(v.allowedSpecialChars, r) {
			return fmt.Errorf("field name contains invalid character '%c' at position %d", r, i)
		}
	}

	return nil
}

// isAlphanumeric checks if a rune is alphanumeric (a-z, A-Z, 0-9)
func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
