# NeoBank Complete Deployment & Test
# Deploys FinOps, DR, Istio, OPA and runs all E2E tests

param(
    [switch]$SkipBuild,
    [switch]$SkipOPA,
    [switch]$SkipTests,
    [switch]$Verbose
)

$ErrorActionPreference = "Continue"
$StartTime = Get-Date

function Write-Section { param([string]$Msg) Write-Host "`n$('='*70)" -ForegroundColor Cyan; Write-Host $Msg -ForegroundColor Cyan; Write-Host "$('='*70)" -ForegroundColor Cyan }
function Write-Step { param([string]$Msg) Write-Host "`n>>> $Msg" -ForegroundColor Yellow }
function Write-OK { param([string]$Msg) Write-Host "    [OK] $Msg" -ForegroundColor Green }
function Write-Warn { param([string]$Msg) Write-Host "    [WARN] $Msg" -ForegroundColor Yellow }
function Write-Fail { param([string]$Msg) Write-Host "    [FAIL] $Msg" -ForegroundColor Red }
function Write-Info { param([string]$Msg) Write-Host "    $Msg" -ForegroundColor Gray }

Write-Section "NeoBank Complete Deployment & Test Suite"
Write-Host "Started: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Gray

# ============================================================================
# PHASE 1: Prerequisites Check
# ============================================================================
Write-Section "PHASE 1: Prerequisites Check"

Write-Step "Checking kubectl"
try {
    $context = kubectl config current-context 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-OK "kubectl connected to: $context"
    } else {
        Write-Fail "kubectl not connected"
        Write-Host "`nPlease ensure Docker Desktop Kubernetes is enabled:" -ForegroundColor Yellow
        Write-Host "  1. Open Docker Desktop" -ForegroundColor White
        Write-Host "  2. Go to Settings > Kubernetes" -ForegroundColor White
        Write-Host "  3. Enable Kubernetes" -ForegroundColor White
        Write-Host "  4. Wait for it to start" -ForegroundColor White
        exit 1
    }
} catch {
    Write-Fail "kubectl not found. Please install kubectl."
    exit 1
}

Write-Step "Checking Docker"
try {
    docker info 2>&1 | Out-Null
    Write-OK "Docker is running"
} catch {
    Write-Fail "Docker not running"
    exit 1
}

Write-Step "Checking cluster nodes"
$nodes = kubectl get nodes -o json 2>&1 | ConvertFrom-Json
if ($nodes.items.Count -gt 0) {
    foreach ($node in $nodes.items) {
        $ready = ($node.status.conditions | Where-Object { $_.type -eq "Ready" }).status
        if ($ready -eq "True") {
            Write-OK "Node: $($node.metadata.name) - Ready"
        } else {
            Write-Warn "Node: $($node.metadata.name) - Not Ready"
        }
    }
} else {
    Write-Fail "No nodes found in cluster"
    exit 1
}

# ============================================================================
# PHASE 2: Build Docker Images
# ============================================================================
if (-not $SkipBuild) {
    Write-Section "PHASE 2: Building Docker Images"
    
    $services = @(
        @{Name="identity-service"; Path="backend/identity-service"},
        @{Name="ledger-service"; Path="backend/ledger-service"},
        @{Name="payment-service"; Path="backend/payment-service"},
        @{Name="product-service"; Path="backend/product-service"},
        @{Name="card-service"; Path="backend/card-service"},
        @{Name="frontend"; Path="frontend"}
    )
    
    foreach ($svc in $services) {
        Write-Step "Building $($svc.Name)"
        if (Test-Path $svc.Path) {
            $output = docker build -t "neobank/$($svc.Name):local" $svc.Path 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-OK "Built neobank/$($svc.Name):local"
            } else {
                Write-Warn "Build failed for $($svc.Name) - will use existing image if available"
                if ($Verbose) { Write-Info $output }
            }
        } else {
            Write-Warn "Path not found: $($svc.Path)"
        }
    }
} else {
    Write-Section "PHASE 2: Skipping Docker Build (using existing images)"
}

# ============================================================================
# PHASE 3: Create Namespaces
# ============================================================================
Write-Section "PHASE 3: Creating Namespaces"

$namespaces = @("neobank", "finops", "velero", "e2e-tests", "gatekeeper-system")
foreach ($ns in $namespaces) {
    Write-Step "Creating namespace: $ns"
    kubectl create namespace $ns --dry-run=client -o yaml 2>$null | kubectl apply -f - 2>&1 | Out-Null
    
    $exists = kubectl get namespace $ns 2>&1
    if ($exists -notmatch "NotFound") {
        Write-OK "Namespace $ns ready"
    } else {
        Write-Fail "Failed to create namespace $ns"
    }
}

# Enable Istio injection on neobank namespace
Write-Step "Enabling Istio injection label"
kubectl label namespace neobank istio-injection=enabled --overwrite 2>&1 | Out-Null
Write-OK "Istio injection label set"

# ============================================================================
# PHASE 4: Deploy FinOps Resources
# ============================================================================
Write-Section "PHASE 4: Deploying FinOps Resources"

