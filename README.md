# rsearch - Phase 2.2: Schema System

This milestone implements the schema registry and field resolution system for rsearch.

## Features Implemented

### 1. Schema Types
- **Field Types**: text, integer, float, boolean, datetime, date, time, json, array
- **Field Properties**: type, column override, indexed hint, aliases
- **Schema Options**: naming conventions, strict operators, strict field names, default field

### 2. Naming Convention Transformations
- **snake_case**: camelCase/PascalCase → snake_case
- **camelCase**: snake_case → camelCase
- **PascalCase**: snake_case → PascalCase
- **none**: No transformation

### 3. Field Resolution
Multi-stage resolution with the following order:
1. Exact match
2. Case-insensitive match (if strictFieldNames: false)
3. Alias lookup
4. Transform via naming convention and match

### 4. Schema Registry
Thread-safe in-memory storage with:
- RWMutex for concurrent access (read-optimized)
- Pre-computed field mappings for fast lookups
- Validation before registration

### 5. API Endpoints
- `POST /api/v1/schemas` - Register a new schema
- `GET /api/v1/schemas` - List all schemas
- `GET /api/v1/schemas/{name}` - Get a specific schema
- `DELETE /api/v1/schemas/{name}` - Delete a schema

## Project Structure

```
.
├── cmd/
│   └── rsearch/
│       └── main.go              # HTTP server application
├── internal/
│   ├── api/
│   │   ├── handler.go           # API handlers
│   │   └── handler_test.go      # API handler tests
│   ├── schema/
│   │   ├── schema.go            # Schema types and field resolution
│   │   ├── schema_test.go       # Schema tests
│   │   ├── naming.go            # Naming convention transformations
│   │   ├── naming_test.go       # Naming tests
│   │   ├── validator.go         # Schema validation
│   │   ├── validator_test.go    # Validator tests
│   │   ├── registry.go          # Thread-safe schema registry
│   │   └── registry_test.go     # Registry tests
│   └── integration_test.go      # Integration tests
├── examples/
│   └── product_schema.json      # Example schema
└── go.mod                       # Go module definition
```

## Running Tests

### Run all tests with coverage:
```bash
go test ./... -v -cover
```

### Run tests with race detector:
```bash
go test ./... -race
```

### Run specific test suites:
```bash
go test ./internal/schema/... -v
go test ./internal/api/... -v
go test ./internal/... -v -run TestIntegration
```

## Test Results

All tests pass with:
- **API Coverage**: 65.1%
- **Schema Coverage**: 88.8%
- **Race Detector**: No race conditions detected
- **Total Tests**: 50+ test cases covering all functionality

## Building and Running

### Build the application:
```bash
go build -o rsearch ./cmd/rsearch
```

### Run the server:
```bash
./rsearch
# Server starts on port 8080 by default
# Set PORT environment variable to change: PORT=3000 ./rsearch
```

### Example Usage

#### 1. Register a schema:
```bash
curl -X POST http://localhost:8080/api/v1/schemas \
  -H "Content-Type: application/json" \
  -d @examples/product_schema.json
```

#### 2. Get a schema:
```bash
curl http://localhost:8080/api/v1/schemas/products
```

#### 3. List all schemas:
```bash
curl http://localhost:8080/api/v1/schemas
```

#### 4. Delete a schema:
```bash
curl -X DELETE http://localhost:8080/api/v1/schemas/products
```

## Field Resolution Examples

Given the example `product_schema.json`:

| Query Field | Resolved Column | Reason |
|------------|-----------------|--------|
| `productCode` | `product_code` | Exact match + snake_case convention |
| `PRODUCTCODE` | `product_code` | Case-insensitive match (strictFieldNames: false) |
| `code` | `product_code` | Alias resolution |
| `SKU` | `product_code` | Alias with case-insensitive match |
| `createdAt` | `created_timestamp` | Explicit column override |

## Key Design Decisions

1. **Thread-Safety**: RWMutex for read-optimized concurrent access
2. **Performance**: Pre-computed lookup caches built at registration time
3. **Flexibility**: Multi-stage field resolution with case-insensitive and alias support
4. **Validation**: Comprehensive schema validation before registration
5. **API Design**: RESTful endpoints following HTTP best practices

## Next Steps (Phase 2.3)

The next phase will implement the query parser to translate user queries into structured query objects using these schemas.

## Go Version

Developed with Go 1.25.4
