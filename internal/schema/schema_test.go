package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	assert.NotNil(t, registry)
	assert.Empty(t, registry.List())
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	schema := &Schema{
		Name: "products",
		Fields: map[string]*Field{
			"product_code": {Name: "product_code", Type: "text", Searchable: true},
		},
	}

	// Test successful registration
	err := registry.Register(schema)
	assert.NoError(t, err)

	// Test duplicate registration
	err = registry.Register(schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Test nil schema
	err = registry.Register(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")

	// Test empty name
	emptySchema := &Schema{Name: ""}
	err = registry.Register(emptySchema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()

	schema := &Schema{
		Name: "products",
		Fields: map[string]*Field{
			"product_code": {Name: "product_code", Type: "text", Searchable: true},
		},
	}

	// Register schema
	err := registry.Register(schema)
	require.NoError(t, err)

	// Test successful retrieval
	retrieved, err := registry.Get("products")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "products", retrieved.Name)

	// Test non-existent schema
	retrieved, err = registry.Get("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, retrieved)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	// Initially empty
	list := registry.List()
	assert.Empty(t, list)

	// Register some schemas
	schema1 := &Schema{Name: "products", Fields: map[string]*Field{}}
	schema2 := &Schema{Name: "users", Fields: map[string]*Field{}}

	registry.Register(schema1)
	registry.Register(schema2)

	// Verify list
	list = registry.List()
	assert.Len(t, list, 2)
	assert.Contains(t, list, "products")
	assert.Contains(t, list, "users")
}

func TestSchema_GetField(t *testing.T) {
	schema := &Schema{
		Name: "products",
		Fields: map[string]*Field{
			"product_code": {Name: "product_code", Type: "text", Searchable: true},
			"region":       {Name: "region", Type: "text", Searchable: true},
		},
	}

	// Test successful retrieval
	field, err := schema.GetField("product_code")
	assert.NoError(t, err)
	assert.NotNil(t, field)
	assert.Equal(t, "product_code", field.Name)
	assert.Equal(t, "text", field.Type)
	assert.True(t, field.Searchable)

	// Test non-existent field
	field, err = schema.GetField("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, field)
	assert.Contains(t, err.Error(), "not found")
}

func TestField_Properties(t *testing.T) {
	field := &Field{
		Name:       "test_field",
		Type:       "number",
		Searchable: true,
		Sortable:   false,
	}

	assert.Equal(t, "test_field", field.Name)
	assert.Equal(t, "number", field.Type)
	assert.True(t, field.Searchable)
	assert.False(t, field.Sortable)
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()

	schema := &Schema{
		Name: "products",
		Fields: map[string]*Field{
			"product_code": {Name: "product_code", Type: "text", Searchable: true},
		},
	}

	// Register schema
	registry.Register(schema)

	// Concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := registry.Get("products")
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
