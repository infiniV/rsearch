#!/bin/bash
# Start the complete rsearch demo environment

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "=========================================="
echo "Starting rsearch demo environment..."
echo "=========================================="
echo ""

# Start databases
echo "Starting Docker services..."
docker-compose -f docker-compose.dev.yaml up -d

echo "Waiting for databases to initialize (15 seconds)..."
sleep 15

# Build rsearch if not exists or source changed
if [ ! -f "bin/rsearch" ]; then
    echo "Building rsearch..."
    export PATH="/usr/local/go/bin:$PATH"
    export GOPATH="$HOME/go"
    go build -o bin/rsearch cmd/rsearch/main.go
fi

# Create demo config with CORS enabled
echo "Creating demo configuration..."
cat > .demo-config.yaml <<'EOF'
server:
  host: "0.0.0.0"
  port: 8080

cors:
  enabled: true
  allowedOrigins:
    - "*"
  allowedMethods:
    - "GET"
    - "POST"
    - "DELETE"
    - "OPTIONS"
  allowedHeaders:
    - "Content-Type"
    - "Authorization"

logging:
  level: "info"
  format: "console"
EOF

# Start rsearch server
echo "Starting rsearch server..."
./bin/rsearch --config .demo-config.yaml > .rsearch.log 2>&1 &
RSEARCH_PID=$!
echo $RSEARCH_PID > .rsearch.pid

sleep 3

# Check if server started successfully
if ! ps -p $RSEARCH_PID > /dev/null; then
    echo "Error: rsearch server failed to start"
    echo "Check .rsearch.log for details"
    exit 1
fi

# Register product schema
echo "Registering product schema..."
curl -s -X POST http://localhost:8080/api/v1/schemas \
  -H "Content-Type: application/json" \
  -d @examples/product_schema.json > /dev/null

if [ $? -eq 0 ]; then
    echo "Schema registered successfully"
else
    echo "Warning: Failed to register schema (server may not be ready yet)"
fi

# Start Node.js demo server for database execution
echo "Starting demo server..."
cd "$PROJECT_ROOT/examples"

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "Installing demo server dependencies..."
    npm install --silent
fi

# Start Node server
node server.js > ../.demo-server.log 2>&1 &
DEMO_PID=$!
echo $DEMO_PID > ../.demo-server.pid
cd "$PROJECT_ROOT"

sleep 2

# Check if demo server started
if ! ps -p $DEMO_PID > /dev/null; then
    echo "Warning: Demo server failed to start"
    echo "Database execution will not be available"
else
    echo "Demo server started successfully"
fi

echo ""
echo "=========================================="
echo "Demo environment is ready!"
echo "=========================================="
echo ""
echo "Services:"
echo "  - rsearch API:    http://localhost:8080"
echo "  - Demo Server:    http://localhost:3000"
echo "  - PostgreSQL:     localhost:5432 (user: rsearch, password: rsearch123)"
echo "  - MySQL:          localhost:3306 (user: rsearch, password: rsearch123)"
echo "  - MongoDB:        localhost:27017"
echo "  - pgAdmin:        http://localhost:5050 (admin@rsearch.local / admin)"
echo "  - Mongo Express:  http://localhost:8081"
echo ""
echo "Open in browser: http://localhost:3000/demo.html"
echo ""
echo "To stop the demo: ./examples/stop-demo.sh"
echo "=========================================="
echo ""
