package translator

import (
	"testing"

	"github.com/infiniv/rsearch/internal/parser"
	"github.com/infiniv/rsearch/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMongoDBTranslator_DatabaseType(t *testing.T) {
	translator := NewMongoDBTranslator()
	assert.Equal(t, "mongodb", translator.DatabaseType())
}

func TestMongoDBTranslator_SimpleFieldQuery(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "product_code",
		Value: &parser.TermValue{Term: "13w42", Pos: parser.Position{}},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, "mongodb", output.Type)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok, "Filter should be map[string]interface{}")
	assert.Equal(t, "13w42", filter["product_code"])
}

func TestMongoDBTranslator_NumberFieldQuery(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"rod_length": {Type: schema.TypeInteger},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "rod_length",
		Value: &parser.TermValue{Term: "100", Pos: parser.Position{}},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	require.NotNil(t, output)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "100", filter["rod_length"])
}

func TestMongoDBTranslator_PhraseValue(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "description",
		Value: &parser.PhraseValue{Phrase: "hello world", Pos: parser.Position{}},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "hello world", filter["description"])
}

func TestMongoDBTranslator_BooleanAND(t *testing.T) {
	translator := NewMongoDBTranslator()

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

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	andArray, ok := filter["$and"].([]interface{})
	require.True(t, ok, "Should have $and operator")
	require.Len(t, andArray, 2)

	left, ok := andArray[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "13w42", left["product_code"])

	right, ok := andArray[1].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "ca", right["region"])
}

func TestMongoDBTranslator_BooleanOR(t *testing.T) {
	translator := NewMongoDBTranslator()

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

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	orArray, ok := filter["$or"].([]interface{})
	require.True(t, ok, "Should have $or operator")
	require.Len(t, orArray, 2)

	left, ok := orArray[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "ca", left["region"])

	right, ok := orArray[1].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "us", right["region"])
}

func TestMongoDBTranslator_UnaryOpNOT(t *testing.T) {
	translator := NewMongoDBTranslator()

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

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	statusFilter, ok := filter["status"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "inactive", statusFilter["$ne"])
}

func TestMongoDBTranslator_UnaryOpNOTWithBinaryOp(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"status": {Type: schema.TypeText},
		"region": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.UnaryOp{
		Op: "NOT",
		Operand: &parser.BinaryOp{
			Op: "AND",
			Left: &parser.FieldQuery{
				Field: "status",
				Value: &parser.TermValue{Term: "active", Pos: parser.Position{}},
			},
			Right: &parser.FieldQuery{
				Field: "region",
				Value: &parser.TermValue{Term: "ca", Pos: parser.Position{}},
			},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	norArray, ok := filter["$nor"].([]interface{})
	require.True(t, ok, "Complex NOT should use $nor")
	require.Len(t, norArray, 1)
}

func TestMongoDBTranslator_RangeQueryInclusive(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"price": {Type: schema.TypeFloat},
	}, schema.SchemaOptions{})

	ast := &parser.RangeQuery{
		Field:          "price",
		Start:          &parser.NumberValue{Number: "50", Pos: parser.Position{}},
		End:            &parser.NumberValue{Number: "500", Pos: parser.Position{}},
		InclusiveStart: true,
		InclusiveEnd:   true,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	priceFilter, ok := filter["price"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "50", priceFilter["$gte"])
	assert.Equal(t, "500", priceFilter["$lte"])
}

func TestMongoDBTranslator_RangeQueryExclusive(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"price": {Type: schema.TypeFloat},
	}, schema.SchemaOptions{})

	ast := &parser.RangeQuery{
		Field:          "price",
		Start:          &parser.NumberValue{Number: "50", Pos: parser.Position{}},
		End:            &parser.NumberValue{Number: "500", Pos: parser.Position{}},
		InclusiveStart: false,
		InclusiveEnd:   false,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	priceFilter, ok := filter["price"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "50", priceFilter["$gt"])
	assert.Equal(t, "500", priceFilter["$lt"])
}

func TestMongoDBTranslator_RangeQueryMixed(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"price": {Type: schema.TypeFloat},
	}, schema.SchemaOptions{})

	ast := &parser.RangeQuery{
		Field:          "price",
		Start:          &parser.NumberValue{Number: "50", Pos: parser.Position{}},
		End:            &parser.NumberValue{Number: "500", Pos: parser.Position{}},
		InclusiveStart: true,
		InclusiveEnd:   false,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	priceFilter, ok := filter["price"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "50", priceFilter["$gte"])
	assert.Equal(t, "500", priceFilter["$lt"])
}

