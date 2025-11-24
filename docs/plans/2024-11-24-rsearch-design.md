# rsearch Design Document

**Date:** 2024-11-24
**Status:** Approved
**Version:** 1.0

## Overview

**rsearch** is a production-grade, lightweight query translation service that converts OpenSearch/Elasticsearch-style query strings into database-specific query formats. It acts as a bridge between application search UIs and databases, enabling advanced search capabilities without requiring full-text search infrastructure.

### Core Concept

```
User Query: "productCode:13w42 AND rodLength:[50 TO 500]"
         ↓
    rsearch API
         ↓
SQL Output: "product_code = $1 AND rod_length BETWEEN $2 AND $3"
Parameters: ["13w42", 50, 500]
         ↓
Application integrates into own database queries
```

### Key Principles

- **Lightweight Bridge**: Translates queries only, never executes them or stores data
- **Database Agnostic**: Supports SQL and NoSQL databases through translator interface
- **Production Grade**: Full observability, security, configurability from v1.0
- **Zero Dependencies**: Standalone binary, no external services required (no Redis, no database connections)
- **Developer Friendly**: Simple HTTP API, works with any language, comprehensive auto-generated docs

---

## Architecture Overview

### High-Level Architecture

**rsearch** is a centralized Go HTTP server that applications connect to via REST API.

**Core Flow:**
1. Application sends query string via HTTP POST
2. rsearch parses query into Abstract Syntax Tree (AST)
3. Database-specific translator converts AST to query format
4. Returns WHERE clause (SQL) or filter object (NoSQL) with parameters
5. Application integrates result into their own queries

**Design Principles:**
- **Stateless translation**: Only stores registered schemas in memory, no query history
- **Return WHERE clause only**: Never executes queries, no database connections
- **Database agnostic**: Interface-based translators, easy to add new databases
- **Comprehensive syntax**: Full OpenSearch query string support from day one
- **Developer-friendly**: Auto-generated docs from tests, clear error messages, sensible defaults

**Technology Stack:**
- Go 1.21+ (simple deployment, good performance, strong typing)
- HTTP REST API (language agnostic)
- In-memory schema storage with RWMutex (thread-safe)
- Standard library + minimal well-established dependencies

**Production-Grade Features:**
- **Configuration**: YAML/JSON config files + environment variables (using viper)
- **Logging**: Structured logging with levels (using zerolog or zap)
- **Metrics**: Prometheus-compatible metrics endpoint
- **Health checks**: `/health` and `/ready` endpoints
- **Graceful shutdown**: Proper signal handling
- **Validation**: Comprehensive input validation and clear error responses
- **Observability**: Request IDs, distributed tracing ready

**External Dependencies (all compile into single binary):**
- **chi** or **gin** - battle-tested HTTP router
- **viper** - configuration management
- **zerolog** or **zap** - structured logging
- **prometheus/client_golang** - metrics (optional, opt-in monitoring)

**Deployment:**
- Single binary, runs anywhere
- Docker image (Alpine-based)
- Kubernetes manifests available
- Systemd service file included

---

## API Design

### Base URL

```
http://localhost:8080
```

### Core Endpoints

#### 1. Register Schema

```
POST /api/v1/schemas
```

**Request:**
```json
{
  "name": "products",
  "fields": {
    "productCode": {"type": "text"},
    "rodLength": {"type": "integer"},
    "price": {"type": "float"},
    "region": {"type": "text"},
    "inStock": {"type": "boolean"},
    "createdAt": {"type": "datetime"},
    "tags": {"type": "array"},
    "metadata": {"type": "json"}
  },
  "options": {
    "namingConvention": "snake_case",
    "strictOperators": false,
    "strictFieldNames": false,
    "defaultField": "name"
  }
}
```

**Response: 201 Created**
```json
{
  "name": "products",
  "fieldCount": 8,
  "registered": "2024-11-24T18:30:00Z"
}
```

#### 2. Translate Query

```
POST /api/v1/translate
```

**Request:**
```json
{
  "schema": "products",
  "database": "postgres",
  "query": "productCode:13w42 AND rodLength:[50 TO 500] AND region:ca"
}
```

**Response for SQL databases (postgres, mysql, sqlite):**
```json
{
  "type": "sql",
  "whereClause": "product_code = $1 AND rod_length BETWEEN $2 AND $3 AND region = $4",
  "parameters": ["13w42", 50, 500, "ca"],
  "parameterTypes": ["text", "integer", "integer", "text"]
}
```

**Response for NoSQL databases (mongodb):**
```json
{
  "type": "mongodb",
  "filter": {
    "$and": [
      {"productCode": "13w42"},
      {"rodLength": {"$gte": 50, "$lte": 500}},
      {"region": "ca"}
    ]
  }
}
```

