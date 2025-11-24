package translator

import (
	"testing"

	"github.com/infiniv/rsearch/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslatorAndOperator(t *testing.T) {
	// Create test schema
	testSchema := schema.NewSchema("test_table", map[string]schema.Field{
		"field1": {Type: schema.TypeText, Column: "field1"},
		"field2": {Type: schema.TypeText, Column: "field2"},
	}, schema.SchemaOptions{})

	tests := []struct {
		name           string
		ast            Node
		expectedSQL    string
		expectedParams []interface{}
	}{
		{
			name: "Simple AND operation",
			ast: &BinaryOp{
				Op: "AND",
				Left: &FieldQuery{
					Field: "field1",
					Value: "value1",
				},
				Right: &FieldQuery{
					Field: "field2",
					Value: "value2",
				},
			},
			expectedSQL:    "field1 = $1 AND field2 = $2",
			expectedParams: []interface{}{"value1", "value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := NewPostgresTranslator()
			output, err := translator.Translate(tt.ast, testSchema)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedSQL, output.WhereClause)
			assert.Equal(t, tt.expectedParams, output.Parameters)
		})
	}
}

func TestTranslatorOrOperator(t *testing.T) {
	// Create test schema
	testSchema := schema.NewSchema("test_table", map[string]schema.Field{
		"field1": {Type: schema.TypeText, Column: "field1"},
		"field2": {Type: schema.TypeText, Column: "field2"},
	}, schema.SchemaOptions{})

	tests := []struct {
		name           string
		ast            Node
		expectedSQL    string
		expectedParams []interface{}
	}{
		{
			name: "Simple OR operation",
			ast: &BinaryOp{
				Op: "OR",
				Left: &FieldQuery{
					Field: "field1",
					Value: "value1",
				},
				Right: &FieldQuery{
					Field: "field2",
					Value: "value2",
				},
			},
			expectedSQL:    "field1 = $1 OR field2 = $2",
			expectedParams: []interface{}{"value1", "value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := NewPostgresTranslator()
			output, err := translator.Translate(tt.ast, testSchema)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedSQL, output.WhereClause)
			assert.Equal(t, tt.expectedParams, output.Parameters)
		})
	}
}

func TestTranslatorNotOperator(t *testing.T) {
	// Create test schema
	testSchema := schema.NewSchema("test_table", map[string]schema.Field{
		"field1": {Type: schema.TypeText, Column: "field1"},
	}, schema.SchemaOptions{})

	tests := []struct {
		name           string
		ast            Node
		expectedSQL    string
		expectedParams []interface{}
	}{
		{
			name: "NOT operation",
			ast: &UnaryOp{
				Op: "NOT",
				Operand: &FieldQuery{
					Field: "field1",
					Value: "value1",
				},
			},
			expectedSQL:    "NOT field1 = $1",
			expectedParams: []interface{}{"value1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := NewPostgresTranslator()
			output, err := translator.Translate(tt.ast, testSchema)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedSQL, output.WhereClause)
			assert.Equal(t, tt.expectedParams, output.Parameters)
		})
	}
}

func TestTranslatorRequiredOperator(t *testing.T) {
	// Create test schema
	testSchema := schema.NewSchema("test_table", map[string]schema.Field{
		"field1": {Type: schema.TypeText, Column: "field1"},
	}, schema.SchemaOptions{})

	tests := []struct {
		name           string
		ast            Node
		expectedSQL    string
		expectedParams []interface{}
	}{
		{
			name: "Required term (+)",
			ast: &UnaryOp{
				Op: "+",
				Operand: &FieldQuery{
					Field: "field1",
					Value: "value1",
				},
			},
			expectedSQL:    "field1 = $1",
			expectedParams: []interface{}{"value1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := NewPostgresTranslator()
			output, err := translator.Translate(tt.ast, testSchema)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedSQL, output.WhereClause)
			assert.Equal(t, tt.expectedParams, output.Parameters)
		})
	}
}

