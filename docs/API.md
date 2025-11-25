# rsearch API Guide

Complete reference for the rsearch query translation API.

## Table of Contents

- [Quick Start](#quick-start)
- [Authentication](#authentication)
- [API Endpoints](#api-endpoints)
  - [Translation](#translation)
  - [Schema Management](#schema-management)
  - [Health & Monitoring](#health--monitoring)
- [Query Syntax](#query-syntax)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Best Practices](#best-practices)

## Quick Start

### 1. Start the server

```bash
# Using Docker
docker run -p 8080:8080 rsearch:latest

# Or build from source
make build
./bin/rsearch
```

### 2. Register a schema

```bash
curl -X POST http://localhost:8080/api/v1/schemas \
  -H "Content-Type: application/json" \
  -d '{
    "name": "users",
    "fields": {
      "id": {"type": "integer", "indexed": true},
      "email": {"type": "text", "indexed": true},
      "status": {"type": "text"},
      "createdAt": {"type": "datetime"}
    },
    "options": {
      "namingConvention": "snake_case",
      "strictFieldNames": false,
      "defaultField": "email"
    }
  }'
```

### 3. Translate a query

```bash
curl -X POST http://localhost:8080/api/v1/translate \
  -H "Content-Type: application/json" \
  -d '{
    "schema": "users",
    "database": "postgres",
    "query": "status:active AND createdAt:>=2024-01-01"
  }'
```

**Response:**
```json
{
  "type": "postgres",
  "whereClause": "status = $1 AND created_at >= $2",
  "parameters": ["active", "2024-01-01"],
  "parameterTypes": ["text", "datetime"]
}
```

### 4. Use in your application

```go
// Execute the translated query
query := fmt.Sprintf("SELECT * FROM users WHERE %s", response.WhereClause)
rows, err := db.Query(query, response.Parameters...)
```

## Authentication

API key authentication is optional and disabled by default.

### Enable Authentication

Configure via environment variables:

```bash
export RSEARCH_SECURITY_AUTH_ENABLED=true
export RSEARCH_SECURITY_AUTH_TYPE=apikey
export RSEARCH_SECURITY_AUTH_APIKEYS=your-secret-key-1,your-secret-key-2
```

Or via config file:

```yaml
security:
  auth:
    enabled: true
    type: apikey
    apiKeys:
      - your-secret-key-1
      - your-secret-key-2
```

### Using API Keys

Include the API key in the `X-API-Key` header:

```bash
curl -X POST http://localhost:8080/api/v1/translate \
  -H "X-API-Key: your-secret-key-1" \
  -H "Content-Type: application/json" \
  -d '{...}'
```

**Unauthorized response (401):**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or missing API key"
  }
}
```

## API Endpoints

### Translation

#### POST /api/v1/translate

Translates an OpenSearch/Elasticsearch query string into database-specific SQL.

**Request:**

```json
{
  "schema": "users",
  "database": "postgres",
  "query": "status:active AND age:>18"
}
```

**Fields:**
- `schema` (required): Name of registered schema
- `database` (required): Target database type (currently only `postgres`)
- `query` (required): Query string in OpenSearch/Elasticsearch syntax

**Response (200 OK):**

```json
{
  "type": "postgres",
  "whereClause": "status = $1 AND age > $2",
  "parameters": ["active", 18],
  "parameterTypes": ["text", "integer"]
}
```

**Fields:**
- `type`: Database type
- `whereClause`: SQL WHERE clause (without the WHERE keyword)
- `parameters`: Ordered parameter values for parameterized query
- `parameterTypes`: Type of each parameter for proper casting

**Error Responses:**

| Status | Code | Description |
|--------|------|-------------|
| 400 | PARSE_ERROR | Invalid query syntax |
| 400 | FIELD_NOT_FOUND | Field not found in schema |
| 400 | TYPE_MISMATCH | Value type doesn't match field type |
| 400 | FEATURE_DISABLED | Using disabled feature (fuzzy, regex, etc.) |
| 404 | SCHEMA_NOT_FOUND | Schema not registered |
| 429 | RATE_LIMITED | Rate limit exceeded |
| 500 | INTERNAL_ERROR | Server error |

**Examples:**

**Simple field query:**
```bash
curl -X POST http://localhost:8080/api/v1/translate \
  -H "Content-Type: application/json" \
  -d '{
    "schema": "users",
    "database": "postgres",
    "query": "email:john@example.com"
  }'
```

**Boolean operators:**
```bash
curl -X POST http://localhost:8080/api/v1/translate \
  -H "Content-Type: application/json" \
  -d '{
    "schema": "products",
    "database": "postgres",
    "query": "(category:electronics OR category:computers) AND price:<1000"
  }'
```

**Range queries:**
```bash
curl -X POST http://localhost:8080/api/v1/translate \
  -H "Content-Type: application/json" \
  -d '{
    "schema": "orders",
    "database": "postgres",
    "query": "createdAt:[2024-01-01 TO 2024-12-31] AND total:>=100"
  }'
```

**Wildcard queries:**
```bash
curl -X POST http://localhost:8080/api/v1/translate \
  -H "Content-Type: application/json" \
  -d '{
    "schema": "users",
    "database": "postgres",
    "query": "email:*@example.com"
  }'
```

**Fuzzy search (requires enabled feature):**
```bash
curl -X POST http://localhost:8080/api/v1/translate \
  -H "Content-Type: application/json" \
  -d '{
    "schema": "products",
    "database": "postgres",
    "query": "name:laptop~2"
  }'