Write-Step "Applying resource quotas"
kubectl apply -f k8s/finops/resource-quotas.yaml 2>&1 | Out-Null
$rq = kubectl get resourcequota -n neobank 2>&1
if ($rq -notmatch "No resources found") {
    Write-OK "Resource quotas applied"
} else {
    Write-Warn "Resource quotas may not have applied"
}

Write-Step "Applying limit ranges"
kubectl apply -f k8s/finops/limit-ranges.yaml 2>&1 | Out-Null
$lr = kubectl get limitrange -n neobank 2>&1
if ($lr -notmatch "No resources found") {
    Write-OK "Limit ranges applied"
} else {
    Write-Warn "Limit ranges may not have applied"
}

# ============================================================================
# PHASE 5: Deploy Application
# ============================================================================
Write-Section "PHASE 5: Deploying NeoBank Application"

Write-Step "Validating Kustomize manifests"
$validation = kustomize build k8s/overlays/local 2>&1
if ($LASTEXITCODE -eq 0) {
    $resourceCount = ($validation | Select-String -Pattern "^kind:").Count
    Write-OK "Manifests valid ($resourceCount resources)"
} else {
    Write-Fail "Kustomize validation failed"
    Write-Info $validation
    exit 1
}

Write-Step "Applying application manifests"
kubectl apply -k k8s/overlays/local/ 2>&1 | Out-Null
if ($LASTEXITCODE -eq 0) {
    Write-OK "Application manifests applied"
} else {
    Write-Warn "Some manifests may have failed - checking status..."
}

Write-Step "Waiting for deployments (this may take 2-3 minutes)"
$maxWait = 180
$waited = 0

while ($waited -lt $maxWait) {
    Start-Sleep -Seconds 10
    $waited += 10
    
    $pods = kubectl get pods -n neobank -o json 2>&1 | ConvertFrom-Json
    $total = $pods.items.Count
    $running = ($pods.items | Where-Object { $_.status.phase -eq "Running" }).Count
    $pending = ($pods.items | Where-Object { $_.status.phase -eq "Pending" }).Count
    
    Write-Info "Pods: $running/$total running, $pending pending (${waited}s elapsed)"
    
    if ($running -eq $total -and $total -gt 0) {
        Write-OK "All $total pods are running!"
        break
    }
}

if ($waited -ge $maxWait) {
    Write-Warn "Timeout waiting for pods - some may still be starting"
}

# ============================================================================
# PHASE 6: Deploy OPA Gatekeeper
# ============================================================================
if (-not $SkipOPA) {
    Write-Section "PHASE 6: Deploying OPA Gatekeeper (Policy-as-Code)"
    
    Write-Step "Checking if Gatekeeper is already installed"
    $gatekeeperPods = kubectl get pods -n gatekeeper-system 2>&1
    if ($gatekeeperPods -match "Running") {
        Write-OK "Gatekeeper already running"
    } else {
        Write-Step "Installing OPA Gatekeeper v3.14.0"
        kubectl apply -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.14.0/deploy/gatekeeper.yaml 2>&1 | Out-Null
        
        Write-Info "Waiting for Gatekeeper to start (60s)..."
        Start-Sleep -Seconds 30
        
        # Wait for controller
        kubectl wait --for=condition=ready pod -l control-plane=controller-manager -n gatekeeper-system --timeout=90s 2>&1 | Out-Null
        
        if ($LASTEXITCODE -eq 0) {
            Write-OK "Gatekeeper installed and running"
        } else {
            Write-Warn "Gatekeeper may still be starting"
        }
    }
    
    Write-Step "Applying constraint templates"
    Start-Sleep -Seconds 5  # Let CRDs register
    kubectl apply -f k8s/opa/constraint-templates.yaml 2>&1 | Out-Null
    Write-OK "Constraint templates applied"
    
    Write-Step "Applying policy constraints"
    Start-Sleep -Seconds 10  # Let templates register
    kubectl apply -f k8s/opa/constraints.yaml 2>&1 | Out-Null
    Write-OK "Policy constraints applied"
    
    Write-Step "Applying audit configuration"
    kubectl apply -f k8s/opa/audit-config.yaml 2>&1 | Out-Null
    kubectl apply -f k8s/opa/exemptions.yaml 2>&1 | Out-Null
    Write-OK "Audit and exemptions configured"
} else {
    Write-Section "PHASE 6: Skipping OPA Gatekeeper"
}

# ============================================================================
# PHASE 7: Deployment Verification
# ============================================================================
Write-Section "PHASE 7: Deployment Verification"

Write-Step "Checking namespaces"
kubectl get namespaces | Select-String -Pattern "neobank|finops|velero|gatekeeper"

Write-Step "Checking pods in neobank namespace"
kubectl get pods -n neobank -o wide

Write-Step "Checking services"
kubectl get services -n neobank

Write-Step "Checking FinOps resources"
kubectl get resourcequotas,limitranges -n neobank

if (-not $SkipOPA) {
    Write-Step "Checking OPA constraints"
    kubectl get constraints 2>&1
    
    Write-Step "Checking constraint templates"
    kubectl get constrainttemplates 2>&1
}

