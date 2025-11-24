package translator

import (
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

	ast := &FieldQuery{
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

	ast := &FieldQuery{
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

	ast := &BinaryOp{
		Op: "AND",
		Left: &FieldQuery{
			Field: "product_code",
			Value: "13w42",
		},
		Right: &FieldQuery{
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

	ast := &BinaryOp{
		Op: "OR",
		Left: &FieldQuery{
			Field: "region",
			Value: "ca",
		},
		Right: &FieldQuery{
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

	ast := &RangeQuery{
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

	ast := &RangeQuery{
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

	ast := &FieldQuery{
		Field: "invalid_field",
		Value: "test",
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "not found")
}

func TestPostgresTranslator_FieldNotSearchable(t *testing.T) {
	translator := NewPostgresTranslator()

	// Note: The current schema design doesn't have a Searchable field
	// All fields in the schema are searchable by default
	// This test is kept for backwards compatibility but should pass now
	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"internal_id": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &FieldQuery{
		Field: "internal_id",
		Value: "test",
	}

	output, err := translator.Translate(ast, testSchema)
	// Since all fields are searchable now, this should succeed
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "internal_id = $1", output.WhereClause)
}

func TestPostgresTranslator_ComplexNestedQuery(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
		"region":       {Type: schema.TypeText},
		"status":       {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	// (productCode:13w42 AND region:ca) OR status:active
	ast := &BinaryOp{
		Op: "OR",
		Left: &BinaryOp{
			Op: "AND",
			Left: &FieldQuery{
				Field: "product_code",
				Value: "13w42",
			},
			Right: &FieldQuery{
				Field: "region",
				Value: "ca",
			},
		},
		Right: &FieldQuery{
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
	ast := &BinaryOp{
		Op: "AND",
		Left: &BinaryOp{
			Op: "AND",
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

func TestPostgresTranslator_FuzzyQuery_AutoDistance(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Fuzzy: true,
		},
	})

	// name:widget~ (auto distance of 2)
	ast := &FuzzyQuery{
		Field:    "name",
		Term:     "widget",
		Distance: 2,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "sql", output.Type)
	assert.Equal(t, "levenshtein(name, $1) <= $2", output.WhereClause)
	assert.Len(t, output.Parameters, 2)
	assert.Equal(t, "widget", output.Parameters[0])
	assert.Equal(t, 2, output.Parameters[1])
	assert.Equal(t, []string{"text", "integer"}, output.ParameterTypes)
}

func TestPostgresTranslator_FuzzyQuery_CustomDistance(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Fuzzy: true,
		},
	})

	// description:fuzzy~1 (custom distance of 1)
	ast := &FuzzyQuery{
		Field:    "description",
		Term:     "fuzzy",
		Distance: 1,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "levenshtein(description, $1) <= $2", output.WhereClause)
	assert.Equal(t, "fuzzy", output.Parameters[0])
	assert.Equal(t, 1, output.Parameters[1])
}

func TestPostgresTranslator_FuzzyQuery_FeatureDisabled(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Fuzzy: false, // Fuzzy search disabled
		},
	})

	ast := &FuzzyQuery{
		Field:    "name",
		Term:     "widget",
		Distance: 2,
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "fuzzy search requires pg_trgm extension")
	assert.Contains(t, err.Error(), "use wildcards instead")
}

func TestPostgresTranslator_FuzzyQuery_InvalidField(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Fuzzy: true,
		},
	})

	ast := &FuzzyQuery{
		Field:    "nonexistent",
		Term:     "widget",
		Distance: 2,
	}

	output, err := translator.Translate(ast, testSchema)
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "not found")
}

func TestPostgresTranslator_FuzzyQuery_WithSnakeCase(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"productName": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		NamingConvention: "snake_case",
		EnabledFeatures: schema.EnabledFeatures{
			Fuzzy: true,
		},
	})

	ast := &FuzzyQuery{
		Field:    "productName",
		Term:     "widget",
		Distance: 2,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	// Should use snake_case column name
	assert.Equal(t, "levenshtein(product_name, $1) <= $2", output.WhereClause)
	assert.Equal(t, "widget", output.Parameters[0])
	assert.Equal(t, 2, output.Parameters[1])
}

func TestPostgresTranslator_FuzzyQuery_CombinedWithOtherQueries(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name":   {Type: schema.TypeText},
		"region": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Fuzzy: true,
		},
	})

	// name:widget~ AND region:ca
	ast := &BinaryOp{
		Op: "AND",
		Left: &FuzzyQuery{
			Field:    "name",
			Term:     "widget",
			Distance: 2,
		},
		Right: &FieldQuery{
			Field: "region",
			Value: "ca",
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "levenshtein(name, $1) <= $2 AND region = $3", output.WhereClause)
	assert.Len(t, output.Parameters, 3)
	assert.Equal(t, "widget", output.Parameters[0])
	assert.Equal(t, 2, output.Parameters[1])
	assert.Equal(t, "ca", output.Parameters[2])
	assert.Equal(t, []string{"text", "integer", "text"}, output.ParameterTypes)
}

func TestPostgresTranslator_FuzzyQuery_MultipleFuzzyQueries(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"name":        {Type: schema.TypeText},
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Fuzzy: true,
		},
	})

	// name:widget~1 OR description:gadget~2
	ast := &BinaryOp{
		Op: "OR",
		Left: &FuzzyQuery{
			Field:    "name",
			Term:     "widget",
			Distance: 1,
		},
		Right: &FuzzyQuery{
			Field:    "description",
			Term:     "gadget",
			Distance: 2,
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "levenshtein(name, $1) <= $2 OR levenshtein(description, $3) <= $4", output.WhereClause)
	assert.Len(t, output.Parameters, 4)
	assert.Equal(t, "widget", output.Parameters[0])
	assert.Equal(t, 1, output.Parameters[1])
	assert.Equal(t, "gadget", output.Parameters[2])
	assert.Equal(t, 2, output.Parameters[3])
}
