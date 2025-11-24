.PHONY: demo build start stop clean test help

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
	@echo "  make demo   - Start demo environment and open browser"
	@echo "  make build  - Build rsearch binary"
	@echo "  make start  - Start services (Docker + rsearch)"
	@echo "  make stop   - Stop all services"
	@echo "  make clean  - Stop services and remove binary"
	@echo "  make test   - Run tests"
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
	@./examples/start-demo.sh
	@sleep 2
	@echo "Opening demo page in browser..."
	@if command -v xdg-open > /dev/null 2>&1; then \
		xdg-open "file://$(CURDIR)/$(DEMO_HTML)" 2>/dev/null & \
	elif command -v open > /dev/null 2>&1; then \
		open "file://$(CURDIR)/$(DEMO_HTML)" & \
	elif command -v wslview > /dev/null 2>&1; then \
		wslview "file://$(CURDIR)/$(DEMO_HTML)" & \
	else \
		echo "Could not detect browser. Please open manually:"; \
		echo "file://$(CURDIR)/$(DEMO_HTML)"; \
	fi
	@echo ""
	@echo "Demo is running! Press Ctrl+C and run 'make stop' when done."

# Start services without opening browser
start: build
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
