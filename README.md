# rsearch

A production-grade query translation service that converts OpenSearch/Elasticsearch-style query strings into database-specific formats. Supports PostgreSQL, MySQL, SQLite, and MongoDB.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)

## Overview

rsearch acts as a lightweight bridge between search interfaces and databases. It parses familiar OpenSearch query syntax and translates it into parameterized queries for your target database.

```
Query String  -->  Parser  -->  AST  -->  Translator  -->  SQL/BSON + Parameters
```

## Key Features

**Multi-Database Support**
- PostgreSQL with parameterized queries ($1, $2...)
- MySQL with prepared statement placeholders (?)
- SQLite with parameter binding
- MongoDB with native BSON filters

**Query Capabilities**
- Boolean operators (AND, OR, NOT)
- Range queries with inclusive/exclusive bounds
- Wildcard and pattern matching
- Fuzzy search with Levenshtein distance
- Proximity search for phrase matching
- Regular expression support
- Field existence checks
- Query boosting for relevance scoring

**Production Ready**
- Rate limiting per client
- Request caching with TTL
- Prometheus metrics
- Health and readiness endpoints
- Structured JSON logging
- Graceful shutdown
- API key authentication

## Quick Start

### Using Docker

```bash
docker run -d -p 8080:8080 rsearch:latest
```

### Using Docker Compose

```bash
docker-compose up -d
```

### Building from Source

```bash
git clone https://github.com/infiniv/rsearch.git
cd rsearch
make build
./bin/rsearch
```

## Usage

### 1. Register a Schema

Schemas define your field types and enable query validation.

```bash
curl -X POST http://localhost:8080/api/v1/schemas \
  -H "Content-Type: application/json" \
  -d '{
    "name": "products",
    "fields": {
      "name": {"type": "text", "indexed": true},
      "price": {"type": "float"},
      "category": {"type": "text"},
      "createdAt": {"type": "datetime"}
    },
    "options": {
      "namingConvention": "snake_case",
      "defaultField": "name",
      "enabledFeatures": {
        "fuzzy": true,
        "proximity": true,
        "regex": true
      }
    }
  }'
```

### 2. Translate a Query

```bash
curl -X POST http://localhost:8080/api/v1/translate \
  -H "Content-Type: application/json" \
  -d '{
    "schema": "products",
    "database": "postgres",
    "query": "category:electronics AND price:[100 TO 500]"
  }'
```

**Response:**

```json
{
  "type": "sql",
  "whereClause": "category = $1 AND price BETWEEN $2 AND $3",
  "parameters": ["electronics", "100", "500"],
  "parameterTypes": ["text", "float", "float"]
}
```

### 3. Execute in Your Application

```go
query := fmt.Sprintf("SELECT * FROM products WHERE %s", response.WhereClause)
rows, err := db.Query(query, response.Parameters...)
```

## Query Syntax

rsearch supports OpenSearch/Elasticsearch query string syntax.

### Field Queries

```
name:laptop                    # Exact match
name:"gaming laptop"           # Phrase match
category:electronics           # Term match
```

### Boolean Operators

```
status:active AND price:<100   # AND operator
category:phones OR category:tablets   # OR operator
NOT status:discontinued        # NOT operator
status:active && inStock:true  # Alternative syntax (&&, ||, !)
```

### Range Queries

```
price:[100 TO 500]            # Inclusive range (BETWEEN)
price:{100 TO 500}            # Exclusive range (> AND <)
price:[100 TO 500}            # Mixed bounds
price:>=100                   # Comparison operators
price:[100 TO *]              # Unbounded range
```

### Wildcards and Patterns

```
name:lap*                     # Suffix wildcard (LIKE 'lap%')
name:*top                     # Prefix wildcard
name:*apt*                    # Contains
name:la?top                   # Single character wildcard
name:/^laptop$/               # Regular expression
```

### Advanced Features

