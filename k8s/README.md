# Kubernetes Deployment for rsearch

This directory contains Kubernetes manifests for deploying rsearch in a production-grade Kubernetes cluster.

## Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured to access your cluster
- Docker image built and pushed to a registry
- (Optional) Prometheus Operator for ServiceMonitor
- (Optional) cert-manager for TLS certificates
- (Optional) Ingress controller (nginx, traefik, etc.)

## Quick Start

### 1. Build and Push Docker Image

```bash
# Build the Docker image
docker build -t your-registry/rsearch:latest .

# Push to your registry
docker push your-registry/rsearch:latest

# Update the image in kustomization.yaml or deployment.yaml
```

### 2. Update Configuration

Edit `configmap.yaml` to adjust configuration values:
- Server settings
- Logging level
- Metrics endpoints
- Rate limits
- CORS settings

Edit `ingress.yaml` to set your domain:
- Replace `rsearch.example.com` with your actual domain
- Update TLS certificate configuration
- Adjust ingress class and annotations for your ingress controller

### 3. Deploy Using kubectl

```bash
# Create namespace
kubectl apply -f namespace.yaml

# Apply all manifests
kubectl apply -f .

# Or use kustomize
kubectl apply -k .
```

### 4. Deploy Using Kustomize

```bash
# Deploy using kustomize
kubectl apply -k /home/raw/rsearch/k8s

# View resources before applying
kubectl kustomize /home/raw/rsearch/k8s
```

### 5. Verify Deployment

```bash
# Check pods
kubectl get pods -n rsearch

# Check services
kubectl get svc -n rsearch

# Check ingress
kubectl get ingress -n rsearch

# View logs
kubectl logs -n rsearch -l app.kubernetes.io/name=rsearch --tail=50 -f

# Check HPA status
kubectl get hpa -n rsearch

# Check pod disruption budget
kubectl get pdb -n rsearch
```

## Manifest Descriptions

### namespace.yaml
- Creates dedicated `rsearch` namespace for isolation
- Labels for resource organization

### configmap.yaml
- Configuration values as environment variables
- All rsearch settings (server, logging, metrics, limits, etc.)
- Can be overridden per environment

### deployment.yaml
- Deployment with 2 default replicas for HA
- Resource limits: 256Mi memory, 100m CPU
- Resource requests: 128Mi memory, 50m CPU
- Health probes (liveness, readiness, startup)
- Security contexts (non-root, read-only root filesystem)
- Pod anti-affinity for distributing replicas across nodes
- Rolling update strategy with zero downtime
- EmptyDir volumes for tmp and cache (read-only root filesystem)

### service.yaml
- ClusterIP service exposing port 80 (maps to container port 8080)
- Metrics port 9090 for Prometheus scraping
- Service discovery within cluster

### ingress.yaml
- External access to rsearch API
- TLS termination (requires cert-manager or manual certificates)
- Path-based routing for /api, /health, /ready endpoints
- Annotations for nginx ingress controller (adjust for your ingress)
- Rate limiting and proxy settings

### hpa.yaml
- HorizontalPodAutoscaler for automatic scaling
- Min replicas: 2, Max replicas: 10
- CPU-based scaling (70% threshold)
- Memory-based scaling (80% threshold)
- Scale-up/scale-down behaviors for stability

### servicemonitor.yaml
- Prometheus ServiceMonitor (requires Prometheus Operator)
- Scrapes /metrics endpoint every 30s
- Relabeling for pod, node, namespace labels
- Integrates with Prometheus for monitoring

### poddisruptionbudget.yaml
- Ensures at least 1 pod available during disruptions
- Protects against voluntary disruptions (node drains, updates)
- Maintains service availability during maintenance

### kustomization.yaml
- Kustomize configuration for managing all resources
- Common labels and annotations
- Image version management
- Easy overlays for different environments (dev, staging, prod)

## Environment-Specific Overlays

You can create overlays for different environments:

```bash
k8s/
├── base/               # Base manifests (current directory)
│   ├── namespace.yaml
│   ├── deployment.yaml
│   └── ...
├── overlays/
│   ├── dev/
│   │   ├── kustomization.yaml
│   │   └── configmap-patch.yaml
│   ├── staging/
│   │   ├── kustomization.yaml
│   │   └── deployment-patch.yaml
│   └── prod/
│       ├── kustomization.yaml
│       ├── replicas-patch.yaml
│       └── resources-patch.yaml
```

Example overlay (overlays/prod/kustomization.yaml):
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

bases:
  - ../../base

patchesStrategicMerge:
  - replicas-patch.yaml

images:
  - name: rsearch
    newName: your-registry/rsearch
    newTag: v1.2.3
```

## Monitoring and Observability

### Prometheus Metrics
- Metrics exposed on port 9090 at /metrics
- ServiceMonitor automatically configures Prometheus scraping
- View metrics in Prometheus or Grafana

### Logs
```bash
# Stream logs from all pods
kubectl logs -n rsearch -l app.kubernetes.io/name=rsearch -f

# Logs from specific pod
kubectl logs -n rsearch rsearch-<pod-id> -f
```

### Health Checks
```bash
# Port-forward to access health endpoints locally
kubectl port-forward -n rsearch svc/rsearch 8080:80

# Check health
curl http://localhost:8080/health

# Check readiness
curl http://localhost:8080/ready
```

## Scaling

### Manual Scaling
```bash
# Scale deployment manually
kubectl scale deployment rsearch -n rsearch --replicas=5
```

### Autoscaling
HPA automatically scales based on CPU/memory utilization:
- Edit `hpa.yaml` to adjust min/max replicas or thresholds
- Monitor HPA: `kubectl get hpa -n rsearch -w`

## Troubleshooting

### Pods not starting
```bash
# Describe pod for events
kubectl describe pod -n rsearch rsearch-<pod-id>

# Check logs
kubectl logs -n rsearch rsearch-<pod-id>

# Check resource constraints
kubectl top pods -n rsearch
```

### Service not accessible
```bash
# Check service endpoints
kubectl get endpoints -n rsearch

# Test service from within cluster
kubectl run -n rsearch test-pod --rm -i --tty --image=curlimages/curl -- sh
curl http://rsearch/health
```

### Ingress not working
```bash
# Check ingress status
kubectl describe ingress -n rsearch rsearch

# Check ingress controller logs
kubectl logs -n ingress-nginx -l app.kubernetes.io/component=controller

# Verify DNS and TLS
kubectl get secret -n rsearch rsearch-tls-secret
```

## Security Considerations

1. **Container Security**
   - Non-root user (UID 1000)
   - Read-only root filesystem
   - No privilege escalation
   - All capabilities dropped

2. **Network Policies**
   - Consider adding NetworkPolicy for ingress/egress control

3. **TLS/Certificates**
   - Use cert-manager for automated certificate management
   - Or manually create TLS secrets

4. **API Security**
   - Enable authentication in configmap if needed
   - Configure CORS appropriately
   - Enable rate limiting

5. **Secrets Management**
   - For API keys, use Kubernetes Secrets or external secret managers
   - Never commit secrets to version control

## Cleanup

```bash
# Delete all resources
kubectl delete -k /home/raw/rsearch/k8s

# Or delete namespace (removes everything)
kubectl delete namespace rsearch
```

## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Kustomize Documentation](https://kubectl.docs.kubernetes.io/guides/introduction/kustomize/)
- [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)
- [cert-manager](https://cert-manager.io/)
- [nginx-ingress](https://kubernetes.github.io/ingress-nginx/)
