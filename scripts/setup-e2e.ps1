# E2E Test Setup Script for Docker Desktop Kubernetes (PowerShell)
# This script sets up the local testing environment on Windows

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir

Write-Host "=== NeoBank E2E Test Setup ===" -ForegroundColor Cyan
Write-Host "Project Root: $ProjectRoot"
Write-Host ""

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
function Test-Prerequisites {
    Write-Info "Checking prerequisites..."
    
    # Check kubectl
    try {
        $kubectlVersion = kubectl version --client --short 2>$null
        if (-not $kubectlVersion) {
            $kubectlVersion = kubectl version --client
        }
        Write-Info "kubectl: $kubectlVersion"
    } catch {
        Write-Err "kubectl is not installed"
        exit 1
    }
    
    # Check Docker
    try {
        $dockerVersion = docker --version
        Write-Info "Docker: $dockerVersion"
    } catch {
        Write-Err "Docker is not installed"
        exit 1
    }
    
    # Check Kubernetes context
    $kubeContext = kubectl config current-context
    Write-Info "Kubernetes context: $kubeContext"
    
    if ($kubeContext -notmatch "docker-desktop|kind|minikube") {
        Write-Warn "Current context is not a local cluster. Proceed with caution."
        $response = Read-Host "Continue? (y/n)"
        if ($response -notmatch "^[Yy]") {
            exit 1
        }
    }
    
    # Check cluster connectivity
    try {
        kubectl cluster-info | Out-Null
        Write-Info "Cluster connectivity: OK"
    } catch {
        Write-Err "Cannot connect to Kubernetes cluster"
        exit 1
    }
}

# Install Istio
function Install-Istio {
    Write-Info "Installing Istio..."
    
    # Check if istioctl is available
    $istioctl = Get-Command istioctl -ErrorAction SilentlyContinue
    
    if (-not $istioctl) {
        Write-Info "Downloading istioctl..."
        # Download Istio for Windows
        $istioVersion = "1.20.0"
        $downloadUrl = "https://github.com/istio/istio/releases/download/$istioVersion/istioctl-$istioVersion-win.zip"
        $zipPath = "$env:TEMP\istioctl.zip"
        $extractPath = "$env:TEMP\istio"
        
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath
        Expand-Archive -Path $zipPath -DestinationPath $extractPath -Force
        
        # Add to PATH temporarily
        $env:PATH = "$extractPath;$env:PATH"
    }
    
    # Check if Istio is already installed
    $istioNs = kubectl get namespace istio-system -o name 2>$null
    if ($istioNs) {
        $istiod = kubectl get deployment istiod -n istio-system -o name 2>$null
        if ($istiod) {
            Write-Info "Istio is already installed"
            return
        }
    }
    
    # Install Istio with demo profile
    Write-Info "Installing Istio with demo profile..."
    istioctl install --set profile=demo -y
    
    Write-Info "Istio installation complete"
}

# Create namespaces
function New-Namespaces {
    Write-Info "Creating namespaces..."
    
    $namespaces = @("neobank", "finops", "velero", "e2e-tests")
    
    foreach ($ns in $namespaces) {
        $existing = kubectl get namespace $ns -o name 2>$null
        if (-not $existing) {
            kubectl create namespace $ns
            Write-Info "Created namespace: $ns"
        } else {
            Write-Info "Namespace exists: $ns"
        }
    }
    
    # Label for Istio injection
    kubectl label namespace neobank istio-injection=enabled --overwrite 2>$null
    kubectl label namespace e2e-tests istio-injection=enabled --overwrite 2>$null
}