```
name:laptop~2                 # Fuzzy search (Levenshtein distance 2)
description:"fast delivery"~5 # Proximity search (words within 5 positions)
name:laptop^2                 # Boost factor (stored in metadata)
_exists_:description          # Field existence (IS NOT NULL)
status:(active OR pending)    # Field grouping
```

For complete syntax documentation, see [Query Syntax Reference](docs/syntax-reference.md).

## Database Translations

### PostgreSQL

```sql
-- Input: category:electronics AND price:[100 TO 500]
-- Output:
WHERE category = $1 AND price BETWEEN $2 AND $3
-- Parameters: ["electronics", "100", "500"]
```

### MySQL

```sql
-- Input: category:electronics AND price:[100 TO 500]
-- Output:
WHERE category = ? AND price BETWEEN ? AND ?
-- Parameters: ["electronics", "100", "500"]
```

### SQLite

```sql
-- Input: category:electronics AND price:[100 TO 500]
-- Output:
WHERE category = ? AND price BETWEEN ? AND ?
-- Parameters: ["electronics", "100", "500"]
```

### MongoDB

```javascript
// Input: category:electronics AND price:[100 TO 500]
// Output:
{
  "$and": [
    {"category": "electronics"},
    {"price": {"$gte": "100", "$lte": "500"}}
  ]
}
```

## API Reference

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/translate` | Translate a query string |
| POST | `/api/v1/schemas` | Register a new schema |
| GET | `/api/v1/schemas` | List all schemas |
| GET | `/api/v1/schemas/{name}` | Get a specific schema |
| DELETE | `/api/v1/schemas/{name}` | Delete a schema |
| GET | `/health` | Health check |
| GET | `/ready` | Readiness check |
| GET | `/metrics` | Prometheus metrics |

For detailed API documentation, see [API Guide](docs/API.md).

## Configuration

rsearch can be configured via YAML file, environment variables, or command-line flags.

**Priority:** Environment variables > Config file > Defaults

### Environment Variables

```bash
# Server
RSEARCH_SERVER_HOST=0.0.0.0
RSEARCH_SERVER_PORT=8080

# Logging
RSEARCH_LOGGING_LEVEL=info
RSEARCH_LOGGING_FORMAT=json

# Metrics
RSEARCH_METRICS_ENABLED=true
RSEARCH_METRICS_PORT=9090

# Rate Limiting
RSEARCH_LIMITS_RATELIMIT_ENABLED=true
RSEARCH_LIMITS_RATELIMIT_REQUESTSPERMINUTE=100

# Caching
RSEARCH_CACHE_ENABLED=true
RSEARCH_CACHE_MAXSIZE=10000
RSEARCH_CACHE_TTL=3600
```

### Configuration File

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  readTimeout: 30s
  writeTimeout: 30s

logging:
  level: "info"
  format: "json"

metrics:
  enabled: true
  port: 9090
  path: "/metrics"

limits:
  maxQueryLength: 10000
  maxParameterCount: 100
  rateLimit:
    enabled: true
    requestsPerMinute: 100

cache:
  enabled: true
  maxSize: 10000
  ttl: 3600
```

See [config.example.yaml](config.example.yaml) for all options.

## Deployment

### Docker

```bash
# Build
docker build -t rsearch:latest .

# Run
docker run -d \
  -p 8080:8080 \
  -p 9090:9090 \
  -e RSEARCH_LOGGING_LEVEL=info \
  rsearch:latest
```

### Kubernetes

rsearch includes production-ready Kubernetes manifests with:

- Deployment with resource limits and health checks
- Service for internal/external access
- ConfigMap for configuration
- HorizontalPodAutoscaler for auto-scaling
- PodDisruptionBudget for availability
- ServiceMonitor for Prometheus integration
- Ingress for external routing

```bash
# Deploy to Kubernetes
kubectl apply -k k8s/overlays/production

# Or use the deploy script
./k8s/deploy.sh production
```

See [Deployment Guide](docs/DEPLOYMENT.md) and [Kubernetes Architecture](k8s/ARCHITECTURE.md) for details.

### CI/CD

