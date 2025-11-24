# rsearch Examples & Demo

This directory contains integration examples and an interactive demo for rsearch.

## Quick Start

The easiest way to try rsearch is using the interactive demo:

```bash
make demo
```

This command will:
1. Build the rsearch binary
2. Start PostgreSQL, MySQL, and MongoDB via Docker Compose
3. Start the rsearch server with CORS enabled
4. Register the product schema
5. Open the demo page in your browser

## Demo Web Page

The `demo.html` file provides an interactive web interface to test rsearch query translation:

**Features:**
- Try pre-configured example queries with one click
- Write custom OpenSearch-style queries
- See real-time SQL translation
- View client-side performance metrics
- Learn query syntax by example

**Example Queries:**
- Simple: `productCode:13w42`
- Boolean: `region:ca AND price:<100`
- Range: `rodLength:[50 TO 200]`
- Wildcard: `name:Widget*`
- Complex: `(region:ca OR region:ny) AND price:>=100`

## Manual Setup

If you prefer to start components individually:

### 1. Start Databases

```bash
docker-compose -f docker-compose.dev.yaml up -d
```

Wait 10-15 seconds for databases to initialize.

### 2. Build and Start rsearch

```bash
# Build
go build -o bin/rsearch cmd/rsearch/main.go

# Start with default config
./bin/rsearch

# Or start with CORS enabled for demo
./bin/rsearch --config .demo-config.yaml
```

### 3. Register Schema

```bash
curl -X POST http://localhost:8080/api/v1/schemas \
  -H "Content-Type: application/json" \
  -d @examples/product_schema.json
```

### 4. Open Demo Page

Open `examples/demo.html` in your browser or use:

```bash
# Linux
xdg-open examples/demo.html

# macOS
open examples/demo.html

# WSL
wslview examples/demo.html
```

## Helper Scripts

### start-demo.sh

Starts the complete demo environment:
- Docker Compose services (PostgreSQL, MySQL, MongoDB)
- rsearch server with CORS enabled
- Registers the product schema

```bash
./examples/start-demo.sh
```

### stop-demo.sh

Stops all demo services:
- rsearch server
- Docker Compose services
- Cleans up temporary files

```bash
./examples/stop-demo.sh
```

## Integration Examples

This directory includes integration examples in multiple languages:

### Node.js (examples/nodejs/search.js)

```bash
cd examples/nodejs
npm install
node search.js
```

**Features:**
- PostgreSQL integration with pg client
- Example queries with result display
- Error handling

### Python (examples/python/search.py)

```bash
cd examples/python
pip install requests psycopg2-binary
python search.py
```

**Features:**
- psycopg2 for PostgreSQL
- Type hints and proper error handling
- Clean output formatting

### Go (examples/go/search.go)

```bash
cd examples/go
go mod init example
go get github.com/lib/pq
go run search.go
```

**Features:**
- Native database/sql package
- Struct-based result parsing
- Idiomatic Go error handling

### PHP (examples/php/search.php)

```bash
cd examples/php
php search.php
```

**Features:**
- PDO for database access
- Prepared statements
- Exception handling

## API Reference

### Health Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready
```

### Schema Management

```bash
# Register schema
curl -X POST http://localhost:8080/api/v1/schemas \
  -H "Content-Type: application/json" \
  -d @examples/product_schema.json

# List schemas
curl http://localhost:8080/api/v1/schemas

# Get specific schema
curl http://localhost:8080/api/v1/schemas/products

# Delete schema
curl -X DELETE http://localhost:8080/api/v1/schemas/products
```

### Query Translation

```bash
curl -X POST http://localhost:8080/api/v1/translate \
  -H "Content-Type: application/json" \
  -d '{
    "schema": "products",
    "database": "postgres",
    "query": "productCode:13w42 AND region:ca"
  }'
```

**Response:**
```json
{
  "type": "sql",
  "whereClause": "product_code = $1 AND region = $2",
  "parameters": ["13w42", "ca"],
  "parameterTypes": ["text", "text"]
}
```

## Database Access

The dev environment includes the following databases:

### PostgreSQL

```bash
psql -h localhost -U rsearch -d rsearch_test
# Password: rsearch123
```

**Web UI:** http://localhost:5050 (pgAdmin)
- Email: admin@rsearch.local
- Password: admin

### MySQL

```bash
mysql -h localhost -u rsearch -p rsearch_test
# Password: rsearch123
```

### MongoDB

```bash
mongosh mongodb://localhost:27017/rsearch_test
```

**Web UI:** http://localhost:8081 (Mongo Express)

## Sample Data

The PostgreSQL database is initialized with 10 sample products:

```sql
SELECT * FROM products;
```

**Fields:**
- product_code: '13w42', '13w43', etc.
- name: 'Widget Pro', 'Gadget One', etc.
- rod_length: 60-450
- price: 49.99-599.99
- region: 'ca', 'ny', 'cb'
- status: 'active', 'discontinued', 'preorder'

## Troubleshooting

### Docker services fail to start

```bash
# Check logs
docker-compose -f docker-compose.dev.yaml logs

# Restart services
docker-compose -f docker-compose.dev.yaml restart
```

### rsearch server won't start

```bash
# Check if port 8080 is already in use
lsof -i :8080

# Check server logs
cat .rsearch.log
```

### Schema registration fails

```bash
# Verify server is running
curl http://localhost:8080/health

# Check schema file syntax
cat examples/product_schema.json | jq .
```

### Demo page shows "Offline" status

1. Verify rsearch server is running on port 8080
2. Check CORS is enabled in config
3. Open browser console for error messages

## Makefile Commands

```bash
make demo   # Start demo and open browser
make build  # Build rsearch binary
make start  # Start services without browser
make stop   # Stop all services
make clean  # Stop and remove binary
make test   # Run tests
```

## Supported Query Syntax

rsearch supports OpenSearch/Elasticsearch query syntax:

**Field Queries:**
- `field:value` - Exact match
- `field:"quoted value"` - Phrase match

**Boolean Operators:**
- `AND`, `OR`, `NOT`
- `&&`, `||`, `!` (symbols)
- `+` (required), `-` (prohibited)

**Range Queries:**
- `field:[min TO max]` - Inclusive
- `field:{min TO max}` - Exclusive
- `field:[min TO max}` - Mixed
- `field:>=value`, `field:<value` - Comparison

**Wildcards:**
- `name:widget*` - Prefix
- `name:*widget` - Suffix
- `name:*widget*` - Contains

**Special Queries:**
- `_exists_:field` - Field exists check
- `field:value^2` - Boost query
- `/regex/` - Regular expression

**Grouping:**
- `(query1 OR query2) AND query3` - Nested boolean logic

## Next Steps

- Try the [interactive demo](demo.html)
- Explore [integration examples](#integration-examples)
- Review [API documentation](../docs/api.md)
- Read [query syntax reference](../docs/syntax-reference.md)
