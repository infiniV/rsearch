package translator

import (
	"testing"

	"github.com/infiniv/rsearch/internal/parser"
	"github.com/infiniv/rsearch/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQLTranslator_DatabaseType(t *testing.T) {
	translator := NewMySQLTranslator()
	assert.Equal(t, "mysql", translator.DatabaseType())
}

func TestMySQLTranslator_SimpleFieldQuery(t *testing.T) {
	translator := NewMySQLTranslator()

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
	assert.Equal(t, "product_code = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "13w42", output.Parameters[0])
	assert.Equal(t, []string{"text"}, output.ParameterTypes)
}

func TestMySQLTranslator_NumberFieldQuery(t *testing.T) {
	translator := NewMySQLTranslator()

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
	assert.Equal(t, "rod_length = ?", output.WhereClause)
	assert.Equal(t, "100", output.Parameters[0])
	assert.Equal(t, []string{"integer"}, output.ParameterTypes)
}

func TestMySQLTranslator_BooleanAND(t *testing.T) {
	translator := NewMySQLTranslator()

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
	assert.Equal(t, "product_code = ? AND region = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "13w42", output.Parameters[0])
	assert.Equal(t, "ca", output.Parameters[1])
	assert.Equal(t, []string{"text", "text"}, output.ParameterTypes)
}

func TestMySQLTranslator_BooleanOR(t *testing.T) {
	translator := NewMySQLTranslator()

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
	assert.Equal(t, "region = ? OR region = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "ca", output.Parameters[0])
	assert.Equal(t, "us", output.Parameters[1])
}

func TestMySQLTranslator_UnaryNOT(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"status": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.UnaryOp{
		Op: "NOT",
		Operand: &parser.FieldQuery{
			Field: "status",
			Value: &parser.TermValue{Term: "inactive", Pos: parser.Position{}},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "NOT status = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "inactive", output.Parameters[0])
}

func TestMySQLTranslator_RangeQuery_Inclusive(t *testing.T) {
	translator := NewMySQLTranslator()

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
	assert.Equal(t, "rod_length BETWEEN ? AND ?", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "50", output.Parameters[0])
	assert.Equal(t, "500", output.Parameters[1])
	assert.Equal(t, []string{"integer", "integer"}, output.ParameterTypes)
}

func TestMySQLTranslator_RangeQuery_Exclusive(t *testing.T) {
	translator := NewMySQLTranslator()

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
	assert.Equal(t, "price > ? AND price < ?", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "10", output.Parameters[0])
	assert.Equal(t, "20", output.Parameters[1])
}

func TestMySQLTranslator_RangeQuery_Mixed(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"price": {Type: schema.TypeFloat},
	}, schema.SchemaOptions{})

	ast := &parser.RangeQuery{
		Field:          "price",
		Start:          &parser.NumberValue{Number: "10", Pos: parser.Position{}},
		End:            &parser.NumberValue{Number: "20", Pos: parser.Position{}},
		InclusiveStart: true,
		InclusiveEnd:   false,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "price >= ? AND price < ?", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "10", output.Parameters[0])
	assert.Equal(t, "20", output.Parameters[1])
}

func TestMySQLTranslator_RangeQuery_Unbounded(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"rod_length": {Type: schema.TypeInteger},
	}, schema.SchemaOptions{})

	ast := &parser.RangeQuery{
		Field:          "rod_length",
		Start:          &parser.NumberValue{Number: "100", Pos: parser.Position{}},
		End:            &parser.TermValue{Term: "*", Pos: parser.Position{}},
		InclusiveStart: true,
		InclusiveEnd:   true,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "rod_length >= ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "100", output.Parameters[0])
}

func TestMySQLTranslator_WildcardQuery_Field(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "product_code",
		Value: &parser.WildcardValue{Pattern: "13w*", Pos: parser.Position{}},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "product_code LIKE ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "13w%", output.Parameters[0])
}

func TestMySQLTranslator_WildcardQuery_Standalone(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		DefaultField: "name",
	})

	ast := &parser.WildcardQuery{
		Pattern: "prod*",
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "name LIKE ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "prod%", output.Parameters[0])
}

func TestMySQLTranslator_RegexQuery(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "product_code",
		Value: &parser.RegexValue{Pattern: "^13w[0-9]+$", Pos: parser.Position{}},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "product_code REGEXP ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "^13w[0-9]+$", output.Parameters[0])
}

func TestMySQLTranslator_PhraseQuery(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "description",
		Value: &parser.PhraseValue{Phrase: "exact phrase", Pos: parser.Position{}},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "description = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "exact phrase", output.Parameters[0])
}

func TestMySQLTranslator_ExistsQuery(t *testing.T) {
	translator := NewMySQLTranslator()

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

func TestMySQLTranslator_ExistsQuery_JSONField(t *testing.T) {
	translator := NewMySQLTranslator()

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
	assert.Equal(t, "metadata IS NOT NULL AND JSON_TYPE(metadata) != 'NULL'", output.WhereClause)
	assert.Len(t, output.Parameters, 0)
}

func TestMySQLTranslator_NotExistsQuery(t *testing.T) {
	translator := NewMySQLTranslator()

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

func TestMySQLTranslator_NotExistsQuery_JSONField(t *testing.T) {
	translator := NewMySQLTranslator()

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
	assert.Equal(t, "NOT (tags IS NOT NULL AND JSON_TYPE(tags) != 'NULL')", output.WhereClause)
	assert.Len(t, output.Parameters, 0)
}

func TestMySQLTranslator_FuzzyQuery(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Fuzzy: true,
		},
	})

	ast := &parser.FuzzyQuery{
		Field:    "name",
		Term:     "product",
		Distance: 2,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "SOUNDEX(name) = SOUNDEX(?)", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "product", output.Parameters[0])
}

