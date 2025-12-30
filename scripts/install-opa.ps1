# Install OPA Gatekeeper for Policy-as-Code
# Deploys Gatekeeper and NeoBank policy configurations

param(
    [switch]$SkipGatekeeper,
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

function Write-Step { param([string]$Msg) Write-Host "`n>>> $Msg" -ForegroundColor Cyan }
function Write-OK { param([string]$Msg) Write-Host "    ✓ $Msg" -ForegroundColor Green }
function Write-Warn { param([string]$Msg) Write-Host "    ⚠ $Msg" -ForegroundColor Yellow }
function Write-Fail { param([string]$Msg) Write-Host "    ✗ $Msg" -ForegroundColor Red }

Write-Host "`n$('='*70)" -ForegroundColor Cyan
Write-Host "OPA Gatekeeper Installation" -ForegroundColor Cyan
Write-Host $('='*70) -ForegroundColor Cyan

# Check prerequisites
Write-Step "Checking prerequisites"
try {
    kubectl cluster-info | Out-Null
    Write-OK "Kubernetes cluster accessible"
} catch {
    Write-Fail "Cannot connect to Kubernetes cluster"
    exit 1
}

# Install Gatekeeper
if (-not $SkipGatekeeper) {
    Write-Step "Installing OPA Gatekeeper"
    
    # Check if already installed
    $existing = kubectl get namespace gatekeeper-system 2>&1
    if ($existing -notmatch "NotFound") {
        Write-OK "Gatekeeper namespace already exists"
        
        $pods = kubectl get pods -n gatekeeper-system -o jsonpath='{.items[*].status.phase}'
        if ($pods -match "Running") {
            Write-OK "Gatekeeper is already running"
        } else {
            Write-Warn "Gatekeeper installed but not running"
        }
    } else {
        Write-Host "    Installing Gatekeeper v3.14.0..." -ForegroundColor Gray
        
        if ($DryRun) {
            Write-OK "Would install Gatekeeper (dry-run)"
        } else {
            # Install Gatekeeper from official release
            kubectl apply -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.14.0/deploy/gatekeeper.yaml
            
            Write-Host "    Waiting for Gatekeeper to be ready..." -ForegroundColor Gray
            Start-Sleep -Seconds 10
            
            # Wait for controller manager
            kubectl wait --for=condition=ready pod -l control-plane=controller-manager -n gatekeeper-system --timeout=120s
            
            Write-OK "Gatekeeper installed successfully"
        }
    }
}

# Wait for CRDs to be available
Write-Step "Waiting for Gatekeeper CRDs"
Start-Sleep -Seconds 5

$crds = @(
    "constrainttemplates.templates.gatekeeper.sh",
    "configs.config.gatekeeper.sh"
)

foreach ($crd in $crds) {
    $exists = kubectl get crd $crd 2>&1
    if ($exists -notmatch "NotFound") {
        Write-OK "CRD: $crd"
    } else {
        Write-Warn "CRD not found: $crd (may need more time)"
    }
}

# Apply constraint templates
Write-Step "Applying constraint templates"
if ($DryRun) {
    kubectl apply -f k8s/opa/constraint-templates.yaml --dry-run=client
    Write-OK "Would apply constraint templates (dry-run)"
} else {
    kubectl apply -f k8s/opa/constraint-templates.yaml
    Write-OK "Constraint templates applied"
}

# Wait for templates to be ready
Write-Host "    Waiting for templates to be established..." -ForegroundColor Gray
Start-Sleep -Seconds 10

# Apply Gatekeeper config
Write-Step "Applying Gatekeeper configuration"
if ($DryRun) {
    kubectl apply -f k8s/opa/gatekeeper-config.yaml --dry-run=client
    Write-OK "Would apply Gatekeeper config (dry-run)"
} else {
    kubectl apply -f k8s/opa/gatekeeper-config.yaml 2>&1 | Out-Null
    Write-OK "Gatekeeper configuration applied"
}

# Apply constraints
Write-Step "Applying policy constraints"
if ($DryRun) {
    kubectl apply -f k8s/opa/constraints.yaml --dry-run=client
    Write-OK "Would apply constraints (dry-run)"
} else {
    # Give templates time to be ready
    Start-Sleep -Seconds 5
    kubectl apply -f k8s/opa/constraints.yaml
    Write-OK "Policy constraints applied"
}

# Apply audit config
Write-Step "Applying audit configuration"
if ($DryRun) {
    kubectl apply -f k8s/opa/audit-config.yaml --dry-run=client
    Write-OK "Would apply audit config (dry-run)"
} else {
    kubectl apply -f k8s/opa/audit-config.yaml 2>&1 | Out-Null
    Write-OK "Audit configuration applied"
}

# Apply exemptions
Write-Step "Applying policy exemptions"
if ($DryRun) {
    kubectl apply -f k8s/opa/exemptions.yaml --dry-run=client
    Write-OK "Would apply exemptions (dry-run)"
} else {
    kubectl apply -f k8s/opa/exemptions.yaml 2>&1 | Out-Null
    Write-OK "Exemptions applied"
}

# Verify installation
Write-Step "Verifying installation"

Write-Host "`n  Gatekeeper Pods:" -ForegroundColor Yellow
kubectl get pods -n gatekeeper-system

Write-Host "`n  Constraint Templates:" -ForegroundColor Yellow
kubectl get constrainttemplates

Write-Host "`n  Active Constraints:" -ForegroundColor Yellow
kubectl get constraints

# Summary
Write-Host "`n$('='*70)" -ForegroundColor Cyan
Write-Host "OPA Gatekeeper Installation Complete!" -ForegroundColor Green
Write-Host $('='*70) -ForegroundColor Cyan

Write-Host "`nPolicies Enforced:" -ForegroundColor Yellow
Write-Host "  • Privileged containers blocked (deny)" -ForegroundColor White
Write-Host "  • Host network access blocked (deny)" -ForegroundColor White
Write-Host "  • Resource limits required (deny)" -ForegroundColor White
Write-Host "  • Non-root containers required (deny)" -ForegroundColor White
Write-Host "  • NodePort services blocked (deny)" -ForegroundColor White
Write-Host "  • Required labels enforced (deny/warn)" -ForegroundColor White
Write-Host "  • Cost allocation labels (warn)" -ForegroundColor White
Write-Host "  • Allowed registries (warn)" -ForegroundColor White
Write-Host "  • Latest tag blocked (warn)" -ForegroundColor White
Write-Host "  • Health probes required (warn)" -ForegroundColor White
Write-Host "  • Read-only root filesystem (warn)" -ForegroundColor White

Write-Host "`nNext Steps:" -ForegroundColor Cyan
Write-Host "  1. Verify policies: kubectl get constraints" -ForegroundColor White
Write-Host "  2. Check violations: kubectl get constraints -o json | jq '.items[].status'" -ForegroundColor White
Write-Host "  3. Run tests: .\scripts\run-e2e-tests.ps1 -OPA" -ForegroundColor White
Write-Host "  4. View audit logs: kubectl logs -n gatekeeper-system -l control-plane=controller-manager`n" -ForegroundColor White
