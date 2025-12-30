# AFT Deployment Automation Script
# This script helps automate the AFT deployment process after Control Tower is set up

[CmdletBinding()]
param(
    [Parameter(Mandatory = $false)]
    [string]$Phase = "check",
    
    [Parameter(Mandatory = $false)]
    [string]$LogArchiveAccountId,
    
    [Parameter(Mandatory = $false)]
    [string]$AuditAccountId
)

$ErrorActionPreference = "Stop"

function Write-Step {
    param([string]$Message)
    Write-Host "`n==> $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor Red
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠ $Message" -ForegroundColor Yellow
}

switch ($Phase) {
    "check" {
        Write-Step "Checking prerequisites for AFT deployment"
        
        # Check AWS CLI
        try {
            $awsVersion = aws --version
            Write-Success "AWS CLI installed: $awsVersion"
        }
        catch {
            Write-Error "AWS CLI not found. Please install it first."
            exit 1
        }
        
        # Check Terraform
        try {
            $tfVersion = terraform version -json | ConvertFrom-Json
            Write-Success "Terraform installed: $($tfVersion.terraform_version)"
        }
        catch {
            Write-Error "Terraform not found. Please install it first."
            exit 1
        }
        
        # Check AWS credentials
        try {
            $identity = aws sts get-caller-identity | ConvertFrom-Json
            Write-Success "AWS credentials configured"
            Write-Host "    Account: $($identity.Account)"
            Write-Host "    User: $($identity.Arn)"
        }
        catch {
            Write-Error "AWS credentials not configured. Run 'aws configure' first."
            exit 1
        }
        
        # Check Control Tower
        Write-Step "Checking Control Tower status"
        try {
            $landingZones = aws controltower list-landing-zones | ConvertFrom-Json
            if ($landingZones.landingZones.Count -gt 0) {
                Write-Success "Control Tower is deployed"
                
                # Get Log Archive account
                $logArchive = aws organizations list-accounts --query 'Accounts[?Name==`Log Archive`].Id' --output text
                if ($logArchive) {
                    Write-Success "Log Archive account found: $logArchive"
                }
                else {
                    Write-Warning "Log Archive account not found"
                }
                
                # Get Audit account
                $audit = aws organizations list-accounts --query 'Accounts[?Name==`Audit`].Id' --output text
                if ($audit) {
                    Write-Success "Audit account found: $audit"
                }
                else {
                    Write-Warning "Audit account not found"
                }
                
                Write-Host "`n✓ Ready to proceed with AFT deployment"
                Write-Host "`nNext step: Run this script with -Phase update-config -LogArchiveAccountId $logArchive -AuditAccountId $audit"
            }
            else {
                Write-Warning "Control Tower is NOT deployed"
                Write-Host "`nYou need to deploy Control Tower first:"
                Write-Host "1. Go to: https://us-east-1.console.aws.amazon.com/controltower/home"
                Write-Host "2. Click 'Set up landing zone'"
                Write-Host "3. Wait 45-60 minutes for deployment"
                Write-Host "4. Run this script again with -Phase check"
            }
        }
        catch {
            Write-Error "Error checking Control Tower: $_"
            exit 1
        }
    }
    
    "update-config" {
        if (-not $LogArchiveAccountId -or -not $AuditAccountId) {
            Write-Error "Both -LogArchiveAccountId and -AuditAccountId are required"
            Write-Host "Usage: .\deploy.ps1 -Phase update-config -LogArchiveAccountId 123456789012 -AuditAccountId 234567890123"
            exit 1
        }
        
        Write-Step "Updating terraform.tfvars with account IDs"
        
        $tfvarsPath = "terraform.tfvars"
        $content = Get-Content $tfvarsPath -Raw
        
        # Update account IDs
        $content = $content -replace 'log_archive_account_id\s*=\s*".*"', "log_archive_account_id = `"$LogArchiveAccountId`""
        $content = $content -replace 'audit_account_id\s*=\s*".*"', "audit_account_id       = `"$AuditAccountId`""
        
        Set-Content -Path $tfvarsPath -Value $content
        Write-Success "Updated terraform.tfvars"
        
        Write-Host "`nNext step: Review the file and run: .\deploy.ps1 -Phase deploy-aft"
    }
    
    "deploy-aft" {
        Write-Step "Deploying AFT infrastructure"
        
        # Verify config is updated
        $tfvars = Get-Content "terraform.tfvars" -Raw
        if ($tfvars -match '"PLACEHOLDER_.*"') {
            Write-Error "terraform.tfvars still contains placeholders. Run -Phase update-config first."
            exit 1
        }
        
        Write-Step "Running terraform init"
        terraform init
        
        Write-Step "Running terraform plan"
        terraform plan -out=tfplan-aft
        
        Write-Host "`nReview the plan above. Press Enter to apply, or Ctrl+C to cancel..."
        Read-Host
        
        Write-Step "Running terraform apply"
        terraform apply tfplan-aft
        
        Write-Success "AFT infrastructure deployed!"
        
        # Get outputs
        Write-Step "AFT Infrastructure Details:"
        terraform output
        
        Write-Host "`nNext steps:"
        Write-Host "1. Run: .\deploy.ps1 -Phase setup-repos"
    }
    
    "setup-repos" {
        Write-Step "Setting up AFT repositories"
        
        # Get repo URLs from Terraform outputs
        $accountRequestRepo = terraform output -raw account_request_repo_url
        $globalCustomRepo = terraform output -raw global_customizations_repo_url
        $accountCustomRepo = terraform output -raw account_customizations_repo_url
        
        $projectRoot = "C:\Users\femil\Documents\PersonalProjects"
        
        # Clone repos
        Write-Step "Cloning AFT repositories"
        
        Set-Location $projectRoot
        
        if (Test-Path "aft-account-request") {
            Write-Warning "aft-account-request already exists, skipping clone"
        }
        else {
            git clone $accountRequestRepo aft-account-request
            Write-Success "Cloned account request repo"
        }
        
        if (Test-Path "aft-global-customizations") {
            Write-Warning "aft-global-customizations already exists, skipping clone"
        }
        else {
            git clone $globalCustomRepo aft-global-customizations
            Write-Success "Cloned global customizations repo"
        }
        
        if (Test-Path "aft-account-customizations") {
            Write-Warning "aft-account-customizations already exists, skipping clone"
        }
        else {
            git clone $accountCustomRepo aft-account-customizations
            Write-Success "Cloned account customizations repo"
        }
        
        # Copy files
        Write-Step "Copying account request files"
        Copy-Item "$projectRoot\new_bank\infra\terraform\aft-account-requests\*.tf" "$projectRoot\aft-account-request\" -Force
        Write-Success "Copied account request files"
        
        Write-Step "Copying global customizations"
        Copy-Item "$projectRoot\new_bank\infra\terraform\aft-global-customizations\*.tf" "$projectRoot\aft-global-customizations\" -Force
        Write-Success "Copied global customizations"
        
        Write-Step "Copying account customizations"
        if (-not (Test-Path "$projectRoot\aft-account-customizations\neobank-dev")) {
            New-Item -ItemType Directory -Path "$projectRoot\aft-account-customizations\neobank-dev" | Out-Null
        }
        Copy-Item "$projectRoot\new_bank\infra\terraform\aft-account-customizations\neobank-dev\*.tf" "$projectRoot\aft-account-customizations\neobank-dev\" -Force
        Write-Success "Copied account customizations"
        
        Write-Warning "IMPORTANT: Update email addresses in account request files before committing!"
        Write-Host "`nEdit these files:"
        Write-Host "  - $projectRoot\aft-account-request\dev-account.tf"
        Write-Host "  - $projectRoot\aft-account-request\staging-account.tf"
        Write-Host "  - $projectRoot\aft-account-request\prod-account.tf"
        Write-Host "`nReplace @example.com with your actual domain."
        Write-Host "`nThen run: .\deploy.ps1 -Phase provision-accounts"
    }
    
    "provision-accounts" {
        Write-Step "Provisioning accounts via AFT"
        
        $projectRoot = "C:\Users\femil\Documents\PersonalProjects"
        
        # Commit and push account requests
        Write-Step "Pushing account requests to trigger provisioning"
        Set-Location "$projectRoot\aft-account-request"
        
        git add *.tf
        git commit -m "Add neobank dev, staging, and prod account requests"
        git push origin main
        
        Write-Success "Account requests pushed!"
        
        # Commit and push global customizations
        Write-Step "Pushing global customizations"
        Set-Location "$projectRoot\aft-global-customizations"
        
        git add *.tf
        git commit -m "Add global security baselines"
        git push origin main
        
        Write-Success "Global customizations pushed!"
        
        # Commit and push account customizations
        Write-Step "Pushing account customizations"
        Set-Location "$projectRoot\aft-account-customizations"
        
        git add .
        git commit -m "Add neobank-dev account customizations"
        git push origin main
        
        Write-Success "Account customizations pushed!"
        
        Write-Host "`nAccount provisioning started. This takes 20-30 minutes per account."
        Write-Host "`nMonitor progress:"
        Write-Host "  aws codepipeline list-pipeline-executions --pipeline-name ct-aft-account-provisioning-pipeline"
        Write-Host "  aws organizations list-accounts"
    }
    
    "status" {
        Write-Step "Checking AFT status"
        
        # List accounts
        Write-Step "Organization Accounts:"
        aws organizations list-accounts --query 'Accounts[*].[Name,Id,Status]' --output table
        
        # Check pipeline
        Write-Step "AFT Pipeline Status:"
        try {
            aws codepipeline get-pipeline-state --name ct-aft-account-provisioning-pipeline --query 'stageStates[*].[stageName,latestExecution.status]' --output table
        }
        catch {
            Write-Warning "AFT pipeline not found. May not be deployed yet."
        }
        
        # Check recent Step Function executions
        Write-Step "Recent AFT Executions:"
        try {
            $sfnArn = terraform output -raw aft_step_function_arn
            aws stepfunctions list-executions --state-machine-arn $sfnArn --max-items 5 --query 'executions[*].[name,status,startDate]' --output table
        }
        catch {
            Write-Warning "Could not retrieve Step Function executions"
        }
    }
    
    default {
        Write-Host "AFT Deployment Script"
        Write-Host "====================="
        Write-Host ""
        Write-Host "Usage: .\deploy.ps1 -Phase <phase>"
        Write-Host ""
        Write-Host "Phases:"
        Write-Host "  check            - Check prerequisites and Control Tower status"
        Write-Host "  update-config    - Update terraform.tfvars with account IDs"
        Write-Host "  deploy-aft       - Deploy AFT infrastructure"
        Write-Host "  setup-repos      - Clone and populate AFT repositories"
        Write-Host "  provision-accounts - Trigger account provisioning"
        Write-Host "  status           - Check AFT deployment status"
        Write-Host ""
        Write-Host "Example workflow:"
        Write-Host "  1. .\deploy.ps1 -Phase check"
        Write-Host "  2. .\deploy.ps1 -Phase update-config -LogArchiveAccountId 123456789012 -AuditAccountId 234567890123"
        Write-Host "  3. .\deploy.ps1 -Phase deploy-aft"
        Write-Host "  4. .\deploy.ps1 -Phase setup-repos"
        Write-Host "  5. (Edit email addresses in account request files)"
        Write-Host "  6. .\deploy.ps1 -Phase provision-accounts"
        Write-Host "  7. .\deploy.ps1 -Phase status"
    }
}
