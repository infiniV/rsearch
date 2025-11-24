package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserAndOperator(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "AND with &&",
			input: "field1:value1 && field2:value2",
		},
		{
			name:  "AND with keyword",
			input: "field1:value1 AND field2:value2",
		},
		{
			name:  "Multiple AND operations",
			input: "a:1 AND b:2 AND c:3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.Parse()
			require.NoError(t, err)
			require.NotNil(t, node)

			// Verify it's a binary operation
			binOp, ok := node.(*BinaryOp)
			require.True(t, ok, "expected BinaryOp node")
			assert.Equal(t, "AND", binOp.Op)
			assert.NotNil(t, binOp.Left)
			assert.NotNil(t, binOp.Right)
		})
	}
}

func TestParserOrOperator(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "OR with ||",
			input: "field1:value1 || field2:value2",
		},
		{
			name:  "OR with keyword",
			input: "field1:value1 OR field2:value2",
		},
		{
			name:  "Multiple OR operations",
			input: "a:1 OR b:2 OR c:3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.Parse()
			require.NoError(t, err)
			require.NotNil(t, node)

			// Verify it's a binary operation
			binOp, ok := node.(*BinaryOp)
			require.True(t, ok, "expected BinaryOp node")
			assert.Equal(t, "OR", binOp.Op)
			assert.NotNil(t, binOp.Left)
			assert.NotNil(t, binOp.Right)
		})
	}
}

func TestParserNotOperator(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "NOT with !",
			input: "!field1:value1",
		},
		{
			name:  "NOT with keyword",
			input: "NOT field1:value1",
		},
		{
			name:  "Double NOT",
			input: "NOT NOT field1:value1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.Parse()
			require.NoError(t, err)
			require.NotNil(t, node)

			// Verify it's a unary operation
			unaryOp, ok := node.(*UnaryOp)
			require.True(t, ok, "expected UnaryOp node")
			assert.Equal(t, "NOT", unaryOp.Op)
			assert.NotNil(t, unaryOp.Operand)
		})
	}
}

func TestParserRequiredOperator(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Required term",
			input: "+field1:value1",
		},
		{
			name:  "Multiple required terms",
			input: "+field1:value1 +field2:value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.Parse()
			require.NoError(t, err)
			require.NotNil(t, node)

			// For single required term
			if tt.name == "Required term" {
				unaryOp, ok := node.(*UnaryOp)
				require.True(t, ok, "expected UnaryOp node")
				assert.Equal(t, "+", unaryOp.Op)
				assert.NotNil(t, unaryOp.Operand)
			}
		})
	}
}

func TestParserProhibitedOperator(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Prohibited term",
			input: "-field1:value1",
		},
		{
			name:  "Multiple prohibited terms",
			input: "-field1:value1 -field2:value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.Parse()
			require.NoError(t, err)
			require.NotNil(t, node)

			// For single prohibited term
			if tt.name == "Prohibited term" {
				unaryOp, ok := node.(*UnaryOp)
				require.True(t, ok, "expected UnaryOp node")
				assert.Equal(t, "-", unaryOp.Op)
				assert.NotNil(t, unaryOp.Operand)
			}
		})
	}
}

func TestParserComplexExpression(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "OR and AND precedence",
			input: "(a:1 OR b:2) AND c:3",
		},
		{
			name:  "AND has higher precedence than OR",
			input: "a:1 OR b:2 AND c:3",
		},
		{
			name:  "NOT with AND",
			input: "NOT a:1 AND b:2",
		},
		{
			name:  "Required and prohibited",
			input: "+required:term -prohibited:term",
		},
		{
			name:  "Complex nested expression",
			input: "(a:1 OR b:2) AND (c:3 OR d:4)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.Parse()
			require.NoError(t, err, "should parse without error")
			require.NotNil(t, node, "should return a node")
		})
	}
}

func TestParserOperatorPrecedence(t *testing.T) {
	// Test: a OR b AND c should parse as a OR (b AND c)
	parser := NewParser("a:1 OR b:2 AND c:3")
	node, err := parser.Parse()
	require.NoError(t, err)
	require.NotNil(t, node)

	// Root should be OR
	rootOp, ok := node.(*BinaryOp)
	require.True(t, ok, "root should be BinaryOp")
	assert.Equal(t, "OR", rootOp.Op)

	// Right side should be AND
	rightOp, ok := rootOp.Right.(*BinaryOp)
	require.True(t, ok, "right should be BinaryOp")
	assert.Equal(t, "AND", rightOp.Op)
}

func TestParserFieldQuery(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedField string
		expectedValue string
	}{
		{
			name:          "Simple field query",
			input:         "field:value",
			expectedField: "field",
			expectedValue: "value",
		},
		{
			name:          "Field with number",
			input:         "age:25",
			expectedField: "age",
			expectedValue: "25",
		},
		{
			name:          "Field with quoted string",
			input:         "name:\"John Doe\"",
			expectedField: "name",
			expectedValue: "John Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			node, err := parser.Parse()
			require.NoError(t, err)
			require.NotNil(t, node)

			fieldQuery, ok := node.(*FieldQuery)
			require.True(t, ok, "expected FieldQuery node")
			assert.Equal(t, tt.expectedField, fieldQuery.Field)
			assert.Equal(t, tt.expectedValue, fieldQuery.Value.Value())
		})
	}
}

func TestParserErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Missing colon",
			input: "field value",
		},
		{
			name:  "Unclosed parenthesis",
			input: "(field:value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			_, err := parser.Parse()
			require.Error(t, err, "should return error for invalid input")
		})
	}
}
