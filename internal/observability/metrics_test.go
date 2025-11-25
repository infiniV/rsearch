package observability

import (
	"net/http"
	"testing"
	"time"

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

func TestRecordQueryComplexity(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic
	m.RecordQueryComplexity(1)
	m.RecordQueryComplexity(5)
	m.RecordQueryComplexity(10)
	m.RecordQueryComplexity(50)
	m.RecordQueryComplexity(100)
}

func TestRecordParameterCount(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic
	m.RecordParameterCount(0)
	m.RecordParameterCount(1)
	m.RecordParameterCount(5)
	m.RecordParameterCount(20)
}

func TestRecordRateLimitHit(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic
	m.RecordRateLimitHit()
	m.RecordRateLimitHit()
}

func TestRecordValidationError(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic with different error types
	m.RecordValidationError("invalid_field")
	m.RecordValidationError("missing_required")
	m.RecordValidationError("type_mismatch")
	m.RecordValidationError("schema_violation")
}

func TestRecordSchemaOperation(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic with different operations
	m.RecordSchemaOperation("create")
	m.RecordSchemaOperation("read")
	m.RecordSchemaOperation("delete")
	m.RecordSchemaOperation("update")
}

func TestRecordQuerySyntax(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic with different syntax features
	m.RecordQuerySyntax("range")
	m.RecordQuerySyntax("wildcard")
	m.RecordQuerySyntax("fuzzy")
	m.RecordQuerySyntax("proximity")
	m.RecordQuerySyntax("regex")
	m.RecordQuerySyntax("boolean")
	m.RecordQuerySyntax("phrase")
	m.RecordQuerySyntax("exists")
	m.RecordQuerySyntax("boost")
}

func TestRecordResponseSize(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic
	m.RecordResponseSize(100)
	m.RecordResponseSize(1024)
	m.RecordResponseSize(10240)
	m.RecordResponseSize(102400)
}

func TestRecordDatabaseTarget(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic with different database types
	m.RecordDatabaseTarget("postgres")
	m.RecordDatabaseTarget("mysql")
	m.RecordDatabaseTarget("sqlite")
}

func TestUpdateSystemMetrics(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Should not panic
	m.UpdateSystemMetrics()
	m.UpdateSystemMetrics()
}

func TestNewMetricsIncludesAllMetrics(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	// Verify new metrics are initialized
	if m.QueryComplexity == nil {
		t.Error("QueryComplexity not initialized")
	}
	if m.ParameterCount == nil {
		t.Error("ParameterCount not initialized")
	}
	if m.RateLimitHits == nil {
		t.Error("RateLimitHits not initialized")
	}
	if m.ValidationErrors == nil {
		t.Error("ValidationErrors not initialized")
	}
	if m.SchemaOperations == nil {
		t.Error("SchemaOperations not initialized")
	}
	if m.QuerySyntaxUsage == nil {
		t.Error("QuerySyntaxUsage not initialized")
	}
	if m.ResponseSize == nil {
		t.Error("ResponseSize not initialized")
	}
	if m.DatabaseTargets == nil {
		t.Error("DatabaseTargets not initialized")
	}
	if m.GoroutineCount == nil {
		t.Error("GoroutineCount not initialized")
	}
	if m.MemoryUsage == nil {
		t.Error("MemoryUsage not initialized")
	}
	if m.Uptime == nil {
		t.Error("Uptime not initialized")
	}
}

func TestNewCollector(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()

	c := NewCollector(m)
	if c == nil {
		t.Fatal("expected collector but got nil")
	}
	if c.metrics != m {
		t.Error("collector metrics not set correctly")
	}
}

func TestCollectorStartStop(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()
	c := NewCollector(m)

	// Start collector
	c.Start()

	// Give it time to run at least once
	time.Sleep(100 * time.Millisecond)

	// Stop collector
	c.Stop()

	// Should not panic
}

func TestCollectorUpdatesMetrics(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()
	c := NewCollector(m)

	// Start collector
	c.Start()

	// Wait for at least one update
	time.Sleep(100 * time.Millisecond)

	// Stop collector
	c.Stop()

	// System metrics should have been updated
	// We can't check exact values, but we verify no panic occurred
}

func TestCollectorStopWithoutStart(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()
	c := NewCollector(m)

	// Should not panic when stopping without starting
	c.Stop()
}

func TestCollectorMultipleStops(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m := NewMetrics()
	c := NewCollector(m)

	c.Start()
	c.Stop()

	// Should not panic when stopping multiple times
	c.Stop()
	c.Stop()
}
