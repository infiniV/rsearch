package translator

import (
	"testing"

	"github.com/infiniv/rsearch/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegistryRegister tests registering translators.
func TestRegistryRegister(t *testing.T) {
	registry := NewRegistry()

	// Create a mock translator
	mockTranslator := &MockTranslator{dbType: "postgres"}

	// Test successful registration
	err := registry.Register("postgres", mockTranslator)
	assert.NoError(t, err)

	// Test duplicate registration
	err = registry.Register("postgres", mockTranslator)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Test empty database type
	err = registry.Register("", mockTranslator)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")

	// Test nil translator
	err = registry.Register("mysql", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestRegistryGet tests retrieving translators.
func TestRegistryGet(t *testing.T) {
	registry := NewRegistry()
	mockTranslator := &MockTranslator{dbType: "postgres"}

	// Register a translator
	err := registry.Register("postgres", mockTranslator)
	require.NoError(t, err)

	// Test successful retrieval
	translator, err := registry.Get("postgres")
	assert.NoError(t, err)
	assert.NotNil(t, translator)
	assert.Equal(t, "postgres", translator.DatabaseType())

	// Test non-existent translator
	translator, err = registry.Get("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, translator)
	assert.Contains(t, err.Error(), "not found")
}

// TestRegistryList tests listing all translators.
func TestRegistryList(t *testing.T) {
	registry := NewRegistry()

	// Initially empty
	list := registry.List()
	assert.Empty(t, list)

	// Register some translators
	registry.Register("postgres", &MockTranslator{dbType: "postgres"})
	registry.Register("mongodb", &MockTranslator{dbType: "mongodb"})

	// Verify list
	list = registry.List()
	assert.Len(t, list, 2)
	assert.Contains(t, list, "postgres")
	assert.Contains(t, list, "mongodb")
}

// TestTranslatorOutput tests the TranslatorOutput structure.
func TestTranslatorOutput(t *testing.T) {
	// Test SQL output
	output := &TranslatorOutput{
		Type:           "sql",
		WhereClause:    "field = $1",
		Parameters:     []interface{}{"value"},
		ParameterTypes: []string{"text"},
	}

	assert.Equal(t, "sql", output.Type)
	assert.Equal(t, "field = $1", output.WhereClause)
	assert.Len(t, output.Parameters, 1)
	assert.Equal(t, "value", output.Parameters[0])

	// Test NoSQL output
	noSQLOutput := &TranslatorOutput{
		Type:   "mongodb",
		Filter: map[string]interface{}{"field": "value"},
	}

	assert.Equal(t, "mongodb", noSQLOutput.Type)
	assert.NotNil(t, noSQLOutput.Filter)
}

// MockTranslator is a mock implementation for testing.
type MockTranslator struct {
	dbType string
	output *TranslatorOutput
	err    error
}

func (m *MockTranslator) Translate(ast Node, schema *schema.Schema) (*TranslatorOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.output, nil
}

func (m *MockTranslator) DatabaseType() string {
	return m.dbType
}
