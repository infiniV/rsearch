# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

rsearch is a production-grade query translation service that converts OpenSearch/Elasticsearch-style query strings into database-specific formats (PostgreSQL SQL with parameterized queries). It acts as a lightweight bridge between application search UIs and databases.

**Core flow:** Query string -> Parser -> AST -> Translator -> WHERE clause + parameters

## Build and Test Commands

```bash
# Build
go build -o bin/rsearch cmd/rsearch/main.go
make build

# Run all tests
go test ./...
make test

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/parser/...
go test ./internal/translator/...
go test ./internal/schema/...

# Run a single test
go test -run TestParsePhraseQuery ./internal/parser/...

# Start demo environment (Docker + rsearch + opens browser)
make demo

# Start/stop services
make start
make stop
```

## Architecture

### Package Structure

```
cmd/rsearch/          Entry point, server setup
internal/
  parser/             Lexer, recursive descent parser, AST nodes
  translator/         Translator interface, PostgreSQL implementation
  schema/             Schema registry with RWMutex, field resolution
  api/                HTTP handlers (chi router), middleware
  config/             Configuration loading (viper)
  observability/      Structured logging (zerolog), Prometheus metrics
pkg/rsearch/          Public types
```

### Data Flow

1. **Parser** (`internal/parser/`): Tokenizes query string via Lexer, builds AST using recursive descent
2. **Schema** (`internal/schema/`): Resolves field names (camelCase -> snake_case), validates against registered schema
3. **Translator** (`internal/translator/`): Converts AST to SQL with parameterized queries ($1, $2...)

### Key Types

**AST Nodes** (`internal/parser/ast.go`):
- `BinaryOp`, `UnaryOp` - boolean operators (AND, OR, NOT)
- `FieldQuery`, `FieldGroupQuery` - field:value and field:(a OR b)
- `RangeQuery` - [start TO end], {start TO end}, >=, <
- `FuzzyQuery`, `ProximityQuery`, `BoostQuery`, `ExistsQuery`
- `TermValue`, `PhraseValue`, `WildcardValue`, `RegexValue`, `NumberValue`

**Translator Interface** (`internal/translator/translator.go`):
```go
type Translator interface {
    Translate(ast parser.Node, schema *schema.Schema) (*TranslatorOutput, error)
    DatabaseType() string
}
```

**Schema** (`internal/schema/schema.go`):
- Field types: text, integer, float, boolean, datetime, date, time, json, array
- Options: namingConvention, strictOperators, strictFieldNames, defaultField

### Parser Operator Precedence (lowest to highest)

OR -> AND -> NOT -> Required/Prohibited (+/-) -> Field (:) -> Boost (^)

### Supported Query Syntax

- Field queries: `field:value`, `field:"phrase"`
- Boolean: `AND`, `OR`, `NOT`, `+`, `-`, `&&`, `||`, `!`
- Ranges: `[50 TO 500]`, `{50 TO 500}`, `>=50`, `<100`
- Wildcards: `widget*`, `wi?get`
- Regex: `/pattern/`
- Boost: `field:value^2`
- Exists: `_exists_:field`
- Grouping: `(a OR b) AND c`, `field:(a OR b)`

## API Endpoints

- `POST /api/v1/translate` - Translate query string
- `POST /api/v1/schemas` - Register schema
- `GET /api/v1/schemas/{name}` - Get schema
- `DELETE /api/v1/schemas/{name}` - Delete schema
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics (when enabled)

## Configuration

Environment variables prefixed with `RSEARCH_`:
```bash
RSEARCH_SERVER_PORT=8080
RSEARCH_LOGGING_LEVEL=debug
RSEARCH_LOGGING_FORMAT=console  # or json
RSEARCH_METRICS_ENABLED=true
```

## Development Notes

- Operators are normalized to keyword form (&&->AND, ||->OR, !->NOT) during parsing
- All SQL uses parameterized queries for security ($1, $2...)
- Schema field names support case-insensitive matching and aliasing
- Naming conventions transform field names (productCode -> product_code)
- Git worktrees in `.worktrees/` are used for parallel feature development
