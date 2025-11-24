#!/bin/bash
# Stop rsearch demo environment

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "Stopping rsearch demo environment..."

# Stop rsearch server
if [ -f .rsearch.pid ]; then
    RSEARCH_PID=$(cat .rsearch.pid)
    if ps -p $RSEARCH_PID > /dev/null 2>&1; then
        kill $RSEARCH_PID 2>/dev/null || true
        echo "Stopped rsearch server (PID: $RSEARCH_PID)"
    fi
    rm -f .rsearch.pid
else
    # Fallback: kill by process name
    pkill -f "bin/rsearch" || true
    echo "Stopped rsearch server"
fi

# Stop Docker services
docker-compose -f docker-compose.dev.yaml down

# Clean up temporary files
rm -f .demo-config.yaml .rsearch.log

echo "Demo environment stopped"
