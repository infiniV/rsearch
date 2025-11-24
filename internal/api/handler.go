package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/infiniv/rsearch/internal/schema"
)

// Handler handles HTTP API requests for schema management
type Handler struct {
	registry *schema.Registry
}

// NewHandler creates a new API handler with the given registry
func NewHandler(registry *schema.Registry) *Handler {
	return &Handler{
		registry: registry,
	}
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// RegisterSchema handles POST /api/v1/schemas
func (h *Handler) RegisterSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var s schema.Schema
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}

	// Register schema
	if err := h.registry.Register(&s); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return created schema
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "schema registered successfully",
		Data:    s,
	})
}

// GetSchema handles GET /api/v1/schemas/{name}
func (h *Handler) GetSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract schema name from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/schemas/")
	schemaName := strings.TrimSpace(path)

	if schemaName == "" {
		h.writeError(w, http.StatusBadRequest, "schema name is required")
		return
	}

	// Get schema
	s, err := h.registry.Get(schemaName)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	// Return schema
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s)
}

// DeleteSchema handles DELETE /api/v1/schemas/{name}
func (h *Handler) DeleteSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract schema name from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/schemas/")
	schemaName := strings.TrimSpace(path)

	if schemaName == "" {
		h.writeError(w, http.StatusBadRequest, "schema name is required")
		return
	}

	// Delete schema
	if err := h.registry.Delete(schemaName); err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// ListSchemas handles GET /api/v1/schemas
func (h *Handler) ListSchemas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// List all schemas
	schemas := h.registry.List()

	// Return schemas
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Data: schemas,
	})
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}

// SetupRoutes sets up HTTP routes for the handler
func (h *Handler) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/schemas", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.RegisterSchema(w, r)
		} else if r.Method == http.MethodGet {
			h.ListSchemas(w, r)
		} else {
			h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/api/v1/schemas/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetSchema(w, r)
		} else if r.Method == http.MethodDelete {
			h.DeleteSchema(w, r)
		} else {
			h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})
}