**Response for search engines (elasticsearch):**
```json
{
  "type": "elasticsearch",
  "query": {
    "bool": {
      "must": [
        {"term": {"productCode": "13w42"}},
        {"range": {"rodLength": {"gte": 50, "lte": 500}}},
        {"term": {"region": "ca"}}
      ]
    }
  }
}
```

#### 3. Get Schema

```
GET /api/v1/schemas/{name}
```

**Response: 200 OK**
```json
{
  "name": "products",
  "fields": {...},
  "options": {...},
  "registered": "2024-11-24T18:30:00Z"
}
```

#### 4. Delete Schema

```
DELETE /api/v1/schemas/{name}
```

**Response: 204 No Content**

#### 5. Health Check

```
GET /health
```

**Response: 200 OK**
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

#### 6. Metrics

```
GET /metrics
```

Returns Prometheus-formatted metrics (opt-in, no Prometheus server required to run rsearch).

---

## Query Syntax Support

rsearch supports the full OpenSearch/Elasticsearch query string syntax:

### Basic Terms

```
widget                     # single term
"blue widget"              # exact phrase
```

### Field-Specific Queries

```
name:widget                # field contains term
name:"blue widget"         # field contains exact phrase
name:wid*                  # wildcard
name:wi?get                # single char wildcard
name:/wi[dg]get/           # regex
```

### Boolean Operators

```
blue AND widget            # both required
blue OR widget             # either one
blue NOT red               # exclude
+blue -red                 # shorthand: must have blue, must not have red
blue && widget             # alternative AND
blue || widget             # alternative OR
!red                       # alternative NOT
```

### Grouping

```
(blue OR red) AND widget
name:(blue OR red)         # field grouping
```

### Ranges

```
rodlength:[50 TO 500]      # inclusive (50 and 500 included)
rodlength:{50 TO 500}      # exclusive (50 and 500 excluded)
rodlength:[50 TO 500}      # mixed: 50 included, 500 excluded
rodlength:>=50
rodlength:>50
rodlength:<=500
rodlength:<500
date:[2023-01-01 TO 2024-01-01]
```

### Boosting (relevance weighting)

```
widget^2                   # boost term 2x
name:widget^4              # boost field match 4x
"blue widget"^2            # boost phrase
```

### Fuzzy Search (typo tolerance)

```
widgit~                    # auto fuzzy (edit distance 2)
widgit~1                   # max 1 edit distance
```

### Proximity Search

```
"blue widget"~5            # words within 5 positions of each other
```

### Existence

```
_exists_:name              # field exists
NOT _exists_:name          # field missing
```

### Special Characters

Escape with backslash: `\+ \- \= \&& \|| \> \< \! \( \) \{ \} \[ \] \^ \" \~ \* \? \: \\ \/`

### Syntax Notes

- **Case sensitive operators**: AND/OR/NOT/TO must be uppercase
- **Default operator**: Without AND/OR, terms are OR by default
- **Configurable strictness**: Can enable case-insensitive operators via schema options
- **Field name case-insensitivity**: Configurable via `strictFieldNames` option

---

## Core Components

### Package Structure

```
rsearch/
├── cmd/
│   ├── rsearch/
│   │   └── main.go                 # Entry point, server setup
│   └── gendocs/
│       └── main.go                 # Documentation generator
├── internal/
│   ├── api/
│   │   ├── handlers.go            # HTTP handlers
│   │   ├── middleware.go          # Logging, recovery, CORS, rate limit
│   │   ├── routes.go              # Route definitions
│   │   └── response.go            # Standard response formats
│   ├── parser/
│   │   ├── lexer.go               # Tokenization
│   │   ├── parser.go              # Recursive descent parser
│   │   ├── ast.go                 # AST node definitions
│   │   ├── errors.go              # Parse error types
│   │   └── escape.go              # Escape sequence handling
│   ├── translator/
│   │   ├── translator.go          # Interface + registry
│   │   ├── postgres.go            # PostgreSQL translator
│   │   ├── mysql.go               # MySQL translator
│   │   ├── mongodb.go             # MongoDB translator
│   │   └── output.go              # TranslatorOutput types
│   ├── schema/
│   │   ├── schema.go              # Schema types
│   │   ├── registry.go            # In-memory storage + RWMutex
│   │   ├── validator.go           # Field validation
│   │   └── naming.go              # Naming convention conversions
│   ├── config/
│   │   └── config.go              # Config loading (viper)
│   ├── observability/
│   │   ├── logger.go              # Structured logging setup
│   │   └── metrics.go             # Prometheus metrics
│   └── cache/
│       └── cache.go                # Query parsing cache
├── pkg/
│   └── rsearch/
│       └── types.go                # Public types
├── docs/
│   └── syntax.md                  # Auto-generated from tests
├── examples/
│   └── queries.json               # Example queries for docs
└── tests/
    ├── parser_test.go             # Comprehensive syntax tests
    ├── translator_test.go         # Translation tests
    └── integration_test.go        # End-to-end API tests
```

