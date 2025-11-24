package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/infiniv/rsearch/internal/parser"
	"github.com/infiniv/rsearch/internal/schema"
	"github.com/infiniv/rsearch/internal/translator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslateHandler_ParserNotImplemented(t *testing.T) {
	// Setup registries
	schemaRegistry := schema.NewRegistry()
	translatorRegistry := translator.NewRegistry()

	// Register test schema
	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})
	schemaRegistry.Register(testSchema)

	// Register postgres translator
	translatorRegistry.Register("postgres", translator.NewPostgresTranslator())

	// Create handler
	handler := NewTranslateHandler(schemaRegistry, translatorRegistry)

	// Create request
	reqBody := TranslateRequest{
		Schema:   "products",
		Database: "postgres",
		Query:    "productCode:13w42",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/translate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, http.StatusNotImplemented, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Parser not implemented")
}

func TestTranslateHandler_InvalidJSON(t *testing.T) {
	schemaRegistry := schema.NewRegistry()
	translatorRegistry := translator.NewRegistry()
	handler := NewTranslateHandler(schemaRegistry, translatorRegistry)

	req := httptest.NewRequest("POST", "/api/v1/translate", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTranslateHandler_SchemaNotFound(t *testing.T) {
	schemaRegistry := schema.NewRegistry()
	translatorRegistry := translator.NewRegistry()
	handler := NewTranslateHandler(schemaRegistry, translatorRegistry)

	reqBody := TranslateRequest{
		Schema:   "nonexistent",
		Database: "postgres",
		Query:    "test:value",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/translate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Schema not found")
}

func TestTranslateHandler_DatabaseNotSupported(t *testing.T) {
	schemaRegistry := schema.NewRegistry()
	translatorRegistry := translator.NewRegistry()

	// Register test schema
	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
	}, schema.SchemaOptions{})
	schemaRegistry.Register(testSchema)

	handler := NewTranslateHandler(schemaRegistry, translatorRegistry)

	reqBody := TranslateRequest{
		Schema:   "products",
		Database: "unsupported",
		Query:    "test:value",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/translate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Database type not supported")
}

func TestTranslateHandler_WithStubAST(t *testing.T) {
	schemaRegistry := schema.NewRegistry()
	translatorRegistry := translator.NewRegistry()

	// Register test schema
	testSchema := schema.NewSchema("products", map[string]schema.Field{
		"product_code": {Type: schema.TypeText},
		"region":       {Type: schema.TypeText},
	}, schema.SchemaOptions{})
	schemaRegistry.Register(testSchema)

	// Register postgres translator
	translatorRegistry.Register("postgres", translator.NewPostgresTranslator())

	// Create a test handler with stub parser
	handler := &TranslateHandler{
		schemaRegistry:     schemaRegistry,
		translatorRegistry: translatorRegistry,
		parseQuery: func(query string) (parser.Node, error) {
			// Stub parser for testing
			// Parse: productCode:13w42 AND region:ca
			return &parser.BinaryOp{
				Op: "AND",
				Left: &parser.FieldQuery{
					Field: "product_code",
					Value: &parser.TermValue{Term: "13w42", Pos: parser.Position{}},
				},
				Right: &parser.FieldQuery{
					Field: "region",
					Value: &parser.TermValue{Term: "ca", Pos: parser.Position{}},
				},
			}, nil
		},
	}

	reqBody := TranslateRequest{
		Schema:   "products",
		Database: "postgres",
		Query:    "productCode:13w42 AND region:ca",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/translate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response TranslateResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "sql", response.Type)
	assert.Equal(t, "product_code = $1 AND region = $2", response.WhereClause)
	assert.Len(t, response.Parameters, 2)
	assert.Equal(t, "13w42", response.Parameters[0])
	assert.Equal(t, "ca", response.Parameters[1])
	assert.Equal(t, []string{"text", "text"}, response.ParameterTypes)
}

func TestTranslateHandler_MissingFields(t *testing.T) {
	schemaRegistry := schema.NewRegistry()
	translatorRegistry := translator.NewRegistry()
	handler := NewTranslateHandler(schemaRegistry, translatorRegistry)

	tests := []struct {
		name    string
		request TranslateRequest
	}{
		{
			name: "missing schema",
			request: TranslateRequest{
				Database: "postgres",
				Query:    "test:value",
			},
		},
		{
			name: "missing database",
			request: TranslateRequest{
				Schema: "products",
				Query:  "test:value",
			},
		},
		{
			name: "missing query",
			request: TranslateRequest{
				Schema:   "products",
				Database: "postgres",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/translate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
