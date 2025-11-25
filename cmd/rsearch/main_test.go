package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/infiniv/rsearch/internal/api"
	"github.com/infiniv/rsearch/internal/config"
	"github.com/infiniv/rsearch/internal/observability"
	"github.com/infiniv/rsearch/internal/ratelimit"
	"github.com/infiniv/rsearch/internal/schema"
	"github.com/infiniv/rsearch/internal/translator"
	"github.com/infiniv/rsearch/pkg/rsearch"
)

func TestServerIntegration(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:            "localhost",
			Port:            18080, // Use different port for testing
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		},
		Logging: config.LoggingConfig{
			Level:  "error",
			Format: "json",
			Output: "stdout",
		},
		Metrics: config.MetricsConfig{
			Enabled: false,
		},
		CORS: config.CORSConfig{
			Enabled: false,
		},
		Features: config.FeaturesConfig{
			RequestIDHeader: "X-Request-ID",
		},
	}

	// Create logger
	logger, err := observability.NewLogger(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.Output)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Setup routes with registries
	schemaRegistry := schema.NewRegistry()
	translatorRegistry := translator.NewRegistry()
	translatorRegistry.Register("postgres", translator.NewPostgresTranslator())
	rateLimiter := ratelimit.NewRateLimiter(100, 10)
	defer rateLimiter.Stop()

	router := api.SetupRoutes(cfg, logger, nil, schemaRegistry, translatorRegistry, rateLimiter)

	// Create server
	server := &http.Server{
		Addr:         cfg.GetAddress(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Ensure server is shut down after test
	defer func() {
		if err := server.Close(); err != nil {
			t.Logf("Error closing server: %v", err)
		}
	}()

	// Test health endpoint
	t.Run("Health endpoint", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://%s/health", cfg.GetAddress()))
		if err != nil {
			t.Fatalf("Failed to request health endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var health rsearch.HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if health.Status != "healthy" {
			t.Errorf("Expected status 'healthy', got '%s'", health.Status)
		}

		if health.Version != rsearch.Version {
			t.Errorf("Expected version '%s', got '%s'", rsearch.Version, health.Version)
		}
	})

	// Test ready endpoint
	t.Run("Ready endpoint", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://%s/ready", cfg.GetAddress()))
		if err != nil {
			t.Fatalf("Failed to request ready endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var ready map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&ready); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if ready["ready"] != true {
			t.Errorf("Expected ready to be true, got %v", ready["ready"])
		}
	})

	// Test request ID header
	t.Run("Request ID header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/health", cfg.GetAddress()), nil)
		req.Header.Set("X-Request-ID", "test-request-123")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to request health endpoint: %v", err)
		}
		defer resp.Body.Close()

		requestID := resp.Header.Get("X-Request-ID")
		if requestID != "test-request-123" {
			t.Errorf("Expected request ID 'test-request-123', got '%s'", requestID)
		}
	})

	// Test auto-generated request ID
	t.Run("Auto-generated request ID", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://%s/health", cfg.GetAddress()))
		if err != nil {
			t.Fatalf("Failed to request health endpoint: %v", err)
		}
		defer resp.Body.Close()

		requestID := resp.Header.Get("X-Request-ID")
		if requestID == "" {
			t.Error("Expected auto-generated request ID, got empty string")
		}
	})
}
