# Verify NeoBank deployment in local Kubernetes cluster
# Checks FinOps, DR, and Istio configurations

$ErrorActionPreference = "Continue"

function Write-Check {
    param(
        [string]$Message,
        [bool]$Success,
        [string]$Details = ""
    )
    
    $symbol = if ($Success) { "✓" } else { "✗" }
    $color = if ($Success) { "Green" } else { "Red" }
    
    Write-Host "[$symbol] $Message" -ForegroundColor $color
    if ($Details) {
        Write-Host "    $Details" -ForegroundColor Gray
    }
}

function Test-Resource {
    param(
        [string]$ResourceType,
        [string]$Name,
        [string]$Namespace
    )
    
    try {
        $resource = kubectl get $ResourceType $Name -n $Namespace 2>&1
        return $resource -notmatch "NotFound"
    } catch {
        return $false
    }
}

Write-Host "`n" + ("=" * 70) -ForegroundColor Cyan
Write-Host "NeoBank Deployment Verification" -ForegroundColor Cyan
Write-Host ("=" * 70) + "`n" -ForegroundColor Cyan

# Check cluster connectivity
Write-Host "`n--- Cluster Connectivity ---" -ForegroundColor Yellow
try {
    $context = kubectl config current-context
    Write-Check "Kubernetes cluster accessible" $true "Context: $context"
} catch {
    Write-Check "Kubernetes cluster accessible" $false "Cannot connect to cluster"
    exit 1
}

# Check namespaces
Write-Host "`n--- Namespaces ---" -ForegroundColor Yellow
$requiredNamespaces = @("neobank", "finops", "e2e-tests")
foreach ($ns in $requiredNamespaces) {
    $exists = kubectl get namespace $ns 2>&1
    $success = $exists -notmatch "NotFound"
    Write-Check "Namespace: $ns" $success
}

# Check Istio
Write-Host "`n--- Istio Installation ---" -ForegroundColor Yellow
$istioNs = kubectl get namespace istio-system 2>&1
$istioInstalled = $istioNs -notmatch "NotFound"
Write-Check "Istio namespace (istio-system)" $istioInstalled

if ($istioInstalled) {
    $istiodPod = kubectl get pods -n istio-system -l app=istiod -o jsonpath='{.items[0].status.phase}' 2>$null
    $istiodRunning = $istiodPod -eq "Running"
    Write-Check "Istiod control plane" $istiodRunning "Status: $istiodPod"
    
    $gateway = kubectl get svc istio-ingressgateway -n istio-system 2>&1
    $gatewayExists = $gateway -notmatch "NotFound"
    Write-Check "Istio ingress gateway" $gatewayExists
}

# Check Velero
Write-Host "`n--- Velero (DR) Installation ---" -ForegroundColor Yellow
$veleroNs = kubectl get namespace velero 2>&1
$veleroInstalled = $veleroNs -notmatch "NotFound"
Write-Check "Velero namespace" $veleroInstalled

if ($veleroInstalled) {
    $veleroPod = kubectl get pods -n velero -l app.kubernetes.io/name=velero -o jsonpath='{.items[0].status.phase}' 2>$null
    $veleroRunning = $veleroPod -eq "Running"
    Write-Check "Velero pod" $veleroRunning "Status: $veleroPod"
    
    $schedules = kubectl get schedules -n velero 2>$null
    $schedulesExist = $schedules -match "neobank"
    Write-Check "Velero backup schedules" $schedulesExist
}

# Check FinOps resources
Write-Host "`n--- FinOps Resources ---" -ForegroundColor Yellow

$resourceQuota = kubectl get resourcequota -n neobank 2>&1
$quotaExists = $resourceQuota -notmatch "No resources found"
Write-Check "Resource quotas in neobank namespace" $quotaExists

$limitRange = kubectl get limitrange -n neobank 2>&1
$limitExists = $limitRange -notmatch "No resources found"
Write-Check "Limit ranges in neobank namespace" $limitExists

$kubecostNs = kubectl get namespace finops 2>&1
$kubecostNsExists = $kubecostNs -notmatch "NotFound"
Write-Check "FinOps namespace" $kubecostNsExists

# Check application deployments
Write-Host "`n--- Application Deployments ---" -ForegroundColor Yellow

$services = @(
    "neobank-identity-service",
    "neobank-ledger-service",
    "neobank-payment-service",
    "neobank-product-service",
    "neobank-card-service",
    "neobank-frontend"
)

