package api

import (
	"net/http"

	"github.com/infiniv/rsearch/internal/config"
	"github.com/infiniv/rsearch/internal/observability"
	"github.com/infiniv/rsearch/pkg/rsearch"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	config  *config.Config
	logger  *observability.Logger
	metrics *observability.Metrics
}

// NewHandlers creates a new handlers instance
func NewHandlers(cfg *config.Config, logger *observability.Logger, metrics *observability.Metrics) *Handlers {
	return &Handlers{
		config:  cfg,
		logger:  logger,
		metrics: metrics,
	}
}

// Health handles the health check endpoint
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	response := rsearch.HealthResponse{
		Status:  "healthy",
		Version: rsearch.Version,
	}
	RespondJSON(w, http.StatusOK, response)
}

// Ready handles the readiness check endpoint
func (h *Handlers) Ready(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"ready":   true,
		"version": rsearch.Version,
	}
	RespondJSON(w, http.StatusOK, response)
}

// Metrics handles the metrics endpoint (wrapped by Prometheus handler in routes)
func (h *Handlers) Metrics() http.Handler {
	if h.metrics == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			RespondError(w, http.StatusNotFound, "METRICS_DISABLED", "Metrics are not enabled")
		})
	}
	return h.metrics.Handler()
}
