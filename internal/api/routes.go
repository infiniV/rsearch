package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/infiniv/rsearch/internal/config"
	"github.com/infiniv/rsearch/internal/observability"
	"github.com/infiniv/rsearch/internal/schema"
	"github.com/infiniv/rsearch/internal/translator"
)

// SetupRoutes sets up all HTTP routes
func SetupRoutes(cfg *config.Config, logger *observability.Logger, metrics *observability.Metrics, schemaRegistry *schema.Registry, translatorRegistry *translator.Registry) *chi.Mux {
	r := chi.NewRouter()

	// Create handlers
	handlers := NewHandlers(cfg, logger, metrics)
	schemaHandler := NewSchemaHandler(schemaRegistry, logger)
	translateHandler := NewTranslateHandler(schemaRegistry, translatorRegistry, logger)

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
		// Schema endpoints
		r.Post("/schemas", schemaHandler.RegisterSchema)
		r.Get("/schemas", schemaHandler.ListSchemas)
		r.Get("/schemas/{name}", schemaHandler.GetSchema)
		r.Delete("/schemas/{name}", schemaHandler.DeleteSchema)

		// Translation endpoint
		r.Post("/translate", translateHandler.Translate)
	})

	return r
}