# Build Docker images
function Build-Images {
    Write-Info "Building Docker images..."
    
    Set-Location $ProjectRoot
    
    # Build backend services
    $services = @("identity-service", "ledger-service", "payment-service", "product-service", "card-service")
    
    foreach ($service in $services) {
        Write-Info "Building $service..."
        docker build -t "neobank/${service}:latest" -f "backend/$service/Dockerfile" ./backend
    }
    
    # Build frontend
    Write-Info "Building frontend..."
    docker build -t "neobank/frontend:latest" -f "frontend/Dockerfile" ./frontend
    
    Write-Info "All images built successfully"
}

# Deploy infrastructure
function Deploy-Infrastructure {
    Write-Info "Deploying infrastructure..."
    
    # Apply base configurations
    kubectl apply -k "$ProjectRoot/k8s/base" 2>$null
    
    # Apply FinOps configurations
    kubectl apply -k "$ProjectRoot/k8s/finops" 2>$null
    
    # Apply DR configurations  
    kubectl apply -k "$ProjectRoot/k8s/dr" 2>$null
    
    # Apply Istio configurations
    kubectl apply -f "$ProjectRoot/k8s/istio/" 2>$null
    
    Write-Info "Infrastructure deployed"
}

# Wait for pods to be ready
function Wait-ForPods {
    Write-Info "Waiting for pods to be ready..."
    
    try {
        kubectl wait --for=condition=ready pod -l app -n neobank --timeout=300s
    } catch {
        Write-Warn "Some pods may not be ready. Checking status..."
        kubectl get pods -n neobank
    }
    
    try {
        kubectl wait --for=condition=ready pod -l app -n istio-system --timeout=300s
    } catch {
        Write-Warn "Istio pods may not be ready. Checking status..."
        kubectl get pods -n istio-system
    }
}

# Setup port forwarding
function Start-PortForwarding {
    Write-Info "Setting up port forwarding..."
    
    # Kill any existing port-forward processes
    Get-Process -Name kubectl -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
    
    # Port forward services (in background jobs)
    Start-Job -ScriptBlock { kubectl port-forward svc/neobank-frontend -n neobank 3001:3000 }
    Start-Job -ScriptBlock { kubectl port-forward svc/neobank-identity-service -n neobank 8081:8081 }
    Start-Job -ScriptBlock { kubectl port-forward svc/neobank-ledger-service -n neobank 8082:8082 }
    Start-Job -ScriptBlock { kubectl port-forward svc/neobank-payment-service -n neobank 8083:8083 }
    Start-Job -ScriptBlock { kubectl port-forward svc/neobank-product-service -n neobank 8084:8084 }
    Start-Job -ScriptBlock { kubectl port-forward svc/neobank-card-service -n neobank 8085:8085 }
    Start-Job -ScriptBlock { kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80 }
    
    Write-Info "Port forwarding established"
    Write-Info "Frontend: http://localhost:3001"
    Write-Info "Istio Gateway: http://localhost:8080"
}

# Main execution
function Main {
    param(
        [switch]$WithIstio,
        [switch]$Build,
        [switch]$PortForward
    )
    
    Test-Prerequisites
    New-Namespaces
    
    if ($WithIstio) {
        Install-Istio
    }
    
    if ($Build) {
        Build-Images
    }
    
    Deploy-Infrastructure
    Wait-ForPods
    
    if ($PortForward) {
        Start-PortForwarding
    }
    
    Write-Info "=== Setup Complete ==="
    Write-Host ""
    Write-Info "To run E2E tests:"
    Write-Info "  cd $ProjectRoot\tests\e2e"
    Write-Info "  go test -v ./..."
    Write-Host ""
    Write-Info "To view cluster status:"
    Write-Info "  kubectl get pods -n neobank"
    Write-Info "  kubectl get pods -n istio-system"
}

# Parse arguments and run
$params = @{}
if ($args -contains "-WithIstio" -or $args -contains "-i") { $params["WithIstio"] = $true }
if ($args -contains "-Build" -or $args -contains "-b") { $params["Build"] = $true }
if ($args -contains "-PortForward" -or $args -contains "-p") { $params["PortForward"] = $true }

Main @params