func TestMySQLTranslator_FuzzyQuery_NotEnabled(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FuzzyQuery{
		Field:    "name",
		Term:     "product",
		Distance: 2,
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "fuzzy search")
}

func TestMySQLTranslator_ProximityQuery(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{
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
	assert.Equal(t, "MATCH(description) AGAINST(? IN BOOLEAN MODE)", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "quick brown fox", output.Parameters[0])
}

func TestMySQLTranslator_ProximityQuery_NotEnabled(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.ProximityQuery{
		Field:    "description",
		Phrase:   "quick brown fox",
		Distance: 5,
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "proximity search")
}

func TestMySQLTranslator_BoostQuery(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.BoostQuery{
		Query: &parser.FieldQuery{
			Field: "name",
			Value: &parser.TermValue{Term: "product", Pos: parser.Position{}},
		},
		Boost: 2.5,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "name = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "product", output.Parameters[0])
	assert.NotNil(t, output.Metadata)
	assert.Contains(t, output.Metadata, "boosts")
	boosts := output.Metadata["boosts"].([]map[string]interface{})
	assert.Len(t, boosts, 1)
	assert.Equal(t, "field_query", boosts[0]["query"])
	assert.Equal(t, 2.5, boosts[0]["boost"])
}

func TestMySQLTranslator_GroupQuery(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.GroupQuery{
		Query: &parser.FieldQuery{
			Field: "name",
			Value: &parser.TermValue{Term: "product", Pos: parser.Position{}},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "(name = ?)", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
}

func TestMySQLTranslator_RequiredQuery(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		DefaultField: "name",
	})

	ast := &parser.RequiredQuery{
		Query: &parser.TermQuery{
			Term: "product",
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "name = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "product", output.Parameters[0])
}

func TestMySQLTranslator_ProhibitedQuery(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		DefaultField: "name",
	})

	ast := &parser.ProhibitedQuery{
		Query: &parser.TermQuery{
			Term: "obsolete",
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "NOT name = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "obsolete", output.Parameters[0])
}

func TestMySQLTranslator_TermQuery(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		DefaultField: "name",
	})

	ast := &parser.TermQuery{
		Term: "widget",
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "name = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "widget", output.Parameters[0])
}

func TestMySQLTranslator_PhraseQuery_Standalone(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		DefaultField: "description",
	})

	ast := &parser.PhraseQuery{
		Phrase: "blue widget",
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "description = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "blue widget", output.Parameters[0])
}

func TestMySQLTranslator_FieldGroupQuery(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"region": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldGroupQuery{
		Field: "region",
		Queries: []parser.Node{
			&parser.TermQuery{Term: "ca"},
			&parser.TermQuery{Term: "us"},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "(region = ? OR region = ?)", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "ca", output.Parameters[0])
	assert.Equal(t, "us", output.Parameters[1])
}

func TestMySQLTranslator_FieldGroupQuery_WithWildcard(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldGroupQuery{
		Field: "product_code",
		Queries: []parser.Node{
			&parser.TermQuery{Term: "13w42"},
			&parser.WildcardQuery{Pattern: "14*"},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "(product_code = ? OR product_code LIKE ?)", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "13w42", output.Parameters[0])
	assert.Equal(t, "14%", output.Parameters[1])
}

func TestMySQLTranslator_FieldGroupQuery_WithBinaryOp(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"status": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldGroupQuery{
		Field: "status",
		Queries: []parser.Node{
			&parser.BinaryOp{
				Op:    "OR",
				Left:  &parser.TermQuery{Term: "active"},
				Right: &parser.TermQuery{Term: "pending"},
			},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "(status = ? OR status = ?)", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "active", output.Parameters[0])
	assert.Equal(t, "pending", output.Parameters[1])
}

func TestMySQLTranslator_ComplexNestedQuery(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
		"region":       {Type: schema.TypeText},
		"status":       {Type: schema.TypeText},
	}, schema.SchemaOptions{})

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
	assert.Equal(t, "(product_code = ? AND region = ?) OR status = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 3)
	assert.Equal(t, "13w42", output.Parameters[0])
	assert.Equal(t, "ca", output.Parameters[1])
	assert.Equal(t, "active", output.Parameters[2])
}

func TestMySQLTranslator_ParameterOrdering(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"a": {Type: schema.TypeText},
		"b": {Type: schema.TypeText},
		"c": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

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
	assert.Equal(t, "(a = ? AND b = ?) AND c = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 3)
	assert.Equal(t, "1", output.Parameters[0])
	assert.Equal(t, "2", output.Parameters[1])
	assert.Equal(t, "3", output.Parameters[2])
}

func TestMySQLTranslator_FieldNotInSchema(t *testing.T) {
	translator := NewMySQLTranslator()

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

func TestMySQLTranslator_DefaultFieldRequired(t *testing.T) {
	translator := NewMySQLTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.TermQuery{
		Term: "widget",
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "default field")
}