### Key Components

#### 1. Lexer

Tokenizes input string into tokens:
- `FIELD`, `COLON`, `AND`, `OR`, `NOT`, `STRING`, `NUMBER`, `LPAREN`, `RPAREN`, etc.
- Handles escape sequences
- Rejects invalid characters (e.g., `;` for SQL injection prevention)
- Tracks position for error messages

#### 2. Parser

Builds AST from tokens using recursive descent with operator precedence:
- **Precedence** (highest to lowest): Boost (^) → Field queries (:) → Required/Prohibited (+/-) → NOT (!, NOT) → AND (&&, AND) → OR (||, OR)
- **Default operator**: Implicit OR between terms
- **Error recovery**: Collect multiple parse errors with line/column positions
- **Grouping**: Parentheses override precedence

#### 3. AST (Abstract Syntax Tree)

Tree structure representing query logic with node types:

**Expression Nodes:**
- `BinaryOp` - AND, OR
- `UnaryOp` - NOT
- `RequiredQuery` - +term
- `ProhibitedQuery` - -term
- `FieldQuery` - field:value
- `FieldGroupQuery` - field:(value1 OR value2)
- `RangeQuery` - field:[start TO end]
- `FuzzyQuery` - field~term
- `ProximityQuery` - "term1 term2"~N
- `ExistsQuery` - _exists_:field
- `BoostQuery` - query^N

**Value Nodes:**
- `TermValue` - simple term
- `PhraseValue` - "quoted phrase"
- `WildcardValue` - wild*card or wild?card
- `RegexValue` - /regex pattern/

#### 4. Schema Registry

Thread-safe in-memory storage:
- `map[string]*Schema` with `sync.RWMutex`
- Field resolution with case-insensitivity (configurable)
- Naming convention transformations (camelCase → snake_case)
- Field alias support
- Pre-compute field mappings at registration time

#### 5. Translator

Database-specific AST → Query converters:

```go
type Translator interface {
    Translate(ast *QueryAST, schema *Schema) (*TranslatorOutput, error)
    DatabaseType() string
}

type TranslatorOutput struct {
    Type       string          // "sql", "mongodb", "elasticsearch", etc.

    // SQL-specific fields
    WhereClause   string
    Parameters    []interface{}
    ParameterTypes []string

    // NoSQL-specific fields
    Filter     interface{}  // MongoDB filter, ES query DSL, etc.
}
```

**Translator Registry:**
- Interface-based pattern
- Register translators by database type
- Easy to add new databases

**PostgreSQL Translator:**
- Always use parameterized queries ($1, $2, etc.)
- Type-safe parameter binding
- SQL injection prevention
- Operator mapping: wildcards → LIKE, ranges → BETWEEN, etc.
- Optional features: fuzzy (requires pg_trgm), proximity (requires full-text search)

#### 6. API Layer

- HTTP handlers with comprehensive validation
- Error responses with clear messages
- Request/response logging
- CORS support
- Rate limiting
- Graceful shutdown

---

## Schema Design

### Schema Definition

```go
type Schema struct {
    Name      string
    Fields    map[string]Field
    Options   SchemaOptions
    CreatedAt time.Time
}

type Field struct {
    Type          FieldType
    Column        string    // Optional: explicit column name override
    Indexed       bool      // Hint for certain translators
    Aliases       []string  // Alternative field names
}

type FieldType string

const (
    TypeText     FieldType = "text"      // VARCHAR, TEXT, string
    TypeInteger  FieldType = "integer"   // INT, BIGINT
    TypeFloat    FieldType = "float"     // FLOAT, DOUBLE, DECIMAL
    TypeBoolean  FieldType = "boolean"   // BOOL
    TypeDateTime FieldType = "datetime"  // TIMESTAMP, DATE
    TypeDate     FieldType = "date"      // DATE only
    TypeTime     FieldType = "time"      // TIME only
    TypeJSON     FieldType = "json"      // JSON/JSONB
    TypeArray    FieldType = "array"     // Array types
)

type SchemaOptions struct {
    NamingConvention string  // "snake_case", "camelCase", "PascalCase", "none"
    StrictOperators  bool    // case-sensitive AND/OR/NOT (default: false)
    StrictFieldNames bool    // case-sensitive field names (default: false)
    DefaultField     string  // field for queries without field specifier
}
```

### Field Name Resolution

Multi-stage resolution process:

1. **Exact match** - Try exact field name match first
2. **Case-insensitive match** - If `strictFieldNames: false`, try case-insensitive
3. **Alias lookup** - Check field aliases
4. **Transform and map** - Apply naming convention transformation

**Example:**

```json
{
  "fields": {
    "productCode": {"type": "text"},
    "SKU": {"type": "text", "column": "sku_number"}
  },
  "options": {
    "namingConvention": "snake_case",
    "strictFieldNames": false
  }
}
```

