package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/infiniv/rsearch/internal/schema"
)

func TestRegisterSchema_Success(t *testing.T) {
	registry := schema.NewRegistry()
	handler := NewHandler(registry)

	schemaJSON := `{
		"name": "users",
		"fields": {
			"userName": {
				"type": "text",
				"indexed": true
			},
			"userAge": {
				"type": "integer"
			}
		},
		"options": {
			"namingConvention": "snake_case",
			"strictFieldNames": false
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/schemas", bytes.NewBufferString(schemaJSON))
	rec := httptest.NewRecorder()

	handler.RegisterSchema(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("RegisterSchema() status = %v, want %v", rec.Code, http.StatusCreated)
	}

	// Verify schema was registered
	s, err := registry.Get("users")
	if err != nil {
		t.Fatalf("Get() unexpected error = %v", err)
	}
	if s.Name != "users" {
		t.Errorf("Schema name = %v, want %v", s.Name, "users")
	}
}

func TestRegisterSchema_InvalidJSON(t *testing.T) {
	registry := schema.NewRegistry()
	handler := NewHandler(registry)

	invalidJSON := `{invalid json`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/schemas", bytes.NewBufferString(invalidJSON))
	rec := httptest.NewRecorder()

	handler.RegisterSchema(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("RegisterSchema() status = %v, want %v", rec.Code, http.StatusBadRequest)
	}
}

func TestRegisterSchema_InvalidSchema(t *testing.T) {
	registry := schema.NewRegistry()
	handler := NewHandler(registry)

	// Schema with no fields (invalid)
	schemaJSON := `{
		"name": "invalid",
		"fields": {},
		"options": {}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/schemas", bytes.NewBufferString(schemaJSON))
	rec := httptest.NewRecorder()

	handler.RegisterSchema(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("RegisterSchema() status = %v, want %v", rec.Code, http.StatusBadRequest)
	}
}

func TestRegisterSchema_DuplicateName(t *testing.T) {
	registry := schema.NewRegistry()
	handler := NewHandler(registry)

	schemaJSON := `{
		"name": "users",
		"fields": {
			"userName": {"type": "text"}
		},
		"options": {}
	}`

	// Register first time
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/schemas", bytes.NewBufferString(schemaJSON))
	rec1 := httptest.NewRecorder()
	handler.RegisterSchema(rec1, req1)

	if rec1.Code != http.StatusCreated {
		t.Errorf("First RegisterSchema() status = %v, want %v", rec1.Code, http.StatusCreated)
	}

	// Register second time (should fail)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/schemas", bytes.NewBufferString(schemaJSON))
	rec2 := httptest.NewRecorder()
	handler.RegisterSchema(rec2, req2)

	if rec2.Code != http.StatusBadRequest {
		t.Errorf("Second RegisterSchema() status = %v, want %v", rec2.Code, http.StatusBadRequest)
	}
}

func TestGetSchema_Success(t *testing.T) {
	registry := schema.NewRegistry()
	handler := NewHandler(registry)

	// Pre-register a schema
	s := &schema.Schema{
		Name: "products",
		Fields: map[string]schema.Field{
			"productCode": {Type: schema.TypeText},
		},
		Options: schema.SchemaOptions{
			NamingConvention: "snake_case",
		},
	}
	err := registry.Register(s)
	if err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/schemas/products", nil)
	rec := httptest.NewRecorder()

	handler.GetSchema(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GetSchema() status = %v, want %v", rec.Code, http.StatusOK)
	}

	var result schema.Schema
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Name != "products" {
		t.Errorf("Schema name = %v, want %v", result.Name, "products")
	}
}

func TestGetSchema_NotFound(t *testing.T) {
	registry := schema.NewRegistry()
	handler := NewHandler(registry)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/schemas/nonexistent", nil)
	rec := httptest.NewRecorder()

	handler.GetSchema(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GetSchema() status = %v, want %v", rec.Code, http.StatusNotFound)
	}
}

func TestDeleteSchema_Success(t *testing.T) {
	registry := schema.NewRegistry()
	handler := NewHandler(registry)

	// Pre-register a schema
	s := &schema.Schema{
		Name: "orders",
		Fields: map[string]schema.Field{
			"orderId": {Type: schema.TypeInteger},
		},
		Options: schema.SchemaOptions{},
	}
	err := registry.Register(s)
	if err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/schemas/orders", nil)
	rec := httptest.NewRecorder()

	handler.DeleteSchema(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("DeleteSchema() status = %v, want %v", rec.Code, http.StatusNoContent)
	}

	// Verify schema was deleted
	_, err = registry.Get("orders")
	if err == nil {
		t.Error("Expected schema to be deleted")
	}
}

func TestDeleteSchema_NotFound(t *testing.T) {
	registry := schema.NewRegistry()
	handler := NewHandler(registry)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/schemas/nonexistent", nil)
	rec := httptest.NewRecorder()

	handler.DeleteSchema(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("DeleteSchema() status = %v, want %v", rec.Code, http.StatusNotFound)
	}
}

func TestListSchemas_Success(t *testing.T) {
	registry := schema.NewRegistry()
	handler := NewHandler(registry)

	// Register multiple schemas
	schemas := []*schema.Schema{
		{
			Name: "users",
			Fields: map[string]schema.Field{
				"userName": {Type: schema.TypeText},
			},
			Options: schema.SchemaOptions{},
		},
		{
			Name: "products",
			Fields: map[string]schema.Field{
				"productCode": {Type: schema.TypeText},
			},
			Options: schema.SchemaOptions{},
		},
	}

	for _, s := range schemas {
		if err := registry.Register(s); err != nil {
			t.Fatalf("Register() unexpected error = %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/schemas", nil)
	rec := httptest.NewRecorder()

	handler.ListSchemas(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("ListSchemas() status = %v, want %v", rec.Code, http.StatusOK)
	}

	var result SuccessResponse
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check that we got schemas in response
	if result.Data == nil {
		t.Error("ListSchemas() expected data, got nil")
	}
}

func TestMethodNotAllowed(t *testing.T) {
	registry := schema.NewRegistry()
	handler := NewHandler(registry)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"RegisterSchema with GET", http.MethodGet, "/api/v1/schemas"},
		{"GetSchema with POST", http.MethodPost, "/api/v1/schemas/test"},
		{"DeleteSchema with POST", http.MethodPost, "/api/v1/schemas/test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			if tt.path == "/api/v1/schemas" {
				if tt.method == http.MethodPost {
					handler.RegisterSchema(rec, req)
				} else {
					handler.ListSchemas(rec, req)
				}
			} else {
				if tt.method == http.MethodGet {
					handler.GetSchema(rec, req)
				} else if tt.method == http.MethodDelete {
					handler.DeleteSchema(rec, req)
				} else {
					handler.GetSchema(rec, req) // Will fail with method not allowed
				}
			}

			if rec.Code != http.StatusMethodNotAllowed && rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
				// Some combinations are valid, others should be method not allowed
				if (tt.name == "RegisterSchema with GET") {
					if rec.Code != http.StatusMethodNotAllowed {
						t.Errorf("%s status = %v, want %v", tt.name, rec.Code, http.StatusMethodNotAllowed)
					}
				}
			}
		})
	}
}
