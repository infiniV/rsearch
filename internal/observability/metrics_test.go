package observability

import (
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewMetrics(t *testing.T) {
	// Unregister any existing metrics first to avoid conflicts in tests
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	m := NewMetrics()
	if m == nil {
		t.Fatal("expected metrics but got nil")
	}

	// Verify all metrics are initialized
	if m.RequestsTotal == nil {
		t.Error("RequestsTotal not initialized")
	}
	if m.RequestDuration == nil {
		t.Error("RequestDuration not initialized")
	}
	if m.ActiveRequests == nil {
		t.Error("ActiveRequests not initialized")
	}
	if m.ErrorsTotal == nil {
		t.Error("ErrorsTotal not initialized")
	}
	if m.ParseDuration == nil {
		t.Error("ParseDuration not initialized")
	}
	if m.TranslateDuration == nil {
		t.Error("TranslateDuration not initialized")
	}
	if m.ActiveSchemas == nil {
		t.Error("ActiveSchemas not initialized")
	}
	if m.CacheHits == nil {
		t.Error("CacheHits not initialized")
	}
	if m.CacheMisses == nil {
		t.Error("CacheMisses not initialized")
	}
}

func TestMetricsHandler(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	handler := m.Handler()
	if handler == nil {
		t.Error("expected handler but got nil")
	}

	// Verify handler is an http.Handler
	var _ http.Handler = handler
}

func TestRecordRequest(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic
	m.RecordRequest("/api/v1/translate", http.StatusOK, 0.123)
	m.RecordRequest("/api/v1/schemas", http.StatusCreated, 0.050)
	m.RecordRequest("/api/v1/translate", http.StatusBadRequest, 0.010)
}

func TestActiveRequests(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic
	m.IncActiveRequests()
	m.IncActiveRequests()
	m.DecActiveRequests()
}

func TestRecordError(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic
	m.RecordError("parse_error")
	m.RecordError("schema_not_found")
	m.RecordError("internal_error")
}

func TestRecordParseDuration(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic
	m.RecordParseDuration(0.005, true)
	m.RecordParseDuration(0.001, false)
}

func TestRecordTranslateDuration(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic
	m.RecordTranslateDuration("postgres", 0.010)
	m.RecordTranslateDuration("mysql", 0.015)
}

func TestSetActiveSchemas(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic
	m.SetActiveSchemas(0)
	m.SetActiveSchemas(5)
	m.SetActiveSchemas(100)
}

func TestRecordCacheHitMiss(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic
	m.RecordCacheHit()
	m.RecordCacheHit()
	m.RecordCacheMiss()
}
