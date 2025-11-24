# rsearch Test Environment

This directory contains test databases and sample data for development and integration testing.

## Quick Start

```bash
# Start all test databases
docker-compose -f docker-compose.dev.yaml up -d

# Wait for databases to be ready
sleep 10

# View logs
docker-compose -f docker-compose.dev.yaml logs -f

# Stop all services
docker-compose -f docker-compose.dev.yaml down
```

## Services

### PostgreSQL
- **Port:** 5432
- **Database:** rsearch_test
- **User:** rsearch
- **Password:** rsearch123
- **Init Script:** `postgres/init.sql`

### MySQL
- **Port:** 3306
- **Database:** rsearch_test
- **User:** rsearch
- **Password:** rsearch123
- **Root Password:** root123
- **Init Script:** `mysql/init.sql`

### MongoDB
- **Port:** 27017
- **Database:** rsearch_test
- **Init Script:** `mongodb/init.js`

### Management Tools

**pgAdmin (PostgreSQL):**
- URL: http://localhost:5050
- Email: admin@rsearch.local
- Password: admin

**Mongo Express (MongoDB):**
- URL: http://localhost:8081

## Sample Data

All databases are initialized with sample product data:
- 10 products with various fields
- Product codes, names, descriptions
- Numeric fields (rod_length, price)
- Text fields (region, status)
- Arrays and JSON (PostgreSQL, MongoDB)

## Connection Examples

### PostgreSQL (psql)
```bash
psql -h localhost -U rsearch -d rsearch_test
```

### MySQL
```bash
mysql -h localhost -u rsearch -prsearch123 rsearch_test
```

### MongoDB
```bash
mongosh mongodb://localhost:27017/rsearch_test
```

## Testing rsearch Queries

Once databases are running, test rsearch query translation:

```bash
# Build rsearch
go build -o bin/rsearch cmd/rsearch/main.go

# Run rsearch
./bin/rsearch &

# Register schema
curl -X POST http://localhost:8080/api/v1/schemas \
  -H "Content-Type: application/json" \
  -d @tests/schemas.json

# Test query translation
curl -X POST http://localhost:8080/api/v1/translate \
  -H "Content-Type: application/json" \
  -d '{
    "schema": "products",
    "database": "postgres",
    "query": "productCode:13w42 AND rodLength:[50 TO 500]"
  }'
```

## Integration Testing

Run integration tests against real databases:

```bash
# Set environment variables
export POSTGRES_DSN="postgres://rsearch:rsearch123@localhost:5432/rsearch_test"
export MYSQL_DSN="rsearch:rsearch123@tcp(localhost:3306)/rsearch_test"
export MONGODB_URI="mongodb://localhost:27017/rsearch_test"

# Run integration tests
go test ./tests/integration_test.go -v
```

## Cleanup

Remove all data and volumes:

```bash
docker-compose -f docker-compose.dev.yaml down -v
```
