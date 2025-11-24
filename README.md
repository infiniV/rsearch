# rsearch - Phase 1: Foundation

This is the foundation implementation of rsearch, a production-grade query translation service.

## What's Implemented

Phase 1 (Foundation) includes:

- Complete HTTP server infrastructure with chi router
- Configuration system with viper (YAML/JSON + environment variables)
- Structured logging with zerolog (JSON and console formats)
- Prometheus metrics integration (optional)
- Health and readiness endpoints
- Graceful shutdown with signal handling
- Comprehensive middleware (logging, recovery, CORS, request ID)
- Standard error response format
- Full test coverage

## Project Structure

```
.
├── cmd/
│   └── rsearch/
│       ├── main.go           # Server entry point
│       └── main_test.go      # Integration tests
├── internal/
│   ├── api/
│   │   ├── handlers.go       # HTTP handlers
│   │   ├── handlers_test.go  # Handler tests
│   │   ├── middleware.go     # HTTP middleware
│   │   ├── response.go       # Response helpers
│   │   └── routes.go         # Route definitions
│   ├── config/
│   │   ├── config.go         # Configuration loading
│   │   └── config_test.go    # Config tests
│   └── observability/
│       ├── logger.go         # Structured logging
│       └── metrics.go        # Prometheus metrics
├── pkg/
│   └── rsearch/
│       └── types.go          # Public types
├── bin/
│   └── rsearch              # Compiled binary
├── config.example.yaml      # Example configuration
└── go.mod                   # Go module definition
```

## Building

```bash
go build -o bin/rsearch cmd/rsearch/main.go
```

## Running Tests

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -cover ./...
```

## Running the Server

```bash
# Run with defaults
./bin/rsearch

# Run with custom config
./bin/rsearch --config config.yaml

# Or use environment variables
export RSEARCH_SERVER_PORT=9000
export RSEARCH_LOGGING_LEVEL=debug
./bin/rsearch
```

## Available Endpoints

### Health Check
```bash
curl http://localhost:8080/health
# Response: {"status":"healthy","version":"1.0.0"}
```

### Readiness Check
```bash
curl http://localhost:8080/ready
# Response: {"ready":true,"version":"1.0.0"}
```

### Metrics (when enabled)
```bash
curl http://localhost:9090/metrics
# Returns Prometheus-formatted metrics
```

### API Endpoints (Not Yet Implemented)
- `POST /api/v1/schemas` - Register schema
- `GET /api/v1/schemas/{name}` - Get schema
- `DELETE /api/v1/schemas/{name}` - Delete schema
- `POST /api/v1/translate` - Translate query

These endpoints currently return 501 Not Implemented.

## Configuration

Configuration can be provided via:

1. Configuration file (YAML or JSON)
2. Environment variables (prefix: `RSEARCH_`)
3. Default values

Priority: Environment variables > Config file > Defaults

### Example Configuration

See `config.example.yaml` for a complete configuration example.

### Environment Variable Examples

```bash
# Server configuration
export RSEARCH_SERVER_PORT=9000
export RSEARCH_SERVER_HOST=0.0.0.0

# Logging
export RSEARCH_LOGGING_LEVEL=debug
export RSEARCH_LOGGING_FORMAT=console

# Metrics
export RSEARCH_METRICS_ENABLED=true
export RSEARCH_METRICS_PORT=9090

# CORS
export RSEARCH_CORS_ENABLED=true
```

## Logging

The server uses structured logging with zerolog. Two formats are supported:

### JSON Format (default)
```json
{"level":"info","time":"2024-11-24T18:30:00+00:00","message":"Server listening on localhost:8080"}
```

### Console Format
```
2024-11-24 18:30:00 INF Server listening on localhost:8080
```

### Log Levels
- `debug` - Detailed debugging information
- `info` - General informational messages
- `warn` - Warning messages
- `error` - Error messages

## Metrics

When metrics are enabled, Prometheus-compatible metrics are exposed:

- `rsearch_requests_total` - Total HTTP requests by endpoint and status
- `rsearch_request_duration_seconds` - Request duration histogram
- `rsearch_active_requests` - Current active requests
- `rsearch_errors_total` - Total errors by type
- `rsearch_parse_duration_seconds` - Query parsing duration (Phase 2)
- `rsearch_translate_duration_seconds` - Translation duration (Phase 4)
- `rsearch_active_schemas` - Number of registered schemas (Phase 3)
- `rsearch_cache_hits_total` - Cache hits
- `rsearch_cache_misses_total` - Cache misses

## Middleware

The following middleware is automatically applied:

1. **Request ID** - Adds unique request ID to each request
2. **Logging** - Logs all requests with timing information
3. **Recovery** - Recovers from panics and returns 500 error
4. **CORS** - Handles CORS headers (when enabled)
5. **Metrics** - Records Prometheus metrics (when enabled)

## Graceful Shutdown

The server handles `SIGTERM` and `SIGINT` signals gracefully:

1. Stops accepting new connections
2. Waits for active requests to complete (up to shutdownTimeout)
3. Closes server cleanly

## Test Results

All tests pass:
```
ok      github.com/infiniv/rsearch/cmd/rsearch          0.109s
ok      github.com/infiniv/rsearch/internal/api         0.006s
ok      github.com/infiniv/rsearch/internal/config      0.005s
```

Coverage includes:
- Configuration loading and validation
- Health and readiness endpoints
- Metrics endpoint (enabled and disabled)
- Server integration (startup, endpoints, shutdown)
- Request ID handling (provided and auto-generated)

## Next Steps

Phase 2 will implement:
- Lexer for tokenizing query strings
- Parser for building Abstract Syntax Tree (AST)
- AST node definitions
- Comprehensive parser tests

## Dependencies

- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/spf13/viper` - Configuration management
- `github.com/rs/zerolog` - Structured logging
- `github.com/prometheus/client_golang` - Metrics
- `github.com/google/uuid` - UUID generation

## Version

Current version: **1.0.0** (Foundation Phase)
