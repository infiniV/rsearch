package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	ActiveRequests   prometheus.Gauge
	ErrorsTotal      *prometheus.CounterVec
	ParseDuration    *prometheus.HistogramVec
	TranslateDuration *prometheus.HistogramVec
	ActiveSchemas    prometheus.Gauge
	CacheHits        prometheus.Counter
	CacheMisses      prometheus.Counter
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
