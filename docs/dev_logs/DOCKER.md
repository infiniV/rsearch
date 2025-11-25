# Docker Deployment Guide

This guide explains how to build, run, and deploy rsearch using Docker.

## Quick Start

Build and run rsearch with Docker Compose:

```bash
make docker-build
make docker-run
```

Access the service:
- Application: http://localhost:8080
- Health check: http://localhost:8080/health
- Metrics: http://localhost:9090/metrics

## Docker Image

The production Docker image is built using a multi-stage build process for optimal size and security.

### Image Characteristics

- Base image: `alpine:3.19` (minimal runtime)
- Size: ~30MB (optimized)
- Non-root user: `rsearch` (UID/GID 1000)
- Static binary: CGO disabled for portability
- Health check: Built-in HTTP health check
- Architecture: linux/amd64

### Build Arguments

The Dockerfile supports build arguments for versioning:

```bash
docker build \
  --build-arg VERSION=1.0.0 \
  --build-arg COMMIT=$(git rev-parse HEAD) \
  --build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  -t rsearch:1.0.0 .
```

## Building the Image

### Using Make

```bash
# Build with default tag (latest)
make docker-build

# Build with custom tag
make docker-build DOCKER_TAG=1.0.0
```

### Using Docker CLI

```bash
# Basic build
docker build -t rsearch:latest .

# Build with version info
docker build \
  --build-arg VERSION=1.0.0 \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  -t rsearch:1.0.0 .
```

## Running the Container

### Using Docker Compose (Recommended)

Docker Compose includes both rsearch and PostgreSQL:

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f rsearch

# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

### Using Docker CLI

```bash
# Run with default configuration
docker run -d \
  --name rsearch \
  -p 8080:8080 \
  -e RSEARCH_SERVER_HOST=0.0.0.0 \
  rsearch:latest

# Run with custom configuration
docker run -d \
  --name rsearch \
  -p 8080:8080 \
  -e RSEARCH_SERVER_HOST=0.0.0.0 \
  -e RSEARCH_LOGGING_LEVEL=debug \
  -e RSEARCH_METRICS_ENABLED=true \
  -e RSEARCH_METRICS_PORT=9090 \
  rsearch:latest

# Run with volume-mounted schemas
docker run -d \
  --name rsearch \
  -p 8080:8080 \
  -e RSEARCH_SERVER_HOST=0.0.0.0 \
  -v $(pwd)/schemas:/app/schemas:ro \
  rsearch:latest
```

## Configuration

### Environment Variables

All configuration can be done via environment variables with the `RSEARCH_` prefix:

```bash
# Server configuration
RSEARCH_SERVER_HOST=0.0.0.0
RSEARCH_SERVER_PORT=8080
RSEARCH_SERVER_READTIMEOUT=30s
RSEARCH_SERVER_WRITETIMEOUT=30s

# Logging
RSEARCH_LOGGING_LEVEL=info        # debug, info, warn, error
RSEARCH_LOGGING_FORMAT=json       # json or console
RSEARCH_LOGGING_OUTPUT=stdout

# Metrics
RSEARCH_METRICS_ENABLED=true
RSEARCH_METRICS_PORT=9090
RSEARCH_METRICS_PATH=/metrics

# CORS
RSEARCH_CORS_ENABLED=false

# Limits
RSEARCH_LIMITS_MAXQUERYLENGTH=10000
RSEARCH_LIMITS_MAXPARAMETERCOUNT=100
RSEARCH_LIMITS_MAXPARSEDEPTH=50

# Cache
RSEARCH_CACHE_ENABLED=true
RSEARCH_CACHE_MAXSIZE=10000
RSEARCH_CACHE_TTL=3600

# Security
RSEARCH_SECURITY_BLOCKSQLKEYWORDS=true
```

### Config File

Alternatively, mount a config file:

```bash
docker run -d \
  --name rsearch \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  rsearch:latest -config /app/config.yaml
```

## Docker Compose Configuration

The `docker-compose.yaml` includes:

1. **rsearch service**: The main application
   - Built from local Dockerfile
   - Exposed on ports 8080 (API) and 9090 (metrics)
   - Health checks enabled
   - Depends on PostgreSQL