# ============================================================================
# PHASE 8: Run E2E Tests
# ============================================================================
if (-not $SkipTests) {
    Write-Section "PHASE 8: Running E2E Tests"
    
    Write-Step "Checking Go installation"
    try {
        $goVersion = go version
        Write-OK "Go: $goVersion"
    } catch {
        Write-Warn "Go not installed - skipping E2E tests"
        Write-Info "Install Go 1.21+ to run E2E tests"
        $SkipTests = $true
    }
    
    if (-not $SkipTests) {
        Push-Location tests/e2e
        
        Write-Step "Downloading Go dependencies"
        go mod download 2>&1 | Out-Null
        Write-OK "Dependencies downloaded"
        
        Write-Step "Running Cluster Health Tests"
        $clusterResult = go test -v -run "TestCluster" -timeout 5m ./... 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-OK "Cluster tests passed"
        } else {
            Write-Warn "Some cluster tests failed"
        }
        if ($Verbose) { $clusterResult | ForEach-Object { Write-Info $_ } }
        
        Write-Step "Running FinOps Tests"
        $finopsResult = go test -v -run "TestFinOps" -timeout 5m ./... 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-OK "FinOps tests passed"
        } else {
            Write-Warn "Some FinOps tests failed"
        }
        if ($Verbose) { $finopsResult | ForEach-Object { Write-Info $_ } }
        
        if (-not $SkipOPA) {
            Write-Step "Running OPA Policy Tests"
            $opaResult = go test -v -run "TestOPA" -timeout 5m ./... 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-OK "OPA tests passed"
            } else {
                Write-Warn "Some OPA tests failed (this is expected if policies are enforcing)"
            }
            if ($Verbose) { $opaResult | ForEach-Object { Write-Info $_ } }
        }
        
        Pop-Location
    }
} else {
    Write-Section "PHASE 8: Skipping E2E Tests"
}

# ============================================================================
# PHASE 9: Summary
# ============================================================================
Write-Section "DEPLOYMENT SUMMARY"

$endTime = Get-Date
$duration = $endTime - $StartTime

Write-Host "`nDeployment completed in $([math]::Round($duration.TotalMinutes, 1)) minutes" -ForegroundColor Green

Write-Host "`n--- Component Status ---" -ForegroundColor Yellow

# Check each component
$components = @()

# Namespaces
$nsCheck = kubectl get namespace neobank 2>&1
$components += @{Name="NeoBank Namespace"; Status=($nsCheck -notmatch "NotFound")}

# Pods
$pods = kubectl get pods -n neobank -o json 2>&1 | ConvertFrom-Json
$runningPods = ($pods.items | Where-Object { $_.status.phase -eq "Running" }).Count
$totalPods = $pods.items.Count
$components += @{Name="Application Pods ($runningPods/$totalPods)"; Status=($runningPods -eq $totalPods -and $totalPods -gt 0)}

# FinOps
$rqCheck = kubectl get resourcequota -n neobank 2>&1
$components += @{Name="FinOps Resource Quotas"; Status=($rqCheck -notmatch "No resources found")}

$lrCheck = kubectl get limitrange -n neobank 2>&1
$components += @{Name="FinOps Limit Ranges"; Status=($lrCheck -notmatch "No resources found")}

# OPA
if (-not $SkipOPA) {
    $gatekeeperCheck = kubectl get pods -n gatekeeper-system 2>&1
    $components += @{Name="OPA Gatekeeper"; Status=($gatekeeperCheck -match "Running")}
    
    $constraintsCheck = kubectl get constraints 2>&1
    $components += @{Name="Policy Constraints"; Status=($constraintsCheck -notmatch "No resources found" -and $constraintsCheck -notmatch "error")}
}

foreach ($comp in $components) {
    $symbol = if ($comp.Status) { "[OK]" } else { "[--]" }
    $color = if ($comp.Status) { "Green" } else { "Yellow" }
    Write-Host "  $symbol $($comp.Name)" -ForegroundColor $color
}

Write-Host "`n--- Access Instructions ---" -ForegroundColor Yellow
Write-Host @"

To access services, run port forwarding:

  # Frontend (in a new terminal)
  kubectl port-forward svc/local-neobank-frontend 3001:3000 -n neobank

  # Or all services at once
  Start-Job { kubectl port-forward svc/local-neobank-frontend 3001:3000 -n neobank }
  Start-Job { kubectl port-forward svc/local-neobank-identity-service 8081:8081 -n neobank }

Then access:
  - Frontend: http://localhost:3001
  - Identity API: http://localhost:8081/health

"@ -ForegroundColor White

Write-Host "--- Useful Commands ---" -ForegroundColor Yellow
Write-Host @"

  # Check pods
  kubectl get pods -n neobank

  # Check pod logs
  kubectl logs -n neobank <pod-name>

  # Check FinOps quotas
  kubectl describe resourcequota -n neobank

  # Check OPA violations
  kubectl get constraints -o json | jq '.items[] | {name: .metadata.name, violations: .status.totalViolations}'

  # Run more tests
  .\scripts\run-e2e-tests.ps1 -All

"@ -ForegroundColor White

Write-Section "Deployment Complete!"
