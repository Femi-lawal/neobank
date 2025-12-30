# NeoBank Complete Local Deployment
# This script deploys everything step-by-step with proper error handling

param(
    [switch]$SkipValidation,
    [switch]$SkipBuild,
    [switch]$Clean
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

function Write-Step { param([string]$Msg) Write-Host "`n>>> $Msg" -ForegroundColor Cyan }
function Write-OK { param([string]$Msg) Write-Host "    ✓ $Msg" -ForegroundColor Green }
function Write-Fail { param([string]$Msg) Write-Host "    ✗ $Msg" -ForegroundColor Red }
function Write-Warn { param([string]$Msg) Write-Host "    ⚠ $Msg" -ForegroundColor Yellow }

Write-Host "`n$('='*70)" -ForegroundColor Cyan
Write-Host "NeoBank Local Kubernetes Deployment" -ForegroundColor Cyan
Write-Host $('='*70) -ForegroundColor Cyan

# Prerequisites check
Write-Step "Checking prerequisites"
try {
    kubectl version --client 2>&1 | Out-Null
    Write-OK "kubectl installed"
} catch {
    Write-Fail "kubectl not found"
    exit 1
}

try {
    $context = kubectl config current-context 2>&1
    Write-OK "Cluster: $context"
} catch {
    Write-Fail "Cannot connect to Kubernetes cluster"
    Write-Host "    Make sure Docker Desktop Kubernetes is enabled" -ForegroundColor Yellow
    exit 1
}

# Clean existing deployment
if ($Clean) {
    Write-Step "Cleaning existing deployment"
    kubectl delete namespace neobank --ignore-not-found=true 2>&1 | Out-Null
    kubectl delete namespace finops --ignore-not-found=true 2>&1 | Out-Null
    kubectl delete namespace e2e-tests --ignore-not-found=true 2>&1 | Out-Null
    Write-OK "Cleanup complete"
    Start-Sleep -Seconds 5
}

# Validate manifests
if (-not $SkipValidation) {
    Write-Step "Validating Kubernetes manifests"
    $validation = kustomize build k8s/overlays/local 2>&1
    if ($LASTEXITCODE -eq 0) {
        $count = ($validation | Select-String -Pattern "^kind:").Count
        Write-OK "Manifests valid ($count resources)"
    } else {
        Write-Fail "Manifest validation failed"
        Write-Host $validation -ForegroundColor Red
        exit 1
    }
}

# Build images
if (-not $SkipBuild) {
    Write-Step "Building Docker images"
    
    $services = @(
        @{Name="identity-service"; Path="backend/identity-service"},
        @{Name="ledger-service"; Path="backend/ledger-service"},
        @{Name="payment-service"; Path="backend/payment-service"},
        @{Name="product-service"; Path="backend/product-service"},
        @{Name="card-service"; Path="backend/card-service"},
        @{Name="frontend"; Path="frontend"}
    )
    
    foreach ($svc in $services) {
        if (Test-Path $svc.Path) {
            Write-Host "    Building $($svc.Name)..." -ForegroundColor Gray
            docker build -t "neobank/$($svc.Name):local" $svc.Path 2>&1 | Out-Null
            if ($LASTEXITCODE -eq 0) {
                Write-OK "$($svc.Name)"
            } else {
                Write-Warn "$($svc.Name) build failed (will use existing image if available)"
            }
        } else {
            Write-Warn "$($svc.Name) path not found: $($svc.Path)"
        }
    }
}

# Create namespaces
Write-Step "Creating namespaces"
kubectl create namespace neobank --dry-run=client -o yaml | kubectl apply -f - 2>&1 | Out-Null
kubectl create namespace finops --dry-run=client -o yaml | kubectl apply -f - 2>&1 | Out-Null
kubectl create namespace e2e-tests --dry-run=client -o yaml | kubectl apply -f - 2>&1 | Out-Null
Write-OK "Namespaces ready"

# Deploy FinOps resources
Write-Step "Deploying FinOps configurations"
kubectl apply -f k8s/finops/resource-quotas.yaml 2>&1 | Out-Null
kubectl apply -f k8s/finops/limit-ranges.yaml 2>&1 | Out-Null

$rq = kubectl get resourcequota -n neobank 2>&1
$lr = kubectl get limitrange -n neobank 2>&1
if ($rq -notmatch "No resources found" -and $lr -notmatch "No resources found") {
    Write-OK "Resource quotas and limits applied"
} else {
    Write-Warn "FinOps resources may not be fully applied"
}

# Deploy application
Write-Step "Deploying NeoBank application"
kubectl apply -k k8s/overlays/local/ 2>&1 | Out-Null
if ($LASTEXITCODE -eq 0) {
    Write-OK "Application deployed"
} else {
    Write-Fail "Deployment failed, retrying with verbose output..."
    kubectl apply -k k8s/overlays/local/
}

# Wait for pods
Write-Step "Waiting for pods to start"
Write-Host "    This may take 2-3 minutes..." -ForegroundColor Gray

$maxWait = 180
$waited = 0
$allReady = $false

while ($waited -lt $maxWait -and -not $allReady) {
    Start-Sleep -Seconds 10
    $waited += 10
    
    $pods = kubectl get pods -n neobank -o json 2>&1 | ConvertFrom-Json
    $total = $pods.items.Count
    $running = ($pods.items | Where-Object { $_.status.phase -eq "Running" }).Count
    
    Write-Host "    Pods: $running/$total running (${waited}s)" -ForegroundColor Gray
    
    if ($running -eq $total -and $total -gt 0) {
        $allReady = $true
    }
}

if ($allReady) {
    Write-OK "All pods are running"
} else {
    Write-Warn "Not all pods are running yet"
}

# Display status
Write-Step "Deployment Status"

Write-Host "`n  Pods:" -ForegroundColor Yellow
kubectl get pods -n neobank -o wide

Write-Host "`n  Services:" -ForegroundColor Yellow
kubectl get services -n neobank

Write-Host "`n  FinOps Resources:" -ForegroundColor Yellow
kubectl get resourcequotas,limitranges -n neobank 2>&1

# Summary
Write-Host "`n$('='*70)" -ForegroundColor Cyan
Write-Host "Deployment Complete!" -ForegroundColor Green
Write-Host $('='*70) -ForegroundColor Cyan

Write-Host "`nNext Steps:" -ForegroundColor Cyan
Write-Host "  1. Verify: kubectl get pods -n neobank" -ForegroundColor White
Write-Host "  2. Port forward frontend:" -ForegroundColor White
Write-Host "     kubectl port-forward svc/local-neobank-frontend 3001:3000 -n neobank" -ForegroundColor Gray
Write-Host "  3. Access: http://localhost:3001" -ForegroundColor White
Write-Host "  4. Run validation: .\scripts\verify-deployment.ps1" -ForegroundColor White
Write-Host "  5. Run tests: .\scripts\run-e2e-tests.ps1 -Cluster`n" -ForegroundColor White

# Show any issues
$failedPods = kubectl get pods -n neobank -o json 2>&1 | ConvertFrom-Json | 
    Select-Object -ExpandProperty items | 
    Where-Object { $_.status.phase -ne "Running" }

if ($failedPods.Count -gt 0) {
    Write-Host "⚠ Some pods are not running:" -ForegroundColor Yellow
    foreach ($pod in $failedPods) {
        Write-Host "  - $($pod.metadata.name): $($pod.status.phase)" -ForegroundColor Yellow
        Write-Host "    Check logs: kubectl logs $($pod.metadata.name) -n neobank" -ForegroundColor Gray
    }
    Write-Host ""
}
