# Simple Step-by-Step Deployment for Local Kubernetes
# Fixes common issues and deploys correctly

$ErrorActionPreference = "Stop"

function Write-Step {
    param([string]$Message)
    Write-Host "`n>>> $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "    ✓ $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "    ✗ $Message" -ForegroundColor Red
}

Write-Host "`n" + ("="*70) -ForegroundColor Cyan
Write-Host "NeoBank Simple Deployment" -ForegroundColor Cyan
Write-Host ("="*70) + "`n" -ForegroundColor Cyan

# Step 1: Check cluster
Write-Step "Step 1: Checking Kubernetes cluster"
try {
    $context = kubectl config current-context
    Write-Success "Connected to: $context"
} catch {
    Write-Error "Cannot connect to Kubernetes"
    Write-Host "    Please ensure Docker Desktop Kubernetes is running" -ForegroundColor Yellow
    exit 1
}

# Step 2: Create namespaces
Write-Step "Step 2: Creating namespaces"
kubectl create namespace neobank --dry-run=client -o yaml | kubectl apply -f - | Out-Null
kubectl create namespace finops --dry-run=client -o yaml | kubectl apply -f - | Out-Null
kubectl create namespace e2e-tests --dry-run=client -o yaml | kubectl apply -f - | Out-Null
Write-Success "Namespaces created"

# Step 3: Apply FinOps configurations
Write-Step "Step 3: Applying FinOps configurations"
kubectl apply -f k8s/finops/resource-quotas.yaml 2>&1 | Out-Null
kubectl apply -f k8s/finops/limit-ranges.yaml 2>&1 | Out-Null
Write-Success "FinOps resources applied to neobank namespace"

# Step 4: Deploy application
Write-Step "Step 4: Deploying application"
kubectl apply -k k8s/overlays/local/ 2>&1 | Out-Null
if ($LASTEXITCODE -eq 0) {
    Write-Success "Application deployed"
} else {
    Write-Error "Application deployment failed"
    Write-Host "`nTrying with more details..." -ForegroundColor Yellow
    kubectl apply -k k8s/overlays/local/
}

# Step 5: Wait for pods
Write-Step "Step 5: Waiting for pods to start (30 seconds)"
Start-Sleep -Seconds 30

# Step 6: Check status
Write-Step "Step 6: Checking deployment status"

Write-Host "`nPods:" -ForegroundColor Yellow
kubectl get pods -n neobank

Write-Host "`nServices:" -ForegroundColor Yellow
kubectl get services -n neobank

Write-Host "`nFinOps Resources:" -ForegroundColor Yellow
kubectl get resourcequotas,limitranges -n neobank

Write-Host "`n" + ("="*70) -ForegroundColor Cyan
Write-Host "Deployment Complete!" -ForegroundColor Cyan
Write-Host ("="*70) -ForegroundColor Cyan

Write-Host "`nNext steps:" -ForegroundColor Green
Write-Host "  1. Wait for all pods to be Running (may take 2-3 minutes)" -ForegroundColor White
Write-Host "  2. Check status: kubectl get pods -n neobank -w" -ForegroundColor White
Write-Host "  3. Port forward: kubectl port-forward svc/local-neobank-frontend 3001:3000 -n neobank" -ForegroundColor White
Write-Host "  4. Access: http://localhost:3001`n" -ForegroundColor White
