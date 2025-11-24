package translator

import (
	"fmt"
	"sync"

	"github.com/infiniv/rsearch/internal/schema"
)

// Translator converts AST nodes to database-specific query formats.
type Translator interface {
	// Translate converts an AST node to database-specific output.
	Translate(ast Node, schema *schema.Schema) (*TranslatorOutput, error)

	// DatabaseType returns the database type this translator targets.
	DatabaseType() string
}

// TranslatorOutput represents the result of translating an AST.
type TranslatorOutput struct {
	// Type indicates the output format ("sql", "mongodb", "elasticsearch")
	Type string

	// SQL-specific fields
	WhereClause    string
	Parameters     []interface{}
	ParameterTypes []string

	// NoSQL-specific fields
	Filter interface{} // MongoDB filter, ES query DSL
}

// Registry manages translator instances.
type Registry struct {
	translators map[string]Translator
	mu          sync.RWMutex
}

// NewRegistry creates a new translator registry.
func NewRegistry() *Registry {
	return &Registry{
		translators: make(map[string]Translator),
	}
}

// Register adds a translator to the registry.
func (r *Registry) Register(dbType string, translator Translator) error {
	if dbType == "" {
		return fmt.Errorf("database type cannot be empty")
	}
	if translator == nil {
		return fmt.Errorf("translator cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.translators[dbType]; exists {
		return fmt.Errorf("translator for %s already registered", dbType)
	}

	r.translators[dbType] = translator
	return nil
}

// Get retrieves a translator by database type.
func (r *Registry) Get(dbType string) (Translator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	translator, exists := r.translators[dbType]
	if !exists {
		return nil, fmt.Errorf("translator for %s not found", dbType)
	}

	return translator, nil
}

// List returns all registered database types.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.translators))
	for dbType := range r.translators {
		types = append(types, dbType)
	}
	return types
}
