package testhelper

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/infiniv/rsearch/internal/schema"
	"github.com/stretchr/testify/require"
)

// TestCase represents a test case from testcases.json
type TestCase struct {
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Query       string   `json:"query"`
	Schema      string   `json:"schema"`
	Expected    Expected `json:"expected"`
}

// Expected represents expected translation output
type Expected struct {
	SQL            string                 `json:"sql"`
	Parameters     []interface{}          `json:"parameters"`
	ParameterTypes []string               `json:"parameterTypes"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// LoadTestCases loads test cases from JSON file
func LoadTestCases(t *testing.T, path string) []TestCase {
	data, err := os.ReadFile(path)
	require.NoError(t, err, "Failed to read test cases file")

	var cases []TestCase
	err = json.Unmarshal(data, &cases)
	require.NoError(t, err, "Failed to parse test cases JSON")

	return cases
}

// LoadSchemas loads schema definitions from JSON file
func LoadSchemas(t *testing.T, path string) map[string]*schema.Schema {
	data, err := os.ReadFile(path)
	require.NoError(t, err, "Failed to read schemas file")

	var rawSchemas map[string]struct {
		Name    string                    `json:"name"`
		Fields  map[string]schema.Field   `json:"fields"`
		Options schema.SchemaOptions      `json:"options"`
	}
	err = json.Unmarshal(data, &rawSchemas)
	require.NoError(t, err, "Failed to parse schemas JSON")

	schemas := make(map[string]*schema.Schema)
	for name, raw := range rawSchemas {
		schemas[name] = schema.NewSchema(raw.Name, raw.Fields, raw.Options)
	}

	return schemas
}

// SetupTestRegistry creates a schema registry with test schemas
func SetupTestRegistry(t *testing.T) *schema.Registry {
	registry := schema.NewRegistry()
	schemas := LoadSchemas(t, "../tests/schemas.json")

	for _, s := range schemas {
		err := registry.Register(s)
		require.NoError(t, err, "Failed to register test schema")
	}

	return registry
}