Query field `PRODUCTCODE` resolves to:
- Case-insensitive match → `productCode`
- Transform via snake_case → column `product_code`

Query field `sku` resolves to:
- Case-insensitive match → `SKU`
- Explicit column override → column `sku_number`

---

## SQL Translation Strategy

### PostgreSQL Translation Examples

**Simple field match:**
```
Query: productCode:13w42
SQL:   product_code = $1
Params: ["13w42"]
```

**Range query:**
```
Query: rodLength:[50 TO 500]
SQL:   rod_length BETWEEN $1 AND $2
Params: [50, 500]
```

**Wildcard:**
```
Query: name:widget*
SQL:   name LIKE $1
Params: ["widget%"]

Query: name:wi?get
SQL:   name LIKE $1
Params: ["wi_get"]
```

**Regex:**
```
Query: name:/wi[dg]get/
SQL:   name ~ $1
Params: ["wi[dg]get"]
```

**Fuzzy search (requires pg_trgm extension):**
```
Query: description:~fuzzy
SQL:   levenshtein(description, $1) <= $2
Params: ["fuzzy", 2]
```

**Proximity search (requires full-text search):**
```
Query: title:"exact phrase"~5
SQL:   to_tsvector('english', title) @@ to_tsquery('english', $1 <5> $2')
Params: ["exact", "phrase"]
```

**Complex boolean:**
```
Query: productCode:13w42 AND (region:ca OR region:ny)
SQL:   product_code = $1 AND (region = $2 OR region = $3)
Params: ["13w42", "ca", "ny"]
```

### Security & Validation

**SQL Injection Prevention:**

1. **Lexer-level rejection** - Semicolons and SQL comments rejected during tokenization
2. **Parameterized queries only** - Never string concatenation
3. **Field name whitelisting** - Only schema-registered field names allowed
4. **Type validation** - Value types must match field types

**Feature Availability:**

Certain features require database extensions:
- **Fuzzy search**: Requires PostgreSQL `pg_trgm` extension
- **Proximity search**: Requires PostgreSQL full-text search

**Handling unavailable features:**

```json
{
  "options": {
    "enabledFeatures": {
      "fuzzy": false,
      "proximity": false,
      "regex": true
    }
  }
}
```

When disabled, translator returns clear error:
```
"Fuzzy search requires pg_trgm extension. Enable in schema or use wildcards instead: name:fuz*"
```

### Default Field Behavior

When query has no field specifier (`cat dog europe`), use `defaultField`:

```
Schema: defaultField: "name"
Query: cat dog europe
Result: name LIKE '%cat%' OR name LIKE '%dog%' OR name LIKE '%europe%'
```

Only searches ONE field (the default), avoiding performance issues. If no default field configured, return error requiring field names.

---

## Configuration System

### Configuration File (config.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  readTimeout: 30s
  writeTimeout: 30s
  shutdownTimeout: 10s

logging:
  level: "info"  # debug, info, warn, error
  format: "json" # json or console
  output: "stdout"

metrics:
  enabled: true
  port: 9090
  path: "/metrics"

cors:
  enabled: true
  allowedOrigins:
    - "http://localhost:3000"
    - "https://app.example.com"
  allowedMethods:
    - "GET"
    - "POST"
    - "DELETE"

schemas:
  loadFromFiles: true
  directory: "./schemas"

limits:
  maxQueryLength: 10000        # 0 = unlimited (not recommended)
  maxParameterCount: 100
  maxParseDepth: 50
  maxSchemaFields: 1000
  maxFieldNameLength: 255
  maxSchemas: 100
  maxRequestBodySize: 1048576  # bytes
  requestTimeout: 30s
  rateLimit:
    enabled: false
    requestsPerMinute: 100
    requestsPerHour: 5000
    burst: 10

cache:
  enabled: true
  maxSize: 10000
  ttl: 3600  # seconds, 0 = no expiry

security:
  allowedSpecialChars: ".-_"
  blockSqlKeywords: true
  auth:
    enabled: false
    type: "apikey"
    apiKeys: []

features:
  querySuggestions: false  # future feature
  maxQueryLength: 1000
  requestIdHeader: "X-Request-ID"

api:
  versions:
    v1:
      enabled: true
      deprecated: false
```

### Environment Variable Overrides

```bash
RSEARCH_SERVER_PORT=8080
RSEARCH_LOG_LEVEL=debug
RSEARCH_METRICS_ENABLED=true
RSEARCH_LIMITS_MAXQUERYLENGTH=50000
```

### Configuration Priority

1. Command line flags: `--max-query-length=50000`
2. Environment variables: `RSEARCH_LIMITS_MAXQUERYLENGTH=50000`
3. Config file: `config.yaml`
4. Default config: `config.default.yaml` (embedded in binary)

### Schema Files (optional pre-registration)

```
schemas/
├── products.json
├── users.json
└── orders.json
```

Loaded at startup if `schemas.loadFromFiles: true`

---

## Testing Strategy

### Test Structure

```
tests/
├── parser_test.go              # Comprehensive syntax tests
├── translator_postgres_test.go # PostgreSQL translation tests
├── translator_mysql_test.go    # MySQL tests
├── schema_test.go              # Schema validation tests
├── integration_test.go         # End-to-end API tests
└── examples/
    └── test_cases.json         # Master test case file
