package schema

import (
	"testing"
)

func TestResolveField_ExactMatch(t *testing.T) {
	fields := map[string]Field{
		"userName": {Type: TypeText, Indexed: true},
		"userId":   {Type: TypeInteger},
	}
	options := SchemaOptions{
		NamingConvention: "snake_case",
		StrictFieldNames: true,
	}
	schema := NewSchema("test", fields, options)

	tests := []struct {
		name          string
		queryField    string
		wantColumn    string
		wantFieldType FieldType
		wantErr       bool
	}{
		{
			name:          "exact match userName",
			queryField:    "userName",
			wantColumn:    "user_name",
			wantFieldType: TypeText,
			wantErr:       false,
		},
		{
			name:          "exact match userId",
			queryField:    "userId",
			wantColumn:    "user_id",
			wantFieldType: TypeInteger,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column, field, err := schema.ResolveField(tt.queryField)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if column != tt.wantColumn {
					t.Errorf("ResolveField() column = %v, want %v", column, tt.wantColumn)
				}
				if field.Type != tt.wantFieldType {
					t.Errorf("ResolveField() field.Type = %v, want %v", field.Type, tt.wantFieldType)
				}
			}
		})
	}
}

func TestResolveField_CaseInsensitive(t *testing.T) {
	fields := map[string]Field{
		"userName": {Type: TypeText},
		"userAge":  {Type: TypeInteger},
	}
	options := SchemaOptions{
		NamingConvention: "snake_case",
		StrictFieldNames: false, // Case-insensitive
	}
	schema := NewSchema("test", fields, options)

	tests := []struct {
		name          string
		queryField    string
		wantColumn    string
		wantFieldType FieldType
		wantErr       bool
	}{
		{
			name:          "lowercase query",
			queryField:    "username",
			wantColumn:    "user_name",
			wantFieldType: TypeText,
			wantErr:       false,
		},
		{
			name:          "uppercase query",
			queryField:    "USERNAME",
			wantColumn:    "user_name",
			wantFieldType: TypeText,
			wantErr:       false,
		},
		{
			name:          "mixed case query",
			queryField:    "UsErNaMe",
			wantColumn:    "user_name",
			wantFieldType: TypeText,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column, field, err := schema.ResolveField(tt.queryField)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if column != tt.wantColumn {
					t.Errorf("ResolveField() column = %v, want %v", column, tt.wantColumn)
				}
				if field.Type != tt.wantFieldType {
					t.Errorf("ResolveField() field.Type = %v, want %v", field.Type, tt.wantFieldType)
				}
			}
		})
	}
}

func TestResolveField_Aliases(t *testing.T) {
	fields := map[string]Field{
		"userName": {
			Type:    TypeText,
			Aliases: []string{"user", "name", "login"},
		},
	}
	options := SchemaOptions{
		NamingConvention: "snake_case",
		StrictFieldNames: false,
	}
	schema := NewSchema("test", fields, options)

	tests := []struct {
		name       string
		queryField string
		wantColumn string
		wantErr    bool
	}{
		{
			name:       "alias: user",
			queryField: "user",
			wantColumn: "user_name",
			wantErr:    false,
		},
		{
			name:       "alias: name",
			queryField: "name",
			wantColumn: "user_name",
			wantErr:    false,
		},
		{
			name:       "alias: login",
			queryField: "login",
			wantColumn: "user_name",
			wantErr:    false,
		},
		{
			name:       "alias case insensitive",
			queryField: "LOGIN",
			wantColumn: "user_name",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column, _, err := schema.ResolveField(tt.queryField)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && column != tt.wantColumn {
				t.Errorf("ResolveField() column = %v, want %v", column, tt.wantColumn)
			}
		})
	}
}

func TestResolveField_NamingConventionTransform(t *testing.T) {
	fields := map[string]Field{
		"productCode": {Type: TypeText},
		"orderDate":   {Type: TypeDate},
	}
	options := SchemaOptions{
		NamingConvention: "snake_case",
		StrictFieldNames: false,
	}
	schema := NewSchema("test", fields, options)

	tests := []struct {
		name       string
		queryField string
		wantColumn string
		wantErr    bool
	}{
		{
			name:       "camelCase to snake_case",
			queryField: "productCode",
			wantColumn: "product_code",
			wantErr:    false,
		},
		{
			name:       "UPPERCASE transformed and matched",
			queryField: "PRODUCTCODE",
			wantColumn: "product_code",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column, _, err := schema.ResolveField(tt.queryField)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && column != tt.wantColumn {
				t.Errorf("ResolveField() column = %v, want %v", column, tt.wantColumn)
			}
		})
	}
}

func TestResolveField_ExplicitColumn(t *testing.T) {
	fields := map[string]Field{
		"userName": {
			Type:   TypeText,
			Column: "usr_nm", // Explicit column override
		},
	}
	options := SchemaOptions{
		NamingConvention: "snake_case",
	}
	schema := NewSchema("test", fields, options)

	column, field, err := schema.ResolveField("userName")
	if err != nil {
		t.Fatalf("ResolveField() unexpected error = %v", err)
	}

	if column != "usr_nm" {
		t.Errorf("ResolveField() column = %v, want %v", column, "usr_nm")
	}
	if field.Type != TypeText {
		t.Errorf("ResolveField() field.Type = %v, want %v", field.Type, TypeText)
	}
}

func TestResolveField_NotFound(t *testing.T) {
	fields := map[string]Field{
		"userName": {Type: TypeText},
	}
	options := SchemaOptions{
		NamingConvention: "snake_case",
		StrictFieldNames: true,
	}
	schema := NewSchema("test", fields, options)

	tests := []struct {
		name       string
		queryField string
	}{
		{
			name:       "non-existent field",
			queryField: "nonExistent",
		},
		{
			name:       "case mismatch with strict",
			queryField: "username",
		},
		{
			name:       "empty field",
			queryField: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := schema.ResolveField(tt.queryField)
			if err == nil {
				t.Errorf("ResolveField() expected error for field %q, got nil", tt.queryField)
			}
		})
	}
}

func TestResolveField_NoNamingConvention(t *testing.T) {
	fields := map[string]Field{
		"userName": {Type: TypeText},
	}
	options := SchemaOptions{
		NamingConvention: "none",
		StrictFieldNames: true,
	}
	schema := NewSchema("test", fields, options)

	column, _, err := schema.ResolveField("userName")
	if err != nil {
		t.Fatalf("ResolveField() unexpected error = %v", err)
	}

	// With "none" convention, column should match field name exactly
	if column != "userName" {
		t.Errorf("ResolveField() column = %v, want %v", column, "userName")
	}
}

func TestIsValidFieldType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		want      bool
	}{
		{"text type", TypeText, true},
		{"integer type", TypeInteger, true},
		{"float type", TypeFloat, true},
		{"boolean type", TypeBoolean, true},
		{"datetime type", TypeDateTime, true},
		{"date type", TypeDate, true},
		{"time type", TypeTime, true},
		{"json type", TypeJSON, true},
		{"array type", TypeArray, true},
		{"invalid type", FieldType("invalid"), false},
		{"empty type", FieldType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidFieldType(tt.fieldType); got != tt.want {
				t.Errorf("IsValidFieldType(%v) = %v, want %v", tt.fieldType, got, tt.want)
			}
		})
	}
}
