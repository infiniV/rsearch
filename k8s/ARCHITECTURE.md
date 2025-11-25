# rsearch Kubernetes Architecture

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Internet/Users                           │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTPS/TLS
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                      Ingress Controller                          │
│                    (nginx/traefik/ALB)                          │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Host: rsearch.example.com                                 │ │
│  │  Paths: /api/*, /health, /ready                           │ │
│  │  TLS: rsearch-tls-secret                                  │ │
│  └────────────────────────────────────────────────────────────┘ │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTP (internal)
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                    Service (ClusterIP)                           │
│                      Port 80 → 8080                             │
│                    Metrics Port 9090                            │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ Load Balanced
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
┌───────▼────────┐  ┌───────▼────────┐  ┌───────▼────────┐
│  rsearch Pod 1 │  │  rsearch Pod 2 │  │  rsearch Pod N │
│                │  │                │  │                │
│  Container:    │  │  Container:    │  │  Container:    │
│  - rsearch     │  │  - rsearch     │  │  - rsearch     │
│  - Port 8080   │  │  - Port 8080   │  │  - Port 8080   │
│  - Port 9090   │  │  - Port 9090   │  │  - Port 9090   │
│                │  │                │  │                │
│  Probes:       │  │  Probes:       │  │  Probes:       │
│  - /health     │  │  - /health     │  │  - /health     │
│  - /ready      │  │  - /ready      │  │  - /ready      │
│                │  │                │  │                │
│  Resources:    │  │  Resources:    │  │  Resources:    │
│  - 128Mi-256Mi │  │  - 128Mi-256Mi │  │  - 128Mi-256Mi │
│  - 50m-100m    │  │  - 50m-100m    │  │  - 50m-100m    │
└────────────────┘  └────────────────┘  └────────────────┘
        │                   │                   │
        └───────────────────┴───────────────────┘
                            │
                            │ Metrics Scraping
                            │
                ┌───────────▼───────────┐
                │   ServiceMonitor      │
                │   (Prometheus CRD)    │
                └───────────┬───────────┘
                            │
                            │
                ┌───────────▼───────────┐
                │   Prometheus Server   │
                │   (Monitoring)        │
                └───────────────────────┘
```

## Component Details

### Namespace: rsearch
Isolated namespace containing all rsearch resources for better organization and security boundaries.

### ConfigMap: rsearch-config
Environment-based configuration:
- Server settings (host, port, timeouts)
- Logging (level, format, output)
- Metrics (enabled, port, path)
- CORS (origins, methods)
- Limits (query length, rate limits, cache)
- Security (SQL keyword blocking, special chars)
- Features (API versions, request IDs)

### Deployment: rsearch
Production-grade deployment with:
- **Replicas**: 2 (default), managed by HPA
- **Update Strategy**: RollingUpdate (maxSurge: 1, maxUnavailable: 0)
- **Security**:
  - Non-root user (UID 1000)
  - Read-only root filesystem
  - No privilege escalation
  - All capabilities dropped
  - Seccomp profile: RuntimeDefault
- **Health Checks**:
  - Liveness: /health (10s delay, 10s period)
  - Readiness: /ready (5s delay, 5s period)
  - Startup: /health (0s delay, 5s period, 12 failures = 60s)
- **Resources**:
  - Requests: 50m CPU, 128Mi memory
  - Limits: 100m CPU, 256Mi memory
- **Volumes**:
  - EmptyDir for /tmp (writable)
  - EmptyDir for /app/cache (writable)
- **Anti-affinity**: Prefers spreading pods across nodes

### Service: rsearch
ClusterIP service for internal load balancing:
- **Type**: ClusterIP (internal only)
- **Ports**:
  - http: 80 → 8080 (API traffic)
  - metrics: 9090 → 9090 (Prometheus scraping)
- **Selector**: Routes to rsearch pods
- **Annotations**: Prometheus scraping configuration

### Ingress: rsearch
External access with TLS:
- **Host**: rsearch.example.com (customize)
- **Paths**:
  - /api/* → rsearch service
  - /health → rsearch service
  - /ready → rsearch service
- **TLS**: Requires rsearch-tls-secret (cert-manager or manual)
- **Annotations**: nginx ingress controller optimized
  - SSL redirect enabled
  - 1MB body size limit
  - 30s timeouts
  - 100 RPS rate limit

### HorizontalPodAutoscaler (HPA)
Automatic scaling based on metrics:
- **Min Replicas**: 2 (high availability)
- **Max Replicas**: 10
- **Metrics**:
  - CPU: 70% target utilization
  - Memory: 80% target utilization
- **Scale-up**: Fast (100% or 2 pods per 30s)
- **Scale-down**: Conservative (50% or 1 pod per 60s, 5min stabilization)

### ServiceMonitor
Prometheus integration (requires Prometheus Operator):
- **Scrape Interval**: 30s
- **Scrape Timeout**: 10s
- **Endpoint**: /metrics on port 9090
- **Relabeling**: Adds pod, node, namespace labels

### PodDisruptionBudget (PDB)
Availability protection:
- **minAvailable**: 1 pod must remain available
- **Protection**: Prevents voluntary disruptions from taking down all pods
- **Policy**: IfHealthyBudget (Kubernetes 1.26+)

### Kustomization
Declarative resource management:
- **Common Labels**: app.kubernetes.io/name=rsearch
- **Resources**: All manifests included
- **Image Management**: Easy version updates
- **Extensible**: Support for environment overlays

## Network Flow

1. **External Request** → Ingress Controller
2. **TLS Termination** → Ingress decrypts HTTPS
3. **Path Routing** → Routes /api/* to rsearch service
4. **Load Balancing** → Service distributes to healthy pods
5. **Pod Processing** → Container handles request
6. **Response** → Flows back through same path

## Monitoring Flow

1. **Metrics Exposure** → Each pod exposes /metrics on port 9090
2. **ServiceMonitor** → Configures Prometheus scraping
3. **Prometheus Scrape** → Pulls metrics every 30s
4. **Alerting** → Prometheus alerts based on metric thresholds
5. **Visualization** → Grafana dashboards display metrics

## Scaling Flow

1. **Metrics Collection** → HPA monitors CPU/memory usage
2. **Threshold Check** → Compares against 70% CPU / 80% memory targets
3. **Scale Decision** → Calculates desired replica count
4. **Scale Action** → Updates deployment replica count
5. **Pod Creation** → New pods start with readiness checks
6. **Load Distribution** → Service includes new pods when ready

## Security Boundaries

```
┌──────────────────────────────────────────────────────────┐
│ Cluster Perimeter                                         │
│  ┌────────────────────────────────────────────────────┐  │
│  │ Namespace: rsearch                                 │  │
│  │  ┌──────────────────────────────────────────────┐ │  │
│  │  │ Pod Security                                  │ │  │
│  │  │  - Non-root user                             │ │  │
│  │  │  - Read-only root filesystem                 │ │  │
│  │  │  - No privilege escalation                   │ │  │
│  │  │  - Capabilities dropped                      │ │  │
│  │  │  - Seccomp profile                           │ │  │
│  │  └──────────────────────────────────────────────┘ │  │
│  │                                                    │  │
│  │  Network Policies (optional):                     │  │
│  │  - Ingress: Only from ingress namespace           │  │
│  │  - Egress: Database, external APIs                │  │
│  └────────────────────────────────────────────────────┘  │
│                                                           │
│  Ingress Controller:                                      │
│  - TLS termination                                        │
│  - WAF integration (optional)                             │
│  - Rate limiting                                          │
│  - DDoS protection                                        │
└──────────────────────────────────────────────────────────┘
```

## High Availability Features

1. **Multiple Replicas**: Minimum 2 pods across nodes
2. **Pod Anti-Affinity**: Distributes pods across nodes
3. **PodDisruptionBudget**: Ensures minimum availability during updates
4. **Rolling Updates**: Zero-downtime deployments
5. **Health Probes**: Automatic unhealthy pod replacement
6. **HPA**: Scales to handle load spikes
7. **Resource Limits**: Prevents resource exhaustion

## Deployment Workflow

```
┌─────────────┐
│ Build Image │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Push to     │
│ Registry    │
└──────┬──────┘
       │
       ▼
┌─────────────┐       ┌──────────────┐
│ Update      │       │ Apply        │
│ Image Tag   │──────▶│ Manifests    │
└─────────────┘       └──────┬───────┘
                             │
                             ▼
                      ┌──────────────┐
                      │ Rolling      │
                      │ Update       │
                      └──────┬───────┘
                             │
                ┌────────────┼────────────┐
                │            │            │
         ┌──────▼──────┐    │    ┌───────▼──────┐
         │ Create New  │    │    │ Keep Old     │
         │ Pod         │    │    │ Running      │
         └──────┬──────┘    │    └──────────────┘
                │           │
         ┌──────▼──────┐    │
         │ Wait Ready  │    │
         └──────┬──────┘    │
                │           │
                └───────────┼────────────┐
                            │            │
                     ┌──────▼──────┐     │
                     │ Terminate   │     │
                     │ Old Pod     │     │
                     └─────────────┘     │
                                         │
                                  ┌──────▼──────┐
                                  │ Deployment  │
                                  │ Complete    │
                                  └─────────────┘
```

## Resource Requirements

### Per Pod
- **CPU**: 50m (request) to 100m (limit)
- **Memory**: 128Mi (request) to 256Mi (limit)

### Cluster Total (minimum 2 replicas)
- **CPU**: 100m request, 200m limit
- **Memory**: 256Mi request, 512Mi limit

### Scaling (10 replicas max)
- **CPU**: 500m request, 1000m limit
- **Memory**: 1280Mi request, 2560Mi limit

## Performance Considerations

1. **Horizontal Scaling**: HPA handles load increases automatically
2. **Resource Efficiency**: Small memory/CPU footprint per pod
3. **Fast Startup**: Startup probe allows 60s for initialization
4. **Request Timeout**: 30s at ingress and application level
5. **Rate Limiting**: 100 RPS at ingress (adjustable)
6. **Connection Pooling**: Application-level (if needed)
7. **Caching**: Enabled in application configuration

## Operational Runbook

### Deploy New Version
```bash
# Update image tag in kustomization.yaml
# Apply changes
kubectl apply -k /home/raw/rsearch/k8s

# Watch rollout
kubectl rollout status deployment/rsearch -n rsearch
```

### Rollback
```bash
kubectl rollout undo deployment/rsearch -n rsearch
```

### Scale Manually
```bash
kubectl scale deployment rsearch -n rsearch --replicas=5
```

### View Logs
```bash
kubectl logs -n rsearch -l app.kubernetes.io/name=rsearch --tail=100 -f
```

### Debug Pod Issues
```bash
kubectl describe pod -n rsearch <pod-name>
kubectl get events -n rsearch --sort-by='.lastTimestamp'
```

### Access Service Locally
```bash
kubectl port-forward -n rsearch svc/rsearch 8080:80
curl http://localhost:8080/health
```

### Check Metrics
```bash
kubectl port-forward -n rsearch svc/rsearch 9090:9090
curl http://localhost:9090/metrics
```

## Future Enhancements

- **NetworkPolicy**: Restrict traffic to/from pods
- **PodSecurityPolicy/PodSecurityStandards**: Enforce security policies
- **Secrets**: Move sensitive config to Kubernetes Secrets
- **External Secrets Operator**: Integrate with vault/cloud secret managers
- **Service Mesh**: Istio/Linkerd for advanced traffic management
- **Distributed Tracing**: OpenTelemetry integration
- **Custom Metrics**: HPA based on application metrics (requests/sec)
- **Multi-region**: Global load balancing across clusters
