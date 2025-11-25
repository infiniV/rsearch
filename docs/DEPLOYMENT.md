# rsearch Deployment Guide

Complete guide for deploying rsearch in production environments.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Configuration Reference](#configuration-reference)
- [Health Checks](#health-checks)
- [Metrics and Monitoring](#metrics-and-monitoring)
- [Scaling](#scaling)
- [Security](#security)
- [Backup and Recovery](#backup-and-recovery)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

- **CPU**: 2+ cores recommended
- **Memory**: 512 MB minimum, 1-2 GB recommended
- **Disk**: 100 MB for binary, additional space for logs
- **Network**: HTTP/HTTPS connectivity
- **OS**: Linux, macOS, or Windows

### Dependencies

- **Go 1.21+** (for building from source)
- **Docker 20.10+** (for Docker deployment)
- **Kubernetes 1.24+** (for K8s deployment)
- **PostgreSQL extensions** (optional):
  - `pg_trgm` - for fuzzy search
  - Full-text search setup - for proximity search

## Docker Deployment

### Build Docker Image

```bash
# Build the image
docker build -t rsearch:1.0.0 .

# Or use multi-stage build for smaller image
docker build -f Dockerfile.multi-stage -t rsearch:1.0.0 .
```

**Dockerfile:**

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' \
    -o rsearch cmd/rsearch/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from build stage
COPY --from=builder /build/rsearch .

# Copy default config (optional)
COPY config.yaml .

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Run as non-root user
RUN addgroup -g 1000 rsearch && \
    adduser -D -u 1000 -G rsearch rsearch && \
    chown -R rsearch:rsearch /app

USER rsearch

ENTRYPOINT ["./rsearch"]
```

### Run with Docker

**Basic deployment:**

```bash
docker run -d \
  --name rsearch \
  -p 8080:8080 \
  -p 9090:9090 \
  rsearch:1.0.0
```

**With environment variables:**

```bash
docker run -d \
  --name rsearch \
  -p 8080:8080 \
  -p 9090:9090 \
  -e RSEARCH_SERVER_PORT=8080 \
  -e RSEARCH_LOGGING_LEVEL=info \
  -e RSEARCH_LOGGING_FORMAT=json \
  -e RSEARCH_METRICS_ENABLED=true \
  -e RSEARCH_CORS_ENABLED=true \
  -e RSEARCH_LIMITS_RATELIMIT_ENABLED=true \
  -e RSEARCH_LIMITS_RATELIMIT_REQUESTSPERMINUTE=100 \
  rsearch:1.0.0
```

**With config file:**

```bash
docker run -d \
  --name rsearch \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  rsearch:1.0.0 --config /app/config.yaml
```

**With schema directory:**

```bash
docker run -d \
  --name rsearch \
  -p 8080:8080 \
  -v $(pwd)/schemas:/app/schemas \
  -e RSEARCH_SCHEMAS_LOADFROMFILES=true \
  -e RSEARCH_SCHEMAS_DIRECTORY=/app/schemas \
  rsearch:1.0.0
```

### Docker Compose

**docker-compose.yml:**

```yaml
version: '3.8'

services:
  rsearch:
    image: rsearch:1.0.0
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      RSEARCH_SERVER_HOST: 0.0.0.0
      RSEARCH_SERVER_PORT: 8080
      RSEARCH_LOGGING_LEVEL: info
      RSEARCH_LOGGING_FORMAT: json
      RSEARCH_METRICS_ENABLED: "true"
      RSEARCH_METRICS_PORT: 9090
      RSEARCH_CORS_ENABLED: "true"
      RSEARCH_CORS_ALLOWEDORIGINS: "*"
      RSEARCH_LIMITS_RATELIMIT_ENABLED: "true"
      RSEARCH_LIMITS_RATELIMIT_REQUESTSPERMINUTE: 100
      RSEARCH_LIMITS_RATELIMIT_REQUESTSPERHOUR: 5000
    volumes:
      - ./config.yaml:/app/config.yaml
      - ./schemas:/app/schemas
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      start_period: 5s
      retries: 3
    restart: unless-stopped
    networks:
      - rsearch-network

  # Optional: Prometheus for metrics
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - rsearch-network

  # Optional: Grafana for visualization
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
    volumes:
      - grafana-data:/var/lib/grafana
    networks:
      - rsearch-network

networks:
  rsearch-network:
    driver: bridge

volumes:
  prometheus-data:
  grafana-data:
```

**Start services:**

```bash
docker-compose up -d
```

## Kubernetes Deployment

### Deployment Manifest

**rsearch-deployment.yaml:**

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: rsearch

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: rsearch-config
  namespace: rsearch
data:
  config.yaml: |
    server:
      host: 0.0.0.0
      port: 8080
      readTimeout: 30s
      writeTimeout: 30s
      shutdownTimeout: 10s
    logging:
      level: info
      format: json
      output: stdout
    metrics:
      enabled: true
      port: 9090
      path: /metrics
    cors:
      enabled: true
      allowedOrigins:
        - "*"
      allowedMethods:
        - GET
        - POST
        - DELETE
    limits:
      maxQueryLength: 10000
      maxParameterCount: 100
      rateLimit:
        enabled: true
        requestsPerMinute: 100
        requestsPerHour: 5000
        burst: 10

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rsearch
  namespace: rsearch
  labels:
    app: rsearch
spec:
  replicas: 3
  selector:
    matchLabels:
      app: rsearch
  template:
    metadata:
      labels:
        app: rsearch
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: rsearch
        image: rsearch:1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        - name: metrics
          containerPort: 9090
          protocol: TCP
        env:
        - name: RSEARCH_SERVER_HOST
          value: "0.0.0.0"
        - name: RSEARCH_SERVER_PORT
          value: "8080"
        - name: RSEARCH_LOGGING_LEVEL
          value: "info"
        - name: RSEARCH_LOGGING_FORMAT
          value: "json"
        - name: RSEARCH_METRICS_ENABLED
          value: "true"
        - name: RSEARCH_METRICS_PORT
          value: "9090"
        volumeMounts:
        - name: config
          mountPath: /app/config.yaml
          subPath: config.yaml
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
          timeoutSeconds: 3
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 3
          failureThreshold: 3
      volumes:
      - name: config
        configMap:
          name: rsearch-config
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000

---
apiVersion: v1
kind: Service
metadata:
  name: rsearch
  namespace: rsearch
  labels:
    app: rsearch
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
  - name: metrics
    port: 9090
    targetPort: 9090
    protocol: TCP
  selector:
    app: rsearch

---
apiVersion: v1
kind: Service
metadata:
  name: rsearch-external
  namespace: rsearch
  labels:
    app: rsearch
spec:
  type: LoadBalancer
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
  selector:
    app: rsearch

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: rsearch-hpa
  namespace: rsearch
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: rsearch
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80

---
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: rsearch-pdb
  namespace: rsearch
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: rsearch
```

### Deploy to Kubernetes

```bash
# Apply the manifests
kubectl apply -f rsearch-deployment.yaml

# Check deployment status
kubectl get pods -n rsearch

# Check service
kubectl get svc -n rsearch

# View logs
kubectl logs -n rsearch -l app=rsearch

# Scale deployment
kubectl scale deployment rsearch -n rsearch --replicas=5
```

### Ingress Configuration

**rsearch-ingress.yaml:**

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: rsearch-ingress
  namespace: rsearch
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - api.example.com
    secretName: rsearch-tls
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: rsearch
            port:
              number: 80
```

```bash
kubectl apply -f rsearch-ingress.yaml
```

## Configuration Reference

### Environment Variables

All configuration can be set via environment variables with the `RSEARCH_` prefix.

#### Server Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| RSEARCH_SERVER_HOST | string | localhost | Server bind address |
| RSEARCH_SERVER_PORT | int | 8080 | Server port |
| RSEARCH_SERVER_READTIMEOUT | duration | 30s | Read timeout |
| RSEARCH_SERVER_WRITETIMEOUT | duration | 30s | Write timeout |
| RSEARCH_SERVER_SHUTDOWNTIMEOUT | duration | 10s | Graceful shutdown timeout |

#### Logging Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| RSEARCH_LOGGING_LEVEL | string | info | Log level (debug, info, warn, error) |
| RSEARCH_LOGGING_FORMAT | string | json | Log format (json, console) |
| RSEARCH_LOGGING_OUTPUT | string | stdout | Log output (stdout, stderr, file path) |

#### Metrics Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| RSEARCH_METRICS_ENABLED | bool | false | Enable Prometheus metrics |
| RSEARCH_METRICS_PORT | int | 9090 | Metrics server port |
| RSEARCH_METRICS_PATH | string | /metrics | Metrics endpoint path |

#### CORS Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| RSEARCH_CORS_ENABLED | bool | false | Enable CORS |
| RSEARCH_CORS_ALLOWEDORIGINS | []string | ["*"] | Allowed origins |
| RSEARCH_CORS_ALLOWEDMETHODS | []string | [GET,POST,DELETE] | Allowed HTTP methods |

#### Schema Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| RSEARCH_SCHEMAS_LOADFROMFILES | bool | false | Load schemas from files on startup |
| RSEARCH_SCHEMAS_DIRECTORY | string | ./schemas | Schema directory path |

#### Limits Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| RSEARCH_LIMITS_MAXQUERYLENGTH | int | 10000 | Maximum query string length |
| RSEARCH_LIMITS_MAXPARAMETERCOUNT | int | 100 | Maximum query parameters |
| RSEARCH_LIMITS_MAXPARSEDEPTH | int | 50 | Maximum parse tree depth |
| RSEARCH_LIMITS_MAXSCHEMAFIELDS | int | 1000 | Maximum fields per schema |
| RSEARCH_LIMITS_MAXFIELDNAMELENGTH | int | 255 | Maximum field name length |
| RSEARCH_LIMITS_MAXSCHEMAS | int | 100 | Maximum registered schemas |
| RSEARCH_LIMITS_MAXREQUESTBODYSIZE | int64 | 1048576 | Maximum request body size (bytes) |
| RSEARCH_LIMITS_REQUESTTIMEOUT | duration | 30s | Request processing timeout |

#### Rate Limiting Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| RSEARCH_LIMITS_RATELIMIT_ENABLED | bool | false | Enable rate limiting |
| RSEARCH_LIMITS_RATELIMIT_REQUESTSPERMINUTE | int | 100 | Requests per minute |
| RSEARCH_LIMITS_RATELIMIT_REQUESTSPERHOUR | int | 5000 | Requests per hour |
| RSEARCH_LIMITS_RATELIMIT_BURST | int | 10 | Burst size |

#### Security Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| RSEARCH_SECURITY_ALLOWEDSPECIALCHARS | string | ".-_" | Allowed special characters |
| RSEARCH_SECURITY_BLOCKSQLKEYWORDS | bool | true | Block SQL keywords in queries |
| RSEARCH_SECURITY_AUTH_ENABLED | bool | false | Enable authentication |
| RSEARCH_SECURITY_AUTH_TYPE | string | apikey | Authentication type |
| RSEARCH_SECURITY_AUTH_APIKEYS | []string | [] | Valid API keys |

#### Cache Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| RSEARCH_CACHE_ENABLED | bool | true | Enable query caching |
| RSEARCH_CACHE_MAXSIZE | int | 10000 | Maximum cache entries |
| RSEARCH_CACHE_TTL | int | 3600 | Cache TTL in seconds |

#### Features Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| RSEARCH_FEATURES_QUERYSUGGESTIONS | bool | false | Enable query suggestions |
| RSEARCH_FEATURES_MAXQUERYLENGTH | int | 1000 | Maximum query length for suggestions |
| RSEARCH_FEATURES_REQUESTIDHEADER | string | X-Request-ID | Request ID header name |

### YAML Configuration File

**config.yaml:**

```yaml
server:
  host: 0.0.0.0
  port: 8080
  readTimeout: 30s
  writeTimeout: 30s
  shutdownTimeout: 10s

logging:
  level: info        # debug, info, warn, error
  format: json       # json, console
  output: stdout     # stdout, stderr, or file path

metrics:
  enabled: true
  port: 9090
  path: /metrics

cors:
  enabled: true
  allowedOrigins:
    - "*"
  allowedMethods:
    - GET
    - POST
    - DELETE

schemas:
  loadFromFiles: true
  directory: ./schemas

limits:
  maxQueryLength: 10000
  maxParameterCount: 100
  maxParseDepth: 50
  maxSchemaFields: 1000
  maxFieldNameLength: 255
  maxSchemas: 100
  maxRequestBodySize: 1048576
  requestTimeout: 30s
  rateLimit:
    enabled: true
    requestsPerMinute: 100
    requestsPerHour: 5000
    burst: 10

cache:
  enabled: true
  maxSize: 10000
  ttl: 3600

security:
  allowedSpecialChars: ".-_"
  blockSqlKeywords: true
  auth:
    enabled: false
    type: apikey
    apiKeys:
      - your-secret-key-1
      - your-secret-key-2

features:
  querySuggestions: false
  maxQueryLength: 1000
  requestIdHeader: X-Request-ID

api:
  versions:
    v1:
      enabled: true
      deprecated: false
```

## Health Checks

### Health Check Endpoint

**Endpoint:** `GET /health`

Returns service health status. Always returns 200 OK when service is running.

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

**Usage:**
- Load balancer health checks
- Basic service availability monitoring
- Container orchestration health probes

### Readiness Check Endpoint

**Endpoint:** `GET /ready`

Returns service readiness status. Indicates if the service is ready to accept requests.

**Response:**
```json
{
  "ready": true,
  "version": "1.0.0"
}
```

**Usage:**
- Kubernetes readiness probes
- Traffic routing decisions
- Startup completion detection

### Kubernetes Health Probes

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30
  timeoutSeconds: 3
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 3
  failureThreshold: 3
```

## Metrics and Monitoring

### Prometheus Configuration

**prometheus.yml:**

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'rsearch'
    static_configs:
      - targets: ['rsearch:9090']
    metrics_path: /metrics
```

### Available Metrics

**Request Metrics:**
- `rsearch_requests_total{endpoint, status}` - Total HTTP requests
- `rsearch_request_duration_seconds{endpoint}` - Request duration histogram
- `rsearch_active_requests` - Current active requests

**Error Metrics:**
- `rsearch_errors_total{type}` - Total errors by type

**Performance Metrics:**
- `rsearch_parse_duration_seconds` - Query parsing duration
- `rsearch_translate_duration_seconds` - Translation duration

**Schema Metrics:**
- `rsearch_active_schemas` - Number of registered schemas

**Cache Metrics:**
- `rsearch_cache_hits_total` - Cache hits
- `rsearch_cache_misses_total` - Cache misses

### Grafana Dashboard

Import the provided Grafana dashboard for rsearch visualization:

**Dashboard JSON:** (Place in `/home/raw/rsearch/docs/grafana-dashboard.json`)

Key panels:
- Request rate and latency
- Error rate by type
- Active connections
- Schema count
- Cache hit ratio
- Parse/translate performance

### Alerting Rules

**alerts.yml:**

```yaml
groups:
  - name: rsearch
    interval: 30s
    rules:
      - alert: HighErrorRate
        expr: rate(rsearch_errors_total[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors/sec"

      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(rsearch_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High request latency"
          description: "95th percentile latency is {{ $value }}s"

      - alert: ServiceDown
        expr: up{job="rsearch"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "rsearch service is down"
```

## Scaling

### Horizontal Scaling

rsearch is stateless and can be scaled horizontally:

**Docker Swarm:**
```bash
docker service scale rsearch=5
```

**Kubernetes:**
```bash
kubectl scale deployment rsearch -n rsearch --replicas=5
```

**Auto-scaling:**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: rsearch-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: rsearch
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Vertical Scaling

Adjust resource limits based on workload:

```yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

### Load Balancing

Use a load balancer for distributing traffic:

**NGINX:**
```nginx
upstream rsearch {
    least_conn;
    server rsearch-1:8080 max_fails=3 fail_timeout=30s;
    server rsearch-2:8080 max_fails=3 fail_timeout=30s;
    server rsearch-3:8080 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name api.example.com;

    location / {
        proxy_pass http://rsearch;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Request-ID $request_id;
    }
}
```

## Security

### TLS/SSL Configuration

**NGINX with Let's Encrypt:**

```nginx
server {
    listen 443 ssl http2;
    server_name api.example.com;

    ssl_certificate /etc/letsencrypt/live/api.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.example.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    location / {
        proxy_pass http://rsearch:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### API Key Authentication

Enable and configure API key authentication:

```bash
export RSEARCH_SECURITY_AUTH_ENABLED=true
export RSEARCH_SECURITY_AUTH_TYPE=apikey
export RSEARCH_SECURITY_AUTH_APIKEYS=key1,key2,key3
```

**Kubernetes Secret:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: rsearch-secrets
  namespace: rsearch
type: Opaque
stringData:
  apiKeys: |
    key1
    key2
    key3
```

### Network Policies

**Kubernetes Network Policy:**

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: rsearch-network-policy
  namespace: rsearch
spec:
  podSelector:
    matchLabels:
      app: rsearch
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 53  # DNS
```

### Security Best Practices

1. **Run as non-root user**
2. **Use read-only file systems**
3. **Enable API key authentication in production**
4. **Use TLS for all external traffic**
5. **Implement rate limiting**
6. **Keep dependencies updated**
7. **Monitor for vulnerabilities**
8. **Use network policies in Kubernetes**
9. **Implement proper CORS policies**
10. **Regular security audits**

## Backup and Recovery

### Schema Backup

**Export all schemas:**

```bash
curl http://localhost:8080/api/v1/schemas | jq '.' > schemas-backup.json
```

**Backup individual schema:**

```bash
curl http://localhost:8080/api/v1/schemas/users > users-schema.json
```

### Schema Restore

**Restore schemas on startup:**

```bash
# Place schemas in directory
mkdir -p ./schemas
cp users-schema.json ./schemas/

# Configure rsearch to load from files
export RSEARCH_SCHEMAS_LOADFROMFILES=true
export RSEARCH_SCHEMAS_DIRECTORY=./schemas

# Start rsearch
./bin/rsearch
```

**Restore via API:**

```bash
for schema in schemas/*.json; do
  curl -X POST http://localhost:8080/api/v1/schemas \
    -H "Content-Type: application/json" \
    -d @$schema
done
```

## Troubleshooting

### Common Issues

**1. Service won't start**

Check logs:
```bash
# Docker
docker logs rsearch

# Kubernetes
kubectl logs -n rsearch -l app=rsearch
```

Verify configuration:
```bash
# Check config syntax
./bin/rsearch --config config.yaml --validate
```

**2. High latency**

- Check resource usage (CPU, memory)
- Review metrics for bottlenecks
- Enable query caching
- Scale horizontally

**3. Parse errors**

- Validate query syntax
- Check schema field definitions
- Review error details in response
- Enable debug logging

**4. Rate limiting issues**

- Adjust rate limits in configuration
- Implement client-side retry logic
- Use backoff strategies

**5. Memory issues**

- Reduce cache size
- Lower max schemas limit
- Decrease max query length
- Scale vertically

### Debug Mode

Enable debug logging:

```bash
export RSEARCH_LOGGING_LEVEL=debug
export RSEARCH_LOGGING_FORMAT=console
```

### Performance Profiling

Enable profiling endpoints:

```bash
# CPU profile
curl http://localhost:8080/debug/pprof/profile > cpu.prof

# Heap profile
curl http://localhost:8080/debug/pprof/heap > heap.prof

# Analyze
go tool pprof cpu.prof
```

### Support

For additional help:

- GitHub Issues: https://github.com/infiniv/rsearch/issues
- Documentation: https://github.com/infiniv/rsearch/docs
- Email: support@example.com
