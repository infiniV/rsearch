package observability

import (
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	RequestsTotal     *prometheus.CounterVec
	RequestDuration   *prometheus.HistogramVec
	ActiveRequests    prometheus.Gauge
	ErrorsTotal       *prometheus.CounterVec
	ParseDuration     *prometheus.HistogramVec
	TranslateDuration *prometheus.HistogramVec
	ActiveSchemas     prometheus.Gauge
	CacheHits         prometheus.Counter
	CacheMisses       prometheus.Counter

	// New metrics
	QueryComplexity   *prometheus.HistogramVec
	ParameterCount    *prometheus.HistogramVec
	RateLimitHits     prometheus.Counter
	ValidationErrors  *prometheus.CounterVec
	SchemaOperations  *prometheus.CounterVec
	QuerySyntaxUsage  *prometheus.CounterVec
	ResponseSize      *prometheus.HistogramVec
	DatabaseTargets   *prometheus.CounterVec

	// System metrics
	GoroutineCount    prometheus.Gauge
	MemoryUsage       prometheus.Gauge
	Uptime            prometheus.Gauge

	startTime         int64
}

// NewMetrics creates and registers Prometheus metrics
func NewMetrics() *Metrics {
	m := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rsearch_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"endpoint", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "rsearch_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"endpoint"},
		),
		ActiveRequests: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "rsearch_active_requests",
				Help: "Number of active HTTP requests",
			},
		),
		ErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rsearch_errors_total",
				Help: "Total number of errors",
			},
			[]string{"type"},
		),
		ParseDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "rsearch_parse_duration_seconds",
				Help:    "Query parsing duration in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"status"},
		),
		TranslateDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "rsearch_translate_duration_seconds",
				Help:    "Query translation duration in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"database"},
		),
		ActiveSchemas: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "rsearch_active_schemas",
				Help: "Number of registered schemas",
			},
		),
		CacheHits: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "rsearch_cache_hits_total",
				Help: "Total number of cache hits",
			},
		),
		CacheMisses: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "rsearch_cache_misses_total",
				Help: "Total number of cache misses",
			},
		),
		QueryComplexity: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "rsearch_query_complexity",
				Help:    "Histogram of query complexity measured by AST node count",
				Buckets: []float64{1, 5, 10, 20, 50, 100, 200, 500},
			},
			[]string{"schema"},
		),
		ParameterCount: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "rsearch_parameter_count",
				Help:    "Histogram of parameter count per translated query",
				Buckets: []float64{0, 1, 2, 5, 10, 20, 50, 100},
			},
			[]string{"schema"},
		),
		RateLimitHits: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "rsearch_rate_limit_hits_total",
				Help: "Total number of rate limit rejections",
			},
		),
		ValidationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rsearch_validation_errors_total",
				Help: "Total number of validation errors by type",
			},
			[]string{"error_type"},
		),
		SchemaOperations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rsearch_schema_operations_total",
				Help: "Total number of schema operations by type",
			},
			[]string{"operation"},
		),
		QuerySyntaxUsage: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rsearch_query_syntax_usage_total",
				Help: "Total usage count of query syntax features",
			},
			[]string{"feature"},
		),
		ResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "rsearch_response_size_bytes",
				Help:    "Histogram of response body sizes in bytes",
				Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000},
			},
			[]string{"endpoint"},
		),
		DatabaseTargets: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rsearch_database_targets_total",
				Help: "Total number of queries by target database type",
			},
			[]string{"database"},
		),
		GoroutineCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "rsearch_goroutines",
				Help: "Current number of goroutines",
			},
		),
		MemoryUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "rsearch_memory_usage_bytes",
				Help: "Current heap memory usage in bytes",
			},
		),
		Uptime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "rsearch_uptime_seconds",
				Help: "Server uptime in seconds since start",
			},
		),
		startTime: time.Now().Unix(),
	}

	// Register all metrics
	prometheus.MustRegister(m.RequestsTotal)
	prometheus.MustRegister(m.RequestDuration)
	prometheus.MustRegister(m.ActiveRequests)
	prometheus.MustRegister(m.ErrorsTotal)
	prometheus.MustRegister(m.ParseDuration)
	prometheus.MustRegister(m.TranslateDuration)
	prometheus.MustRegister(m.ActiveSchemas)
	prometheus.MustRegister(m.CacheHits)
	prometheus.MustRegister(m.CacheMisses)
	prometheus.MustRegister(m.QueryComplexity)
	prometheus.MustRegister(m.ParameterCount)
	prometheus.MustRegister(m.RateLimitHits)
	prometheus.MustRegister(m.ValidationErrors)
	prometheus.MustRegister(m.SchemaOperations)
	prometheus.MustRegister(m.QuerySyntaxUsage)
	prometheus.MustRegister(m.ResponseSize)
	prometheus.MustRegister(m.DatabaseTargets)
	prometheus.MustRegister(m.GoroutineCount)
	prometheus.MustRegister(m.MemoryUsage)
	prometheus.MustRegister(m.Uptime)

	return m
}

