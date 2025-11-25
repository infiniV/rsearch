#!/bin/bash
# Deployment script for rsearch Kubernetes manifests
# Usage: ./deploy.sh [apply|delete|status|logs]

set -e

NAMESPACE="rsearch"
KUSTOMIZE_DIR="/home/raw/rsearch/k8s"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

function info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

function warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

function error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

function check_prerequisites() {
    info "Checking prerequisites..."

    if ! command -v kubectl &> /dev/null; then
        error "kubectl not found. Please install kubectl."
        exit 1
    fi

    if ! kubectl cluster-info &> /dev/null; then
        error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi

    info "Prerequisites OK"
}

function apply() {
    info "Deploying rsearch to Kubernetes..."

    # Apply using kustomize
    kubectl apply -k "$KUSTOMIZE_DIR"

    info "Deployment initiated. Waiting for rollout..."
    kubectl rollout status deployment/rsearch -n "$NAMESPACE" --timeout=5m

    info "Deployment complete!"
    show_status
}

function delete() {
    warn "This will delete all rsearch resources including the namespace."
    read -p "Are you sure? (yes/no): " confirm

    if [ "$confirm" != "yes" ]; then
        info "Deletion cancelled."
        exit 0
    fi

    info "Deleting rsearch resources..."
    kubectl delete -k "$KUSTOMIZE_DIR"

    info "Resources deleted."
}

function show_status() {
    info "Checking deployment status..."

    echo ""
    echo "=== Namespace ==="
    kubectl get namespace "$NAMESPACE" 2>/dev/null || warn "Namespace not found"

    echo ""
    echo "=== Pods ==="
    kubectl get pods -n "$NAMESPACE" -o wide 2>/dev/null || warn "No pods found"

    echo ""
    echo "=== Deployment ==="
    kubectl get deployment -n "$NAMESPACE" 2>/dev/null || warn "No deployment found"

    echo ""
    echo "=== Service ==="
    kubectl get svc -n "$NAMESPACE" 2>/dev/null || warn "No service found"

    echo ""
    echo "=== Ingress ==="
    kubectl get ingress -n "$NAMESPACE" 2>/dev/null || warn "No ingress found"

    echo ""
    echo "=== HPA ==="
    kubectl get hpa -n "$NAMESPACE" 2>/dev/null || warn "No HPA found"

    echo ""
    echo "=== PodDisruptionBudget ==="
    kubectl get pdb -n "$NAMESPACE" 2>/dev/null || warn "No PDB found"

    echo ""
    echo "=== Recent Events ==="
    kubectl get events -n "$NAMESPACE" --sort-by='.lastTimestamp' 2>/dev/null | tail -10 || warn "No events found"
}

function show_logs() {
    info "Fetching logs from rsearch pods..."
    kubectl logs -n "$NAMESPACE" -l app.kubernetes.io/name=rsearch --tail=100 -f
}

function port_forward() {
    info "Port-forwarding rsearch service to localhost:8080..."
    info "Press Ctrl+C to stop"
    kubectl port-forward -n "$NAMESPACE" svc/rsearch 8080:80
}

function describe_pods() {
    info "Describing rsearch pods..."
    kubectl describe pods -n "$NAMESPACE" -l app.kubernetes.io/name=rsearch
}

function validate() {
    info "Validating Kubernetes manifests..."
    kubectl kustomize "$KUSTOMIZE_DIR" > /dev/null
    info "Manifests are valid!"
}

function show_help() {
    cat << EOF
rsearch Kubernetes Deployment Script

Usage: $0 <command>

Commands:
    apply           Deploy or update rsearch to Kubernetes
    delete          Delete all rsearch resources
    status          Show status of all rsearch resources
    logs            Stream logs from rsearch pods
    port-forward    Port-forward service to localhost:8080
    describe        Describe all rsearch pods
    validate        Validate Kubernetes manifests
    help            Show this help message

Examples:
    $0 apply        # Deploy rsearch
    $0 status       # Check deployment status
    $0 logs         # View logs
    $0 delete       # Remove deployment

Environment Variables:
    KUBECONFIG      Path to kubeconfig file (default: ~/.kube/config)

EOF
}

# Main script
case "${1:-}" in
    apply)
        check_prerequisites
        apply
        ;;
    delete)
        check_prerequisites
        delete
        ;;
    status)
        check_prerequisites
        show_status
        ;;
    logs)
        check_prerequisites
        show_logs
        ;;
    port-forward)
        check_prerequisites
        port_forward
        ;;
    describe)
        check_prerequisites
        describe_pods
        ;;
    validate)
        check_prerequisites
        validate
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        error "Unknown command: ${1:-}"
        echo ""
        show_help
        exit 1
        ;;
esac