func TestMongoDBTranslator_RangeQueryUnboundedStart(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"price": {Type: schema.TypeFloat},
	}, schema.SchemaOptions{})

	ast := &parser.RangeQuery{
		Field:          "price",
		Start:          &parser.TermValue{Term: "*", Pos: parser.Position{}},
		End:            &parser.NumberValue{Number: "500", Pos: parser.Position{}},
		InclusiveStart: true,
		InclusiveEnd:   true,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	priceFilter, ok := filter["price"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "500", priceFilter["$lte"])
	assert.NotContains(t, priceFilter, "$gte")
}

func TestMongoDBTranslator_RangeQueryUnboundedEnd(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"price": {Type: schema.TypeFloat},
	}, schema.SchemaOptions{})

	ast := &parser.RangeQuery{
		Field:          "price",
		Start:          &parser.NumberValue{Number: "50", Pos: parser.Position{}},
		End:            &parser.TermValue{Term: "*", Pos: parser.Position{}},
		InclusiveStart: true,
		InclusiveEnd:   true,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	priceFilter, ok := filter["price"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "50", priceFilter["$gte"])
	assert.NotContains(t, priceFilter, "$lte")
}

func TestMongoDBTranslator_WildcardValue(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "product_code",
		Value: &parser.WildcardValue{Pattern: "13*", Pos: parser.Position{}},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	productFilter, ok := filter["product_code"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "^13.*$", productFilter["$regex"])
}

func TestMongoDBTranslator_WildcardValueWithQuestionMark(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "product_code",
		Value: &parser.WildcardValue{Pattern: "1?w42", Pos: parser.Position{}},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	productFilter, ok := filter["product_code"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "^1.w42$", productFilter["$regex"])
}

func TestMongoDBTranslator_RegexValue(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "product_code",
		Value: &parser.RegexValue{Pattern: "^[0-9]+$", Pos: parser.Position{}},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	productFilter, ok := filter["product_code"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "^[0-9]+$", productFilter["$regex"])
	assert.Equal(t, "", productFilter["$options"])
}

func TestMongoDBTranslator_ExistsQuery(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"optional_field": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.ExistsQuery{
		Field: "optional_field",
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	fieldFilter, ok := filter["optional_field"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, fieldFilter["$exists"])
	assert.Equal(t, nil, fieldFilter["$ne"])
}

func TestMongoDBTranslator_BoostQuery(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.BoostQuery{
		Query: &parser.FieldQuery{
			Field: "product_code",
			Value: &parser.TermValue{Term: "13w42", Pos: parser.Position{}},
		},
		Boost: 2.5,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "13w42", filter["product_code"])

	require.NotNil(t, output.Metadata)
	boosts, ok := output.Metadata["boosts"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, boosts, 1)
	assert.Equal(t, "field_query", boosts[0]["query"])
	assert.Equal(t, 2.5, boosts[0]["boost"])
}

func TestMongoDBTranslator_GroupQuery(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"status": {Type: schema.TypeText},
		"region": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.GroupQuery{
		Query: &parser.BinaryOp{
			Op: "OR",
			Left: &parser.FieldQuery{
				Field: "status",
				Value: &parser.TermValue{Term: "active", Pos: parser.Position{}},
			},
			Right: &parser.FieldQuery{
				Field: "region",
				Value: &parser.TermValue{Term: "ca", Pos: parser.Position{}},
			},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	orArray, ok := filter["$or"].([]interface{})
	require.True(t, ok)
	require.Len(t, orArray, 2)
}

func TestMongoDBTranslator_RequiredQuery(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"status": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.RequiredQuery{
		Query: &parser.FieldQuery{
			Field: "status",
			Value: &parser.TermValue{Term: "active", Pos: parser.Position{}},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "active", filter["status"])
}

func TestMongoDBTranslator_ProhibitedQuery(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"status": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.ProhibitedQuery{
		Query: &parser.FieldQuery{
			Field: "status",
			Value: &parser.TermValue{Term: "inactive", Pos: parser.Position{}},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	statusFilter, ok := filter["status"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "inactive", statusFilter["$ne"])
}

func TestMongoDBTranslator_TermQuery(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		DefaultField: "description",
	})

	ast := &parser.TermQuery{
		Term: "laptop",
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "laptop", filter["description"])
}

func TestMongoDBTranslator_PhraseQuery(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		DefaultField: "description",
	})

	ast := &parser.PhraseQuery{
		Phrase: "gaming laptop",
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "gaming laptop", filter["description"])
}

func TestMongoDBTranslator_WildcardQuery(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		DefaultField: "description",
	})

	ast := &parser.WildcardQuery{
		Pattern: "lap*",
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	descFilter, ok := filter["description"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "^lap.*$", descFilter["$regex"])
}

func TestMongoDBTranslator_FuzzyQuery(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Fuzzy: true,
		},
	})

	ast := &parser.FuzzyQuery{
		Field:    "product_name",
		Term:     "laptop",
		Distance: 2,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	textFilter, ok := filter["$text"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "laptop", textFilter["$search"])

	require.NotNil(t, output.Metadata)
	assert.Equal(t, 2, output.Metadata["fuzzy_distance"])
}