```

### Schema Management

#### POST /api/v1/schemas

Register a new schema in the registry.

**Request:**

```json
{
  "name": "products",
  "fields": {
    "productId": {
      "type": "integer",
      "column": "product_id",
      "indexed": true
    },
    "productName": {
      "type": "text",
      "column": "product_name",
      "aliases": ["name", "title"]
    },
    "price": {
      "type": "float",
      "indexed": true
    },
    "category": {
      "type": "text"
    },
    "tags": {
      "type": "array"
    },
    "createdAt": {
      "type": "datetime"
    }
  },
  "options": {
    "namingConvention": "snake_case",
    "strictFieldNames": false,
    "strictOperators": false,
    "defaultField": "productName",
    "enabledFeatures": {
      "fuzzy": true,
      "proximity": true,
      "regex": true
    }
  }
}
```

**Field Types:**
- `text` - String/varchar fields
- `integer` - Integer numbers
- `float` - Floating-point numbers
- `boolean` - Boolean values (true/false)
- `datetime` - Date and time
- `date` - Date only
- `time` - Time only
- `json` - JSON fields
- `array` - Array fields

**Schema Options:**
- `namingConvention`: Transform field names (`snake_case`, `camelCase`, `PascalCase`, `none`)
- `strictFieldNames`: Case-sensitive field name matching (default: false)
- `strictOperators`: Case-sensitive operators (default: false)
- `defaultField`: Field to use for queries without field specifier
- `enabledFeatures`: Optional database features

**Enabled Features:**
- `fuzzy`: Fuzzy search using Levenshtein distance (requires `pg_trgm`)
- `proximity`: Proximity search (requires full-text search setup)
- `regex`: Regular expression matching

**Response (201 Created):**

```json
{
  "message": "schema registered successfully",
  "data": {
    "name": "products",
    "fields": {...},
    "options": {...},
    "createdAt": "2024-11-25T10:30:00Z"
  }
}
```

**Error Responses:**

| Status | Code | Description |
|--------|------|-------------|
| 400 | INVALID_SCHEMA | Schema validation failed |
| 409 | SCHEMA_EXISTS | Schema name already exists |

#### GET /api/v1/schemas

List all registered schemas.

**Response (200 OK):**

```json
{
  "data": ["users", "products", "orders"]
}
```

#### GET /api/v1/schemas/{name}

Retrieve a specific schema by name.

**Response (200 OK):**

```json
{
  "name": "users",
  "fields": {
    "id": {"type": "integer", "indexed": true},
    "email": {"type": "text", "indexed": true}
  },
  "options": {...},
  "createdAt": "2024-11-25T10:30:00Z"
}
```

**Error Responses:**

| Status | Code | Description |
|--------|------|-------------|
| 404 | SCHEMA_NOT_FOUND | Schema not found |

#### DELETE /api/v1/schemas/{name}

Delete a schema from the registry.

**Response (204 No Content)**

**Error Responses:**

| Status | Code | Description |
|--------|------|-------------|
| 404 | SCHEMA_NOT_FOUND | Schema not found |

**Example:**

```bash
curl -X DELETE http://localhost:8080/api/v1/schemas/users
```

### Health & Monitoring

#### GET /health

Health check endpoint. Returns service health status.

**Response (200 OK):**

```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

#### GET /ready

Readiness check endpoint. Indicates if the service is ready to accept requests.

**Response (200 OK):**

```json
{
  "ready": true,
  "version": "1.0.0"
}
```

#### GET /metrics