```

### Test-Driven Documentation

```go
type TestCase struct {
    Category    string   `json:"category"`    // "Basic", "Boolean", "Ranges", etc.
    Description string   `json:"description"` // Human-readable
    Query       string   `json:"query"`       // Input query
    Expected    Expected `json:"expected"`    // Expected output
    Schema      string   `json:"schema"`      // Schema name
}

type Expected struct {
    SQL        string        `json:"sql"`
    Parameters []interface{} `json:"parameters"`
    MongoDB    interface{}   `json:"mongodb,omitempty"`
}
```

**Example test cases:**

```json
[
  {
    "category": "Field Queries",
    "description": "Simple field match",
    "query": "productCode:13w42",
    "expected": {
      "sql": "product_code = $1",
      "parameters": ["13w42"]
    }
  },
  {
    "category": "Ranges",
    "description": "Inclusive range query",
    "query": "rodLength:[50 TO 500]",
    "expected": {
      "sql": "rod_length BETWEEN $1 AND $2",
      "parameters": [50, 500]
    }
  }
]
```

### Documentation Generator

```bash
# Generate docs from test cases
go run cmd/gendocs/main.go

# Output:
# - docs/syntax.md (complete syntax reference)
# - docs/postgres.md (PostgreSQL-specific examples)
# - docs/mongodb.md (MongoDB-specific examples)
```

Generated documentation format:

```markdown
# rsearch Query Syntax Reference

*Auto-generated from test suite - Last updated: 2024-11-24*

## Field Queries

### Simple field match
**Query:** `productCode:13w42`
**PostgreSQL:** `product_code = $1` with parameters: `["13w42"]`

