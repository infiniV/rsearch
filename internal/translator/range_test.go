package translator

import (
	"testing"

	"github.com/infiniv/rsearch/internal/parser"
	"github.com/infiniv/rsearch/internal/schema"
)

// TestRangeQueryTranslation tests comprehensive range query translation
func TestRangeQueryTranslation(t *testing.T) {
	// Create a test schema
	testSchema := schema.NewSchema("test_schema", map[string]schema.Field{
		"age":         {Type: schema.TypeInteger, Column: "age"},
		"price":       {Type: schema.TypeFloat, Column: "price"},
		"created":     {Type: schema.TypeText, Column: "created_at"},
		"score":       {Type: schema.TypeInteger, Column: "score"},
		"rating":      {Type: schema.TypeFloat, Column: "rating"},
		"updated":     {Type: schema.TypeText, Column: "updated_at"},
		"temperature": {Type: schema.TypeInteger, Column: "temp"},
		"name":        {Type: schema.TypeText, Column: "name"},
		"count":       {Type: schema.TypeInteger, Column: "count"},
		"salary":      {Type: schema.TypeFloat, Column: "salary"},
	}, schema.SchemaOptions{})

	tests := []struct {
		name           string
		input          string
		wantSQL        string
		wantParamCount int
		wantParams     []interface{}
	}{
		{
			name:           "Inclusive both sides - BETWEEN",
			input:          "age:[18 TO 65]",
			wantSQL:        "age BETWEEN $1 AND $2",
			wantParamCount: 2,
			wantParams:     []interface{}{"18", "65"},
		},
		{
			name:           "Exclusive both sides - comparison operators",
			input:          "price:{100 TO 1000}",
			wantSQL:        "price > $1 AND price < $2",
			wantParamCount: 2,
			wantParams:     []interface{}{"100", "1000"},
		},
		{
			name:           "Mixed - inclusive start, exclusive end",
			input:          "score:[50 TO 100}",
			wantSQL:        "score >= $1 AND score < $2",
			wantParamCount: 2,
			wantParams:     []interface{}{"50", "100"},
		},
		{
			name:           "Mixed - exclusive start, inclusive end",
			input:          "rating:{0 TO 5]",
			wantSQL:        "rating > $1 AND rating <= $2",
			wantParamCount: 2,
			wantParams:     []interface{}{"0", "5"},
		},
		{
			name:           "Greater than or equal - comparison syntax",
			input:          "age:>=18",
			wantSQL:        "age >= $1",
			wantParamCount: 1,
			wantParams:     []interface{}{"18"},
		},
		{
			name:           "Greater than - comparison syntax",
			input:          "price:>100",
			wantSQL:        "price > $1",
			wantParamCount: 1,
			wantParams:     []interface{}{"100"},
		},
		{
			name:           "Less than or equal - comparison syntax",
			input:          "age:<=65",
			wantSQL:        "age <= $1",
			wantParamCount: 1,
			wantParams:     []interface{}{"65"},
		},
		{
			name:           "Less than - comparison syntax",
			input:          "score:<100",
			wantSQL:        "score < $1",
			wantParamCount: 1,
			wantParams:     []interface{}{"100"},
		},
		{
			name:           "Date range - inclusive",
			input:          "created:[2024-01-01 TO 2024-12-31]",
			wantSQL:        "created_at BETWEEN $1 AND $2",
			wantParamCount: 2,
			wantParams:     []interface{}{"2024-01-01", "2024-12-31"},
		},
		{
			name:           "Date range - exclusive",
			input:          "updated:{2024-01-01 TO 2024-12-31}",
			wantSQL:        "updated_at > $1 AND updated_at < $2",
			wantParamCount: 2,
			wantParams:     []interface{}{"2024-01-01", "2024-12-31"},
		},
		{
			name:           "Negative numbers - inclusive",
			input:          `temperature:["-10" TO "40"]`,
			wantSQL:        "temp BETWEEN $1 AND $2",
			wantParamCount: 2,
			wantParams:     []interface{}{"-10", "40"},
		},
		{
			name:           "String range - alphabetical",
			input:          "name:[alice TO zoe]",
			wantSQL:        "name BETWEEN $1 AND $2",
			wantParamCount: 2,
			wantParams:     []interface{}{"alice", "zoe"},
		},
		{
			name:           "Unbounded range - open end",
			input:          "price:[100 TO *]",
			wantSQL:        "price >= $1",
			wantParamCount: 1,
			wantParams:     []interface{}{"100"},
		},
		{
			name:           "Unbounded range - open start",
			input:          "age:[* TO 18]",
			wantSQL:        "age <= $1",
			wantParamCount: 1,
			wantParams:     []interface{}{"18"},
		},
		{
			name:           "Unbounded range - open start, exclusive end",
			input:          "score:[* TO 100}",
			wantSQL:        "score < $1",
			wantParamCount: 1,
			wantParams:     []interface{}{"100"},
		},
		{
			name:           "Zero value ranges",
			input:          "count:[0 TO 100]",
			wantSQL:        "count BETWEEN $1 AND $2",
			wantParamCount: 2,
			wantParams:     []interface{}{"0", "100"},
		},
		{
			name:           "Decimal values",
			input:          "rating:[0.0 TO 5.0]",
			wantSQL:        "rating BETWEEN $1 AND $2",
			wantParamCount: 2,
			wantParams:     []interface{}{"0.0", "5.0"},
		},
		{
			name:           "Large numbers",
			input:          "salary:[50000 TO 150000]",
			wantSQL:        "salary BETWEEN $1 AND $2",
			wantParamCount: 2,
			wantParams:     []interface{}{"50000", "150000"},
		},
		{
			name:           "Same start and end - inclusive",
			input:          "age:[18 TO 18]",
			wantSQL:        "age BETWEEN $1 AND $2",
			wantParamCount: 2,
			wantParams:     []interface{}{"18", "18"},
		},
		{
			name:           "Same start and end - exclusive",
			input:          "age:{18 TO 18}",
			wantSQL:        "age > $1 AND age < $2",
			wantParamCount: 2,
			wantParams:     []interface{}{"18", "18"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the query
			p := parser.NewParser(tt.input)
			ast, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			// Translate to SQL
			translator := NewPostgresTranslator()
			output, err := translator.Translate(ast, testSchema)
			if err != nil {
				t.Fatalf("Translation error: %v", err)
			}

			// Check SQL output
			if output.WhereClause != tt.wantSQL {
				t.Errorf("SQL mismatch:\ngot:  %q\nwant: %q", output.WhereClause, tt.wantSQL)
			}

			// Check parameter count
			if len(output.Parameters) != tt.wantParamCount {
				t.Errorf("Parameter count = %d, want %d", len(output.Parameters), tt.wantParamCount)
			}

			// Check parameter values
			if len(tt.wantParams) > 0 {
				for i, wantParam := range tt.wantParams {
					if i >= len(output.Parameters) {
						t.Errorf("Missing parameter at index %d", i)
						continue
					}
					if output.Parameters[i] != wantParam {
						t.Errorf("Parameter[%d] = %v, want %v", i, output.Parameters[i], wantParam)
					}
				}
			}
		})
	}
}