Prometheus metrics endpoint. Only available when metrics are enabled.

**Configuration:**

```bash
export RSEARCH_METRICS_ENABLED=true
export RSEARCH_METRICS_PORT=9090
export RSEARCH_METRICS_PATH=/metrics
```

**Response (200 OK):**

```
# HELP rsearch_requests_total Total HTTP requests
# TYPE rsearch_requests_total counter
rsearch_requests_total{endpoint="/api/v1/translate",status="200"} 150

# HELP rsearch_request_duration_seconds Request duration in seconds
# TYPE rsearch_request_duration_seconds histogram
rsearch_request_duration_seconds_bucket{endpoint="/api/v1/translate",le="0.005"} 100
```

**Available Metrics:**
- `rsearch_requests_total` - Total HTTP requests by endpoint and status
- `rsearch_request_duration_seconds` - Request duration histogram
- `rsearch_active_requests` - Current active requests
- `rsearch_errors_total` - Total errors by type
- `rsearch_parse_duration_seconds` - Query parsing duration
- `rsearch_translate_duration_seconds` - Translation duration
- `rsearch_active_schemas` - Number of registered schemas
- `rsearch_cache_hits_total` - Cache hits
- `rsearch_cache_misses_total` - Cache misses

## Query Syntax

rsearch supports OpenSearch/Elasticsearch query string syntax. See the [full syntax reference](syntax-reference.md) for complete documentation.

### Quick Reference

**Field queries:**
```
field:value              # Exact match
field:"phrase query"     # Phrase match
field:wild*              # Wildcard
field:/regex/            # Regex (if enabled)
```

**Boolean operators:**
```
term1 AND term2          # Conjunction (also: &&)
term1 OR term2           # Disjunction (also: ||)
NOT term                 # Negation (also: !)
+term                    # Required
-term                    # Prohibited
(term1 OR term2)         # Grouping
```

**Range queries:**
```
[50 TO 500]              # Inclusive range
{50 TO 500}              # Exclusive range
[50 TO 500}              # Mixed
>=50                     # Greater than or equal
<100                     # Less than
[100 TO *]               # Unbounded range
```

**Advanced queries:**
```
field:term~2             # Fuzzy search (Levenshtein distance)
"phrase"~5               # Proximity search
field:value^2            # Boost (metadata only)
_exists_:field           # Existence check
field:(a OR b)           # Field group
```

## Error Handling

All errors follow a standard format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": [
      {
        "position": 10,
        "message": "Unexpected token"
      }
    ],
    "query": "original query string"
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| PARSE_ERROR | 400 | Query parsing failed |
| SCHEMA_NOT_FOUND | 404 | Schema not registered |
| FIELD_NOT_FOUND | 400 | Field not found in schema |
| TYPE_MISMATCH | 400 | Value type doesn't match field type |
| FEATURE_DISABLED | 400 | Feature not enabled for schema |
| INVALID_RANGE | 400 | Invalid range query |
| UNSUPPORTED_SYNTAX | 400 | Unsupported query syntax |
| SCHEMA_EXISTS | 409 | Schema name already exists |
| INVALID_SCHEMA | 400 | Schema validation failed |
| INTERNAL_ERROR | 500 | Internal server error |
| RATE_LIMITED | 429 | Rate limit exceeded |
| UNAUTHORIZED | 401 | Invalid or missing API key |
| FORBIDDEN | 403 | Access forbidden |
| QUERY_TOO_LONG | 400 | Query exceeds maximum length |
| TOO_MANY_PARAMETERS | 400 | Too many query parameters |
| TIMEOUT | 408 | Request timeout |
| SERVICE_UNAVAILABLE | 503 | Service temporarily unavailable |

### Example Error Response

```json
{
  "error": {
    "code": "PARSE_ERROR",
    "message": "Failed to parse query: unexpected token at position 17",
    "details": [
      {
        "position": 17,
        "line": 1,
        "column": 17,
        "message": "Expected closing parenthesis"
      }
    ],
    "query": "status:active AND ("
  }
}
```

## Rate Limiting

Rate limiting is optional and can be configured to protect against abuse.

### Configuration

```bash
# Enable rate limiting
export RSEARCH_LIMITS_RATELIMIT_ENABLED=true
export RSEARCH_LIMITS_RATELIMIT_REQUESTSPERMINUTE=100
export RSEARCH_LIMITS_RATELIMIT_REQUESTSPERHOUR=5000
export RSEARCH_LIMITS_RATELIMIT_BURST=10
```

Or via config file:

