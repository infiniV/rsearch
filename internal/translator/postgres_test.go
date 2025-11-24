package translator

import (
	"testing"

	"github.com/infiniv/rsearch/internal/parser"
	"github.com/infiniv/rsearch/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresTranslator_DatabaseType(t *testing.T) {
	translator := NewPostgresTranslator()
	assert.Equal(t, "postgres", translator.DatabaseType())
}

func TestPostgresTranslator_SimpleFieldQuery(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText, Column: "product_code"},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "product_code",
		Value: &parser.TermValue{Term: "13w42"},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "sql", output.Type)
	assert.Equal(t, "product_code = $1", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "13w42", output.Parameters[0])
	assert.Equal(t, []string{"text"}, output.ParameterTypes)
}

func TestPostgresTranslator_NumberFieldQuery(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"rod_length": {Type: schema.TypeInteger, Column: "rod_length"},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "rod_length",
		Value: &parser.NumberValue{Number: "100"},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "rod_length = $1", output.WhereClause)
	assert.Equal(t, "100", output.Parameters[0])
	assert.Equal(t, []string{"integer"}, output.ParameterTypes)
}

func TestPostgresTranslator_BooleanAND(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText, Column: "product_code"},
		"region":       {Type: schema.TypeText, Column: "region"},
	}, schema.SchemaOptions{})

	ast := &parser.BinaryOp{
		Op: "AND",
		Left: &parser.FieldQuery{
			Field: "product_code",
			Value: &parser.TermValue{Term: "13w42"},
		},
		Right: &parser.FieldQuery{
			Field: "region",
			Value: &parser.TermValue{Term: "ca"},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "product_code = $1 AND region = $2", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "13w42", output.Parameters[0])
	assert.Equal(t, "ca", output.Parameters[1])
	assert.Equal(t, []string{"text", "text"}, output.ParameterTypes)
}

func TestPostgresTranslator_BooleanOR(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"region": {Type: schema.TypeText, Column: "region"},
	}, schema.SchemaOptions{})

	ast := &parser.BinaryOp{
		Op: "OR",
		Left: &parser.FieldQuery{
			Field: "region",
			Value: &parser.TermValue{Term: "ca"},
		},
		Right: &parser.FieldQuery{
			Field: "region",
			Value: &parser.TermValue{Term: "us"},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "region = $1 OR region = $2", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "ca", output.Parameters[0])
	assert.Equal(t, "us", output.Parameters[1])
}

func TestPostgresTranslator_RangeQuery(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"rod_length": {Type: schema.TypeInteger, Column: "rod_length"},
	}, schema.SchemaOptions{})

	ast := &parser.RangeQuery{
		Field:          "rod_length",
		Start:          &parser.NumberValue{Number: "50"},
		End:            &parser.NumberValue{Number: "500"},
		InclusiveStart: true,
		InclusiveEnd:   true,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "rod_length BETWEEN $1 AND $2", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "50", output.Parameters[0])
	assert.Equal(t, "500", output.Parameters[1])
	assert.Equal(t, []string{"integer", "integer"}, output.ParameterTypes)
}

func TestPostgresTranslator_RangeQuery_Exclusive(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"price": {Type: schema.TypeFloat, Column: "price"},
	}, schema.SchemaOptions{})

	ast := &parser.RangeQuery{
		Field:          "price",
		Start:          &parser.NumberValue{Number: "10"},
		End:            &parser.NumberValue{Number: "20"},
		InclusiveStart: false,
		InclusiveEnd:   false,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "price > $1 AND price < $2", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "10", output.Parameters[0])
	assert.Equal(t, "20", output.Parameters[1])
}

func TestPostgresTranslator_FieldNotInSchema(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText, Column: "product_code"},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "invalid_field",
		Value: &parser.TermValue{Term: "test"},
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "not found")
}

func TestPostgresTranslator_ComplexNestedQuery(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText, Column: "product_code"},
		"region":       {Type: schema.TypeText, Column: "region"},
		"status":       {Type: schema.TypeText, Column: "status"},
	}, schema.SchemaOptions{})

	// (productCode:13w42 AND region:ca) OR status:active
	ast := &parser.BinaryOp{
		Op: "OR",
		Left: &parser.BinaryOp{
			Op: "AND",
			Left: &parser.FieldQuery{
				Field: "product_code",
				Value: &parser.TermValue{Term: "13w42"},
			},
			Right: &parser.FieldQuery{
				Field: "region",
				Value: &parser.TermValue{Term: "ca"},
			},
		},
		Right: &parser.FieldQuery{
			Field: "status",
			Value: &parser.TermValue{Term: "active"},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "(product_code = $1 AND region = $2) OR status = $3", output.WhereClause)
	assert.Len(t, output.Parameters, 3)
	assert.Equal(t, "13w42", output.Parameters[0])
	assert.Equal(t, "ca", output.Parameters[1])
	assert.Equal(t, "active", output.Parameters[2])
}

func TestPostgresTranslator_ParameterNumbering(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"a": {Type: schema.TypeText, Column: "a"},
		"b": {Type: schema.TypeText, Column: "b"},
		"c": {Type: schema.TypeText, Column: "c"},
	}, schema.SchemaOptions{})

	// a:1 AND b:2 AND c:3
	ast := &parser.BinaryOp{
		Op: "AND",
		Left: &parser.BinaryOp{
			Op: "AND",
			Left: &parser.FieldQuery{
				Field: "a",
				Value: &parser.TermValue{Term: "1"},
			},
			Right: &parser.FieldQuery{
				Field: "b",
				Value: &parser.TermValue{Term: "2"},
			},
		},
		Right: &parser.FieldQuery{
			Field: "c",
			Value: &parser.TermValue{Term: "3"},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "(a = $1 AND b = $2) AND c = $3", output.WhereClause)
	assert.Len(t, output.Parameters, 3)
	assert.Equal(t, "1", output.Parameters[0])
	assert.Equal(t, "2", output.Parameters[1])
	assert.Equal(t, "3", output.Parameters[2])
}
