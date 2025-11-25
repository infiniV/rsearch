package testhelper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTestCases(t *testing.T) {
	// Find the testcases.json file
	testCasesPath := filepath.Join("..", "testcases.json")
	if _, err := os.Stat(testCasesPath); os.IsNotExist(err) {
		t.Skip("testcases.json not found, skipping")
	}

	cases := LoadTestCases(t, testCasesPath)
	if len(cases) == 0 {
		t.Error("expected at least one test case")
	}

	// Verify test case structure
	for i, tc := range cases {
		if tc.Category == "" {
			t.Errorf("test case %d: Category is empty", i)
		}
		if tc.Description == "" {
			t.Errorf("test case %d: Description is empty", i)
		}
		if tc.Query == "" {
			t.Errorf("test case %d: Query is empty", i)
		}
		if tc.Schema == "" {
			t.Errorf("test case %d: Schema is empty", i)
		}
		if tc.Expected.SQL == "" {
			t.Errorf("test case %d: Expected.SQL is empty", i)
		}
	}
}

func TestLoadSchemas(t *testing.T) {
	// Find the schemas.json file
	schemasPath := filepath.Join("..", "schemas.json")
	if _, err := os.Stat(schemasPath); os.IsNotExist(err) {
		t.Skip("schemas.json not found, skipping")
	}

	schemas := LoadSchemas(t, schemasPath)
	if len(schemas) == 0 {
		t.Error("expected at least one schema")
	}

	// Verify schema structure
	for name, s := range schemas {
		if name == "" {
			t.Error("schema name is empty")
		}
		if s == nil {
			t.Errorf("schema %q is nil", name)
		}
	}
}

func TestTestCaseStruct(t *testing.T) {
	tc := TestCase{
		Category:    "test",
		Description: "test description",
		Query:       "field:value",
		Schema:      "products",
		Expected: Expected{
			SQL:            "field = $1",
			Parameters:     []interface{}{"value"},
			ParameterTypes: []string{"text"},
		},
	}

	if tc.Category != "test" {
		t.Error("Category not set correctly")
	}
	if tc.Description != "test description" {
		t.Error("Description not set correctly")
	}
	if tc.Query != "field:value" {
		t.Error("Query not set correctly")
	}
	if tc.Schema != "products" {
		t.Error("Schema not set correctly")
	}
	if tc.Expected.SQL != "field = $1" {
		t.Error("Expected.SQL not set correctly")
	}
}

func TestExpectedStruct(t *testing.T) {
	expected := Expected{
		SQL:            "field = $1",
		Parameters:     []interface{}{"value", 123},
		ParameterTypes: []string{"text", "integer"},
		Metadata: map[string]interface{}{
			"boost": float64(2),
		},
	}

	if expected.SQL != "field = $1" {
		t.Error("SQL not set correctly")
	}
	if len(expected.Parameters) != 2 {
		t.Error("Parameters length incorrect")
	}
	if len(expected.ParameterTypes) != 2 {
		t.Error("ParameterTypes length incorrect")
	}
	if expected.Metadata["boost"] != float64(2) {
		t.Error("Metadata not set correctly")
	}
}
