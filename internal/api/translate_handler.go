package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/raw/rsearch/internal/schema"
	"github.com/raw/rsearch/internal/translator"
)

// TranslateRequest represents the request body for the translate endpoint.
type TranslateRequest struct {
	Schema   string `json:"schema"`
	Database string `json:"database"`
	Query    string `json:"query"`
}

// TranslateResponse represents the response body for the translate endpoint.
type TranslateResponse struct {
	Type           string        `json:"type"`
	WhereClause    string        `json:"whereClause,omitempty"`
	Parameters     []interface{} `json:"parameters,omitempty"`
	ParameterTypes []string      `json:"parameterTypes,omitempty"`
	Filter         interface{}   `json:"filter,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// TranslateHandler handles translation requests.
type TranslateHandler struct {
	schemaRegistry     *schema.Registry
	translatorRegistry *translator.Registry
	parseQuery         func(string) (translator.Node, error)
}

// NewTranslateHandler creates a new translate handler.
func NewTranslateHandler(schemaRegistry *schema.Registry, translatorRegistry *translator.Registry) *TranslateHandler {
	return &TranslateHandler{
		schemaRegistry:     schemaRegistry,
		translatorRegistry: translatorRegistry,
		parseQuery:         nil, // Parser not implemented yet
	}
}

// ServeHTTP handles HTTP requests.
func (h *TranslateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse request body
	var req TranslateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Schema == "" {
		h.sendError(w, http.StatusBadRequest, "Schema is required")
		return
	}
	if req.Database == "" {
		h.sendError(w, http.StatusBadRequest, "Database is required")
		return
	}
	if req.Query == "" {
		h.sendError(w, http.StatusBadRequest, "Query is required")
		return
	}

	// Lookup schema
	sch, err := h.schemaRegistry.Get(req.Schema)
	if err != nil {
		h.sendError(w, http.StatusNotFound, fmt.Sprintf("Schema not found: %s", req.Schema))
		return
	}

	// Get translator
	trans, err := h.translatorRegistry.Get(req.Database)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("Database type not supported: %s", req.Database))
		return
	}

	// Parse query (not implemented yet)
	if h.parseQuery == nil {
		h.sendError(w, http.StatusNotImplemented, "Parser not implemented yet")
		return
	}

	ast, err := h.parseQuery(req.Query)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse query: %s", err.Error()))
		return
	}

	// Translate AST
	output, err := trans.Translate(ast, sch)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("Translation failed: %s", err.Error()))
		return
	}

	// Build response
	response := TranslateResponse{
		Type:           output.Type,
		WhereClause:    output.WhereClause,
		Parameters:     output.Parameters,
		ParameterTypes: output.ParameterTypes,
		Filter:         output.Filter,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// sendError sends an error response.
func (h *TranslateHandler) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
