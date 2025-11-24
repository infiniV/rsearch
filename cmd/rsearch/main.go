package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/raw/rsearch/internal/api"
	"github.com/raw/rsearch/internal/schema"
)

func main() {
	// Create schema registry
	registry := schema.NewRegistry()

	// Create API handler
	handler := api.NewHandler(registry)

	// Setup routes
	mux := http.NewServeMux()
	handler.SetupRoutes(mux)

	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
	})

	// Add root endpoint with API documentation
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "rsearch Schema API",
			"version": "0.1.0",
			"endpoints": map[string]string{
				"POST /api/v1/schemas":           "Register a new schema",
				"GET /api/v1/schemas":            "List all schemas",
				"GET /api/v1/schemas/{name}":     "Get a specific schema",
				"DELETE /api/v1/schemas/{name}":  "Delete a schema",
				"GET /health":                    "Health check",
			},
		})
	})

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting rsearch Schema API server on %s", addr)
	log.Printf("Health check: http://localhost%s/health", addr)
	log.Printf("API documentation: http://localhost%s/", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