```yaml
limits:
  rateLimit:
    enabled: true
    requestsPerMinute: 100
    requestsPerHour: 5000
    burst: 10
```

### Rate Limit Headers

When rate limiting is enabled, responses include headers:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1732540800
```

### Rate Limit Exceeded Response (429)

```json
{
  "error": {
    "code": "RATE_LIMITED",
    "message": "Rate limit exceeded. Please try again later."
  }
}
```

## Best Practices

### 1. Register Schemas on Startup

Register all required schemas when your application starts, not on every request:

```go
func initSchemas(client *http.Client) error {
    schemas := []Schema{usersSchema, productsSchema, ordersSchema}
    for _, schema := range schemas {
        if err := registerSchema(client, schema); err != nil {
            return err
        }
    }
    return nil
}
```

### 2. Cache Schemas Locally

Fetch and cache schema definitions to reduce API calls:

```go
type SchemaCache struct {
    schemas map[string]*Schema
    mu      sync.RWMutex
}

func (c *SchemaCache) Get(name string) (*Schema, error) {
    c.mu.RLock()
    schema, ok := c.schemas[name]
    c.mu.RUnlock()

    if !ok {
        // Fetch from API and cache
        schema, err := fetchSchema(name)
        if err != nil {
            return nil, err
        }
        c.mu.Lock()
        c.schemas[name] = schema
        c.mu.Unlock()
    }
    return schema, nil
}
```

### 3. Use Connection Pooling

Configure HTTP client with appropriate connection pooling:

```go
client := &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

### 4. Handle Errors Gracefully

Always check error codes and handle appropriately:

```go
resp, err := translateQuery(client, req)
if err != nil {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        switch apiErr.Code {
        case "PARSE_ERROR":
            return fmt.Errorf("invalid query syntax: %w", err)
        case "SCHEMA_NOT_FOUND":
            return fmt.Errorf("schema not found: %w", err)
        case "RATE_LIMITED":
            time.Sleep(time.Second)
            return translateQuery(client, req) // Retry
        default:
            return err
        }
    }
    return err
}
```

### 5. Validate Input

Validate query strings before sending to API:

```go
func validateQuery(query string, maxLength int) error {
    if len(query) == 0 {
        return errors.New("query cannot be empty")
    }
    if len(query) > maxLength {
        return fmt.Errorf("query too long: %d > %d", len(query), maxLength)
    }
    return nil
}
```

### 6. Use Request IDs

Include request IDs for tracing:

```go
req, _ := http.NewRequest("POST", url, body)
req.Header.Set("X-Request-ID", requestID)
req.Header.Set("Content-Type", "application/json")
```

### 7. Monitor Metrics

Set up Prometheus to scrape metrics and create alerts:

```yaml
scrape_configs:
  - job_name: 'rsearch'
    static_configs:
      - targets: ['localhost:9090']
```

### 8. Implement Circuit Breaker

Use circuit breaker pattern for resilience:

```go
type CircuitBreaker struct {
    maxFailures int
    failures    int
    lastFailure time.Time
    timeout     time.Duration
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.failures >= cb.maxFailures {
        if time.Since(cb.lastFailure) < cb.timeout {
            return errors.New("circuit breaker open")
        }
        cb.failures = 0 // Reset
    }

    if err := fn(); err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        return err
    }

    cb.failures = 0
    return nil
}
```

### 9. Use Structured Logging

Log API interactions with structured fields:

```go
logger.WithFields(map[string]interface{}{
    "request_id": requestID,
    "schema":     "users",
    "query":      query,
    "duration":   duration.Milliseconds(),
}).Info("Query translated successfully")
```

### 10. Test with Real Data

Test with realistic query patterns and data:

```go
func TestTranslation(t *testing.T) {
    tests := []struct {
        name     string
        query    string
        expected string
    }{
        {"simple", "status:active", "status = $1"},
        {"boolean", "status:active AND age:>18", "status = $1 AND age > $2"},
        {"range", "createdAt:[2024-01-01 TO 2024-12-31]", "created_at BETWEEN $1 AND $2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp, err := translate(tt.query)
            require.NoError(t, err)
            assert.Equal(t, tt.expected, resp.WhereClause)
        })
    }
}
```

## Additional Resources

- [Query Syntax Reference](syntax-reference.md) - Complete query syntax documentation
- [Deployment Guide](DEPLOYMENT.md) - Production deployment instructions
- [OpenAPI Specification](openapi.yaml) - Machine-readable API spec
- [GitHub Repository](https://github.com/infiniv/rsearch) - Source code and issues