foreach ($svc in $services) {
    # Check both with and without local- prefix
    $deployments = kubectl get deployment -n neobank 2>$null | Select-String -Pattern $svc
    $exists = $null -ne $deployments
    
    if ($exists) {
        $depName = ($deployments -split '\s+')[0]
        $ready = kubectl get deployment $depName -n neobank -o jsonpath='{.status.readyReplicas}' 2>$null
        $replicas = kubectl get deployment $depName -n neobank -o jsonpath='{.status.replicas}' 2>$null
        $success = $ready -eq $replicas -and $ready -gt 0
        Write-Check "Deployment: $svc" $success "Ready: $ready/$replicas"
    } else {
        Write-Check "Deployment: $svc" $false "Not found"
    }
}

# Check Istio configuration
if ($istioInstalled) {
    Write-Host "`n--- Istio Configuration ---" -ForegroundColor Yellow
    
    $gateway = kubectl get gateway -n neobank 2>&1
    $gatewayExists = $gateway -notmatch "No resources found"
    Write-Check "Istio Gateway" $gatewayExists
    
    $vs = kubectl get virtualservices -n neobank 2>&1
    $vsExists = $vs -notmatch "No resources found"
    Write-Check "Virtual Services" $vsExists
    
    $dr = kubectl get destinationrules -n neobank 2>&1
    $drExists = $dr -notmatch "No resources found"
    Write-Check "Destination Rules" $drExists
    
    $pa = kubectl get peerauthentication -n neobank 2>&1
    $paExists = $pa -notmatch "No resources found"
    Write-Check "Peer Authentication (mTLS)" $paExists
    
    $ap = kubectl get authorizationpolicies -n neobank 2>&1
    $apExists = $ap -notmatch "No resources found"
    Write-Check "Authorization Policies" $apExists
}

# Check pod status
Write-Host "`n--- Pod Status ---" -ForegroundColor Yellow

$allPods = kubectl get pods -n neobank -o json | ConvertFrom-Json
$totalPods = $allPods.items.Count
$runningPods = ($allPods.items | Where-Object { $_.status.phase -eq "Running" }).Count
$podsReady = $runningPods -eq $totalPods -and $totalPods -gt 0

Write-Check "All pods running" $podsReady "Running: $runningPods/$totalPods"

if (-not $podsReady) {
    Write-Host "`nPod details:" -ForegroundColor Gray
    kubectl get pods -n neobank
}

# Check services
Write-Host "`n--- Services ---" -ForegroundColor Yellow

$svcList = kubectl get services -n neobank -o json | ConvertFrom-Json
$svcCount = $svcList.items.Count
$svcExists = $svcCount -gt 0

Write-Check "Services created" $svcExists "Count: $svcCount"

# Summary
Write-Host "`n" + ("=" * 70) -ForegroundColor Cyan
Write-Host "Verification Summary" -ForegroundColor Cyan
Write-Host ("=" * 70) -ForegroundColor Cyan

$components = @{
    "Kubernetes Cluster" = $true
    "NeoBank Namespace" = (kubectl get namespace neobank 2>&1) -notmatch "NotFound"
    "FinOps Resources" = $quotaExists -and $limitExists
    "Istio Service Mesh" = $istioInstalled
    "Velero DR" = $veleroInstalled
    "Application Pods" = $podsReady
}

Write-Host ""
foreach ($comp in $components.GetEnumerator() | Sort-Object Name) {
    $symbol = if ($comp.Value) { "✓" } else { "✗" }
    $color = if ($comp.Value) { "Green" } else { "Yellow" }
    Write-Host "  [$symbol] $($comp.Key)" -ForegroundColor $color
}

Write-Host "`n--- Recommendations ---" -ForegroundColor Cyan

if (-not $istioInstalled) {
    Write-Host "• Install Istio: .\scripts\deploy-local.ps1 -InstallIstio -BuildImages" -ForegroundColor Yellow
}

if (-not $veleroInstalled) {
    Write-Host "• Install Velero: .\scripts\deploy-local.ps1 -InstallVelero" -ForegroundColor Yellow
}

if (-not $podsReady) {
    Write-Host "• Deploy application: kubectl apply -k k8s/overlays/local/" -ForegroundColor Yellow
    Write-Host "• Check pod logs: kubectl logs -n neobank <pod-name>" -ForegroundColor Yellow
}

if ($podsReady) {
    Write-Host "• Set up port forwarding: .\scripts\setup-e2e.ps1 -PortForward" -ForegroundColor Green
    Write-Host "• Run E2E tests: .\scripts\run-e2e-tests.ps1 -All" -ForegroundColor Green
}

Write-Host ""
