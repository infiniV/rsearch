package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/infiniv/rsearch/internal/config"
	"github.com/infiniv/rsearch/internal/observability"
)

// SetupRoutes sets up all HTTP routes
func SetupRoutes(cfg *config.Config, logger *observability.Logger, metrics *observability.Metrics) *chi.Mux {
	r := chi.NewRouter()

	// Create handlers
	handlers := NewHandlers(cfg, logger, metrics)

	// Global middleware
	r.Use(RequestIDMiddleware(cfg))
	r.Use(LoggingMiddleware(logger))
	r.Use(RecoveryMiddleware(logger))
	r.Use(CORSMiddleware(cfg))

	// Add metrics middleware if enabled
	if metrics != nil {
		r.Use(MetricsMiddleware(metrics))
	}

	// Health and readiness endpoints (no /api prefix)
	r.Get("/health", handlers.Health)
	r.Get("/ready", handlers.Ready)

	// Metrics endpoint (only if enabled)
	if cfg.Metrics.Enabled && metrics != nil {
		r.Handle(cfg.Metrics.Path, handlers.Metrics())
	}

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Schema endpoints (to be implemented in Phase 3)
		r.Post("/schemas", notImplementedHandler)
		r.Get("/schemas/{name}", notImplementedHandler)
		r.Delete("/schemas/{name}", notImplementedHandler)

		// Translation endpoint (to be implemented in Phase 4)
		r.Post("/translate", notImplementedHandler)
	})

	return r
}

// notImplementedHandler is a placeholder for endpoints to be implemented
func notImplementedHandler(w http.ResponseWriter, r *http.Request) {
	RespondError(w, 501, "NOT_IMPLEMENTED", "This endpoint is not yet implemented")
}
