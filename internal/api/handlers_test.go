package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/infiniv/rsearch/internal/config"
	"github.com/infiniv/rsearch/internal/observability"
	"github.com/infiniv/rsearch/pkg/rsearch"
)

func setupTestHandlers(t *testing.T, withMetrics bool) *Handlers {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:            "localhost",
			Port:            8080,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Metrics: config.MetricsConfig{
			Enabled: true,
			Port:    9090,
			Path:    "/metrics",
		},
		Features: config.FeaturesConfig{
			RequestIDHeader: "X-Request-ID",
		},
	}

	logger, err := observability.NewLogger("error", "json", "stdout")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	var metrics *observability.Metrics
	if withMetrics {
		metrics = observability.NewMetrics()
	}

	return NewHandlers(cfg, logger, metrics)
}

func TestHealthHandler(t *testing.T) {
	handlers := setupTestHandlers(t, false)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handlers.Health(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
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
}

func TestReadyHandler(t *testing.T) {
	handlers := setupTestHandlers(t, false)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	handlers.Ready(w, req)

	resp := w.Result()
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

	if ready["version"] != rsearch.Version {
		t.Errorf("Expected version '%s', got '%v'", rsearch.Version, ready["version"])
	}
}

func TestMetricsHandler(t *testing.T) {
	handlers := setupTestHandlers(t, true)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	metricsHandler := handlers.Metrics()
	metricsHandler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check that we get Prometheus metrics format
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/plain; version=0.0.4; charset=utf-8" {
		t.Logf("Warning: Expected Prometheus content type, got %s", contentType)
	}
}

func TestMetricsHandlerWhenDisabled(t *testing.T) {
	cfg := &config.Config{
		Features: config.FeaturesConfig{
			RequestIDHeader: "X-Request-ID",
		},
	}

	logger, _ := observability.NewLogger("error", "json", "stdout")
	handlers := NewHandlers(cfg, logger, nil)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	metricsHandler := handlers.Metrics()
	metricsHandler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 when metrics disabled, got %d", resp.StatusCode)
	}

	var errResp rsearch.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errResp.Error.Code != "METRICS_DISABLED" {
		t.Errorf("Expected error code 'METRICS_DISABLED', got '%s'", errResp.Error.Code)
	}
}
