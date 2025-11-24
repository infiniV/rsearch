.PHONY: demo build start stop clean test help generate-docs

# Go configuration (standard install location)
export PATH := /usr/local/go/bin:$(PATH)
export GOPATH := $(HOME)/go

# Project variables
BINARY_NAME = rsearch
BINARY_PATH = bin/$(BINARY_NAME)
DEMO_HTML = examples/demo.html

help:
	@echo "rsearch Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make demo          - Start demo environment and open browser"
	@echo "  make build         - Build rsearch binary"
	@echo "  make start         - Start services (Docker + rsearch)"
	@echo "  make stop          - Stop all services"
	@echo "  make clean         - Stop services and remove binary"
	@echo "  make test          - Run tests"
	@echo "  make generate-docs - Generate syntax documentation from test cases"
	@echo ""

# Build the rsearch binary
build:
	@echo "Building rsearch..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) cmd/rsearch/main.go
	@echo "Build complete: $(BINARY_PATH)"

# Start demo environment and open browser
demo: build
	@echo "Starting demo environment..."
	@-pkill -f "bin/rsearch" 2>/dev/null || true
	@./examples/start-demo.sh
	@sleep 2
	@echo "Check file://///wsl.localhost/Ubuntu/home/raw/rsearch/examples/demo.html"
	@echo ""
	@echo "Demo is running! Press Ctrl+C and run 'make stop' when done."

# Start services without opening browser
start: build
	@-pkill -f "bin/rsearch" 2>/dev/null || true
	@./examples/start-demo.sh

# Stop all services
stop:
	@./examples/stop-demo.sh

# Clean up everything
clean: stop
	@echo "Removing binary..."
	@rm -f $(BINARY_PATH)
	@echo "Cleanup complete"

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Generate syntax documentation from test cases
generate-docs:
	@echo "Generating documentation..."
	@go run cmd/gendocs/main.go
