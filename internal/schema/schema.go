package schema

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// FieldType represents the data type of a field
type FieldType string

const (
	TypeText     FieldType = "text"
	TypeInteger  FieldType = "integer"
	TypeFloat    FieldType = "float"
	TypeBoolean  FieldType = "boolean"
	TypeDateTime FieldType = "datetime"
	TypeDate     FieldType = "date"
	TypeTime     FieldType = "time"
	TypeJSON     FieldType = "json"
	TypeArray    FieldType = "array"
)

// Field represents a schema field definition
type Field struct {
	Type    FieldType `json:"type"`
	Column  string    `json:"column,omitempty"`   // Optional: explicit column name override
	Indexed bool      `json:"indexed"`            // Hint for translators
	Aliases []string  `json:"aliases,omitempty"`  // Alternative field names
}

// EnabledFeatures contains flags for optional database features
type EnabledFeatures struct {
	Fuzzy     bool `json:"fuzzy"`     // Enable fuzzy search (requires pg_trgm for PostgreSQL)
	Proximity bool `json:"proximity"` // Enable proximity search (requires full-text search)
	Regex     bool `json:"regex"`     // Enable regex search
}

// SchemaOptions contains configuration options for a schema
type SchemaOptions struct {
	NamingConvention string           `json:"namingConvention"` // "snake_case", "camelCase", "PascalCase", "none"
	StrictOperators  bool             `json:"strictOperators"`  // case-sensitive AND/OR/NOT
	StrictFieldNames bool             `json:"strictFieldNames"` // case-sensitive field names
	DefaultField     string           `json:"defaultField"`     // field for queries without field specifier
	EnabledFeatures  EnabledFeatures  `json:"enabledFeatures"`  // Optional database features
}

// Schema represents a schema definition
type Schema struct {
	Name      string        `json:"name"`
	Fields    map[string]Field `json:"fields"`
	Options   SchemaOptions `json:"options"`
	CreatedAt time.Time     `json:"createdAt"`

	// Internal cache for fast lookups
	lowerFieldMap map[string]string   // lowercase field name -> actual field name
	aliasMap      map[string]string   // alias (normalized) -> field name
}

// NewSchema creates a new schema with the given name and fields
func NewSchema(name string, fields map[string]Field, options SchemaOptions) *Schema {
	s := &Schema{
		Name:      name,
		Fields:    fields,
		Options:   options,
		CreatedAt: time.Now(),
	}
	s.buildLookupCache()
	return s
}

// buildLookupCache pre-computes field mappings for fast resolution
func (s *Schema) buildLookupCache() {
	s.lowerFieldMap = make(map[string]string)
	s.aliasMap = make(map[string]string)

	for fieldName, field := range s.Fields {
		// Build case-insensitive lookup
		s.lowerFieldMap[strings.ToLower(fieldName)] = fieldName

		// Build alias lookup
		for _, alias := range field.Aliases {
			normalizedAlias := alias
			if !s.Options.StrictFieldNames {
				normalizedAlias = strings.ToLower(alias)
			}
			s.aliasMap[normalizedAlias] = fieldName
		}
	}
}

// ResolveField resolves a query field name to its actual column name and field definition
// Resolution order:
// 1. Exact match
// 2. Case-insensitive match (if strictFieldNames: false)
// 3. Alias lookup
// 4. Transform via naming convention and match
func (s *Schema) ResolveField(queryField string) (columnName string, field *Field, err error) {
	if queryField == "" {
		return "", nil, errors.New("empty field name")
	}

	// Stage 1: Exact match
	if f, exists := s.Fields[queryField]; exists {
		return s.getColumnName(queryField, &f), &f, nil
	}

	// Stage 2: Case-insensitive match (if enabled)
	if !s.Options.StrictFieldNames {
		if actualField, exists := s.lowerFieldMap[strings.ToLower(queryField)]; exists {
			f := s.Fields[actualField]
			return s.getColumnName(actualField, &f), &f, nil
		}
	}

	// Stage 3: Alias lookup
	lookupKey := queryField
	if !s.Options.StrictFieldNames {
		lookupKey = strings.ToLower(queryField)
	}
	if actualField, exists := s.aliasMap[lookupKey]; exists {
		f := s.Fields[actualField]
		return s.getColumnName(actualField, &f), &f, nil
	}

	// Stage 4: Transform via naming convention and match
	if s.Options.NamingConvention != "" && s.Options.NamingConvention != "none" {
		transformed := s.transformFieldName(queryField)
		if transformed != queryField {
			// Try exact match with transformed name
			if f, exists := s.Fields[transformed]; exists {
				return s.getColumnName(transformed, &f), &f, nil
			}
			// Try case-insensitive match with transformed name
			if !s.Options.StrictFieldNames {
				if actualField, exists := s.lowerFieldMap[strings.ToLower(transformed)]; exists {
					f := s.Fields[actualField]
					return s.getColumnName(actualField, &f), &f, nil
				}
			}
		}
	}

	return "", nil, fmt.Errorf("field %q not found in schema %q", queryField, s.Name)
}

// getColumnName returns the column name for a field (using explicit column or field name)
func (s *Schema) getColumnName(fieldName string, field *Field) string {
	if field.Column != "" {
		return field.Column
	}

	// Apply naming convention transformation if needed
	if s.Options.NamingConvention == "" || s.Options.NamingConvention == "none" {
		return fieldName
	}

	return s.transformFieldName(fieldName)
}

// transformFieldName applies the schema's naming convention to transform a field name
func (s *Schema) transformFieldName(fieldName string) string {
	switch s.Options.NamingConvention {
	case "snake_case":
		return ToSnakeCase(fieldName)
	case "camelCase":
		return ToCamelCase(fieldName)
	case "PascalCase":
		return ToPascalCase(fieldName)
	default:
		return fieldName
	}
}

// ValidFieldTypes returns a list of all valid field types
func ValidFieldTypes() []FieldType {
	return []FieldType{
		TypeText,
		TypeInteger,
		TypeFloat,
		TypeBoolean,
		TypeDateTime,
		TypeDate,
		TypeTime,
		TypeJSON,
		TypeArray,
	}
}

// IsValidFieldType checks if a field type is valid
func IsValidFieldType(ft FieldType) bool {
	switch ft {
	case TypeText, TypeInteger, TypeFloat, TypeBoolean,
		TypeDateTime, TypeDate, TypeTime, TypeJSON, TypeArray:
		return true
	default:
		return false
	}
}