// TestRangeQueryWithBooleanOperators tests range queries combined with AND/OR
func TestRangeQueryWithBooleanOperators(t *testing.T) {
	testSchema := schema.NewSchema("test_schema", map[string]schema.Field{
		"age":        {Type: schema.TypeInteger, Column: "age"},
		"salary":     {Type: schema.TypeFloat, Column: "salary"},
		"experience": {Type: schema.TypeInteger, Column: "experience"},
		"status":     {Type: schema.TypeText, Column: "status"},
		"price":      {Type: schema.TypeFloat, Column: "price"},
		"category":   {Type: schema.TypeText, Column: "category"},
	}, schema.SchemaOptions{})

	tests := []struct {
		name           string
		input          string
		wantSQL        string
		wantParamCount int
	}{
		{
			name:           "Range AND field query",
			input:          "age:[18 TO 65] AND status:active",
			wantSQL:        "age BETWEEN $1 AND $2 AND status = $3",
			wantParamCount: 3,
		},
		{
			name:           "Range OR range",
			input:          "age:[18 TO 30] OR age:[60 TO 100]",
			wantSQL:        "age BETWEEN $1 AND $2 OR age BETWEEN $3 AND $4",
			wantParamCount: 4,
		},
		{
			name:           "Multiple ranges with AND",
			input:          "age:[18 TO 65] AND salary:[50000 TO 150000] AND experience:[2 TO 10]",
			wantSQL:        "(age BETWEEN $1 AND $2 AND salary BETWEEN $3 AND $4) AND experience BETWEEN $5 AND $6",
			wantParamCount: 6,
		},
		{
			name:           "Comparison operator with field query",
			input:          "price:>=100 AND category:electronics",
			wantSQL:        "price >= $1 AND category = $2",
			wantParamCount: 2,
		},
		{
			name:           "Mixed range notations with OR",
			input:          "age:[18 TO 30] OR age:{60 TO 100}",
			wantSQL:        "age BETWEEN $1 AND $2 OR age > $3 AND age < $4",
			wantParamCount: 4,
		},
		{
			name:           "Exclusive range AND inclusive range",
			input:          "price:{100 TO 500} AND salary:[50000 TO 100000]",
			wantSQL:        "price > $1 AND price < $2 AND salary BETWEEN $3 AND $4",
			wantParamCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the query
			p := parser.NewParser(tt.input)
			ast, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			// Translate to SQL
			translator := NewPostgresTranslator()
			output, err := translator.Translate(ast, testSchema)
			if err != nil {
				t.Fatalf("Translation error: %v", err)
			}

			// Check SQL output
			if output.WhereClause != tt.wantSQL {
				t.Errorf("SQL mismatch:\ngot:  %q\nwant: %q", output.WhereClause, tt.wantSQL)
			}

			// Check parameter count
			if len(output.Parameters) != tt.wantParamCount {
				t.Errorf("Parameter count = %d, want %d", len(output.Parameters), tt.wantParamCount)
			}
		})
	}
}

