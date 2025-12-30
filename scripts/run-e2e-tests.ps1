# NeoBank E2E Test Runner for Docker Desktop Kubernetes
# Run comprehensive E2E tests for FinOps, DR, Istio, and OPA

param(
    [switch]$Setup,
    [switch]$All,
    [switch]$FinOps,
    [switch]$DR,
    [switch]$Istio,
    [switch]$Chaos,
    [switch]$Integration,
    [switch]$Cluster,
    [switch]$OPA,
    [switch]$Verbose,
    [int]$Timeout = 30
)

$ErrorActionPreference = "Continue"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$E2EDir = Join-Path $ScriptDir "..\tests\e2e"

function Write-Header {
    param([string]$Message)
    Write-Host ""
    Write-Host "=" * 60 -ForegroundColor Cyan
    Write-Host $Message -ForegroundColor Cyan
    Write-Host "=" * 60 -ForegroundColor Cyan
    Write-Host ""
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

function Test-Prerequisites {
    Write-Header "Checking Prerequisites"
    
    # Check Go
    try {
        $goVersion = go version
        Write-Info "Go: $goVersion"
    } catch {
        Write-Err "Go is not installed. Please install Go 1.21+"
        return $false
    }
    
    # Check kubectl
    try {
        $kubectlVersion = kubectl version --client --short 2>$null
        if (-not $kubectlVersion) {
            $kubectlVersion = kubectl version --client
        }
        Write-Info "kubectl: $kubectlVersion"
    } catch {
        Write-Err "kubectl is not installed"
        return $false
    }
    
    # Check cluster connectivity
    try {
        kubectl cluster-info 2>$null | Out-Null
        Write-Info "Cluster connectivity: OK"
    } catch {
        Write-Err "Cannot connect to Kubernetes cluster"
        return $false
    }
    
    # Check context
    $context = kubectl config current-context
    Write-Info "Current context: $context"
    
    return $true
}

function Setup-Environment {
    Write-Header "Setting Up Test Environment"
    
    # Ensure namespaces exist
    $namespaces = @("neobank", "finops", "velero", "e2e-tests")
    foreach ($ns in $namespaces) {
        $exists = kubectl get namespace $ns 2>$null
        if (-not $exists) {
            Write-Info "Creating namespace: $ns"
            kubectl create namespace $ns
        } else {
            Write-Info "Namespace exists: $ns"
        }
    }
    
    # Apply configurations
    Write-Info "Applying base configurations..."
    kubectl apply -k "$ScriptDir\..\k8s\base" 2>$null
    
    Write-Info "Applying FinOps configurations..."
    kubectl apply -k "$ScriptDir\..\k8s\finops" 2>$null
    
    Write-Info "Applying DR configurations..."
    kubectl apply -k "$ScriptDir\..\k8s\dr" 2>$null
    
    Write-Info "Applying Istio configurations..."
    kubectl apply -f "$ScriptDir\..\k8s\istio\" 2>$null
    
    # Wait for pods
    Write-Info "Waiting for pods to be ready..."
    kubectl wait --for=condition=ready pod -l app -n neobank --timeout=120s 2>$null
    
    Write-Info "Environment setup complete"
}

function Start-PortForwarding {
    Write-Header "Starting Port Forwarding"
    
    # Kill existing port forwards
    Get-Process -Name kubectl -ErrorAction SilentlyContinue | 
        Where-Object { $_.CommandLine -like "*port-forward*" } | 
        Stop-Process -Force -ErrorAction SilentlyContinue
    
    # Start port forwarding in background
    $jobs = @()
    
    $portForwards = @(
        @{Service = "neobank-frontend"; Namespace = "neobank"; Ports = "3001:3000"},
        @{Service = "neobank-identity-service"; Namespace = "neobank"; Ports = "8081:8081"},
        @{Service = "neobank-ledger-service"; Namespace = "neobank"; Ports = "8082:8082"},
        @{Service = "neobank-payment-service"; Namespace = "neobank"; Ports = "8083:8083"},
        @{Service = "neobank-product-service"; Namespace = "neobank"; Ports = "8084:8084"},
        @{Service = "neobank-card-service"; Namespace = "neobank"; Ports = "8085:8085"},
        @{Service = "istio-ingressgateway"; Namespace = "istio-system"; Ports = "8080:80"}
    )
    
    foreach ($pf in $portForwards) {
        $exists = kubectl get svc $pf.Service -n $pf.Namespace 2>$null
        if ($exists) {
            $job = Start-Job -ScriptBlock {
                param($svc, $ns, $ports)
                kubectl port-forward svc/$svc -n $ns $ports
            } -ArgumentList $pf.Service, $pf.Namespace, $pf.Ports
            $jobs += $job
            Write-Info "Port forwarding: $($pf.Service) ($($pf.Ports))"
        } else {
            Write-Warn "Service not found: $($pf.Service)"
        }
    }
    
    # Give port forwarding time to establish
    Start-Sleep -Seconds 3
    
    return $jobs
}

function Run-Tests {
    param(
        [string]$TestPattern = "",
        [string]$TestName = "All"
    )
    
    Write-Header "Running $TestName Tests"
    
    Push-Location $E2EDir
    
    try {
        # Install dependencies
        Write-Info "Installing Go dependencies..."
        go mod download 2>$null
        
        # Build tests
        Write-Info "Building tests..."
        go build -v ./... 2>$null
        
        # Run tests
        Write-Info "Executing tests..."
        
        $testArgs = @("-v", "-timeout", "${Timeout}m")
        if ($TestPattern) {
            $testArgs += @("-run", $TestPattern)
        }
        $testArgs += "./..."
        
        if ($Verbose) {
            go test @testArgs
        } else {
            go test @testArgs 2>&1 | ForEach-Object {
                if ($_ -match "^(PASS|FAIL|---)" -or $_ -match "^\s+(✓|○|⚠)") {
                    Write-Host $_
                } elseif ($_ -match "^ok|^FAIL" -or $_ -match "coverage:") {
                    Write-Host $_ -ForegroundColor $(if ($_ -match "^ok") { "Green" } else { "Red" })
                }
            }
        }
        
        $exitCode = $LASTEXITCODE
        return $exitCode
    }
    finally {
        Pop-Location
    }
}

function Show-TestSummary {
    param([int]$ExitCode)
    
    Write-Header "Test Summary"
    
    if ($ExitCode -eq 0) {
        Write-Host "All tests passed!" -ForegroundColor Green
    } else {
        Write-Host "Some tests failed (exit code: $ExitCode)" -ForegroundColor Red
    }
    
    Write-Host ""
    Write-Host "Available test commands:"
    Write-Host "  .\run-e2e-tests.ps1 -All        # Run all tests"
    Write-Host "  .\run-e2e-tests.ps1 -FinOps     # Run FinOps tests"
    Write-Host "  .\run-e2e-tests.ps1 -DR         # Run DR tests"
    Write-Host "  .\run-e2e-tests.ps1 -Istio      # Run Istio tests"
    Write-Host "  .\run-e2e-tests.ps1 -Chaos      # Run Chaos tests"
    Write-Host "  .\run-e2e-tests.ps1 -Integration # Run Integration tests"
    Write-Host "  .\run-e2e-tests.ps1 -Cluster    # Run Cluster tests"
    Write-Host "  .\run-e2e-tests.ps1 -OPA        # Run OPA/Policy tests"
}

# Main execution
Write-Header "NeoBank E2E Test Runner"

if (-not (Test-Prerequisites)) {
    exit 1
}

if ($Setup) {
    Setup-Environment
}

$jobs = Start-PortForwarding

try {
    $exitCode = 0
    
    if ($All -or (-not $FinOps -and -not $DR -and -not $Istio -and -not $Chaos -and -not $Integration -and -not $Cluster -and -not $OPA)) {
        $result = Run-Tests -TestName "All"
        if ($result -ne 0) { $exitCode = $result }
    } else {
        if ($FinOps) {
            $result = Run-Tests -TestPattern "TestFinOps" -TestName "FinOps"
            if ($result -ne 0) { $exitCode = $result }
        }
        if ($DR) {
            $result = Run-Tests -TestPattern "TestDR" -TestName "DR"
            if ($result -ne 0) { $exitCode = $result }
        }
        if ($Istio) {
            $result = Run-Tests -TestPattern "TestIstio" -TestName "Istio"
            if ($result -ne 0) { $exitCode = $result }
        }
        if ($Chaos) {
            $result = Run-Tests -TestPattern "TestChaos" -TestName "Chaos"
            if ($result -ne 0) { $exitCode = $result }
        }
        if ($Integration) {
            $result = Run-Tests -TestPattern "TestIntegration" -TestName "Integration"
            if ($result -ne 0) { $exitCode = $result }
        }
        if ($Cluster) {
            $result = Run-Tests -TestPattern "TestCluster" -TestName "Cluster"
            if ($result -ne 0) { $exitCode = $result }
        }
        if ($OPA) {
            $result = Run-Tests -TestPattern "TestOPA" -TestName "OPA/Policy"
            if ($result -ne 0) { $exitCode = $result }
        }
    }
    
    Show-TestSummary -ExitCode $exitCode
    exit $exitCode
}
finally {
    # Cleanup port forwarding jobs
    $jobs | Stop-Job -ErrorAction SilentlyContinue
    $jobs | Remove-Job -ErrorAction SilentlyContinue
}
