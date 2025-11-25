package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/infiniv/rsearch/internal/api"
	"github.com/infiniv/rsearch/internal/config"
	"github.com/infiniv/rsearch/internal/observability"
	"github.com/infiniv/rsearch/internal/ratelimit"
	"github.com/infiniv/rsearch/internal/schema"
	"github.com/infiniv/rsearch/internal/translator"
	"github.com/infiniv/rsearch/pkg/rsearch"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := observability.NewLogger(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.Output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Infof("Starting rsearch v%s", rsearch.Version)
	logger.Infof("Server will listen on %s", cfg.GetAddress())

	// Initialize metrics if enabled
	var metrics *observability.Metrics
	if cfg.Metrics.Enabled {
		metrics = observability.NewMetrics()
		logger.Infof("Metrics enabled on %s%s", cfg.GetMetricsAddress(), cfg.Metrics.Path)
	}

	// Initialize schema registry
	schemaRegistry := schema.NewRegistry()
	logger.Info("Schema registry initialized")

	// Initialize translator registry
	translatorRegistry := translator.NewRegistry()
	translatorRegistry.Register("postgres", translator.NewPostgresTranslator())
	logger.Info("Translator registry initialized with PostgreSQL support")

	// Initialize rate limiter
	rateLimiter := ratelimit.NewRateLimiter(cfg.Limits.RateLimit.RequestsPerMinute, cfg.Limits.RateLimit.Burst)
	defer rateLimiter.Stop()
	if cfg.Limits.RateLimit.Enabled {
		logger.Infof("Rate limiting enabled: %d requests/min with burst of %d",
			cfg.Limits.RateLimit.RequestsPerMinute, cfg.Limits.RateLimit.Burst)
	}

	// Setup routes
	router := api.SetupRoutes(cfg, logger, metrics, schemaRegistry, translatorRegistry, rateLimiter)

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetAddress(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Infof("Server listening on %s", cfg.GetAddress())
		serverErrors <- server.ListenAndServe()
	}()

	// Start metrics server if enabled
	var metricsServer *http.Server
	if cfg.Metrics.Enabled && metrics != nil {
		metricsServer = &http.Server{
			Addr:    cfg.GetMetricsAddress(),
			Handler: metrics.Handler(),
		}
		go func() {
			logger.Infof("Metrics server listening on %s", cfg.GetMetricsAddress())
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.ErrorWithErr(err, "Metrics server error")
			}
		}()
	}

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			logger.ErrorWithErr(err, "Server error")
			os.Exit(1)
		}
	case sig := <-shutdown:
		logger.Infof("Received signal: %v. Starting graceful shutdown...", sig)

		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			logger.ErrorWithErr(err, "Error during server shutdown")
			if err := server.Close(); err != nil {
				logger.ErrorWithErr(err, "Error closing server")
			}
		}

		// Shutdown metrics server if running
		if metricsServer != nil {
			if err := metricsServer.Shutdown(ctx); err != nil {
				logger.ErrorWithErr(err, "Error during metrics server shutdown")
			}
		}

		logger.Info("Server stopped gracefully")
	}
}