// TestRangeQueryParameterTypes tests that parameter types are correctly set
func TestRangeQueryParameterTypes(t *testing.T) {
	testSchema := schema.NewSchema("test_schema", map[string]schema.Field{
		"age":   {Type: schema.TypeInteger, Column: "age"},
		"price": {Type: schema.TypeFloat, Column: "price"},
		"name":  {Type: schema.TypeText, Column: "name"},
	}, schema.SchemaOptions{})

	tests := []struct {
		name      string
		input     string
		wantTypes []string
	}{
		{
			name:      "Integer range",
			input:     "age:[18 TO 65]",
			wantTypes: []string{"integer", "integer"},
		},
		{
			name:      "Float range",
			input:     "price:[100.0 TO 999.99]",
			wantTypes: []string{"float", "float"},
		},
		{
			name:      "String range",
			input:     "name:[alice TO zoe]",
			wantTypes: []string{"text", "text"},
		},
		{
			name:      "Comparison operator - integer",
			input:     "age:>=18",
			wantTypes: []string{"integer"},
		},
		{
			name:      "Comparison operator - float",
			input:     "price:<=999.99",
			wantTypes: []string{"float"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the query
			p := parser.NewParser(tt.input)
			ast, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			// Translate to SQL
			translator := NewPostgresTranslator()
			output, err := translator.Translate(ast, testSchema)
			if err != nil {
				t.Fatalf("Translation error: %v", err)
			}

			// Check parameter types
			if len(output.ParameterTypes) != len(tt.wantTypes) {
				t.Fatalf("Parameter type count = %d, want %d", len(output.ParameterTypes), len(tt.wantTypes))
			}

			for i, wantType := range tt.wantTypes {
				if output.ParameterTypes[i] != wantType {
					t.Errorf("ParameterType[%d] = %q, want %q", i, output.ParameterTypes[i], wantType)
				}
			}
		})
	}
}

// TestRangeQueryFieldMapping tests that field names are correctly mapped to column names
func TestRangeQueryFieldMapping(t *testing.T) {
	testSchema := schema.NewSchema("test_schema", map[string]schema.Field{
		"userAge":      {Type: schema.TypeInteger, Column: "user_age"},
		"productPrice": {Type: schema.TypeFloat, Column: "product_price"},
		"createdAt":    {Type: schema.TypeText, Column: "created_timestamp"},
	}, schema.SchemaOptions{})

	tests := []struct {
		name    string
		input   string
		wantSQL string
	}{
		{
			name:    "CamelCase to snake_case",
			input:   "userAge:[18 TO 65]",
			wantSQL: "user_age BETWEEN $1 AND $2",
		},
		{
			name:    "Different column name",
			input:   "productPrice:[100 TO 1000]",
			wantSQL: "product_price BETWEEN $1 AND $2",
		},
		{
			name:    "Timestamp field mapping",
			input:   "createdAt:[2024-01-01 TO 2024-12-31]",
			wantSQL: "created_timestamp BETWEEN $1 AND $2",
		},
		{
			name:    "Comparison with field mapping",
			input:   "userAge:>=18",
			wantSQL: "user_age >= $1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the query
			p := parser.NewParser(tt.input)
			ast, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			// Translate to SQL
			translator := NewPostgresTranslator()
			output, err := translator.Translate(ast, testSchema)
			if err != nil {
				t.Fatalf("Translation error: %v", err)
			}

			// Check SQL output
			if output.WhereClause != tt.wantSQL {
				t.Errorf("SQL mismatch:\ngot:  %q\nwant: %q", output.WhereClause, tt.wantSQL)
			}
		})
	}
}

// TestRangeQueryErrors tests error conditions
func TestRangeQueryErrors(t *testing.T) {
	testSchema := schema.NewSchema("test_schema", map[string]schema.Field{
		"age": {Type: schema.TypeInteger, Column: "age"},
	}, schema.SchemaOptions{})

	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "Non-existent field",
			input:     "nonexistent:[1 TO 10]",
			wantError: true,
		},
		{
			name:      "Valid query",
			input:     "age:[18 TO 65]",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the query
			p := parser.NewParser(tt.input)
			ast, err := p.Parse()
			if err != nil {
				if !tt.wantError {
					t.Fatalf("Unexpected parse error: %v", err)
				}
				return
			}

			// Translate to SQL
			translator := NewPostgresTranslator()
			_, err = translator.Translate(ast, testSchema)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
