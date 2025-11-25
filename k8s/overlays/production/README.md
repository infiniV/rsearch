# Production Overlay

This overlay contains production-specific configurations for rsearch.

## Changes from Base

1. **Replicas**: Increased to 3 for higher availability
2. **Resources**: Higher limits and requests
   - CPU: 250m request, 500m limit (vs 50m/100m base)
   - Memory: 512Mi request, 1Gi limit (vs 128Mi/256Mi base)
3. **Image**: Specific version tag (v1.0.0) instead of latest
4. **Rate Limits**: Higher limits for production traffic
5. **Environment Labels**: Added environment=production

## Deployment

```bash
# Deploy to production
kubectl apply -k /home/raw/rsearch/k8s/overlays/production

# Preview changes
kubectl kustomize /home/raw/rsearch/k8s/overlays/production

# Dry run
kubectl apply -k /home/raw/rsearch/k8s/overlays/production --dry-run=client
```

## Customization

Before deploying, update:

1. **Image Registry**: Edit `kustomization.yaml` and change `newName` to your registry
2. **Image Tag**: Update `newTag` to your specific version
3. **Domain**: Edit `../../ingress.yaml` to set your production domain
4. **TLS Certificate**: Ensure TLS secret exists or configure cert-manager

## Resource Requirements

With 3 replicas:
- **Total CPU**: 750m request, 1500m limit
- **Total Memory**: 1536Mi request, 3Gi limit

Ensure your cluster has sufficient capacity.
