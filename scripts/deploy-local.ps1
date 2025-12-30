# Deploy NeoBank to Local Docker Desktop Kubernetes
# Includes FinOps, DR, and Istio configurations

param(
    [switch]$InstallIstio,
    [switch]$InstallVelero,
    [switch]$BuildImages,
    [switch]$SkipWait,
    [switch]$Clean
)

$ErrorActionPreference = "Stop"

function Write-Header {
    param([string]$Message)
    Write-Host "`n$('=' * 70)" -ForegroundColor Cyan
    Write-Host $Message -ForegroundColor Cyan
    Write-Host "$('=' * 70)`n" -ForegroundColor Cyan
}

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Err {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Check prerequisites
Write-Header "Checking Prerequisites"

try {
    kubectl version --client | Out-Null
    Write-Info "✓ kubectl installed"
} catch {
    Write-Err "kubectl not found. Please install kubectl"
    exit 1
}

try {
    kubectl cluster-info | Out-Null
    Write-Info "✓ Kubernetes cluster accessible"
} catch {
    Write-Err "Cannot connect to Kubernetes cluster. Is Docker Desktop running?"
    exit 1
}

$context = kubectl config current-context
Write-Info "Current context: $context"

if ($Clean) {
    Write-Header "Cleaning Up Existing Resources"
    
    kubectl delete namespace neobank --ignore-not-found=true
    kubectl delete namespace finops --ignore-not-found=true
    kubectl delete namespace e2e-tests --ignore-not-found=true
    
    Write-Info "Cleanup complete. Waiting for namespaces to terminate..."
    Start-Sleep -Seconds 10
}

# Install Istio if requested
if ($InstallIstio) {
    Write-Header "Installing Istio"
    
    try {
        istioctl version | Out-Null
        Write-Info "✓ istioctl found"
        
        Write-Info "Installing Istio with demo profile..."
        istioctl install --set profile=demo -y
        
        Write-Info "Waiting for Istio to be ready..."
        kubectl wait --for=condition=ready pod -l app=istiod -n istio-system --timeout=300s
        
        Write-Info "✓ Istio installed successfully"
    } catch {
        Write-Warn "istioctl not found. Skipping Istio installation"
        Write-Warn "Install with: https://istio.io/latest/docs/setup/getting-started/"
    }
}

# Install Velero if requested
if ($InstallVelero) {
    Write-Header "Installing Velero"
    
    try {
        velero version --client-only | Out-Null
        Write-Info "✓ velero CLI found"
        
        Write-Info "Installing Velero with local storage..."
        
        # Create velero namespace
        kubectl create namespace velero --dry-run=client -o yaml | kubectl apply -f -
        
        # Apply Velero configuration
        kubectl apply -k k8s/dr/
        
        Write-Info "✓ Velero configuration applied"
        Write-Warn "Note: Velero requires additional setup for backup storage"
    } catch {
        Write-Warn "velero CLI not found. Skipping Velero installation"
        Write-Warn "Install with: https://velero.io/docs/v1.12/basic-install/"
    }
}

# Build Docker images if requested
if ($BuildImages) {
    Write-Header "Building Docker Images"
    
    $services = @(
        @{Name="identity-service"; Path="backend/identity-service"},
        @{Name="ledger-service"; Path="backend/ledger-service"},
        @{Name="payment-service"; Path="backend/payment-service"},
        @{Name="product-service"; Path="backend/product-service"},
        @{Name="card-service"; Path="backend/card-service"},
        @{Name="frontend"; Path="frontend"}
    )
    
    foreach ($svc in $services) {
        Write-Info "Building $($svc.Name)..."
        docker build -t "neobank/$($svc.Name):local" $svc.Path
    }
    
    Write-Info "✓ All images built"
}

# Create namespaces
Write-Header "Creating Namespaces"

$namespaces = @(
    @{Name="neobank"; Labels=@("istio-injection=enabled")},
    @{Name="finops"; Labels=@()},
    @{Name="e2e-tests"; Labels=@()}
)

foreach ($ns in $namespaces) {
    Write-Info "Creating namespace: $($ns.Name)"
    
    $labels = ""
    if ($ns.Labels.Count -gt 0) {
        $labels = "--labels=" + ($ns.Labels -join ",")
    }
    
    if ($labels) {
        kubectl create namespace $ns.Name $labels --dry-run=client -o yaml | kubectl apply -f -
    } else {
        kubectl create namespace $ns.Name --dry-run=client -o yaml | kubectl apply -f -
    }
}

Write-Info "✓ Namespaces created"

# Apply FinOps configurations
Write-Header "Deploying FinOps Configurations"

try {
    # First create the finops namespace
    kubectl create namespace finops --dry-run=client -o yaml | kubectl apply -f -
    
    # Apply FinOps resources to their namespace
    kubectl apply -f k8s/finops/namespace.yaml 2>$null
    kubectl apply -f k8s/finops/kubecost-config.yaml 2>$null
    kubectl apply -f k8s/finops/metrics-exporter.yaml 2>$null
    kubectl apply -f k8s/finops/budget-alerts.yaml 2>$null
    
    # Apply FinOps resources to neobank namespace
    kubectl apply -f k8s/finops/resource-quotas.yaml -n neobank 2>$null
    kubectl apply -f k8s/finops/limit-ranges.yaml -n neobank 2>$null
    kubectl apply -f k8s/finops/cost-allocation-labels.yaml 2>$null
    
    Write-Info "✓ FinOps configurations applied"
    
    Write-Info "Verifying FinOps resources..."
    kubectl get resourcequotas -n neobank 2>$null
    kubectl get limitranges -n neobank 2>$null
} catch {
    Write-Warn "Error applying FinOps configurations: $_"
}

# Apply DR configurations (Velero)
if ($InstallVelero) {
    Write-Header "Deploying DR Configurations"
    
    try {
        # Create velero namespace first
        kubectl create namespace velero --dry-run=client -o yaml | kubectl apply -f -
        
        # Apply Velero configurations
        kubectl apply -f k8s/dr/namespace.yaml 2>$null
        kubectl apply -f k8s/dr/velero-config.yaml 2>$null
        kubectl apply -f k8s/dr/backup-schedules.yaml 2>$null
        kubectl apply -f k8s/dr/restore-procedures.yaml 2>$null
        kubectl apply -f k8s/dr/dr-testing.yaml 2>$null
        kubectl apply -f k8s/dr/failover-config.yaml 2>$null
        
        Write-Info "✓ DR configurations applied"
        
        Write-Info "Verifying DR resources..."
        kubectl get schedules -n velero 2>$null
        kubectl get cronjobs -n velero 2>$null
    } catch {
        Write-Warn "Error applying DR configurations: $_"
    }
}

# Apply Istio configurations
if ($InstallIstio) {
    Write-Header "Deploying Istio Configurations"
    
    try {
        # Wait a moment for Istio to be fully ready
        Start-Sleep -Seconds 5
        
        kubectl apply -f k8s/istio/gateway.yaml
        kubectl apply -f k8s/istio/peer-authentication.yaml
        kubectl apply -f k8s/istio/destination-rules.yaml
        kubectl apply -f k8s/istio/virtual-services.yaml
        kubectl apply -f k8s/istio/rate-limiting.yaml
        kubectl apply -f k8s/istio/authorization-policies.yaml
        kubectl apply -f k8s/istio/service-mesh-config.yaml
        
        Write-Info "✓ Istio configurations applied"
        
        Write-Info "Verifying Istio resources..."
        kubectl get gateway -n neobank
        kubectl get virtualservices -n neobank
        kubectl get destinationrules -n neobank
    } catch {
        Write-Warn "Error applying Istio configurations: $_"
    }
}

# Deploy application
Write-Header "Deploying NeoBank Application"

try {
    kubectl apply -k k8s/overlays/local/
    Write-Info "✓ Application deployed"
} catch {
    Write-Err "Error deploying application: $_"
    exit 1
}

# Wait for pods to be ready
if (-not $SkipWait) {
    Write-Header "Waiting for Pods to be Ready"
    
    Write-Info "Waiting for deployments to be available (this may take a few minutes)..."
    
    $deployments = kubectl get deployments -n neobank -o name
    foreach ($dep in $deployments) {
        $depName = $dep -replace "deployment.apps/", ""
        Write-Info "Waiting for $depName..."
        kubectl wait --for=condition=available $dep -n neobank --timeout=300s
    }
    
    Write-Info "✓ All deployments ready"
}

# Display status
Write-Header "Deployment Status"

Write-Host "`n--- Namespaces ---"
kubectl get namespaces | Select-String -Pattern "neobank|finops|velero|istio-system"

Write-Host "`n--- Pods in neobank namespace ---"
kubectl get pods -n neobank -o wide

Write-Host "`n--- Services in neobank namespace ---"
kubectl get services -n neobank

if ($InstallIstio) {
    Write-Host "`n--- Istio Gateway ---"
    kubectl get gateway -n neobank
    kubectl get svc istio-ingressgateway -n istio-system
}

Write-Host "`n--- FinOps Resources ---"
kubectl get resourcequotas,limitranges -n neobank

if ($InstallVelero) {
    Write-Host "`n--- DR Resources (Velero) ---"
    kubectl get schedules,cronjobs -n velero 2>$null
}

# Port forwarding instructions
Write-Header "Access Instructions"

Write-Host @"
To access services, run port forwarding:

# Frontend
kubectl port-forward svc/local-neobank-frontend 3001:3000 -n neobank

# Istio Gateway (if installed)
kubectl port-forward svc/istio-ingressgateway 8080:80 -n istio-system

# Individual services
kubectl port-forward svc/local-neobank-identity-service 8081:8081 -n neobank
kubectl port-forward svc/local-neobank-ledger-service 8082:8082 -n neobank
kubectl port-forward svc/local-neobank-payment-service 8083:8083 -n neobank
kubectl port-forward svc/local-neobank-product-service 8084:8084 -n neobank
kubectl port-forward svc/local-neobank-card-service 8085:8085 -n neobank

Or use the automated script:
.\scripts\setup-e2e.ps1 -PortForward

Service URLs:
- Frontend: http://localhost:3001
- Istio Gateway: http://localhost:8080
- Identity Service: http://localhost:8081/health
"@

Write-Header "Deployment Complete!"

Write-Info "Next steps:"
Write-Info "1. Set up port forwarding (see above)"
Write-Info "2. Run E2E tests: .\scripts\run-e2e-tests.ps1 -All"
Write-Info "3. Monitor resources: kubectl get all -n neobank"