### Wildcard search
**Query:** `name:widget*`
**PostgreSQL:** `name LIKE $1` with parameters: `["widget%"]`
```

### Test Coverage Requirements

- Every syntax feature has ≥3 test cases
- Edge cases: empty values, special chars, nested groups
- Error cases: invalid syntax, type mismatches
- All tests must pass before docs are generated

### CI/CD Integration

```bash
make test           # Run all tests
make generate-docs  # Regenerate docs from tests
```

Documentation is always accurate because it's generated from passing tests.

---

## Error Handling

### Error Response Format

```json
{
  "error": {
    "code": "PARSE_ERROR",
    "message": "Invalid query syntax",
    "details": [
      {
        "position": 15,
        "line": 1,
        "column": 15,
        "message": "Expected field name before ':'"
      }
    ],
    "query": "productCode:13w42 AND :value",
    "suggestion": "Check syntax near position 15"
  }
}
```

### Error Types

```
PARSE_ERROR        - Invalid query syntax
SCHEMA_NOT_FOUND   - Schema doesn't exist
FIELD_NOT_FOUND    - Field not in schema
TYPE_MISMATCH      - Value doesn't match field type
FEATURE_DISABLED   - Feature not enabled (fuzzy, etc.)
INVALID_RANGE      - Start > End in range
UNSUPPORTED_SYNTAX - Valid syntax, unsupported by DB
SCHEMA_EXISTS      - Schema already registered
INVALID_SCHEMA     - Invalid schema definition
```

### Validation Layers

1. **Lexer**: Invalid characters, malformed tokens
2. **Parser**: Grammar violations, unbalanced parentheses
3. **Schema**: Field existence, type checking
4. **Translator**: Database compatibility, feature availability

### User-Friendly Error Messages

**Bad:**
```
syntax error at token 15
```

**Good:**
```
Missing field name before ':' at position 15. Example: productCode:value
```

**Bad:**
```
field not found
```

**Good:**
```
Field 'productcode' not found. Did you mean 'productCode'? Available fields: productCode, region, price
```

**Bad:**
```
type error
```

**Good:**
```
Field 'price' expects number but got 'abc'. Example: price:100 or price:[50 TO 100]
```

---

## Performance & Security

### Performance Optimizations

**1. Query Parsing Cache**

```go
type ParserCache struct {
    cache *lru.Cache
    mu    sync.RWMutex
}
```

Cache frequently used queries to avoid re-parsing.

**2. Schema Registry**

- In-memory storage with RWMutex (read-optimized)
- No disk I/O on query translation
- Pre-compute field name mappings at registration

**3. Benchmarking Targets**

- Simple query parsing: <1ms
- Complex query (10+ operators): <5ms
- Schema lookup: <0.1ms
- End-to-end API request: <10ms

### Security Measures

**1. Input Validation**

All limits configurable (see Configuration section). Defaults:
- Max query length: 10,000 characters
- Max schema fields: 1,000
- Max parameter count: 100

**2. SQL Injection Prevention**

- Parameterized queries only (never string concat)
- Lexer rejects invalid characters (`;`, `--`, `/*`)
- No dynamic SQL generation from user input
- Field names validated against schema (whitelist)

**3. Rate Limiting**

Per-IP rate limiting (configurable, disabled by default):
```yaml
rateLimit:
  enabled: true
  requestsPerMinute: 100
  requestsPerHour: 5000
```

**4. Authentication (optional)**

```yaml
auth:
  enabled: false
  type: "apikey"  # apikey, jwt, basic
  apiKeys:
    - "key123"
```

**5. CORS Configuration**

Configurable allowed origins, restrict in production.

**6. Request Size Limits**

Configurable max request body size (default: 1MB).

**7. Secure Defaults**

- Bind to localhost by default
- Require explicit configuration for public binding
- No default API keys
- All errors sanitized (no stack traces in production)

### Monitoring Metrics

```
# Performance metrics
rsearch_parse_duration_seconds{quantile="0.5"}
rsearch_parse_duration_seconds{quantile="0.99"}
rsearch_translate_duration_seconds{database="postgres"}

# Error metrics
rsearch_errors_total{type="parse_error"}
rsearch_errors_total{type="schema_not_found"}

# Usage metrics
rsearch_requests_total{endpoint="/translate",status="200"}
rsearch_active_schemas
rsearch_cache_hit_ratio
```

---

## Deployment

### Deployment Options

**1. Standalone Binary**

```bash
wget https://github.com/yourusername/rsearch/releases/latest/rsearch
chmod +x rsearch
./rsearch --config config.yaml
```

**2. Docker**

```dockerfile
FROM alpine:latest
COPY rsearch /usr/local/bin/
COPY config.yaml /etc/rsearch/
EXPOSE 8080 9090
ENTRYPOINT ["rsearch", "--config", "/etc/rsearch/config.yaml"]
```

```bash
docker run -p 8080:8080 -v ./schemas:/schemas rsearch:latest
```

**3. Docker Compose**

```yaml
version: '3.8'
services:
  rsearch:
    image: rsearch:latest
    ports:
      - "8080:8080"
      - "9090:9090"
    volumes:
      - ./config.yaml:/etc/rsearch/config.yaml
      - ./schemas:/schemas
    environment:
      - RSEARCH_LOG_LEVEL=info
```

**4. Kubernetes**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rsearch
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: rsearch
        image: rsearch:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
```

**5. Systemd Service**

```ini
[Unit]
Description=rsearch query translation service
After=network.target

[Service]
Type=simple
User=rsearch
ExecStart=/usr/local/bin/rsearch --config /etc/rsearch/config.yaml
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

---

## Local Development Setup

### Docker Compose for Testing

```yaml
# docker-compose.dev.yaml
version: '3.8'

services:
  rsearch:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"
    volumes:
      - ./config.dev.yaml:/etc/rsearch/config.yaml
      - ./schemas:/schemas
    environment:
      - RSEARCH_LOG_LEVEL=debug
    depends_on:
      - postgres
      - mysql
      - mongodb

  postgres:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_DB=rsearch_test
      - POSTGRES_USER=rsearch
      - POSTGRES_PASSWORD=rsearch123
    volumes:
      - ./testdata/postgres/init.sql:/docker-entrypoint-initdb.d/init.sql

  mysql:
    image: mysql:8.0
    ports:
      - "3306:3306"
    environment:
      - MYSQL_DATABASE=rsearch_test
      - MYSQL_USER=rsearch
      - MYSQL_PASSWORD=rsearch123
      - MYSQL_ROOT_PASSWORD=root123
    volumes:
      - ./testdata/mysql/init.sql:/docker-entrypoint-initdb.d/init.sql

  mongodb:
    image: mongo:7
    ports:
      - "27017:27017"
    environment:
      - MONGO_INITDB_DATABASE=rsearch_test
    volumes:
      - ./testdata/mongodb/init.js:/docker-entrypoint-initdb.d/init.js

  pgadmin:
    image: dpage/pgadmin4:latest
    ports:
      - "5050:80"
    environment:
      - PGADMIN_DEFAULT_EMAIL=admin@rsearch.local
      - PGADMIN_DEFAULT_PASSWORD=admin

  mongo-express:
    image: mongo-express:latest
    ports:
      - "8081:8081"
    environment:
      - ME_CONFIG_MONGODB_URL=mongodb://mongodb:27017/
```

### Sample Database Init

```sql
-- testdata/postgres/init.sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    product_code VARCHAR(50),
    name VARCHAR(255),
    description TEXT,
    rod_length INTEGER,
    price DECIMAL(10,2),
    region VARCHAR(50),
    in_stock BOOLEAN,
    tags TEXT[],
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO products (product_code, name, description, rod_length, price, region, in_stock, tags) VALUES
('13w42', 'Widget Pro', 'High quality widget', 150, 99.99, 'ca', true, ARRAY['premium', 'bestseller']),
('13w43', 'Widget Basic', 'Standard widget', 200, 49.99, 'ny', true, ARRAY['basic']),
('13w44', 'Widget Premium', 'Premium quality widget', 450, 199.99, 'ca', false, ARRAY['premium', 'new']),
('15x20', 'Gadget One', 'Multi-purpose gadget', 75, 79.99, 'cb', true, ARRAY['gadget']),
('15x21', 'Gadget Two', 'Advanced gadget', 125, 129.99, 'ca', true, ARRAY['gadget', 'advanced']);

CREATE INDEX idx_product_code ON products(product_code);
CREATE INDEX idx_region ON products(region);
CREATE INDEX idx_rod_length ON products(rod_length);
```

### Quick Start Script

```bash
#!/bin/bash
# scripts/dev-setup.sh

echo "Starting rsearch development environment..."

docker-compose -f docker-compose.dev.yaml up -d postgres mysql mongodb

echo "Waiting for databases to be ready..."
sleep 10

echo "Building rsearch..."
go build -o bin/rsearch cmd/rsearch/main.go

echo "Starting rsearch..."
./bin/rsearch --config config.dev.yaml &

sleep 3

echo "Registering test schema..."
curl -X POST http://localhost:8080/api/v1/schemas \
  -H "Content-Type: application/json" \
  -d @examples/schemas/products.json

echo "Development environment ready!"
echo ""
echo "Services:"
echo "  rsearch:      http://localhost:8080"
echo "  metrics:      http://localhost:9090/metrics"
echo "  postgres:     localhost:5432"
echo "  mysql:        localhost:3306"
echo "  mongodb:      localhost:27017"
echo "  pgAdmin:      http://localhost:5050"
echo "  mongo-express: http://localhost:8081"
```

### Integration Test with Real Database

```go
// tests/integration_db_test.go
func TestPostgresIntegration(t *testing.T) {
    db := connectPostgres()
    defer db.Close()

    resp := callRsearchAPI("productCode:13w42 AND region:ca")

    query := fmt.Sprintf("SELECT * FROM products WHERE %s", resp.WhereClause)
    rows, err := db.Query(query, resp.Parameters...)

    assert.NoError(t, err)
    assert.True(t, rows.Next())

    var product Product
    rows.Scan(&product.ID, &product.ProductCode, ...)
    assert.Equal(t, "13w42", product.ProductCode)
    assert.Equal(t, "ca", product.Region)
}
```

---

## Integration Examples

Applications integrate rsearch using standard HTTP clients. No libraries required.

### JavaScript/Node.js

```javascript
async function searchProducts(userQuery) {
  const response = await fetch('http://localhost:8080/api/v1/translate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      schema: 'products',
      database: 'postgres',
      query: userQuery
    })
  });

  const result = await response.json();

  const products = await db.query(
    `SELECT * FROM products WHERE ${result.whereClause}`,
    result.parameters
  );

  return products;
}
```

### Python

```python
import requests

def search_products(user_query):
    response = requests.post('http://localhost:8080/api/v1/translate', json={
        'schema': 'products',
        'database': 'postgres',
        'query': user_query
    })

    result = response.json()

    cur.execute(
        f"SELECT * FROM products WHERE {result['whereClause']}",
        result['parameters']
    )

    return cur.fetchall()
```

### PHP

```php
function searchProducts($userQuery) {
    $response = file_get_contents('http://localhost:8080/api/v1/translate', false,
        stream_context_create([
            'http' => [
                'method' => 'POST',
                'header' => 'Content-Type: application/json',
                'content' => json_encode([
                    'schema' => 'products',
                    'database' => 'postgres',
                    'query' => $userQuery
                ])
            ]
        ])
    );

    $result = json_decode($response, true);

    $stmt = $pdo->prepare("SELECT * FROM products WHERE {$result['whereClause']}");
    $stmt->execute($result['parameters']);

    return $stmt->fetchAll();
}
```

### Go

```go
func searchProducts(userQuery string) ([]Product, error) {
    reqBody, _ := json.Marshal(map[string]string{
        "schema":   "products",
        "database": "postgres",
        "query":    userQuery,
    })

    resp, _ := http.Post("http://localhost:8080/api/v1/translate",
        "application/json", bytes.NewBuffer(reqBody))

    var result struct {
        WhereClause string        `json:"whereClause"`
        Parameters  []interface{} `json:"parameters"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    query := fmt.Sprintf("SELECT * FROM products WHERE %s", result.WhereClause)
    rows, _ := db.Query(query, result.Parameters...)

    // Scan results...
}
```

---

## Versioning & Roadmap

### API Versioning

- API versions in URL path (`/api/v1/`, `/api/v2/`)
- Backwards compatibility within major versions
- Multiple versions can run simultaneously
- Deprecation notices 6+ months before removal

### Semantic Versioning

- `v1.0.0` - Initial production release
- `v1.x.x` - Backwards compatible features, bug fixes
- `v2.0.0` - Breaking changes

### Implementation Roadmap

**Phase 1: Core Features (v1.0)**
- Full OpenSearch query syntax support
- PostgreSQL translator
- Schema registration
- Test-driven documentation
- Docker deployment
- Production-grade config/logging/metrics

**Phase 2: Additional Databases (v1.1-v1.3)**
- MySQL translator
- SQLite translator
- MongoDB translator
- Elasticsearch translator

**Phase 3: Advanced Features (v1.4-v2.0)**
- Query suggestions/auto-complete
- Query validation before execution
- Query cost estimation
- Query optimization hints
- Schema introspection (auto-generate from DB)
- Multi-tenancy support

**Phase 4: Enhanced DX (v2.x)**
- GraphQL API endpoint
- WebSocket streaming for large results
- Query builder UI (optional web interface)
- Query analytics dashboard

**Optional Features (community-driven):**
- Language-specific client libraries
- Framework integrations (Rails, Django, Express)
- ORM plugins (Sequelize, TypeORM, SQLAlchemy)
- Query caching strategies
- Distributed tracing (OpenTelemetry)

### Extensibility

```go
// Custom translator plugin
type CustomTranslator struct{}

func (t *CustomTranslator) Translate(ast *QueryAST) (*TranslatorOutput, error) {
    // Custom database translation logic
}

// Register at startup
translator.Register("customdb", &CustomTranslator{})
```

### Feature Flags

```yaml
features:
  querySuggestions: false
  costEstimation: false
  schemaIntrospection: false
```

New features added behind flags, enabled when stable.

---

## Implementation Order

**Milestone 1: Foundation (Week 1-2)**
1. Project setup (go mod init, directory structure)
2. Configuration system (viper, config loading)
3. Basic HTTP server (chi/gin router, middleware)
4. Health/metrics endpoints
5. Logging setup

**Milestone 2: Parser & AST (Week 2-3)**
1. Lexer implementation
2. Token types and tokenization
3. AST node definitions
4. Recursive descent parser
5. Parser tests (basic syntax)

**Milestone 3: Schema System (Week 3-4)**
1. Schema types and validation
2. In-memory registry with RWMutex
3. Field resolution (naming conventions, case-insensitivity)
4. Schema API endpoints (register, get, delete)
5. Schema tests

**Milestone 4: PostgreSQL Translator (Week 4-5)**
1. Translator interface
2. PostgreSQL translator implementation
3. SQL generation (parameterized queries)
4. Type mapping and validation
5. Translator tests

**Milestone 5: Complete Query Syntax (Week 5-7)**
1. All operator types (AND, OR, NOT, +, -, etc.)
2. Range queries (inclusive, exclusive, mixed)
3. Wildcards and regex
4. Fuzzy search (optional, with pg_trgm check)
5. Proximity search (optional, with FTS check)
6. Boost queries
7. Field grouping
8. Comprehensive parser tests (100+ test cases)

**Milestone 6: Documentation & Testing (Week 7-8)**
1. Test case JSON format
2. Documentation generator
3. Integration tests with real databases
4. Docker compose dev environment
5. Example integration code (multiple languages)

**Milestone 7: Production Readiness (Week 8-9)**
1. Error handling and validation
2. Rate limiting
3. Query caching
4. Security hardening
5. Performance optimization
6. Metrics and monitoring

**Milestone 8: Additional Databases (Week 9-10)**
1. MySQL translator
2. SQLite translator
3. MongoDB translator

**Milestone 9: Release Prep (Week 10-11)**
1. Docker images
2. Kubernetes manifests
3. Release automation
4. Complete documentation
5. Example applications
6. v1.0.0 release

---

## Summary

**rsearch** is a production-grade query translation service that bridges the gap between user-friendly search syntax and database queries. By providing a lightweight, standalone service with comprehensive syntax support and multi-database compatibility, rsearch enables advanced search capabilities without requiring full-text search infrastructure or complex client libraries.

**Key strengths:**
- Simple HTTP API, works with any language
- Full OpenSearch query syntax from day one
- Production-ready: observability, security, configurability
- Extensible: easy to add new databases and features
- Developer-friendly: auto-generated docs, clear errors, sensible defaults
- Zero external dependencies: single binary, runs anywhere

**Target use cases:**
- Applications needing advanced search without Elasticsearch
- Multi-tenant SaaS with user-facing search
- Internal tools with complex filtering requirements
- Database-agnostic search layers
- Migration from Elasticsearch to simpler solutions

The design prioritizes simplicity, extensibility, and production-readiness from the start, ensuring rsearch can grow from initial release to enterprise-scale deployments.
