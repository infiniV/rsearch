package translator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSQLOutput(t *testing.T) {
	whereClause := "field = $1 AND other = $2"
	params := []interface{}{"value1", 42}
	types := []string{"text", "integer"}

	output := NewSQLOutput(whereClause, params, types)

	assert.Equal(t, "sql", output.Type)
	assert.Equal(t, whereClause, output.WhereClause)
	assert.Equal(t, params, output.Parameters)
	assert.Equal(t, types, output.ParameterTypes)
	assert.Nil(t, output.Filter)
}

func TestNewMongoDBOutput(t *testing.T) {
	filter := map[string]interface{}{
		"field": "value",
		"age":   map[string]interface{}{"$gt": 18},
	}

	output := NewMongoDBOutput(filter)

	assert.Equal(t, "mongodb", output.Type)
	assert.Equal(t, filter, output.Filter)
	assert.Empty(t, output.WhereClause)
	assert.Nil(t, output.Parameters)
	assert.Nil(t, output.ParameterTypes)
}

func TestNewElasticsearchOutput(t *testing.T) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"field": "value",
			},
		},
	}

	output := NewElasticsearchOutput(query)

	assert.Equal(t, "elasticsearch", output.Type)
	assert.Equal(t, query, output.Filter)
	assert.Empty(t, output.WhereClause)
	assert.Nil(t, output.Parameters)
	assert.Nil(t, output.ParameterTypes)
}
