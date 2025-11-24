package internal

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/infiniv/rsearch/internal/api"
	"github.com/infiniv/rsearch/internal/schema"
)

// TestIntegration_CompleteWorkflow tests the complete schema registration and field resolution workflow
func TestIntegration_CompleteWorkflow(t *testing.T) {
	// Setup: Create registry and API handler
	registry := schema.NewRegistry()
	handler := api.NewHandler(registry)

	// Step 1: Register a schema via API
	schemaJSON := `{
		"name": "products",
		"fields": {
			"productCode": {
				"type": "text",
				"indexed": true,
				"aliases": ["code", "sku"]
			},
			"productName": {
				"type": "text",
				"aliases": ["name"]
			},
			"price": {
				"type": "float"
			},
			"inStock": {
				"type": "boolean"
			},
			"createdAt": {
				"type": "datetime",
				"column": "created_timestamp"
			}
		},
		"options": {
			"namingConvention": "snake_case",
			"strictFieldNames": false,
			"strictOperators": true,
			"defaultField": "productName"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/schemas", bytes.NewBufferString(schemaJSON))
	rec := httptest.NewRecorder()
	handler.RegisterSchema(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Failed to register schema: status = %v, body = %s", rec.Code, rec.Body.String())
	}
	t.Log("Step 1: Schema registered successfully")

	// Step 2: Retrieve the schema via API
	req = httptest.NewRequest(http.MethodGet, "/api/v1/schemas/products", nil)
	rec = httptest.NewRecorder()
	handler.GetSchema(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Failed to get schema: status = %v", rec.Code)
	}

	var retrievedSchema schema.Schema
	if err := json.NewDecoder(rec.Body).Decode(&retrievedSchema); err != nil {
		t.Fatalf("Failed to decode schema: %v", err)
	}
	t.Log("Step 2: Schema retrieved successfully")

	// Step 3: Test field resolution scenarios
	testCases := []struct {
		name           string
		queryField     string
		expectedColumn string
		shouldSucceed  bool
	}{
		{
			name:           "exact match with naming convention",
			queryField:     "productCode",
			expectedColumn: "product_code",
			shouldSucceed:  true,
		},
		{
			name:           "case insensitive match",
			queryField:     "PRODUCTCODE",
			expectedColumn: "product_code",
			shouldSucceed:  true,
		},
		{
			name:           "alias resolution",
			queryField:     "code",
			expectedColumn: "product_code",
			shouldSucceed:  true,
		},
		{
			name:           "alias case insensitive",
			queryField:     "SKU",
			expectedColumn: "product_code",
			shouldSucceed:  true,
		},
		{
			name:           "explicit column override",
			queryField:     "createdAt",
			expectedColumn: "created_timestamp",
			shouldSucceed:  true,
		},
		{
			name:          "non-existent field",
			queryField:    "nonExistent",
			shouldSucceed: false,
		},
	}

	s, err := registry.Get("products")
	if err != nil {
		t.Fatalf("Failed to get schema from registry: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			column, field, err := s.ResolveField(tc.queryField)

			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("Expected success, got error: %v", err)
					return
				}
				if column != tc.expectedColumn {
					t.Errorf("Expected column %q, got %q", tc.expectedColumn, column)
				}
				if field == nil {
					t.Error("Expected field, got nil")
				}
			} else {
				if err == nil {
					t.Error("Expected error, got success")
				}
			}
		})
	}
	t.Log("Step 3: Field resolution tests passed")

	// Step 4: List all schemas
	req = httptest.NewRequest(http.MethodGet, "/api/v1/schemas", nil)
	rec = httptest.NewRecorder()
	handler.ListSchemas(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Failed to list schemas: status = %v", rec.Code)
	}

	var listResponse api.SuccessResponse
	if err := json.NewDecoder(rec.Body).Decode(&listResponse); err != nil {
		t.Fatalf("Failed to decode list response: %v", err)
	}
	t.Log("Step 4: Schema list retrieved successfully")

	// Step 5: Delete the schema
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/schemas/products", nil)
	rec = httptest.NewRecorder()
	handler.DeleteSchema(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("Failed to delete schema: status = %v", rec.Code)
	}
	t.Log("Step 5: Schema deleted successfully")

	// Step 6: Verify schema is gone
	req = httptest.NewRequest(http.MethodGet, "/api/v1/schemas/products", nil)
	rec = httptest.NewRecorder()
	handler.GetSchema(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Expected schema to be deleted, got status = %v", rec.Code)
	}
	t.Log("Step 6: Verified schema deletion")
}

// TestIntegration_MultipleSchemas tests working with multiple schemas
func TestIntegration_MultipleSchemas(t *testing.T) {
	registry := schema.NewRegistry()
	handler := api.NewHandler(registry)

	// Register multiple schemas
	schemas := []string{
		`{
			"name": "users",
			"fields": {
				"userId": {"type": "integer"},
				"userName": {"type": "text"}
			},
			"options": {"namingConvention": "snake_case"}
		}`,
		`{
			"name": "orders",
			"fields": {
				"orderId": {"type": "integer"},
				"orderDate": {"type": "date"}
			},
			"options": {"namingConvention": "snake_case"}
		}`,
		`{
			"name": "inventory",
			"fields": {
				"itemCode": {"type": "text"},
				"quantity": {"type": "integer"}
			},
			"options": {"namingConvention": "camelCase"}
		}`,
	}

	// Register all schemas
	for i, schemaJSON := range schemas {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/schemas", bytes.NewBufferString(schemaJSON))
		rec := httptest.NewRecorder()
		handler.RegisterSchema(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Failed to register schema %d: status = %v, body = %s", i+1, rec.Code, rec.Body.String())
		}
	}
	t.Logf("Registered %d schemas", len(schemas))

	// Verify all schemas are present
	req := httptest.NewRequest(http.MethodGet, "/api/v1/schemas", nil)
	rec := httptest.NewRecorder()
	handler.ListSchemas(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Failed to list schemas: status = %v", rec.Code)
	}

	var listResponse api.SuccessResponse
	if err := json.NewDecoder(rec.Body).Decode(&listResponse); err != nil {
		t.Fatalf("Failed to decode list response: %v", err)
	}

	// List response should contain all registered schemas
	t.Logf("Successfully listed all schemas")

	// Test field resolution for different schemas
	testCases := []struct {
		schemaName string
		queryField string
		wantColumn string
	}{
		{"users", "userName", "user_name"},
		{"orders", "orderDate", "order_date"},
		{"inventory", "item_code", "itemCode"}, // camelCase convention
	}

	for _, tc := range testCases {
		t.Run(tc.schemaName+"/"+tc.queryField, func(t *testing.T) {
			s, err := registry.Get(tc.schemaName)
			if err != nil {
				t.Fatalf("Failed to get schema %q: %v", tc.schemaName, err)
			}

			column, _, err := s.ResolveField(tc.queryField)
			if err != nil {
				t.Errorf("Failed to resolve field %q: %v", tc.queryField, err)
				return
			}

			if column != tc.wantColumn {
				t.Errorf("Expected column %q, got %q", tc.wantColumn, column)
			}
		})
	}
}

// TestIntegration_ConcurrentSchemaOperations tests concurrent access to schemas
func TestIntegration_ConcurrentSchemaOperations(t *testing.T) {
	registry := schema.NewRegistry()

	// Pre-register a schema
	s := &schema.Schema{
		Name: "concurrent_test",
		Fields: map[string]schema.Field{
			"field1": {Type: schema.TypeText},
			"field2": {Type: schema.TypeInteger},
		},
		Options: schema.SchemaOptions{
			NamingConvention: "snake_case",
		},
	}

	if err := registry.Register(s); err != nil {
		t.Fatalf("Failed to register schema: %v", err)
	}

	// Perform concurrent field resolutions
	const numGoroutines = 100
	const numOperations = 50

	errCh := make(chan error, numGoroutines*numOperations)
	doneCh := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { doneCh <- true }()

			s, err := registry.Get("concurrent_test")
			if err != nil {
				errCh <- err
				return
			}

			for j := 0; j < numOperations; j++ {
				_, _, err := s.ResolveField("field1")
				if err != nil {
					errCh <- err
					return
				}
			}
		}()
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-doneCh
	}
	close(errCh)

	// Check for errors
	for err := range errCh {
		t.Errorf("Concurrent operation error: %v", err)
	}
}
