package translator

import (
	"testing"

	"github.com/infiniv/rsearch/internal/parser"
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
		ast            parser.Node
		expectedSQL    string
		expectedParams []interface{}
	}{
		{
			name: "Simple AND operation",
			ast: &parser.BinaryOp{
				Op: "AND",
				Left: &parser.FieldQuery{
					Field: "field1",
					Value: &parser.TermValue{Term: "value1", Pos: parser.Position{}},
				},
				Right: &parser.FieldQuery{
					Field: "field2",
					Value: &parser.TermValue{Term: "value2", Pos: parser.Position{}},
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
		ast            parser.Node
		expectedSQL    string
		expectedParams []interface{}
	}{
		{
			name: "Simple OR operation",
			ast: &parser.BinaryOp{
				Op: "OR",
				Left: &parser.FieldQuery{
					Field: "field1",
					Value: &parser.TermValue{Term: "value1", Pos: parser.Position{}},
				},
				Right: &parser.FieldQuery{
					Field: "field2",
					Value: &parser.TermValue{Term: "value2", Pos: parser.Position{}},
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
		ast            parser.Node
		expectedSQL    string
		expectedParams []interface{}
	}{
		{
			name: "NOT operation",
			ast: &parser.UnaryOp{
				Op: "NOT",
				Operand: &parser.FieldQuery{
					Field: "field1",
					Value: &parser.TermValue{Term: "value1", Pos: parser.Position{}},
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
		ast            parser.Node
		expectedSQL    string
		expectedParams []interface{}
	}{
		{
			name: "Required term (+)",
			ast: &parser.UnaryOp{
				Op: "+",
				Operand: &parser.FieldQuery{
					Field: "field1",
					Value: &parser.TermValue{Term: "value1", Pos: parser.Position{}},
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
		ast            parser.Node
		expectedSQL    string
		expectedParams []interface{}
	}{
		{
			name: "Prohibited term (-)",
			ast: &parser.UnaryOp{
				Op: "-",
				Operand: &parser.FieldQuery{
					Field: "field1",
					Value: &parser.TermValue{Term: "value1", Pos: parser.Position{}},
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
		ast            parser.Node
		expectedSQL    string
		expectedParams []interface{}
	}{
		{
			name: "(a OR b) AND c",
			ast: &parser.BinaryOp{
				Op: "AND",
				Left: &parser.BinaryOp{
					Op: "OR",
					Left: &parser.FieldQuery{
						Field: "a",
						Value: &parser.TermValue{Term: "1", Pos: parser.Position{}},
					},
					Right: &parser.FieldQuery{
						Field: "b",
						Value: &parser.TermValue{Term: "2", Pos: parser.Position{}},
					},
				},
				Right: &parser.FieldQuery{
					Field: "c",
					Value: &parser.TermValue{Term: "3", Pos: parser.Position{}},
				},
			},
			expectedSQL:    "(a = $1 OR b = $2) AND c = $3",
			expectedParams: []interface{}{"1", "2", "3"},
		},
		{
			name: "NOT a AND b",
			ast: &parser.BinaryOp{
				Op: "AND",
				Left: &parser.UnaryOp{
					Op: "NOT",
					Operand: &parser.FieldQuery{
						Field: "a",
						Value: &parser.TermValue{Term: "1", Pos: parser.Position{}},
					},
				},
				Right: &parser.FieldQuery{
					Field: "b",
					Value: &parser.TermValue{Term: "2", Pos: parser.Position{}},
				},
			},
			expectedSQL:    "NOT a = $1 AND b = $2",
			expectedParams: []interface{}{"1", "2"},
		},
		{
			name: "+required -prohibited",
			ast: &parser.BinaryOp{
				Op: "AND",
				Left: &parser.UnaryOp{
					Op: "+",
					Operand: &parser.FieldQuery{
						Field: "a",
						Value: &parser.TermValue{Term: "required", Pos: parser.Position{}},
					},
				},
				Right: &parser.UnaryOp{
					Op: "-",
					Operand: &parser.FieldQuery{
						Field: "b",
						Value: &parser.TermValue{Term: "prohibited", Pos: parser.Position{}},
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
		ast            parser.Node
		expectedSQL    string
		expectedParams []interface{}
	}{
		{
			name: "a OR (b AND c) - nested binary ops",
			ast: &parser.BinaryOp{
				Op: "OR",
				Left: &parser.FieldQuery{
					Field: "a",
					Value: &parser.TermValue{Term: "1", Pos: parser.Position{}},
				},
				Right: &parser.BinaryOp{
					Op: "AND",
					Left: &parser.FieldQuery{
						Field: "b",
						Value: &parser.TermValue{Term: "2", Pos: parser.Position{}},
					},
					Right: &parser.FieldQuery{
						Field: "c",
						Value: &parser.TermValue{Term: "3", Pos: parser.Position{}},
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
