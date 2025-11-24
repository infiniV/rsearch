package schema

import (
	"fmt"
	"sync"
)

// Schema represents a search schema with field definitions.
type Schema struct {
	Name   string
	Fields map[string]*Field
}

// Field represents a field in a schema.
type Field struct {
	Name       string
	Type       string // "text", "number", "date", "boolean"
	Searchable bool
	Sortable   bool
}

// Registry manages schemas.
type Registry struct {
	schemas map[string]*Schema
	mu      sync.RWMutex
}

// NewRegistry creates a new schema registry.
func NewRegistry() *Registry {
	return &Registry{
		schemas: make(map[string]*Schema),
	}
}

// Register adds a schema to the registry.
func (r *Registry) Register(schema *Schema) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}
	if schema.Name == "" {
		return fmt.Errorf("schema name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.schemas[schema.Name]; exists {
		return fmt.Errorf("schema %s already registered", schema.Name)
	}

	r.schemas[schema.Name] = schema
	return nil
}

// Get retrieves a schema by name.
func (r *Registry) Get(name string) (*Schema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schema, exists := r.schemas[name]
	if !exists {
		return nil, fmt.Errorf("schema %s not found", name)
	}

	return schema, nil
}

// List returns all registered schema names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.schemas))
	for name := range r.schemas {
		names = append(names, name)
	}
	return names
}

// GetField returns a field by name from the schema.
func (s *Schema) GetField(name string) (*Field, error) {
	field, exists := s.Fields[name]
	if !exists {
		return nil, fmt.Errorf("field %s not found in schema %s", name, s.Name)
	}
	return field, nil
}
