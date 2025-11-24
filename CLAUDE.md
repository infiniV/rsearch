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

# Run specific package tests
go test ./internal/parser/...
go test ./internal/translator/...
go test ./internal/schema/...

# Run a single test
go test -run TestParsePhraseQuery ./internal/parser/...

# Regenerate syntax documentation from test cases
go run cmd/gendocs/main.go

# Start demo environment (Docker + rsearch + opens browser)
make demo

# Start/stop services
make start
make stop
```

## Architecture

### Package Structure

```
cmd/
  rsearch/              Entry point, server setup
  gendocs/              Documentation generator from test cases
internal/
  parser/               Lexer, recursive descent parser, AST nodes
  translator/           Translator interface, PostgreSQL implementation
  schema/               Schema registry with RWMutex, field resolution
  api/                  HTTP handlers (chi router), middleware
  config/               Configuration loading (viper)
  observability/        Structured logging (zerolog), Prometheus metrics
pkg/rsearch/            Public types
tests/
  testcases.json        Test cases used for docs generation
  schemas.json          Test schemas
examples/
  demo.html             Interactive web demo UI
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
- `GroupQuery` - parenthesized expressions (a OR b)
- `RequiredQuery`, `ProhibitedQuery` - +term, -term
- `TermQuery`, `PhraseQuery`, `WildcardQuery` - standalone queries
- `FuzzyQuery`, `ProximityQuery`, `BoostQuery`, `ExistsQuery`
- Value nodes: `TermValue`, `PhraseValue`, `WildcardValue`, `RegexValue`, `NumberValue`

**Translator** (`internal/translator/postgres.go`):
Supported node translations:
- `FieldQuery` - handles term, phrase, wildcard, regex values
- `BinaryOp` - AND/OR with proper parentheses
- `UnaryOp` - NOT operator
- `RangeQuery` - BETWEEN, comparison operators, unbounded ranges
- `GroupQuery` - parenthesized SQL
- `RequiredQuery` - pass-through
- `ProhibitedQuery` - NOT prefix
- `ExistsQuery` - IS NOT NULL
- `BoostQuery` - SQL unchanged, boost stored in metadata
- `FuzzyQuery` - levenshtein() (requires pg_trgm)
- `ProximityQuery` - phraseto_tsquery() (requires FTS)
- `FieldGroupQuery` - field:(a OR b) expansion

**Schema** (`internal/schema/schema.go`):
- Field types: text, integer, float, boolean, datetime, date, time, json, array
- Options: namingConvention, strictOperators, strictFieldNames, defaultField
- EnabledFeatures: fuzzy (pg_trgm), proximity (FTS), regex

### Parser Operator Precedence (lowest to highest)

OR -> AND -> NOT -> Required/Prohibited (+/-) -> Field (:) -> Boost (^) -> Fuzzy (~)

### Supported Query Syntax

**Field queries:**
- `field:value` - exact match
- `field:"phrase"` - phrase match
- `field:wild*` - wildcard (LIKE)
- `field:/regex/` - regex match (~)

**Boolean operators:**
- `AND`, `&&` - conjunction
- `OR`, `||` - disjunction
- `NOT`, `!` - negation
- `+term` - required (pass-through)
- `-term` - prohibited (NOT)
- Implicit OR: `term1 term2` -> `term1 OR term2`

**Ranges:**
- `[50 TO 500]` - inclusive (BETWEEN)
- `{50 TO 500}` - exclusive (> AND <)
- `[50 TO 500}` - mixed
- `>=50`, `>50`, `<=100`, `<100` - comparison
- `[100 TO *]` - unbounded

**Advanced:**
- `field:term~2` - fuzzy search (levenshtein)
- `"phrase"~5` - proximity search (FTS)
- `field:value^2` - boost (metadata only for SQL)
- `_exists_:field` - existence check (IS NOT NULL)

**Grouping:**
- `(a OR b) AND c` - parentheses
- `field:(a OR b)` - field group

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
- Standalone terms/wildcards require `defaultField` in schema options
- Fuzzy search requires `enabledFeatures.fuzzy: true` and pg_trgm extension
- Proximity search requires `enabledFeatures.proximity: true` and FTS setup
- Documentation auto-generated from `tests/testcases.json` via `cmd/gendocs`
