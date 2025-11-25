package schema

import (
	"fmt"
	"sync"
	"testing"
)

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	schema := &Schema{
		Name: "users",
		Fields: map[string]Field{
			"userName": {Type: TypeText},
		},
		Options: SchemaOptions{
			NamingConvention: "snake_case",
		},
	}

	err := registry.Register(schema)
	if err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}

	// Verify schema can be retrieved
	retrieved, err := registry.Get("users")
	if err != nil {
		t.Fatalf("Get() unexpected error = %v", err)
	}
	if retrieved.Name != "users" {
		t.Errorf("Get() name = %v, want %v", retrieved.Name, "users")
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	registry := NewRegistry()

	schema1 := &Schema{
		Name: "users",
		Fields: map[string]Field{
			"userName": {Type: TypeText},
		},
		Options: SchemaOptions{},
	}

	schema2 := &Schema{
		Name: "users",
		Fields: map[string]Field{
			"userAge": {Type: TypeInteger},
		},
		Options: SchemaOptions{},
	}

	err := registry.Register(schema1)
	if err != nil {
		t.Fatalf("Register() first schema unexpected error = %v", err)
	}

	err = registry.Register(schema2)
	if err == nil {
		t.Error("Register() expected error for duplicate schema, got nil")
	}
}

func TestRegistry_RegisterInvalid(t *testing.T) {
	registry := NewRegistry()

	schema := &Schema{
		Name:    "invalid",
		Fields:  map[string]Field{}, // Empty fields - invalid
		Options: SchemaOptions{},
	}

	err := registry.Register(schema)
	if err == nil {
		t.Error("Register() expected error for invalid schema, got nil")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()

	schema := &Schema{
		Name: "products",
		Fields: map[string]Field{
			"productCode": {Type: TypeText},
		},
		Options: SchemaOptions{},
	}

	err := registry.Register(schema)
	if err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}

	// Test successful get
	retrieved, err := registry.Get("products")
	if err != nil {
		t.Fatalf("Get() unexpected error = %v", err)
	}
	if retrieved.Name != "products" {
		t.Errorf("Get() name = %v, want %v", retrieved.Name, "products")
	}

	// Test get non-existent
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Get() expected error for non-existent schema, got nil")
	}
}

func TestRegistry_Delete(t *testing.T) {
	registry := NewRegistry()

	schema := &Schema{
		Name: "orders",
		Fields: map[string]Field{
			"orderId": {Type: TypeInteger},
		},
		Options: SchemaOptions{},
	}

	err := registry.Register(schema)
	if err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}

	// Delete schema
	err = registry.Delete("orders")
	if err != nil {
		t.Fatalf("Delete() unexpected error = %v", err)
	}

	// Verify it's gone
	_, err = registry.Get("orders")
	if err == nil {
		t.Error("Get() expected error after delete, got nil")
	}
}

func TestRegistry_DeleteNonExistent(t *testing.T) {
	registry := NewRegistry()

	err := registry.Delete("nonexistent")
	if err == nil {
		t.Error("Delete() expected error for non-existent schema, got nil")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	schema1 := &Schema{
		Name: "users",
		Fields: map[string]Field{
			"userName": {Type: TypeText},
		},
		Options: SchemaOptions{},
	}

	schema2 := &Schema{
		Name: "products",
		Fields: map[string]Field{
			"productCode": {Type: TypeText},
		},
		Options: SchemaOptions{},
	}

	err := registry.Register(schema1)
	if err != nil {
		t.Fatalf("Register() schema1 unexpected error = %v", err)
	}

	err = registry.Register(schema2)
	if err != nil {
		t.Fatalf("Register() schema2 unexpected error = %v", err)
	}

	// List all schemas
	schemas := registry.List()
	if len(schemas) != 2 {
		t.Errorf("List() length = %v, want %v", len(schemas), 2)
	}

	// Verify both schemas are present
	names := make(map[string]bool)
	for _, s := range schemas {
		names[s.Name] = true
	}
	if !names["users"] || !names["products"] {
		t.Error("List() missing expected schemas")
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()

	// Pre-register a schema
	baseSchema := &Schema{
		Name: "base",
		Fields: map[string]Field{
			"field": {Type: TypeText},
		},
		Options: SchemaOptions{},
	}
	err := registry.Register(baseSchema)
	if err != nil {
		t.Fatalf("Register() base schema unexpected error = %v", err)
	}

	const numGoroutines = 100
	const numOperations = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				_, err := registry.Get("base")
				if err != nil {
					t.Errorf("Concurrent Get() error = %v", err)
				}
				_ = registry.List()
			}
		}(i)
	}

	wg.Wait()
}

func TestRegistry_ConcurrentWrites(t *testing.T) {
	registry := NewRegistry()

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent writes (each goroutine writes unique schemas)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			schema := &Schema{
				Name: fmt.Sprintf("schema_%d", id),
				Fields: map[string]Field{
					"field": {Type: TypeText},
				},
				Options: SchemaOptions{},
			}
			err := registry.Register(schema)
			if err != nil {
				t.Errorf("Concurrent Register() error = %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all schemas were registered
	schemas := registry.List()
	if len(schemas) != numGoroutines {
		t.Errorf("List() length = %v, want %v", len(schemas), numGoroutines)
	}
}
