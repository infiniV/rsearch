package translator

import (
	"testing"

	"github.com/infiniv/rsearch/internal/parser"
	"github.com/infiniv/rsearch/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteTranslator_DatabaseType(t *testing.T) {
	translator := NewSQLiteTranslator()
	assert.Equal(t, "sqlite", translator.DatabaseType())
}

func TestSQLiteTranslator_SimpleFieldQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

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

func TestSQLiteTranslator_NumberFieldQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

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

func TestSQLiteTranslator_BooleanAND(t *testing.T) {
	translator := NewSQLiteTranslator()

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

func TestSQLiteTranslator_BooleanOR(t *testing.T) {
	translator := NewSQLiteTranslator()

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

func TestSQLiteTranslator_UnaryOpNOT(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"status": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.UnaryOp{
		Op: "NOT",
		Operand: &parser.FieldQuery{
			Field: "status",
			Value: &parser.TermValue{Term: "active", Pos: parser.Position{}},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "NOT status = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "active", output.Parameters[0])
}

func TestSQLiteTranslator_RangeQuery_Inclusive(t *testing.T) {
	translator := NewSQLiteTranslator()

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

func TestSQLiteTranslator_RangeQuery_Exclusive(t *testing.T) {
	translator := NewSQLiteTranslator()

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

func TestSQLiteTranslator_RangeQuery_Unbounded(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"price": {Type: schema.TypeFloat},
	}, schema.SchemaOptions{})

	ast := &parser.RangeQuery{
		Field:          "price",
		Start:          &parser.NumberValue{Number: "100", Pos: parser.Position{}},
		End:            &parser.TermValue{Term: "*", Pos: parser.Position{}},
		InclusiveStart: true,
		InclusiveEnd:   true,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "price >= ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "100", output.Parameters[0])
}

func TestSQLiteTranslator_WildcardQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "name",
		Value: &parser.WildcardValue{Pattern: "test*", Pos: parser.Position{}},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "name LIKE ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "test%", output.Parameters[0])
}

func TestSQLiteTranslator_RegexQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"email": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "email",
		Value: &parser.RegexValue{Pattern: "^[a-z]+@.*\\.com$", Pos: parser.Position{}},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "email REGEXP ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "^[a-z]+@.*\\.com$", output.Parameters[0])
}

func TestSQLiteTranslator_PhraseQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "description",
		Value: &parser.PhraseValue{Phrase: "quick brown fox", Pos: parser.Position{}},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "description = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "quick brown fox", output.Parameters[0])
}

func TestSQLiteTranslator_FuzzyQuery_NotSupported(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Fuzzy: true,
		},
	})

	ast := &parser.FuzzyQuery{
		Field:    "name",
		Term:     "test",
		Distance: 2,
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "fuzzy search not supported")
	assert.Contains(t, err.Error(), "wildcard")
}

func TestSQLiteTranslator_ProximityQuery_WithFTS(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Proximity: true,
		},
	})

	ast := &parser.ProximityQuery{
		Field:    "description",
		Phrase:   "quick brown",
		Distance: 5,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "description MATCH ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "NEAR(quick brown, 5)", output.Parameters[0])
}

func TestSQLiteTranslator_ProximityQuery_WithoutFTS(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.ProximityQuery{
		Field:    "description",
		Phrase:   "quick brown",
		Distance: 5,
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "proximity search requires FTS5")
}

func TestSQLiteTranslator_ExistsQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

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
}

func TestSQLiteTranslator_ExistsQuery_JSONField(t *testing.T) {
	translator := NewSQLiteTranslator()

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
	assert.Equal(t, "metadata IS NOT NULL AND json_extract(metadata, '$') IS NOT NULL", output.WhereClause)
	assert.Len(t, output.Parameters, 0)
}

func TestSQLiteTranslator_NotExistsQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

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

func TestSQLiteTranslator_NotExistsQuery_JSONField(t *testing.T) {
	translator := NewSQLiteTranslator()

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
	assert.Equal(t, "NOT (tags IS NOT NULL AND json_extract(tags, '$') IS NOT NULL)", output.WhereClause)
	assert.Len(t, output.Parameters, 0)
}

func TestSQLiteTranslator_BoostQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.BoostQuery{
		Query: &parser.FieldQuery{
			Field: "name",
			Value: &parser.TermValue{Term: "test", Pos: parser.Position{}},
		},
		Boost: 2.5,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "name = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "test", output.Parameters[0])
	assert.NotNil(t, output.Metadata)
	assert.Contains(t, output.Metadata, "boosts")
}

func TestSQLiteTranslator_GroupQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"status": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.GroupQuery{
		Query: &parser.FieldQuery{
			Field: "status",
			Value: &parser.TermValue{Term: "active", Pos: parser.Position{}},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "(status = ?)", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "active", output.Parameters[0])
}

func TestSQLiteTranslator_RequiredQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.RequiredQuery{
		Query: &parser.FieldQuery{
			Field: "name",
			Value: &parser.TermValue{Term: "test", Pos: parser.Position{}},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "name = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "test", output.Parameters[0])
}

func TestSQLiteTranslator_ProhibitedQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"status": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.ProhibitedQuery{
		Query: &parser.FieldQuery{
			Field: "status",
			Value: &parser.TermValue{Term: "deleted", Pos: parser.Position{}},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "NOT status = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "deleted", output.Parameters[0])
}

func TestSQLiteTranslator_TermQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		DefaultField: "name",
	})

	ast := &parser.TermQuery{
		Term: "test",
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "name = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "test", output.Parameters[0])
}

func TestSQLiteTranslator_StandalonePhraseQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		DefaultField: "description",
	})

	ast := &parser.PhraseQuery{
		Phrase: "quick brown fox",
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "description = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "quick brown fox", output.Parameters[0])
}

func TestSQLiteTranslator_StandaloneWildcardQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		DefaultField: "name",
	})

	ast := &parser.WildcardQuery{
		Pattern: "test*",
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "name LIKE ?", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "test%", output.Parameters[0])
}

func TestSQLiteTranslator_FieldGroupQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"status": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

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
	assert.Equal(t, "(status = ? OR status = ?)", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "active", output.Parameters[0])
	assert.Equal(t, "pending", output.Parameters[1])
}

func TestSQLiteTranslator_FieldGroupQuery_WithWildcard(t *testing.T) {
	translator := NewSQLiteTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldGroupQuery{
		Field: "name",
		Queries: []parser.Node{
			&parser.TermQuery{Term: "test"},
			&parser.WildcardQuery{Pattern: "prod*"},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "(name = ? OR name LIKE ?)", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "test", output.Parameters[0])
	assert.Equal(t, "prod%", output.Parameters[1])
}

func TestSQLiteTranslator_ComplexNestedQuery(t *testing.T) {
	translator := NewSQLiteTranslator()

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
	assert.Equal(t, "(product_code = ? AND region = ?) OR status = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 3)
	assert.Equal(t, "13w42", output.Parameters[0])
	assert.Equal(t, "ca", output.Parameters[1])
	assert.Equal(t, "active", output.Parameters[2])
}

func TestSQLiteTranslator_ParameterOrdering(t *testing.T) {
	translator := NewSQLiteTranslator()

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
	assert.Equal(t, "(a = ? AND b = ?) AND c = ?", output.WhereClause)
	assert.Len(t, output.Parameters, 3)
	assert.Equal(t, "1", output.Parameters[0])
	assert.Equal(t, "2", output.Parameters[1])
	assert.Equal(t, "3", output.Parameters[2])
}

func TestSQLiteTranslator_FieldNotInSchema(t *testing.T) {
	translator := NewSQLiteTranslator()

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
