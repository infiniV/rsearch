package schema

import (
	"fmt"
	"sync"
)

// Registry is a thread-safe in-memory storage for schemas
type Registry struct {
	schemas map[string]*Schema
	mu      sync.RWMutex
}

// NewRegistry creates a new schema registry
func NewRegistry() *Registry {
	return &Registry{
		schemas: make(map[string]*Schema),
	}
}

// Register adds a schema to the registry after validation
// Returns an error if the schema is invalid or already exists
func (r *Registry) Register(schema *Schema) error {
	if schema == nil {
		return fmt.Errorf("schema is nil")
	}

	// Validate schema before registration
	if err := ValidateSchema(schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate
	if _, exists := r.schemas[schema.Name]; exists {
		return fmt.Errorf("schema %q already exists", schema.Name)
	}

	// Pre-compute field mappings for fast lookups
	schema.buildLookupCache()

	// Store schema
	r.schemas[schema.Name] = schema

	return nil
}

// Get retrieves a schema by name
// Returns an error if the schema does not exist
func (r *Registry) Get(name string) (*Schema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schema, exists := r.schemas[name]
	if !exists {
		return nil, fmt.Errorf("schema %q not found", name)
	}

	return schema, nil
}

// Delete removes a schema from the registry
// Returns an error if the schema does not exist
func (r *Registry) Delete(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.schemas[name]; !exists {
		return fmt.Errorf("schema %q not found", name)
	}

	delete(r.schemas, name)
	return nil
}

// List returns all registered schemas
// Returns a copy of the schema list to prevent external modification
func (r *Registry) List() []*Schema {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schemas := make([]*Schema, 0, len(r.schemas))
	for _, schema := range r.schemas {
		schemas = append(schemas, schema)
	}

	return schemas
}

// Count returns the number of registered schemas
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.schemas)
}

// Exists checks if a schema with the given name exists
func (r *Registry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.schemas[name]
	return exists
}