GitHub Actions workflows are included for:

- **ci.yaml** - Build, test, and lint on every push
- **release.yaml** - Automated releases with Docker image publishing
- **codeql.yaml** - Security analysis

See [CI/CD Guide](docs/CI_CD.md) for setup instructions.

## Project Structure

```
rsearch/
├── cmd/
│   ├── rsearch/          # Server entry point
│   └── gendocs/          # Documentation generator
├── internal/
│   ├── api/              # HTTP handlers and middleware
│   ├── parser/           # Lexer and recursive descent parser
│   ├── translator/       # Database translators (postgres, mysql, sqlite, mongodb)
│   ├── schema/           # Schema registry and validation
│   ├── config/           # Configuration management
│   ├── cache/            # Query caching
│   ├── ratelimit/        # Rate limiting
│   ├── validation/       # Input validation
│   └── observability/    # Logging and metrics
├── pkg/rsearch/          # Public types and interfaces
├── docs/                 # Documentation
├── k8s/                  # Kubernetes manifests
├── examples/             # Client examples (Go, Python, Node.js, PHP)
└── tests/                # Integration tests
```

## Schema Configuration

### Field Types

| Type | Description | PostgreSQL | MySQL | MongoDB |
|------|-------------|------------|-------|---------|
| `text` | String values | VARCHAR/TEXT | VARCHAR | String |
| `integer` | Whole numbers | INTEGER | INT | Int32/Int64 |
| `float` | Decimal numbers | NUMERIC | DECIMAL | Double |
| `boolean` | True/false | BOOLEAN | TINYINT | Boolean |
| `datetime` | Timestamps | TIMESTAMP | DATETIME | Date |
| `date` | Date only | DATE | DATE | Date |
| `json` | JSON objects | JSONB | JSON | Object |
| `array` | Arrays | ARRAY | JSON | Array |

### Schema Options

| Option | Description | Default |
|--------|-------------|---------|
| `namingConvention` | Field name transformation (snake_case, camelCase) | none |
| `strictFieldNames` | Reject unknown fields | false |
| `strictOperators` | Reject unsupported operators | false |
| `defaultField` | Field for unqualified terms | none |
| `enabledFeatures.fuzzy` | Enable fuzzy search | false |
| `enabledFeatures.proximity` | Enable proximity search | false |
| `enabledFeatures.regex` | Enable regex matching | false |

## Performance

rsearch is optimized for high throughput and low latency:

- Sub-millisecond p99 latency for typical queries
- Lock-free read paths for schema lookups
- Query result caching with configurable TTL
- Connection pooling for database execution

**Benchmarks (approximate):**

| Operation | Throughput | Latency (p99) |
|-----------|------------|---------------|
| Parse simple query | ~100,000/sec | <50us |
| Parse complex query | ~50,000/sec | <100us |
| Translate to SQL | ~200,000/sec | <25us |

## Development

```bash
# Install dependencies
go mod download

# Run tests
make test

# Run with coverage
make test-coverage

# Build
make build

# Run locally
make run

# Start demo environment
make demo

# Lint
make lint

# Format
make fmt
```

## Documentation

- [API Guide](docs/API.md) - Complete API reference
- [Query Syntax Reference](docs/syntax-reference.md) - All supported query patterns
- [Deployment Guide](docs/DEPLOYMENT.md) - Production deployment instructions
- [CI/CD Guide](docs/CI_CD.md) - Continuous integration setup
- [Kubernetes Architecture](k8s/ARCHITECTURE.md) - Kubernetes deployment details
- [OpenAPI Specification](docs/openapi.yaml) - Machine-readable API spec

## Requirements

- Go 1.21 or later
- PostgreSQL 12+ (for fuzzy search: fuzzystrmatch and pg_trgm extensions)
- MySQL 8.0+ (optional)
- MongoDB 5.0+ (optional)

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome. Please read the contribution guidelines and submit pull requests to the main repository.

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

---

**rsearch** - Query translation for modern applications
