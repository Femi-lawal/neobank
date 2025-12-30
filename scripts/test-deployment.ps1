# Test NeoBank Local Deployment
# This script tests the deployment process step by step

param(
    [switch]$DryRun,
    [switch]$SkipBuild
)

$ErrorActionPreference = "Continue"

Write-Host "`n=== NeoBank Deployment Test ===" -ForegroundColor Cyan

# Test 1: Check cluster connectivity
Write-Host "`n[TEST 1] Checking Kubernetes cluster..." -ForegroundColor Yellow
try {
    $context = kubectl config current-context
    Write-Host "  ✓ Connected to: $context" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Cannot connect to Kubernetes cluster" -ForegroundColor Red
    Write-Host "    Make sure Docker Desktop Kubernetes is enabled" -ForegroundColor Gray
    exit 1
}

# Test 2: Validate Kustomize manifests
Write-Host "`n[TEST 2] Validating Kustomize configurations..." -ForegroundColor Yellow

$overlayPath = "k8s\overlays\local"
if (Test-Path $overlayPath) {
    Write-Host "  ✓ Local overlay exists" -ForegroundColor Green
    
    # Test kustomize build (dry-run)
    Write-Host "  Testing kustomize build..." -ForegroundColor Gray
    $buildOutput = kustomize build $overlayPath 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        $resourceCount = ($buildOutput | Select-String -Pattern "^kind:").Count
        Write-Host "  ✓ Kustomize build successful ($resourceCount resources)" -ForegroundColor Green
    } else {
        Write-Host "  ✗ Kustomize build failed:" -ForegroundColor Red
        Write-Host "    $buildOutput" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "  ✗ Local overlay not found at: $overlayPath" -ForegroundColor Red
    exit 1
}

# Test 3: Check Docker images
if (-not $SkipBuild) {
    Write-Host "`n[TEST 3] Checking Docker images..." -ForegroundColor Yellow
    
    $requiredImages = @(
        "neobank/identity-service:local",
        "neobank/ledger-service:local",
        "neobank/payment-service:local",
        "neobank/product-service:local",
        "neobank/card-service:local",
        "neobank/frontend:local"
    )
    
    $missingImages = @()
    foreach ($image in $requiredImages) {
        $exists = docker images --format "{{.Repository}}:{{.Tag}}" | Select-String -Pattern "^$([regex]::Escape($image))$" -Quiet
        if ($exists) {
            Write-Host "  ✓ $image" -ForegroundColor Green
        } else {
            Write-Host "  ✗ $image (missing)" -ForegroundColor Red
            $missingImages += $image
        }
    }
    
    if ($missingImages.Count -gt 0) {
        Write-Host "`n  Building missing images..." -ForegroundColor Yellow
        
        # Build images
        $imageMap = @{
            "neobank/identity-service:local" = "backend\identity-service"
            "neobank/ledger-service:local" = "backend\ledger-service"
            "neobank/payment-service:local" = "backend\payment-service"
            "neobank/product-service:local" = "backend\product-service"
            "neobank/card-service:local" = "backend\card-service"
            "neobank/frontend:local" = "frontend"
        }
        
        foreach ($img in $missingImages) {
            $path = $imageMap[$img]
            Write-Host "  Building $img from $path..." -ForegroundColor Gray
            docker build -t $img $path -q
            if ($LASTEXITCODE -eq 0) {
                Write-Host "    ✓ Built successfully" -ForegroundColor Green
            } else {
                Write-Host "    ✗ Build failed" -ForegroundColor Red
            }
        }
    }
}

# Test 4: Validate YAML files
Write-Host "`n[TEST 4] Validating YAML files..." -ForegroundColor Yellow

$yamlFiles = Get-ChildItem -Path "k8s\finops\*.yaml" -File
foreach ($file in $yamlFiles) {
    $content = Get-Content $file.FullName -Raw
    if ($content -match "apiVersion:" -and $content -match "kind:") {
        Write-Host "  ✓ $($file.Name)" -ForegroundColor Green
    } else {
        Write-Host "  ✗ $($file.Name) (invalid YAML)" -ForegroundColor Red
    }
}

# Test 5: Deploy (or dry-run)
if ($DryRun) {
    Write-Host "`n[TEST 5] Dry-run deployment..." -ForegroundColor Yellow
    kubectl apply -k $overlayPath --dry-run=client
    Write-Host "  ✓ Dry-run successful" -ForegroundColor Green
} else {
    Write-Host "`n[TEST 5] Deploying to cluster..." -ForegroundColor Yellow
    
    # Create namespaces first
    Write-Host "  Creating namespaces..." -ForegroundColor Gray
    kubectl create namespace neobank --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace finops --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace e2e-tests --dry-run=client -o yaml | kubectl apply -f -
    
    # Apply configurations
    Write-Host "  Applying configurations..." -ForegroundColor Gray
    kubectl apply -k $overlayPath
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✓ Deployment successful" -ForegroundColor Green
        
        # Wait a bit for pods to start
        Write-Host "`n  Waiting for pods to start..." -ForegroundColor Gray
        Start-Sleep -Seconds 10
        
        # Check pod status
        Write-Host "`n  Pod Status:" -ForegroundColor Yellow
        kubectl get pods -n neobank
        
        Write-Host "`n  Services:" -ForegroundColor Yellow
        kubectl get services -n neobank
        
        Write-Host "`n  FinOps Resources:" -ForegroundColor Yellow
        kubectl get resourcequotas,limitranges -n neobank
        
    } else {
        Write-Host "  ✗ Deployment failed" -ForegroundColor Red
        exit 1
    }
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Cyan

if (-not $DryRun) {
    Write-Host "`nNext steps:" -ForegroundColor Green
    Write-Host "1. Verify deployment: .\scripts\verify-deployment.ps1" -ForegroundColor White
    Write-Host "2. Set up port forwarding: .\scripts\setup-e2e.ps1 -PortForward" -ForegroundColor White
    Write-Host "3. Run tests: .\scripts\run-e2e-tests.ps1 -All" -ForegroundColor White
}
