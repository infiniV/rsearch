package schema

import (
	"testing"
)

func TestValidateSchema_Valid(t *testing.T) {
	tests := []struct {
		name   string
		schema *Schema
	}{
		{
			name: "simple valid schema",
			schema: &Schema{
				Name: "users",
				Fields: map[string]Field{
					"userName": {Type: TypeText},
					"userAge":  {Type: TypeInteger},
				},
				Options: SchemaOptions{
					NamingConvention: "snake_case",
				},
			},
		},
		{
			name: "schema with aliases",
			schema: &Schema{
				Name: "products",
				Fields: map[string]Field{
					"productCode": {
						Type:    TypeText,
						Aliases: []string{"code", "sku"},
					},
				},
				Options: SchemaOptions{
					NamingConvention: "snake_case",
				},
			},
		},
		{
			name: "schema with explicit columns",
			schema: &Schema{
				Name: "orders",
				Fields: map[string]Field{
					"orderId": {
						Type:   TypeInteger,
						Column: "order_id",
					},
				},
				Options: SchemaOptions{
					NamingConvention: "none",
				},
			},
		},
		{
			name: "schema with default field",
			schema: &Schema{
				Name: "search",
				Fields: map[string]Field{
					"title":   {Type: TypeText},
					"content": {Type: TypeText},
				},
				Options: SchemaOptions{
					NamingConvention: "snake_case",
					DefaultField:     "content",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSchema(tt.schema)
			if err != nil {
				t.Errorf("ValidateSchema() error = %v, want nil", err)
			}
		})
	}
}

func TestValidateSchema_InvalidName(t *testing.T) {
	schema := &Schema{
		Name: "",
		Fields: map[string]Field{
			"field1": {Type: TypeText},
		},
		Options: SchemaOptions{},
	}

	err := ValidateSchema(schema)
	if err == nil {
		t.Error("ValidateSchema() expected error for empty name, got nil")
	}
}

func TestValidateSchema_NoFields(t *testing.T) {
	schema := &Schema{
		Name:    "test",
		Fields:  map[string]Field{},
		Options: SchemaOptions{},
	}

	err := ValidateSchema(schema)
	if err == nil {
		t.Error("ValidateSchema() expected error for no fields, got nil")
	}
}

func TestValidateSchema_EmptyFieldName(t *testing.T) {
	schema := &Schema{
		Name: "test",
		Fields: map[string]Field{
			"": {Type: TypeText},
		},
		Options: SchemaOptions{},
	}

	err := ValidateSchema(schema)
	if err == nil {
		t.Error("ValidateSchema() expected error for empty field name, got nil")
	}
}

func TestValidateSchema_InvalidFieldType(t *testing.T) {
	schema := &Schema{
		Name: "test",
		Fields: map[string]Field{
			"field1": {Type: FieldType("invalid")},
		},
		Options: SchemaOptions{},
	}

	err := ValidateSchema(schema)
	if err == nil {
		t.Error("ValidateSchema() expected error for invalid field type, got nil")
	}
}

func TestValidateSchema_DuplicateAliases(t *testing.T) {
	schema := &Schema{
		Name: "test",
		Fields: map[string]Field{
			"field1": {
				Type:    TypeText,
				Aliases: []string{"alias1", "alias2"},
			},
			"field2": {
				Type:    TypeText,
				Aliases: []string{"alias2", "alias3"}, // "alias2" is duplicate
			},
		},
		Options: SchemaOptions{},
	}

	err := ValidateSchema(schema)
	if err == nil {
		t.Error("ValidateSchema() expected error for duplicate aliases, got nil")
	}
}

func TestValidateSchema_AliasMatchesFieldName(t *testing.T) {
	schema := &Schema{
		Name: "test",
		Fields: map[string]Field{
			"field1": {
				Type:    TypeText,
				Aliases: []string{"field2"}, // Alias matches another field name
			},
			"field2": {
				Type: TypeText,
			},
		},
		Options: SchemaOptions{},
	}

	err := ValidateSchema(schema)
	if err == nil {
		t.Error("ValidateSchema() expected error for alias matching field name, got nil")
	}
}

func TestValidateSchema_InvalidColumnName(t *testing.T) {
	tests := []struct {
		name       string
		columnName string
	}{
		{"column with spaces", "user name"},
		{"column with special chars", "user@name"},
		{"column with dash", "user-name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := &Schema{
				Name: "test",
				Fields: map[string]Field{
					"field1": {
						Type:   TypeText,
						Column: tt.columnName,
					},
				},
				Options: SchemaOptions{},
			}

			err := ValidateSchema(schema)
			if err == nil {
				t.Errorf("ValidateSchema() expected error for column name %q, got nil", tt.columnName)
			}
		})
	}
}

func TestValidateSchema_InvalidNamingConvention(t *testing.T) {
	schema := &Schema{
		Name: "test",
		Fields: map[string]Field{
			"field1": {Type: TypeText},
		},
		Options: SchemaOptions{
			NamingConvention: "invalid_convention",
		},
	}

	err := ValidateSchema(schema)
	if err == nil {
		t.Error("ValidateSchema() expected error for invalid naming convention, got nil")
	}
}

func TestValidateSchema_DefaultFieldNotFound(t *testing.T) {
	schema := &Schema{
		Name: "test",
		Fields: map[string]Field{
			"field1": {Type: TypeText},
		},
		Options: SchemaOptions{
			DefaultField: "nonexistent",
		},
	}

	err := ValidateSchema(schema)
	if err == nil {
		t.Error("ValidateSchema() expected error for non-existent default field, got nil")
	}
}
