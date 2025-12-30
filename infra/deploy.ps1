# NeoBank AWS Deployment Script for Windows
# ============================================================================
# Usage: .\deploy.ps1 -Action <plan|apply|destroy|status> [-Environment <dev|prod>]
# ============================================================================

param(
    [Parameter(Mandatory = $false)]
    [ValidateSet("plan", "apply", "destroy", "status", "init", "setup-aws")]
    [string]$Action = "plan",
    
    [Parameter(Mandatory = $false)]
    [ValidateSet("dev", "staging", "prod")]
    [string]$Environment = "dev",
    
    [Parameter(Mandatory = $false)]
    [string]$AWSRegion = "us-east-1"
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$TerraformDir = Join-Path $ScriptDir "terraform"
$K8sDir = Join-Path (Split-Path -Parent $ScriptDir) "k8s\overlays\aws"

function Write-Info { 
    Write-Host "[INFO] $args" -ForegroundColor Green 
}

function Write-Warning { 
    Write-Host "[WARN] $args" -ForegroundColor Yellow 
}

function Write-Error { 
    Write-Host "[ERROR] $args" -ForegroundColor Red 
}

function Check-Prerequisites {
    Write-Info "Checking prerequisites..."
    $missing = @()
    
    if (!(Get-Command terraform -ErrorAction SilentlyContinue)) {
        $missing += "terraform (https://developer.hashicorp.com/terraform/downloads)"
    }
    
    if (!(Get-Command aws -ErrorAction SilentlyContinue)) {
        $missing += "aws-cli (https://aws.amazon.com/cli/)"
    }
    
    if (!(Get-Command kubectl -ErrorAction SilentlyContinue)) {
        $missing += "kubectl (https://kubernetes.io/docs/tasks/tools/)"
    }
    
    if ($missing.Count -gt 0) {
        Write-Error "Missing required tools:"
        $missing | ForEach-Object { Write-Host "  - $_" -ForegroundColor Red }
        Write-Host ""
        Write-Host "Installation options:" -ForegroundColor Yellow
        Write-Host "  winget install Hashicorp.Terraform"
        Write-Host "  winget install Amazon.AWSCLI"
        Write-Host "  winget install Kubernetes.kubectl"
        Write-Host ""
        Write-Host "Or using Chocolatey:"
        Write-Host "  choco install terraform awscli kubectl"
        return $false
    }
    
    return $true
}

function Check-AWSCredentials {
    Write-Info "Checking AWS credentials..."
    
    try {
        $identity = aws sts get-caller-identity 2>&1
        if ($LASTEXITCODE -ne 0) {
            throw "AWS credentials not configured"
        }
        Write-Info "AWS credentials configured successfully"
        Write-Host $identity | ConvertFrom-Json | Format-List
        return $true
    }
    catch {
        Write-Warning "AWS credentials not found or invalid"
        Write-Host ""
        Write-Host "To configure AWS credentials, use one of these methods:" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Method 1 - AWS CLI configure:" -ForegroundColor Cyan
        Write-Host "  aws configure"
        Write-Host ""
        Write-Host "Method 2 - Environment variables:" -ForegroundColor Cyan
        Write-Host '  $env:AWS_ACCESS_KEY_ID = "your-access-key"'
        Write-Host '  $env:AWS_SECRET_ACCESS_KEY = "your-secret-key"'
        Write-Host '  $env:AWS_DEFAULT_REGION = "us-east-1"'
        Write-Host ""
        Write-Host "Method 3 - AWS SSO (recommended for organizations):" -ForegroundColor Cyan
        Write-Host "  aws configure sso"
        Write-Host "  aws sso login --profile your-profile"
        Write-Host ""
        return $false
    }
}

function Setup-AWS {
    Write-Info "Setting up AWS credentials..."
    
    # Check if winget is available
    if (Get-Command winget -ErrorAction SilentlyContinue) {
        Write-Info "Installing AWS CLI via winget..."
        winget install Amazon.AWSCLI --silent
    }
    else {
        Write-Warning "Please install AWS CLI manually from: https://aws.amazon.com/cli/"
    }
    
    Write-Host ""
    Write-Host "After installing AWS CLI, configure your credentials:" -ForegroundColor Yellow
    Write-Host "  aws configure"
    Write-Host ""
    Write-Host "You will need:" -ForegroundColor Cyan
    Write-Host "  - AWS Access Key ID"
    Write-Host "  - AWS Secret Access Key"
    Write-Host "  - Default region (e.g., us-east-1)"
    Write-Host "  - Default output format (json)"
}

function Run-TerraformInit {
    Write-Info "Initializing Terraform..."
    Push-Location $TerraformDir
    try {
        terraform init
        if ($LASTEXITCODE -ne 0) { throw "Terraform init failed" }
    }
    finally {
        Pop-Location
    }
}

function Run-TerraformPlan {
    Write-Info "Running Terraform plan for $Environment environment..."
    Push-Location $TerraformDir
    try {
        $tfvarsFile = "environments\$Environment.tfvars"
        if (!(Test-Path $tfvarsFile)) {
            Write-Error "Environment file not found: $tfvarsFile"
            return
        }
        terraform plan -var-file="$tfvarsFile" -out="tfplan-$Environment"
        if ($LASTEXITCODE -ne 0) { throw "Terraform plan failed" }
        Write-Info "Plan saved to tfplan-$Environment"
    }
    finally {
        Pop-Location
    }
}

function Run-TerraformApply {
    Write-Info "Applying Terraform plan for $Environment environment..."
    Write-Warning "This will create/modify AWS resources and may incur costs!"
    
    $confirm = Read-Host "Are you sure you want to proceed? (yes/no)"
    if ($confirm -ne "yes") {
        Write-Info "Aborted."
        return
    }
    
    Push-Location $TerraformDir
    try {
        $planFile = "tfplan-$Environment"
        if (Test-Path $planFile) {
            terraform apply $planFile
        }
        else {
            $tfvarsFile = "environments\$Environment.tfvars"
            terraform apply -var-file="$tfvarsFile" -auto-approve
        }
        
        if ($LASTEXITCODE -ne 0) { throw "Terraform apply failed" }
        
        Write-Info "Infrastructure deployed successfully!"
        Write-Host ""
        Write-Info "To configure kubectl, run:"
        Write-Host "  aws eks update-kubeconfig --name neobank-$Environment-eks --region $AWSRegion"
    }
    finally {
        Pop-Location
    }
}

function Run-TerraformDestroy {
    Write-Error "WARNING: This will DESTROY all AWS resources!"
    Write-Warning "This action cannot be undone!"
    
    $confirm = Read-Host "Type 'destroy' to confirm destruction"
    if ($confirm -ne "destroy") {
        Write-Info "Aborted."
        return
    }
    
    Push-Location $TerraformDir
    try {
        $tfvarsFile = "environments\$Environment.tfvars"
        terraform destroy -var-file="$tfvarsFile"
        if ($LASTEXITCODE -ne 0) { throw "Terraform destroy failed" }
    }
    finally {
        Pop-Location
    }
}

function Show-Status {
    Write-Info "Current deployment status:"
    
    Push-Location $TerraformDir
    try {
        Write-Host ""
        Write-Host "Terraform State:" -ForegroundColor Cyan
        terraform show -json 2>$null | ConvertFrom-Json | Select-Object -ExpandProperty values -ErrorAction SilentlyContinue
        
        Write-Host ""
        Write-Host "Outputs:" -ForegroundColor Cyan
        terraform output 2>$null
    }
    finally {
        Pop-Location
    }
}

# Main execution
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "NeoBank AWS Infrastructure Deployment" -ForegroundColor Cyan  
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Environment: $Environment" -ForegroundColor Yellow
Write-Host "AWS Region:  $AWSRegion" -ForegroundColor Yellow
Write-Host "Action:      $Action" -ForegroundColor Yellow
Write-Host ""

switch ($Action) {
    "setup-aws" {
        Setup-AWS
    }
    "init" {
        if (!(Check-Prerequisites)) { exit 1 }
        if (!(Check-AWSCredentials)) { exit 1 }
        Run-TerraformInit
    }
    "plan" {
        if (!(Check-Prerequisites)) { exit 1 }
        if (!(Check-AWSCredentials)) { exit 1 }
        Run-TerraformInit
        Run-TerraformPlan
    }
    "apply" {
        if (!(Check-Prerequisites)) { exit 1 }
        if (!(Check-AWSCredentials)) { exit 1 }
        Run-TerraformInit
        Run-TerraformApply
    }
    "destroy" {
        if (!(Check-Prerequisites)) { exit 1 }
        if (!(Check-AWSCredentials)) { exit 1 }
        Run-TerraformDestroy
    }
    "status" {
        Show-Status
    }
}