2. **postgres service**: PostgreSQL database
   - Version: 16-alpine
   - Default credentials: rsearch/rsearch123
   - Persistent volume for data
   - Health checks enabled

### Customizing Docker Compose

Edit `docker-compose.yaml` to customize:

```yaml
services:
  rsearch:
    environment:
      # Add or modify environment variables
      - RSEARCH_LOGGING_LEVEL=debug
      - RSEARCH_CORS_ENABLED=true
    volumes:
      # Mount custom schemas
      - ./schemas:/app/schemas:ro
    ports:
      # Change port mapping
      - "9000:8080"
```

## Health Checks

The container includes built-in health checks:

```bash
# Check container health status
docker ps --filter name=rsearch

# View health check logs
docker inspect rsearch | jq '.[0].State.Health'

# Manually test health endpoint
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

## Pushing to Registry

### Using Make

```bash
# Set registry in Makefile, then:
make docker-push DOCKER_REGISTRY=docker.io/username

# Or inline:
make docker-push DOCKER_REGISTRY=ghcr.io/username
```

### Using Docker CLI

```bash
# Tag for registry
docker tag rsearch:latest docker.io/username/rsearch:1.0.0

# Push to registry
docker push docker.io/username/rsearch:1.0.0

# Push multiple tags
docker tag rsearch:latest docker.io/username/rsearch:latest
docker push docker.io/username/rsearch:latest
docker push docker.io/username/rsearch:1.0.0
```

## Production Deployment

### Kubernetes

Example Kubernetes deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rsearch
spec:
  replicas: 3
  selector:
    matchLabels:
      app: rsearch
  template:
    metadata:
      labels:
        app: rsearch
    spec:
      containers:
      - name: rsearch
        image: rsearch:1.0.0
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: RSEARCH_SERVER_HOST
          value: "0.0.0.0"
        - name: RSEARCH_LOGGING_FORMAT
          value: "json"
        - name: RSEARCH_METRICS_ENABLED
          value: "true"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
```

### Docker Swarm

```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.yaml rsearch

# View services
docker service ls

# Scale service
docker service scale rsearch_rsearch=3

# Remove stack
docker stack rm rsearch
```

## Troubleshooting

### Container won't start

Check logs:
```bash
docker logs rsearch
```

Common issues:
- Port already in use: Change port mapping
- Permission denied: Ensure volumes have correct permissions
- Configuration error: Check environment variables

### Health check failing

```bash
# Check if service is listening
docker exec rsearch wget -O- http://localhost:8080/health

# Check if bound to correct interface
docker exec rsearch netstat -tlnp
```

### High memory usage

```bash
# Check container stats
docker stats rsearch

# Adjust cache limits
docker run -e RSEARCH_CACHE_MAXSIZE=5000 rsearch:latest
```

### Viewing metrics

```bash
# Access Prometheus metrics
curl http://localhost:9090/metrics

# Or from inside container
docker exec rsearch wget -O- http://localhost:9090/metrics
```

## Security Best Practices

1. **Non-root user**: Container runs as `rsearch` user (UID 1000)
2. **Read-only volumes**: Mount schemas as read-only (`:ro`)
3. **No secrets in environment**: Use Docker secrets or volume-mounted files
4. **Network isolation**: Use Docker networks to isolate services
5. **Resource limits**: Set memory and CPU limits in production
6. **Image scanning**: Scan images for vulnerabilities regularly

## Development vs Production

### Development
- Use `docker-compose.dev.yaml` for development databases
- Enable debug logging
- Mount source code for live reload
- Use console logging format

### Production
- Use `docker-compose.yaml` for production
- Use JSON logging format
- Enable metrics collection
- Set resource limits
- Use health checks
- Run multiple replicas
- Use persistent volumes

## Make Targets Reference

```bash
make docker-build     # Build Docker image
make docker-run       # Start services with docker-compose
make docker-stop      # Stop docker-compose services
make docker-push      # Push image to registry
make docker-clean     # Remove containers and images
```

## Additional Resources

- [Dockerfile](./Dockerfile) - Multi-stage production Dockerfile
- [docker-compose.yaml](./docker-compose.yaml) - Production compose file
- [config.example.yaml](./config.example.yaml) - Configuration reference
- [.dockerignore](./.dockerignore) - Build context exclusions
