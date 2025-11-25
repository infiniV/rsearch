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
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "product_code",
		Value: &parser.TermValue{Term: "13w42", Pos: parser.Position{}},
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
		Value: &parser.TermValue{Term: "100", Pos: parser.Position{}},
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
			Value: &parser.TermValue{Term: "13w42", Pos: parser.Position{}},
		},
		Right: &parser.FieldQuery{
			Field: "region",
			Value: &parser.TermValue{Term: "ca", Pos: parser.Position{}},
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
			Value: &parser.TermValue{Term: "ca", Pos: parser.Position{}},
		},
		Right: &parser.FieldQuery{
			Field: "region",
			Value: &parser.TermValue{Term: "us", Pos: parser.Position{}},
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
		Start:          &parser.NumberValue{Number: "50", Pos: parser.Position{}},
		End:            &parser.NumberValue{Number: "500", Pos: parser.Position{}},
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
		"price": {Type: schema.TypeFloat},
	}, schema.SchemaOptions{})

	ast := &parser.RangeQuery{
		Field:          "price",
		Start:          &parser.NumberValue{Number: "10", Pos: parser.Position{}},
		End:            &parser.NumberValue{Number: "20", Pos: parser.Position{}},
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
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "invalid_field",
		Value: &parser.TermValue{Term: "test", Pos: parser.Position{}},
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
				Value: &parser.TermValue{Term: "13w42", Pos: parser.Position{}},
			},
			Right: &parser.FieldQuery{
				Field: "region",
				Value: &parser.TermValue{Term: "ca", Pos: parser.Position{}},
			},
		},
		Right: &parser.FieldQuery{
			Field: "status",
			Value: &parser.TermValue{Term: "active", Pos: parser.Position{}},
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
			Value: &parser.TermValue{Term: "ca", Pos: parser.Position{}},
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
			Value: &parser.TermValue{Term: "100", Pos: parser.Position{}},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "(name IS NOT NULL AND description IS NOT NULL) OR price = $1", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "100", output.Parameters[0])
}

// ProximityQuery Tests

func TestPostgresTranslator_ProximityQuery(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		DefaultField: "description",
		EnabledFeatures: schema.EnabledFeatures{
			Proximity: true,
		},
	})

	ast := &parser.ProximityQuery{
		Field:    "description",
		Phrase:   "quick brown fox",
		Distance: 5,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "sql", output.Type)
	assert.Contains(t, output.WhereClause, "to_tsvector")
	assert.Contains(t, output.WhereClause, "phraseto_tsquery")
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "quick brown fox", output.Parameters[0])
}

func TestPostgresTranslator_ProximityQuery_DefaultField(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"content": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		DefaultField: "content",
		EnabledFeatures: schema.EnabledFeatures{
			Proximity: true,
		},
	})

	// Proximity without explicit field uses default
	ast := &parser.ProximityQuery{
		Field:    "",
		Phrase:   "hello world",
		Distance: 3,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Contains(t, output.WhereClause, "content")
}

func TestPostgresTranslator_ProximityQuery_NotEnabled(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Proximity: false,
		},
	})

	ast := &parser.ProximityQuery{
		Field:    "description",
		Phrase:   "quick fox",
		Distance: 5,
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "full-text search")
}

func TestPostgresTranslator_ProximityQuery_SingleWord(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Proximity: true,
		},
	})

	// Single word falls back to simple match
	ast := &parser.ProximityQuery{
		Field:    "name",
		Phrase:   "widget",
		Distance: 3,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "name = $1", output.WhereClause)
	assert.Equal(t, "widget", output.Parameters[0])
}

func TestPostgresTranslator_ProximityQuery_InvalidField(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Proximity: true,
		},
	})

	ast := &parser.ProximityQuery{
		Field:    "invalid_field",
		Phrase:   "test phrase",
		Distance: 3,
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "not found")
}

// FieldGroupQuery Tests

func TestPostgresTranslator_FieldGroupQuery(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"status": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	// status:(active OR pending)
	ast := &parser.FieldGroupQuery{
		Field: "status",
		Queries: []parser.Node{
			&parser.TermQuery{Term: "active"},
			&parser.TermQuery{Term: "pending"},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "sql", output.Type)
	assert.Equal(t, "(status = $1 OR status = $2)", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "active", output.Parameters[0])
	assert.Equal(t, "pending", output.Parameters[1])
}

func TestPostgresTranslator_FieldGroupQuery_SingleValue(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"category": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldGroupQuery{
		Field: "category",
		Queries: []parser.Node{
			&parser.TermQuery{Term: "electronics"},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "category = $1", output.WhereClause)
	assert.Equal(t, "electronics", output.Parameters[0])
}

func TestPostgresTranslator_FieldGroupQuery_WithWildcard(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	// name:(widget* OR gadget*)
	ast := &parser.FieldGroupQuery{
		Field: "name",
		Queries: []parser.Node{
			&parser.WildcardQuery{Pattern: "widget*"},
			&parser.WildcardQuery{Pattern: "gadget*"},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "(name LIKE $1 OR name LIKE $2)", output.WhereClause)
	assert.Equal(t, "widget%", output.Parameters[0])
	assert.Equal(t, "gadget%", output.Parameters[1])
}

func TestPostgresTranslator_FieldGroupQuery_WithBinaryOp(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"tags": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	// tags:(scala AND functional)
	ast := &parser.FieldGroupQuery{
		Field: "tags",
		Queries: []parser.Node{
			&parser.BinaryOp{
				Op:    "AND",
				Left:  &parser.TermQuery{Term: "scala"},
				Right: &parser.TermQuery{Term: "functional"},
			},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "(tags = $1 AND tags = $2)", output.WhereClause)
	assert.Equal(t, "scala", output.Parameters[0])
	assert.Equal(t, "functional", output.Parameters[1])
}

func TestPostgresTranslator_FieldGroupQuery_Empty(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldGroupQuery{
		Field:   "name",
		Queries: []parser.Node{},
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "empty")
}

func TestPostgresTranslator_FieldGroupQuery_InvalidField(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldGroupQuery{
		Field: "invalid_field",
		Queries: []parser.Node{
			&parser.TermQuery{Term: "test"},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "not found")
}

func TestPostgresTranslator_FieldGroupQuery_CombinedWithOther(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"status": {Type: schema.TypeText},
		"region": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	// status:(active OR pending) AND region:ca
	ast := &parser.BinaryOp{
		Op: "AND",
		Left: &parser.FieldGroupQuery{
			Field: "status",
			Queries: []parser.Node{
				&parser.TermQuery{Term: "active"},
				&parser.TermQuery{Term: "pending"},
			},
		},
		Right: &parser.FieldQuery{
			Field: "region",
			Value: &parser.TermValue{Term: "ca"},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "(status = $1 OR status = $2) AND region = $3", output.WhereClause)
	assert.Len(t, output.Parameters, 3)
}