func TestMongoDBTranslator_FuzzyQueryDisabled(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_name": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Fuzzy: false,
		},
	})

	ast := &parser.FuzzyQuery{
		Field:    "product_name",
		Term:     "laptop",
		Distance: 2,
	}

	_, err := translator.Translate(ast, testSchema)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fuzzy search requires text index")
}

func TestMongoDBTranslator_ProximityQuery(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Proximity: true,
		},
	})

	ast := &parser.ProximityQuery{
		Field:    "description",
		Phrase:   "gaming laptop",
		Distance: 5,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	textFilter, ok := filter["$text"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "\"gaming laptop\"", textFilter["$search"])

	require.NotNil(t, output.Metadata)
	assert.Equal(t, 5, output.Metadata["proximity_distance"])
}

func TestMongoDBTranslator_ProximityQueryDisabled(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{
		EnabledFeatures: schema.EnabledFeatures{
			Proximity: false,
		},
	})

	ast := &parser.ProximityQuery{
		Field:    "description",
		Phrase:   "gaming laptop",
		Distance: 5,
	}

	_, err := translator.Translate(ast, testSchema)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "proximity search requires text index")
}

func TestMongoDBTranslator_FieldGroupQuery(t *testing.T) {
	translator := NewMongoDBTranslator()

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

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	orArray, ok := filter["$or"].([]interface{})
	require.True(t, ok)
	require.Len(t, orArray, 2)

	first, ok := orArray[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "active", first["status"])

	second, ok := orArray[1].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "pending", second["status"])
}

func TestMongoDBTranslator_FieldGroupQueryWithWildcard(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"status": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldGroupQuery{
		Field: "status",
		Queries: []parser.Node{
			&parser.TermQuery{Term: "active"},
			&parser.WildcardQuery{Pattern: "pend*"},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	orArray, ok := filter["$or"].([]interface{})
	require.True(t, ok)
	require.Len(t, orArray, 2)

	second, ok := orArray[1].(map[string]interface{})
	require.True(t, ok)
	statusFilter, ok := second["status"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "^pend.*$", statusFilter["$regex"])
}

func TestMongoDBTranslator_ComplexNestedQuery(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"status":       {Type: schema.TypeText},
		"region":       {Type: schema.TypeText},
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.BinaryOp{
		Op: "AND",
		Left: &parser.BinaryOp{
			Op: "OR",
			Left: &parser.FieldQuery{
				Field: "status",
				Value: &parser.TermValue{Term: "active", Pos: parser.Position{}},
			},
			Right: &parser.FieldQuery{
				Field: "status",
				Value: &parser.TermValue{Term: "pending", Pos: parser.Position{}},
			},
		},
		Right: &parser.BinaryOp{
			Op: "AND",
			Left: &parser.FieldQuery{
				Field: "region",
				Value: &parser.TermValue{Term: "ca", Pos: parser.Position{}},
			},
			Right: &parser.FieldQuery{
				Field: "product_code",
				Value: &parser.WildcardValue{Pattern: "13*", Pos: parser.Position{}},
			},
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)

	filter, ok := output.Filter.(map[string]interface{})
	require.True(t, ok)

	andArray, ok := filter["$and"].([]interface{})
	require.True(t, ok)
	require.Len(t, andArray, 2)

	leftFilter, ok := andArray[0].(map[string]interface{})
	require.True(t, ok)
	orArray, ok := leftFilter["$or"].([]interface{})
	require.True(t, ok)
	require.Len(t, orArray, 2)

	rightFilter, ok := andArray[1].(map[string]interface{})
	require.True(t, ok)
	rightAndArray, ok := rightFilter["$and"].([]interface{})
	require.True(t, ok)
	require.Len(t, rightAndArray, 2)
}

func TestMongoDBTranslator_NoDefaultField(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"description": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.TermQuery{
		Term: "laptop",
	}

	_, err := translator.Translate(ast, testSchema)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires a default field")
}

func TestMongoDBTranslator_InvalidFieldName(t *testing.T) {
	translator := NewMongoDBTranslator()

	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})

	ast := &parser.FieldQuery{
		Field: "invalid_field",
		Value: &parser.TermValue{Term: "test", Pos: parser.Position{}},
	}

	_, err := translator.Translate(ast, testSchema)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in schema")
}