func TestTranslatorProhibitedOperator(t *testing.T) {
	// Create test schema
	testSchema := schema.NewSchema("test_table", map[string]schema.Field{
		"field1": {Type: schema.TypeText, Column: "field1"},
	}, schema.SchemaOptions{})

	tests := []struct {
		name           string
		ast            Node
		expectedSQL    string
		expectedParams []interface{}
	}{
		{
			name: "Prohibited term (-)",
			ast: &UnaryOp{
				Op: "-",
				Operand: &FieldQuery{
					Field: "field1",
					Value: "value1",
				},
			},
			expectedSQL:    "NOT field1 = $1",
			expectedParams: []interface{}{"value1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := NewPostgresTranslator()
			output, err := translator.Translate(tt.ast, testSchema)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedSQL, output.WhereClause)
			assert.Equal(t, tt.expectedParams, output.Parameters)
		})
	}
}

func TestTranslatorComplexExpressions(t *testing.T) {
	// Create test schema
	testSchema := schema.NewSchema("test_table", map[string]schema.Field{
		"a": {Type: schema.TypeText, Column: "a"},
		"b": {Type: schema.TypeText, Column: "b"},
		"c": {Type: schema.TypeText, Column: "c"},
	}, schema.SchemaOptions{})

	tests := []struct {
		name           string
		ast            Node
		expectedSQL    string
		expectedParams []interface{}
	}{
		{
			name: "(a OR b) AND c",
			ast: &BinaryOp{
				Op: "AND",
				Left: &BinaryOp{
					Op: "OR",
					Left: &FieldQuery{
						Field: "a",
						Value: "1",
					},
					Right: &FieldQuery{
						Field: "b",
						Value: "2",
					},
				},
				Right: &FieldQuery{
					Field: "c",
					Value: "3",
				},
			},
			expectedSQL:    "(a = $1 OR b = $2) AND c = $3",
			expectedParams: []interface{}{"1", "2", "3"},
		},
		{
			name: "NOT a AND b",
			ast: &BinaryOp{
				Op: "AND",
				Left: &UnaryOp{
					Op: "NOT",
					Operand: &FieldQuery{
						Field: "a",
						Value: "1",
					},
				},
				Right: &FieldQuery{
					Field: "b",
					Value: "2",
				},
			},
			expectedSQL:    "NOT a = $1 AND b = $2",
			expectedParams: []interface{}{"1", "2"},
		},
		{
			name: "+required -prohibited",
			ast: &BinaryOp{
				Op: "AND",
				Left: &UnaryOp{
					Op: "+",
					Operand: &FieldQuery{
						Field: "a",
						Value: "required",
					},
				},
				Right: &UnaryOp{
					Op: "-",
					Operand: &FieldQuery{
						Field: "b",
						Value: "prohibited",
					},
				},
			},
			expectedSQL:    "a = $1 AND NOT b = $2",
			expectedParams: []interface{}{"required", "prohibited"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := NewPostgresTranslator()
			output, err := translator.Translate(tt.ast, testSchema)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedSQL, output.WhereClause)
			assert.Equal(t, tt.expectedParams, output.Parameters)
		})
	}
}

func TestTranslatorOperatorPrecedence(t *testing.T) {
	// Create test schema
	testSchema := schema.NewSchema("test_table", map[string]schema.Field{
		"a": {Type: schema.TypeText, Column: "a"},
		"b": {Type: schema.TypeText, Column: "b"},
		"c": {Type: schema.TypeText, Column: "c"},
	}, schema.SchemaOptions{})

	tests := []struct {
		name           string
		ast            Node
		expectedSQL    string
		expectedParams []interface{}
	}{
		{
			name: "a OR (b AND c) - nested binary ops",
			ast: &BinaryOp{
				Op: "OR",
				Left: &FieldQuery{
					Field: "a",
					Value: "1",
				},
				Right: &BinaryOp{
					Op: "AND",
					Left: &FieldQuery{
						Field: "b",
						Value: "2",
					},
					Right: &FieldQuery{
						Field: "c",
						Value: "3",
					},
				},
			},
			expectedSQL:    "a = $1 OR (b = $2 AND c = $3)",
			expectedParams: []interface{}{"1", "2", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := NewPostgresTranslator()
			output, err := translator.Translate(tt.ast, testSchema)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedSQL, output.WhereClause)
			assert.Equal(t, tt.expectedParams, output.Parameters)
		})
	}
}
