package rsearch

// Version is the current version of rsearch
const Version = "1.0.0"

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// ErrorResponse represents the standard error response format
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorInfo   `json:"details,omitempty"`
	Query   string        `json:"query,omitempty"`
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Position int    `json:"position,omitempty"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
	Message  string `json:"message"`
}

// Error codes
const (
	ErrorCodeParseError      = "PARSE_ERROR"
	ErrorCodeSchemaNotFound  = "SCHEMA_NOT_FOUND"
	ErrorCodeFieldNotFound   = "FIELD_NOT_FOUND"
	ErrorCodeTypeMismatch    = "TYPE_MISMATCH"
	ErrorCodeFeatureDisabled = "FEATURE_DISABLED"
	ErrorCodeInvalidRange    = "INVALID_RANGE"
	ErrorCodeUnsupportedSyntax = "UNSUPPORTED_SYNTAX"
	ErrorCodeSchemaExists    = "SCHEMA_EXISTS"
	ErrorCodeInvalidSchema   = "INVALID_SCHEMA"
	ErrorCodeInternalError   = "INTERNAL_ERROR"
)