// Handler returns the Prometheus HTTP handler
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

// RecordRequest records an HTTP request
func (m *Metrics) RecordRequest(endpoint string, status int, duration float64) {
	m.RequestsTotal.WithLabelValues(endpoint, http.StatusText(status)).Inc()
	m.RequestDuration.WithLabelValues(endpoint).Observe(duration)
}

// IncActiveRequests increments active requests
func (m *Metrics) IncActiveRequests() {
	m.ActiveRequests.Inc()
}

// DecActiveRequests decrements active requests
func (m *Metrics) DecActiveRequests() {
	m.ActiveRequests.Dec()
}

// RecordError records an error
func (m *Metrics) RecordError(errorType string) {
	m.ErrorsTotal.WithLabelValues(errorType).Inc()
}

// RecordParseDuration records parsing duration
func (m *Metrics) RecordParseDuration(duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	m.ParseDuration.WithLabelValues(status).Observe(duration)
}

// RecordTranslateDuration records translation duration
func (m *Metrics) RecordTranslateDuration(database string, duration float64) {
	m.TranslateDuration.WithLabelValues(database).Observe(duration)
}

// SetActiveSchemas sets the number of active schemas
func (m *Metrics) SetActiveSchemas(count int) {
	m.ActiveSchemas.Set(float64(count))
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit() {
	m.CacheHits.Inc()
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss() {
	m.CacheMisses.Inc()
}

// RecordQueryComplexity records query complexity by AST node count
func (m *Metrics) RecordQueryComplexity(nodeCount int) {
	m.QueryComplexity.WithLabelValues("").Observe(float64(nodeCount))
}

// RecordParameterCount records the number of parameters in a translated query
func (m *Metrics) RecordParameterCount(count int) {
	m.ParameterCount.WithLabelValues("").Observe(float64(count))
}

// RecordRateLimitHit records a rate limit rejection
func (m *Metrics) RecordRateLimitHit() {
	m.RateLimitHits.Inc()
}

// RecordValidationError records a validation error by type
func (m *Metrics) RecordValidationError(errorType string) {
	m.ValidationErrors.WithLabelValues(errorType).Inc()
}

// RecordSchemaOperation records a schema operation
func (m *Metrics) RecordSchemaOperation(operation string) {
	m.SchemaOperations.WithLabelValues(operation).Inc()
}

// RecordQuerySyntax records usage of a query syntax feature
func (m *Metrics) RecordQuerySyntax(feature string) {
	m.QuerySyntaxUsage.WithLabelValues(feature).Inc()
}

// RecordResponseSize records the size of a response body
func (m *Metrics) RecordResponseSize(bytes int) {
	m.ResponseSize.WithLabelValues("").Observe(float64(bytes))
}

// RecordDatabaseTarget records a query by target database type
func (m *Metrics) RecordDatabaseTarget(dbType string) {
	m.DatabaseTargets.WithLabelValues(dbType).Inc()
}

// UpdateSystemMetrics updates system-level metrics
func (m *Metrics) UpdateSystemMetrics() {
	// Update goroutine count
	m.GoroutineCount.Set(float64(runtime.NumGoroutine()))

	// Update memory usage
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.MemoryUsage.Set(float64(memStats.HeapAlloc))

	// Update uptime
	uptime := time.Now().Unix() - m.startTime
	m.Uptime.Set(float64(uptime))
}
