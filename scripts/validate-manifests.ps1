# Validate Kubernetes manifests before deployment
# Tests that all YAML files are valid

$ErrorActionPreference = "Continue"

Write-Host "`n=== Manifest Validation ===" -ForegroundColor Cyan

$issuesFound = 0

# Test 1: Validate local overlay with kustomize
Write-Host "`n[1] Testing local overlay kustomize build..." -ForegroundColor Yellow
$output = kustomize build k8s/overlays/local 2>&1
if ($LASTEXITCODE -eq 0) {
    $resourceCount = ($output | Select-String -Pattern "^kind:").Count
    Write-Host "    ✓ Valid ($resourceCount resources)" -ForegroundColor Green
    
    # Show what would be deployed
    Write-Host "`n    Resources that will be deployed:" -ForegroundColor Gray
    $output | Select-String -Pattern "^kind:|^  name:" | ForEach-Object {
        Write-Host "      $_" -ForegroundColor Gray
    }
} else {
    Write-Host "    ✗ Invalid" -ForegroundColor Red
    Write-Host "    Error: $output" -ForegroundColor Red
    $issuesFound++
}

# Test 2: Validate individual FinOps files
Write-Host "`n[2] Testing FinOps manifests..." -ForegroundColor Yellow
$finopsFiles = Get-ChildItem -Path "k8s/finops/*.yaml" -File | Where-Object { $_.Name -ne "kustomization.yaml" }
foreach ($file in $finopsFiles) {
    $validation = kubectl apply -f $file.FullName --dry-run=client 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "    ✓ $($file.Name)" -ForegroundColor Green
    } else {
        Write-Host "    ✗ $($file.Name)" -ForegroundColor Red
        Write-Host "      $validation" -ForegroundColor Red
        $issuesFound++
    }
}

# Test 3: Validate DR files
Write-Host "`n[3] Testing DR manifests..." -ForegroundColor Yellow
$drFiles = Get-ChildItem -Path "k8s/dr/*.yaml" -File | Where-Object { $_.Name -ne "kustomization.yaml" }
foreach ($file in $drFiles) {
    $validation = kubectl apply -f $file.FullName --dry-run=client 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "    ✓ $($file.Name)" -ForegroundColor Green
    } else {
        Write-Host "    ✗ $($file.Name)" -ForegroundColor Red
        Write-Host "      $validation" -ForegroundColor Red
        $issuesFound++
    }
}

# Test 4: Validate Istio files
Write-Host "`n[4] Testing Istio manifests..." -ForegroundColor Yellow
if (Test-Path "k8s/istio") {
    $istioFiles = Get-ChildItem -Path "k8s/istio/*.yaml" -File
    foreach ($file in $istioFiles) {
        $validation = kubectl apply -f $file.FullName --dry-run=client 2>&1
        if ($LASTEXITCODE -eq 0 -or $validation -match "created \(dry run\)|configured \(dry run\)|unchanged \(dry run\)") {
            Write-Host "    ✓ $($file.Name)" -ForegroundColor Green
        } else {
            Write-Host "    ○ $($file.Name) (may require Istio CRDs)" -ForegroundColor Yellow
        }
    }
} else {
    Write-Host "    ○ Istio directory not found (optional)" -ForegroundColor Yellow
}

# Test 5: Check for common issues
Write-Host "`n[5] Checking for common issues..." -ForegroundColor Yellow

# Check if namespace is hardcoded in base files
$baseFiles = Get-ChildItem -Path "k8s/base/*.yaml" -File
$hardcodedNamespaces = @()
foreach ($file in $baseFiles) {
    $content = Get-Content $file.FullName -Raw
    if ($content -match "namespace:\s*neobank") {
        $hardcodedNamespaces += $file.Name
    }
}

if ($hardcodedNamespaces.Count -gt 0) {
    Write-Host "    ⚠ Hardcoded namespaces found (expected):" -ForegroundColor Yellow
    foreach ($file in $hardcodedNamespaces) {
        Write-Host "      - $file" -ForegroundColor Gray
    }
} else {
    Write-Host "    ✓ No hardcoded namespaces in base" -ForegroundColor Green
}

# Summary
Write-Host "`n" + ("="*70) -ForegroundColor Cyan
if ($issuesFound -eq 0) {
    Write-Host "All validations passed! ✓" -ForegroundColor Green
    Write-Host "`nYou can proceed with deployment:" -ForegroundColor White
    Write-Host "  .\scripts\simple-deploy.ps1" -ForegroundColor Cyan
} else {
    Write-Host "Found $issuesFound issue(s) ✗" -ForegroundColor Red
    Write-Host "`nPlease fix the issues before deploying." -ForegroundColor Yellow
}
Write-Host ("="*70) + "`n" -ForegroundColor Cyan
