package translator

import (
        "github.com/infiniv/rsearch/internal/parser"
	"testing"

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
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "product_code",
		Value: "13w42",
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
		"rod_length": {Type: schema.TypeInteger},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "rod_length",
		Value: "100",
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
		"product_code": {Type: schema.TypeText},
		"region":       {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.BinaryOp{
		Op: "AND",
		Left: &parser.FieldQuery{
			Field: "product_code",
			Value: "13w42",
		},
		Right: &parser.FieldQuery{
			Field: "region",
			Value: "ca",
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
		"region": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.BinaryOp{
		Op: "OR",
		Left: &parser.FieldQuery{
			Field: "region",
			Value: "ca",
		},
		Right: &parser.FieldQuery{
			Field: "region",
			Value: "us",
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
		"rod_length": {Type: schema.TypeInteger},
	}, schema.SchemaOptions{})

	ast := &parser.RangeQuery{
		Field:          "rod_length",
		Start:          50,
		End:            500,
		InclusiveStart: true,
		InclusiveEnd:   true,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "rod_length BETWEEN $1 AND $2", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, 50, output.Parameters[0])
	assert.Equal(t, 500, output.Parameters[1])
	assert.Equal(t, []string{"integer", "integer"}, output.ParameterTypes)
}

func TestPostgresTranslator_RangeQuery_Exclusive(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"price": {Type: schema.TypeFloat},
	}, schema.SchemaOptions{})

	ast := &parser.RangeQuery{
		Field:          "price",
		Start:          10,
		End:            20,
		InclusiveStart: false,
		InclusiveEnd:   false,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "price > $1 AND price < $2", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, 10, output.Parameters[0])
	assert.Equal(t, 20, output.Parameters[1])
}

func TestPostgresTranslator_FieldNotInSchema(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "invalid_field",
		Value: "test",
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "not found")
}

func TestPostgresTranslator_ComplexNestedQuery(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
		"region":       {Type: schema.TypeText},
		"status":       {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	// (productCode:13w42 AND region:ca) OR status:active
	ast := &parser.BinaryOp{
		Op: "OR",
		Left: &parser.BinaryOp{
			Op: "AND",
			Left: &parser.FieldQuery{
				Field: "product_code",
				Value: "13w42",
			},
			Right: &parser.FieldQuery{
				Field: "region",
				Value: "ca",
			},
		},
		Right: &parser.FieldQuery{
			Field: "status",
			Value: "active",
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
		"a": {Type: schema.TypeText},
		"b": {Type: schema.TypeText},
		"c": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	// a:1 AND b:2 AND c:3
	ast := &parser.BinaryOp{
		Op: "AND",
		Left: &parser.BinaryOp{
			Op: "AND",
			Left: &parser.FieldQuery{
				Field: "a",
				Value: "1",
			},
			Right: &parser.FieldQuery{
				Field: "b",
				Value: "2",
			},
		},
		Right: &parser.FieldQuery{
			Field: "c",
			Value: "3",
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

func TestPostgresTranslator_ExistsQuery(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.ExistsQuery{
		Field: "name",
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "sql", output.Type)
	assert.Equal(t, "name IS NOT NULL", output.WhereClause)
	assert.Len(t, output.Parameters, 0)
	assert.Len(t, output.ParameterTypes, 0)
}

func TestPostgresTranslator_ExistsQuery_JSONField(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"metadata": {Type: schema.TypeJSON},
	}, schema.SchemaOptions{})

	ast := &parser.ExistsQuery{
		Field: "metadata",
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "sql", output.Type)
	assert.Equal(t, "metadata IS NOT NULL AND metadata != 'null'::jsonb", output.WhereClause)
	assert.Len(t, output.Parameters, 0)
}

func TestPostgresTranslator_NotExistsQuery(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.UnaryOp{
		Op: "NOT",
		Operand: &parser.ExistsQuery{
			Field: "description",
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "sql", output.Type)
	assert.Equal(t, "NOT description IS NOT NULL", output.WhereClause)
	assert.Len(t, output.Parameters, 0)
}

func TestPostgresTranslator_NotExistsQuery_JSONField(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"tags": {Type: schema.TypeJSON},
	}, schema.SchemaOptions{})

	ast := &parser.UnaryOp{
		Op: "NOT",
		Operand: &parser.ExistsQuery{
			Field: "tags",
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "sql", output.Type)
	assert.Equal(t, "NOT (tags IS NOT NULL AND tags != 'null'::jsonb)", output.WhereClause)
	assert.Len(t, output.Parameters, 0)
}

func TestPostgresTranslator_ExistsQuery_WithOtherConditions(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name":   {Type: schema.TypeText},
		"region": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	// _exists_:name AND region:ca
	ast := &parser.BinaryOp{
		Op: "AND",
		Left: &parser.ExistsQuery{
			Field: "name",
		},
		Right: &parser.FieldQuery{
			Field: "region",
			Value: "ca",
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "name IS NOT NULL AND region = $1", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "ca", output.Parameters[0])
}

func TestPostgresTranslator_ExistsQuery_InvalidField(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.ExistsQuery{
		Field: "invalid_field",
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "not found")
}

func TestPostgresTranslator_ComplexExistsQuery(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name":        {Type: schema.TypeText},
		"description": {Type: schema.TypeText},
		"price":       {Type: schema.TypeFloat},
	}, schema.SchemaOptions{})

	// (_exists_:name AND _exists_:description) OR price:100
	ast := &parser.BinaryOp{
		Op: "OR",
		Left: &parser.BinaryOp{
			Op: "AND",
			Left: &parser.ExistsQuery{
				Field: "name",
			},
			Right: &parser.ExistsQuery{
				Field: "description",
			},
		},
		Right: &parser.FieldQuery{
			Field: "price",
			Value: "100",
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "(name IS NOT NULL AND description IS NOT NULL) OR price = $1", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "100", output.Parameters[0])
}
