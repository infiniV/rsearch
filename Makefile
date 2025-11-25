.PHONY: demo build start stop clean test help generate-docs docker-build docker-run docker-stop docker-push docker-clean status restart kill-all

# Go configuration (standard install location)
export PATH := /usr/local/go/bin:$(PATH)
export GOPATH := $(HOME)/go

# Project variables
BINARY_NAME = rsearch
BINARY_PATH = bin/$(BINARY_NAME)
DEMO_HTML = examples/demo.html
DOCKER_IMAGE = rsearch
DOCKER_TAG = latest
DOCKER_REGISTRY = # Set your registry here (e.g., docker.io/username)

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
CYAN := \033[0;36m
NC := \033[0m # No Color

help:
	@echo ""
	@echo "$(CYAN)rsearch$(NC) - Query Translation Service"
	@echo ""
	@echo "$(GREEN)Development:$(NC)"
	@echo "  make demo             - Full restart: stop all, rebuild, start fresh"
	@echo "  make start            - Start services (calls stop first)"
	@echo "  make stop             - Stop ALL services (aggressive kill)"
	@echo "  make restart          - Alias for demo"
	@echo "  make status           - Show running services status"
	@echo "  make kill-all         - Nuclear option: kill everything rsearch-related"
	@echo ""
	@echo "$(GREEN)Build & Test:$(NC)"
	@echo "  make build            - Build rsearch binary"
	@echo "  make test             - Run tests"
	@echo "  make clean            - Stop services, remove binary and temp files"
	@echo "  make generate-docs    - Generate syntax documentation"
	@echo ""
	@echo "$(GREEN)Docker:$(NC)"
	@echo "  make docker-build     - Build Docker image"
	@echo "  make docker-run       - Run in Docker (production)"
	@echo "  make docker-stop      - Stop Docker containers"
	@echo "  make docker-push      - Push to registry"
	@echo "  make docker-clean     - Remove containers and images"
	@echo ""

# Build the rsearch binary
build:
	@echo "$(CYAN)[BUILD]$(NC) Compiling rsearch..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) cmd/rsearch/main.go
	@echo "$(GREEN)[BUILD]$(NC) Complete: $(BINARY_PATH)"

# Show status of all services
status:
	@echo ""
	@echo "$(CYAN)[STATUS]$(NC) Checking services..."
	@echo ""
	@printf "  rsearch server:  "; \
	if [ -f .rsearch.pid ] && ps -p $$(cat .rsearch.pid 2>/dev/null) > /dev/null 2>&1; then \
		echo "$(GREEN)Running$(NC) (PID: $$(cat .rsearch.pid))"; \
	elif pgrep -f "bin/rsearch" > /dev/null 2>&1; then \
		echo "$(YELLOW)Running$(NC) (PID: $$(pgrep -f 'bin/rsearch' | head -1))"; \
	else \
		echo "$(RED)Stopped$(NC)"; \
	fi
	@printf "  Demo server:     "; \
	if [ -f .demo-server.pid ] && ps -p $$(cat .demo-server.pid 2>/dev/null) > /dev/null 2>&1; then \
		echo "$(GREEN)Running$(NC) (PID: $$(cat .demo-server.pid))"; \
	elif pgrep -f "examples/server.js" > /dev/null 2>&1; then \
		echo "$(YELLOW)Running$(NC) (PID: $$(pgrep -f 'examples/server.js' | head -1))"; \
	else \
		echo "$(RED)Stopped$(NC)"; \
	fi
	@printf "  Docker services: "; \
	if docker-compose -f docker-compose.dev.yaml ps 2>/dev/null | grep -q "Up"; then \
		echo "$(GREEN)Running$(NC)"; \
	else \
		echo "$(RED)Stopped$(NC)"; \
	fi
	@echo ""
	@echo "$(CYAN)[ENDPOINTS]$(NC)"
	@echo "  API:          http://localhost:8080"
	@echo "  Demo:         http://localhost:3000/demo.html"
	@echo "  Health:       http://localhost:8080/health"
	@echo ""

# Nuclear option - kill everything
kill-all:
	@echo "$(RED)[KILL]$(NC) Terminating all rsearch processes..."
	@-pkill -9 -f "bin/rsearch" 2>/dev/null || true
	@-pkill -9 -f "examples/server.js" 2>/dev/null || true
	@-pkill -9 -f "node.*server.js" 2>/dev/null || true
	@-docker-compose -f docker-compose.dev.yaml down --remove-orphans 2>/dev/null || true
	@rm -f .rsearch.pid .demo-server.pid
	@echo "$(GREEN)[KILL]$(NC) All processes terminated"

# Stop all services - aggressive version
stop:
	@echo "$(YELLOW)[STOP]$(NC) Stopping all services..."
	@# Kill rsearch server
	@if [ -f .rsearch.pid ]; then \
		PID=$$(cat .rsearch.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			kill $$PID 2>/dev/null && echo "  Stopped rsearch (PID: $$PID)"; \
		fi; \
		rm -f .rsearch.pid; \
	fi
	@-pkill -f "bin/rsearch" 2>/dev/null && echo "  Killed stray rsearch processes" || true
	@# Kill demo server
	@if [ -f .demo-server.pid ]; then \
		PID=$$(cat .demo-server.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			kill $$PID 2>/dev/null && echo "  Stopped demo server (PID: $$PID)"; \
		fi; \
		rm -f .demo-server.pid; \
	fi
	@-pkill -f "examples/server.js" 2>/dev/null && echo "  Killed stray demo processes" || true
	@-pkill -f "node.*server.js" 2>/dev/null || true
	@# Stop Docker
	@echo "  Stopping Docker services..."
	@-docker-compose -f docker-compose.dev.yaml down 2>/dev/null || true
	@# Cleanup temp files
	@rm -f .demo-config.yaml .rsearch.log .demo-server.log
	@echo "$(GREEN)[STOP]$(NC) All services stopped"

# Start demo - always does full restart
demo: stop build
	@echo ""
	@echo "$(CYAN)[DEMO]$(NC) Starting fresh demo environment..."
	@./examples/start-demo.sh
	@echo ""
	@echo "$(GREEN)[DEMO]$(NC) Ready! Open: http://localhost:3000/demo.html"
	@echo "$(YELLOW)[DEMO]$(NC) Stop with: make stop"

# Restart alias
restart: demo

# Start services (stop first to ensure clean state)
start: stop build
	@echo "$(CYAN)[START]$(NC) Starting services..."
	@./examples/start-demo.sh

# Clean up everything
clean: stop
	@echo "$(YELLOW)[CLEAN]$(NC) Removing build artifacts..."
	@rm -f $(BINARY_PATH)
	@rm -f .rsearch.pid .demo-server.pid
	@rm -f .demo-config.yaml .rsearch.log .demo-server.log
	@echo "$(GREEN)[CLEAN]$(NC) Complete"

# Run tests
test:
	@echo "$(CYAN)[TEST]$(NC) Running tests..."
	@go test ./...

# Generate syntax documentation from test cases
generate-docs:
	@echo "Generating documentation..."
	@go run cmd/gendocs/main.go

# Docker targets

# Build Docker image
docker-build:
	@echo "Building Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built successfully"

# Run production Docker compose
docker-run:
	@echo "Starting rsearch in Docker (production mode)..."
	@docker-compose up -d
	@echo ""
	@echo "rsearch is running!"
	@echo "  Application: http://localhost:8080"
	@echo "  Health check: http://localhost:8080/health"
	@echo "  Metrics: http://localhost:9090/metrics"
	@echo ""
	@echo "View logs: docker-compose logs -f rsearch"
	@echo "Stop services: make docker-stop"

# Stop Docker compose
docker-stop:
	@echo "Stopping Docker containers..."
	@docker-compose down
	@echo "Containers stopped"

# Push Docker image to registry
docker-push:
	@if [ -z "$(DOCKER_REGISTRY)" ]; then \
		echo "Error: DOCKER_REGISTRY not set"; \
		echo "Set it in Makefile or use: make docker-push DOCKER_REGISTRY=your-registry"; \
		exit 1; \
	fi
	@echo "Tagging image for registry..."
	@docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Pushing to $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)"
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Image pushed successfully"

# Clean Docker resources
docker-clean:
	@echo "Cleaning Docker resources..."
	@docker-compose down -v
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	@echo "Docker resources cleaned"
