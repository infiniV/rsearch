package rsearch

import (
	"encoding/json"
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
	if Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", Version, "1.0.0")
	}
}

func TestErrorCodes(t *testing.T) {
	codes := []string{
		ErrorCodeParseError,
		ErrorCodeSchemaNotFound,
		ErrorCodeFieldNotFound,
		ErrorCodeTypeMismatch,
		ErrorCodeFeatureDisabled,
		ErrorCodeInvalidRange,
		ErrorCodeUnsupportedSyntax,
		ErrorCodeSchemaExists,
		ErrorCodeInvalidSchema,
		ErrorCodeInternalError,
	}

	// Verify all codes are non-empty
	for _, code := range codes {
		if code == "" {
			t.Error("Error code should not be empty")
		}
	}

	// Verify codes are unique
	seen := make(map[string]bool)
	for _, code := range codes {
		if seen[code] {
			t.Errorf("Duplicate error code: %s", code)
		}
		seen[code] = true
	}
}

func TestHealthResponseJSON(t *testing.T) {
	response := HealthResponse{
		Status:  "healthy",
		Version: Version,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal HealthResponse: %v", err)
	}

	var decoded HealthResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal HealthResponse: %v", err)
	}

	if decoded.Status != response.Status {
		t.Errorf("Status = %q, want %q", decoded.Status, response.Status)
	}
	if decoded.Version != response.Version {
		t.Errorf("Version = %q, want %q", decoded.Version, response.Version)
	}
}

func TestErrorResponseJSON(t *testing.T) {
	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    ErrorCodeParseError,
			Message: "Unexpected token",
			Query:   "field:value AND",
			Details: []ErrorInfo{
				{
					Position: 15,
					Line:     1,
					Column:   16,
					Message:  "Expected term after AND",
				},
			},
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal ErrorResponse: %v", err)
	}

	var decoded ErrorResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ErrorResponse: %v", err)
	}

	if decoded.Error.Code != response.Error.Code {
		t.Errorf("Error.Code = %q, want %q", decoded.Error.Code, response.Error.Code)
	}
	if decoded.Error.Message != response.Error.Message {
		t.Errorf("Error.Message = %q, want %q", decoded.Error.Message, response.Error.Message)
	}
	if decoded.Error.Query != response.Error.Query {
		t.Errorf("Error.Query = %q, want %q", decoded.Error.Query, response.Error.Query)
	}
	if len(decoded.Error.Details) != 1 {
		t.Fatalf("Error.Details length = %d, want 1", len(decoded.Error.Details))
	}
	if decoded.Error.Details[0].Position != 15 {
		t.Errorf("Error.Details[0].Position = %d, want 15", decoded.Error.Details[0].Position)
	}
}

func TestErrorResponseWithoutDetails(t *testing.T) {
	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    ErrorCodeSchemaNotFound,
			Message: "Schema 'unknown' not found",
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal ErrorResponse: %v", err)
	}

	// Details should be omitted when empty
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	errorObj := raw["error"].(map[string]interface{})
	if _, exists := errorObj["details"]; exists {
		t.Error("details should be omitted when empty")
	}
	if _, exists := errorObj["query"]; exists {
		t.Error("query should be omitted when empty")
	}
}

func TestErrorInfoJSON(t *testing.T) {
	info := ErrorInfo{
		Position: 10,
		Line:     2,
		Column:   5,
		Message:  "Invalid character",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal ErrorInfo: %v", err)
	}

	var decoded ErrorInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ErrorInfo: %v", err)
	}

	if decoded.Position != info.Position {
		t.Errorf("Position = %d, want %d", decoded.Position, info.Position)
	}
	if decoded.Line != info.Line {
		t.Errorf("Line = %d, want %d", decoded.Line, info.Line)
	}
	if decoded.Column != info.Column {
		t.Errorf("Column = %d, want %d", decoded.Column, info.Column)
	}
	if decoded.Message != info.Message {
		t.Errorf("Message = %q, want %q", decoded.Message, info.Message)
	}
}
