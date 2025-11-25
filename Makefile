.PHONY: demo build start stop clean test help generate-docs docker-build docker-run docker-stop docker-push docker-clean

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

help:
	@echo "rsearch Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make demo             - Start demo environment and open browser"
	@echo "  make build            - Build rsearch binary"
	@echo "  make start            - Start services (Docker + rsearch)"
	@echo "  make stop             - Stop all services"
	@echo "  make clean            - Stop services and remove binary"
	@echo "  make test             - Run tests"
	@echo "  make generate-docs    - Generate syntax documentation from test cases"
	@echo ""
	@echo "Docker commands:"
	@echo "  make docker-build     - Build Docker image"
	@echo "  make docker-run       - Run rsearch in Docker (production compose)"
	@echo "  make docker-stop      - Stop Docker containers"
	@echo "  make docker-push      - Push Docker image to registry"
	@echo "  make docker-clean     - Remove Docker containers and images"
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
