# deploy_aws_mgmt.ps1 - Deploy to AWS Management Account
# Usage: ./deploy_aws_mgmt.ps1 [apply|destroy]

param (
    [string]$Action = "apply"
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Join-Path $ScriptDir ".."
$TerraformDir = Join-Path $ScriptDir "terraform"

# 1. Terraform Infrastructure
Write-Host ">>> API: Initializing Terraform..." -ForegroundColor Cyan
Push-Location $TerraformDir

terraform init

if ($Action -eq "destroy") {
    terraform destroy -var-file="environments/dev.tfvars" -auto-approve
    Exit
}

Write-Host ">>> API: Applying Terraform..." -ForegroundColor Cyan
terraform apply -var-file="environments/dev.tfvars" -auto-approve

# 2. Get Outputs
Write-Host ">>> API: Fetching Infrastructure Outputs..." -ForegroundColor Cyan
$EcrUrls = terraform output -json ecr_repository_urls | ConvertFrom-Json
$ConfigureKubectl = terraform output -raw configure_kubectl
$RdsEndpoint = terraform output -raw rds_endpoint
$RdsPort = terraform output -raw rds_port

# IRSA ARNs
$IdentityArn = terraform output -raw identity_service_role_arn
$LedgerArn = terraform output -raw ledger_service_role_arn
$PaymentArn = terraform output -raw payment_service_role_arn
$ProductArn = terraform output -raw product_service_role_arn
$CardArn = terraform output -raw card_service_role_arn

# Secret ARNs
$DbSecretArn = terraform output -raw db_secret_arn
$JwtSecretArn = terraform output -raw jwt_secret_arn

$Region = "us-east-1" 

Pop-Location 
Push-Location $ProjectRoot

# 3. Docker Build & Push
Write-Host ">>> API: Authenticating to ECR..." -ForegroundColor Cyan
aws ecr get-login-password --region $Region | docker login --username AWS --password-stdin $EcrUrls.PSObject.Properties.Value[0].Split("/")[0]

$Timestamp = Get-Date -Format "yyyyMMddHHmm"
$Services = @{
    "identity-service" = "backend/identity-service/Dockerfile"
    "ledger-service"   = "backend/ledger-service/Dockerfile"
    "payment-service"  = "backend/payment-service/Dockerfile"
    "product-service"  = "backend/product-service/Dockerfile"
    "card-service"     = "backend/card-service/Dockerfile"
    "frontend"         = "frontend/Dockerfile"
}

# 4. Fetch/Prepare Secrets (Bypass ESO)
Write-Host ">>> API: Fetching Secrets from AWS Secrets Manager..." -ForegroundColor Cyan

# Parse DB Secret
$DbSecretJsonString = aws secretsmanager get-secret-value --secret-id $DbSecretArn --query SecretString --output text
$DbSecretObj = $DbSecretJsonString | ConvertFrom-Json
$DbUser = $DbSecretObj.username
$DbPass = $DbSecretObj.password

# Parse JWT Secret
$JwtSecretJsonString = aws secretsmanager get-secret-value --secret-id $JwtSecretArn --query SecretString --output text
$JwtSecretObj = $JwtSecretJsonString | ConvertFrom-Json
$JwtSecret = $JwtSecretObj.secret

# 5. Generate Kustomization
$KustomizationPath = "k8s/overlays/aws-dev/kustomization.yaml"
Write-Host ">>> API: Generating Kustomization..." -ForegroundColor Cyan

# Build & Push Loop
$ImageConfig = ""
foreach ($SvcName in $Services.Keys) {
    $Dockerfile = $Services[$SvcName]
    $RepoKey = "neobank/$SvcName"
    $EcrUrl = $EcrUrls."$RepoKey"

    if (-not $EcrUrl) { Write-Error "ECR URL for $RepoKey not found." }

    $ImageTag = "$EcrUrl`:$Timestamp"
    $ImageLatest = "$EcrUrl`:latest"

    Write-Host "   Building $SvcName -> $ImageTag"
    if ($SvcName -eq "frontend") {
        docker build -t $ImageTag -t $ImageLatest -f $Dockerfile "frontend"
    }
    else {
        docker build -t $ImageTag -t $ImageLatest -f $Dockerfile "backend"
    }

    Write-Host "   Pushing $SvcName..."
    docker push $ImageTag
    docker push $ImageLatest
    
    $ImageConfig += "`n  - name: $RepoKey`n    newName: $EcrUrl`n    newTag: $Timestamp"
}

$KustomizationContent = @"
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: neobank

resources:
  - ../../base
  - ../../istio

namePrefix: dev-

commonLabels:
  environment: dev
  cost-center: development
  team: platform

images:$ImageConfig

patches:
  - patch: |-
      - op: replace
        path: /metadata/annotations/eks.amazonaws.com~1role-arn
        value: "$IdentityArn"
    target:
      kind: ServiceAccount
      name: identity-service-sa
  - patch: |-
      - op: replace
        path: /metadata/annotations/eks.amazonaws.com~1role-arn
        value: "$LedgerArn"
    target:
      kind: ServiceAccount
      name: ledger-service-sa
  - patch: |-
      - op: replace
        path: /metadata/annotations/eks.amazonaws.com~1role-arn
        value: "$PaymentArn"
    target:
      kind: ServiceAccount
      name: payment-service-sa
  - patch: |-
      - op: replace
        path: /metadata/annotations/eks.amazonaws.com~1role-arn
        value: "$ProductArn"
    target:
      kind: ServiceAccount
      name: product-service-sa
  - patch: |-
      - op: replace
        path: /metadata/annotations/eks.amazonaws.com~1role-arn
        value: "$CardArn"
    target:
      kind: ServiceAccount
      name: card-service-sa
  
  # Patch ConfigMap for RDS
  - patch: |-
      - op: replace
        path: /data/DB_HOST
        value: "$RdsEndpoint"
      - op: replace
        path: /data/DB_PORT
        value: "$RdsPort"
    target:
      kind: ConfigMap
      name: neobank-config

# ConfigMap generator for dev settings
configMapGenerator:
  - name: neobank-config
    behavior: merge
    literals:
      - LOG_LEVEL=info
      - ENVIRONMENT=dev
      - ENABLE_FINOPS=true
      - ENABLE_DR_TESTING=true
      - DB_SSLMODE=require

# Secret Generators (from AWS values)
secretGenerator:
  - name: neobank-db-credentials
    literals:
      - username=$DbUser
      - password=$DbPass
  - name: neobank-jwt
    literals:
      - secret=$JwtSecret
"@

Set-Content -Path $KustomizationPath -Value $KustomizationContent

# 6. Deploy
Write-Host ">>> API: Configuring Kubectl..." -ForegroundColor Cyan
Invoke-Expression $ConfigureKubectl

Write-Host ">>> API: Deploying to EKS..." -ForegroundColor Cyan
kubectl apply -k k8s/overlays/aws-dev

Write-Host ">>> Deployment Complete!" -ForegroundColor Green
Write-Host "Check status: kubectl get pods -n neobank"
