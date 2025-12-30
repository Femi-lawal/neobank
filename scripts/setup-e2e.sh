#!/bin/bash
# E2E Test Setup Script for Docker Desktop Kubernetes
# This script sets up the local testing environment

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=== NeoBank E2E Test Setup ==="
echo "Project Root: $PROJECT_ROOT"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed"
        exit 1
    fi
    log_info "kubectl: $(kubectl version --client --short 2>/dev/null || kubectl version --client)"
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    log_info "Docker: $(docker --version)"
    
    # Check Kubernetes context
    KUBE_CONTEXT=$(kubectl config current-context)
    log_info "Kubernetes context: $KUBE_CONTEXT"
    
    if [[ "$KUBE_CONTEXT" != *"docker-desktop"* ]] && [[ "$KUBE_CONTEXT" != *"kind"* ]] && [[ "$KUBE_CONTEXT" != *"minikube"* ]]; then
        log_warn "Current context is not a local cluster. Proceed with caution."
        read -p "Continue? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    # Check cluster connectivity
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    log_info "Cluster connectivity: OK"
}

# Install Istio
install_istio() {
    log_info "Installing Istio..."
    
    # Check if istioctl is installed
    if ! command -v istioctl &> /dev/null; then
        log_info "Downloading istioctl..."
        curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.20.0 sh -
        export PATH="$PWD/istio-1.20.0/bin:$PATH"
    fi
    
    # Check if Istio is already installed
    if kubectl get namespace istio-system &> /dev/null; then
        log_info "Istio namespace exists, checking installation..."
        if kubectl get deployment istiod -n istio-system &> /dev/null; then
            log_info "Istio is already installed"
            return
        fi
    fi
    
    # Install Istio with demo profile (includes all features for testing)
    log_info "Installing Istio with demo profile..."
    istioctl install --set profile=demo -y
    
    # Enable sidecar injection for neobank namespace
    kubectl label namespace neobank istio-injection=enabled --overwrite 2>/dev/null || true
    
    log_info "Istio installation complete"
}

# Create namespaces
create_namespaces() {
    log_info "Creating namespaces..."
    
    for ns in neobank finops velero e2e-tests; do
        if ! kubectl get namespace $ns &> /dev/null; then
            kubectl create namespace $ns
            log_info "Created namespace: $ns"
        else
            log_info "Namespace exists: $ns"
        fi
    done
    
    # Label for Istio injection
    kubectl label namespace neobank istio-injection=enabled --overwrite
    kubectl label namespace e2e-tests istio-injection=enabled --overwrite
}

# Build Docker images
build_images() {
    log_info "Building Docker images..."
    
    cd "$PROJECT_ROOT"
    
    # Build backend services
    SERVICES=("identity-service" "ledger-service" "payment-service" "product-service" "card-service")
    
    for service in "${SERVICES[@]}"; do
        log_info "Building $service..."
        docker build -t neobank/$service:latest -f backend/$service/Dockerfile ./backend
    done
    
    # Build frontend
    log_info "Building frontend..."
    docker build -t neobank/frontend:latest -f frontend/Dockerfile ./frontend
    
    log_info "All images built successfully"
}

# Deploy infrastructure
deploy_infrastructure() {
    log_info "Deploying infrastructure..."
    
    # Apply base configurations
    kubectl apply -k "$PROJECT_ROOT/k8s/base" || true
    
    # Apply FinOps configurations
    kubectl apply -k "$PROJECT_ROOT/k8s/finops" || true
    
    # Apply DR configurations
    kubectl apply -k "$PROJECT_ROOT/k8s/dr" || true
    
    # Apply Istio configurations
    kubectl apply -f "$PROJECT_ROOT/k8s/istio/" || true
    
    log_info "Infrastructure deployed"
}

# Wait for pods to be ready
wait_for_pods() {
    log_info "Waiting for pods to be ready..."
    
    # Wait for neobank pods
    kubectl wait --for=condition=ready pod -l app -n neobank --timeout=300s || {
        log_warn "Some pods may not be ready. Checking status..."
        kubectl get pods -n neobank
    }
    
    # Wait for Istio pods
    kubectl wait --for=condition=ready pod -l app -n istio-system --timeout=300s || {
        log_warn "Istio pods may not be ready. Checking status..."
        kubectl get pods -n istio-system
    }
}

# Setup port forwarding
setup_port_forwarding() {
    log_info "Setting up port forwarding..."
    
    # Kill any existing port-forward processes
    pkill -f "kubectl port-forward" || true
    
    # Port forward services
    kubectl port-forward svc/neobank-frontend -n neobank 3001:3000 &
    kubectl port-forward svc/neobank-identity-service -n neobank 8081:8081 &
    kubectl port-forward svc/neobank-ledger-service -n neobank 8082:8082 &
    kubectl port-forward svc/neobank-payment-service -n neobank 8083:8083 &
    kubectl port-forward svc/neobank-product-service -n neobank 8084:8084 &
    kubectl port-forward svc/neobank-card-service -n neobank 8085:8085 &
    
    # Port forward Istio ingress gateway
    kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80 &
    
    log_info "Port forwarding established"
    log_info "Frontend: http://localhost:3001"
    log_info "Istio Gateway: http://localhost:8080"
}

# Main execution
main() {
    check_prerequisites
    create_namespaces
    
    if [[ "$1" == "--with-istio" ]] || [[ "$1" == "-i" ]]; then
        install_istio
    fi
    
    if [[ "$1" == "--build" ]] || [[ "$1" == "-b" ]]; then
        build_images
    fi
    
    deploy_infrastructure
    wait_for_pods
    
    if [[ "$1" == "--port-forward" ]] || [[ "$1" == "-p" ]]; then
        setup_port_forwarding
    fi
    
    log_info "=== Setup Complete ==="
    log_info ""
    log_info "To run E2E tests:"
    log_info "  cd $PROJECT_ROOT/tests/e2e"
    log_info "  go test -v ./..."
    log_info ""
    log_info "To view cluster status:"
    log_info "  kubectl get pods -n neobank"
    log_info "  kubectl get pods -n istio-system"
}

main "$@"
