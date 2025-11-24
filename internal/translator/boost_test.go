package translator

import (
        "github.com/infiniv/rsearch/internal/parser"
	"testing"

	"github.com/infiniv/rsearch/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresTranslator_BoostQuery_SimpleField(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema(
		"products",
		map[string]schema.Field{
			"name": {Type: schema.TypeText},
		},
		schema.SchemaOptions{})

	// name:widget^2
	ast := &parser.BoostQuery{
		Query: &parser.FieldQuery{
			Field: "name",
			Value: "widget",
		},
		Boost: 2.0,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)

	// SQL should be the same as without boost
	assert.Equal(t, "name = $1", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "widget", output.Parameters[0])

	// Boost should be stored in metadata
	assert.NotNil(t, output.Metadata)
	assert.Contains(t, output.Metadata, "boosts")
	boosts := output.Metadata["boosts"].([]map[string]interface{})
	require.Len(t, boosts, 1)
	assert.Equal(t, "field_query", boosts[0]["query"])
	assert.Equal(t, 2.0, boosts[0]["boost"])
}

func TestPostgresTranslator_BoostQuery_HighBoost(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema(
		"products",
		map[string]schema.Field{
			"name": {Type: schema.TypeText},
		},
		schema.SchemaOptions{})

	// name:widget^4
	ast := &parser.BoostQuery{
		Query: &parser.FieldQuery{
			Field: "name",
			Value: "widget",
		},
		Boost: 4.0,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)

	// Check boost value in metadata
	boosts := output.Metadata["boosts"].([]map[string]interface{})
	require.Len(t, boosts, 1)
	assert.Equal(t, 4.0, boosts[0]["boost"])
}

func TestPostgresTranslator_BoostQuery_WithBinaryOp(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema(
		"products",
		map[string]schema.Field{
			"name":   {Type: schema.TypeText},
			"status": {Type: schema.TypeText},
		},
		schema.SchemaOptions{})

	// name:widget^2 AND status:active
	ast := &parser.BinaryOp{
		Op: "AND",
		Left: &parser.BoostQuery{
			Query: &parser.FieldQuery{
				Field: "name",
				Value: "widget",
			},
			Boost: 2.0,
		},
		Right: &parser.FieldQuery{
			Field: "status",
			Value: "active",
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)

	// SQL should be normal AND query
	assert.Equal(t, "name = $1 AND status = $2", output.WhereClause)
	assert.Len(t, output.Parameters, 2)

	// Boost should be in metadata
	boosts := output.Metadata["boosts"].([]map[string]interface{})
	require.Len(t, boosts, 1)
	assert.Equal(t, 2.0, boosts[0]["boost"])
}

func TestPostgresTranslator_BoostQuery_MultipleBoosts(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema(
		"products",
		map[string]schema.Field{
			"name":        {Type: schema.TypeText},
			"description": {Type: schema.TypeText},
		},
		schema.SchemaOptions{})

	// name:widget^2 OR description:gadget^3
	ast := &parser.BinaryOp{
		Op: "OR",
		Left: &parser.BoostQuery{
			Query: &parser.FieldQuery{
				Field: "name",
				Value: "widget",
			},
			Boost: 2.0,
		},
		Right: &parser.BoostQuery{
			Query: &parser.FieldQuery{
				Field: "description",
				Value: "gadget",
			},
			Boost: 3.0,
		},
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)

	// SQL should be normal OR query
	assert.Equal(t, "name = $1 OR description = $2", output.WhereClause)

	// Should have two boosts in metadata
	boosts := output.Metadata["boosts"].([]map[string]interface{})
	require.Len(t, boosts, 2)
	assert.Equal(t, 2.0, boosts[0]["boost"])
	assert.Equal(t, 3.0, boosts[1]["boost"])
}

func TestPostgresTranslator_BoostQuery_NestedBoost(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema(
		"products",
		map[string]schema.Field{
			"name":   {Type: schema.TypeText},
			"region": {Type: schema.TypeText},
		},
		schema.SchemaOptions{})

	// (name:widget AND region:us)^2
	ast := &parser.BoostQuery{
		Query: &parser.BinaryOp{
			Op: "AND",
			Left: &parser.FieldQuery{
				Field: "name",
				Value: "widget",
			},
			Right: &parser.FieldQuery{
				Field: "region",
				Value: "us",
			},
		},
		Boost: 2.0,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)

	// SQL should be normal AND query
	assert.Equal(t, "name = $1 AND region = $2", output.WhereClause)

	// Boost should apply to the whole binary operation
	boosts := output.Metadata["boosts"].([]map[string]interface{})
	require.Len(t, boosts, 1)
	assert.Equal(t, "binary_op", boosts[0]["query"])
	assert.Equal(t, 2.0, boosts[0]["boost"])
}

func TestPostgresTranslator_BoostQuery_RangeQuery(t *testing.T) {
	translator := NewPostgresTranslator()

	testSchema := schema.NewSchema(
		"products",
		map[string]schema.Field{
			"price": {Type: schema.TypeInteger},
		},
		schema.SchemaOptions{})

	// price:[10 TO 100]^1.5
	ast := &parser.BoostQuery{
		Query: &parser.RangeQuery{
			Field:          "price",
			Start:          10,
			End:            100,
			InclusiveStart: true,
			InclusiveEnd:   true,
		},
		Boost: 1.5,
	}

	output, err := translator.Translate(ast, testSchema)
	require.NoError(t, err)
	assert.NotNil(t, output)

	// SQL should be normal range query
	assert.Equal(t, "price BETWEEN $1 AND $2", output.WhereClause)

	// Boost should be in metadata
	boosts := output.Metadata["boosts"].([]map[string]interface{})
	require.Len(t, boosts, 1)
	assert.Equal(t, "range_query", boosts[0]["query"])
	assert.Equal(t, 1.5, boosts[0]["boost"])
}
