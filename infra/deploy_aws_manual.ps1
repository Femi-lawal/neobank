# deploy_aws_manual.ps1 - Manual Deploy to AWS Management Account (Pure CLI Discovery)
Write-Host "SCRIPT VERSION: 2.0 (Secrets & ECR Fix)"

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Join-Path $ScriptDir ".."

$Region = "us-east-1" 
$NamePrefix = "neobank-dev"

# 1. Discover Infrastructure (Bypass Terraform State)
Write-Host ">>> API: Discovering Infrastructure via AWS CLI..." -ForegroundColor Cyan

# Disable AWS Pager to prevent hanging/errors on Windows
$env:AWS_PAGER = ""

# Get Account ID
$Identity = aws sts get-caller-identity --output json | ConvertFrom-Json
$AccountId = $Identity.Account
Write-Host "   Account ID: $AccountId"

# ECR Domain
$EcrDomain = "$AccountId.dkr.ecr.$Region.amazonaws.com"
Write-Host "   ECR Domain: $EcrDomain"

# RDS Endpoint
Write-Host "   Fetching RDS Endpoint..."
try {
  # List all instances to handle naming mismatches
  $RdsJson = aws rds describe-db-instances --region $Region --output json
  if (-not $RdsJson) { throw "Empty output from aws rds describe-db-instances" }
    
  $RdsObj = $RdsJson | ConvertFrom-Json
  $TargetDb = $RdsObj.DBInstances | Where-Object { $_.DBInstanceIdentifier -like "*neobank*" } | Select-Object -First 1

  if (-not $TargetDb) {
    Write-Warning "No RDS instance found matching '*neobank*'. Available: $($RdsObj.DBInstances.DBInstanceIdentifier -join ', ')"
    throw "RDS Discovery Failed"
  }

  $RdsEndpoint = $TargetDb.Endpoint.Address
  $RdsPort = $TargetDb.Endpoint.Port
  Write-Host "   RDS Instance: $($TargetDb.DBInstanceIdentifier)"
  Write-Host "   RDS Endpoint: $RdsEndpoint`:$RdsPort"
}
catch {
  Write-Warning "Could not fetch RDS endpoint: $_"
  throw
}

# IAM Role ARNs (Constructed by convention)
$RoleBase = "arn:aws:iam::$AccountId`:role/$NamePrefix"
$IdentityArn = "$RoleBase-identity-service-role"
$LedgerArn = "$RoleBase-ledger-service-role"
$PaymentArn = "$RoleBase-payment-service-role"
$ProductArn = "$RoleBase-product-service-role"
$CardArn = "$RoleBase-card-service-role"

# Secret Names (Constructed by convention)
$DbSecretId = "$NamePrefix/database/credentials"
$JwtSecretId = "$NamePrefix/auth/jwt-secret"

# Redis Endpoint
Write-Host "   Fetching Redis Endpoint..."
$RedisAddr = "neobank-dev-redis:6379" # Default
try {
  $RedisJson = aws elasticache describe-replication-groups --region $Region --output json
  if ($RedisJson) {
    $RedisObj = $RedisJson | ConvertFrom-Json
    $TargetRedis = $RedisObj.ReplicationGroups | Where-Object { $_.ReplicationGroupId -like "*neobank*" } | Select-Object -First 1
    if ($TargetRedis) {
      $RedisAddr = "$($TargetRedis.NodeGroups[0].PrimaryEndpoint.Address):$($TargetRedis.NodeGroups[0].PrimaryEndpoint.Port)"
      Write-Host "   Redis Endpoint: $RedisAddr"
    }
  }
}
catch {
  Write-Warning "Could not fetch Redis endpoint: $_"
}

# 2. Docker Build & Push
Write-Host ">>> API: Authenticating to ECR..." -ForegroundColor Cyan
aws ecr get-login-password --region $Region | docker login --username AWS --password-stdin $EcrDomain

$Timestamp = Get-Date -Format "yyyyMMddHHmm"
$Services = @{
  "identity-service" = "backend/identity-service/Dockerfile"
  "ledger-service"   = "backend/ledger-service/Dockerfile"
  "payment-service"  = "backend/payment-service/Dockerfile"
  "product-service"  = "backend/product-service/Dockerfile"
  "card-service"     = "backend/card-service/Dockerfile"
  "frontend"         = "frontend/Dockerfile"
}

Push-Location $ProjectRoot

# 3. Generate Kustomization
$KustomizationPath = "k8s/overlays/aws-dev/kustomization.yaml"
Write-Host ">>> API: Generating Kustomization..." -ForegroundColor Cyan

# Build & Push Loop
$ImageConfig = ""
try {
  foreach ($SvcName in $Services.Keys) {
    $Dockerfile = $Services[$SvcName]
    $RepoKey = "neobank/$SvcName"
    
    $EcrUrl = "$EcrDomain/$RepoKey"

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
    $MaxRetries = 3
    for ($i = 1; $i -le $MaxRetries; $i++) {
      try {
        docker push $ImageTag
        docker push $ImageLatest
        break
      }
      catch {
        if ($i -eq $MaxRetries) { throw $_ }
        Write-Warning "   Push failed (attempt $i), retrying in 5s..."
        Start-Sleep -Seconds 5
      }
    }
    
    # Quote the tag to ensure string type in YAML
    $ImageConfig += "`n  - name: $RepoKey`n    newName: $EcrUrl`n    newTag: 'latest'"
  }
}
catch {
  Write-Warning "Build/Push loop failed: $_. Continuing with latest images."
  # Ensure ImageConfig is populated if catch happens mid-loop
  foreach ($SvcName in $Services.Keys) {
    $RepoKey = "neobank/$SvcName"
    $EcrUrl = "$EcrDomain/$RepoKey"
    if ($ImageConfig -notlike "*$RepoKey*") {
      $ImageConfig += "`n  - name: $RepoKey`n    newName: $EcrUrl`n    newTag: 'latest'"
    }
  }
}

# 4. Fetch Secrets
Write-Host ">>> API: Fetching Secrets from AWS Secrets Manager..." -ForegroundColor Cyan
Write-Host "   DB Secret ID: $DbSecretId"
# Parse DB Secret
try {
  $DbSecretJsonString = aws secretsmanager get-secret-value --secret-id $DbSecretId --query SecretString --output text
  if (-not $DbSecretJsonString) { throw "Empty DB Secret output" }
  $DbSecretObj = $DbSecretJsonString | ConvertFrom-Json
  $DbUser = $DbSecretObj.username
  $DbPass = $DbSecretObj.password
  $DbName = $DbSecretObj.database
}
catch {
  Write-Warning "Failed to fetch/parse DB secret: $_"
  throw
}

Write-Host "   JWT Secret ID: $JwtSecretId"
# Parse JWT Secret
try {
  $JwtSecretJsonString = aws secretsmanager get-secret-value --secret-id $JwtSecretId --query SecretString --output text
  if (-not $JwtSecretJsonString) { throw "Empty JWT Secret output" }
  $JwtSecretObj = $JwtSecretJsonString | ConvertFrom-Json
  $JwtSecret = $JwtSecretObj.secret
}
catch {
  Write-Warning "Failed to fetch/parse JWT secret: $_"
  throw
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
      - op: replace
        path: /data/DB_NAME
        value: "$DbName"
      - op: replace
        path: /data/REDIS_ADDR
        value: "$RedisAddr"
    target:
      kind: ConfigMap
      name: neobank-config

# ConfigMap patches for dev settings (Replacing generator to avoid merge errors)
  - patch: |-
      - op: replace
        path: /data/LOG_LEVEL
        value: "info"
      - op: replace
        path: /data/ENVIRONMENT
        value: "dev"
      - op: add
        path: /data/ENABLE_FINOPS
        value: "true"
      - op: add
        path: /data/ENABLE_DR_TESTING
        value: "true"
      - op: replace
        path: /data/DB_SSLMODE
        value: "require"
    target:
      kind: ConfigMap
      name: neobank-config

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
aws eks update-kubeconfig --region $Region --name "$NamePrefix-eks"

Write-Host ">>> API: Installing Addons via Helm..." -ForegroundColor Cyan
try {
  helm repo add external-secrets https://charts.external-secrets.io
  helm repo add istio https://istio-release.storage.googleapis.com/charts
  helm repo update

  Write-Host "   Installing External Secrets Operator..."
  helm upgrade --install external-secrets external-secrets/external-secrets -n external-secrets --create-namespace --set installCRDs=true --wait

  Write-Host "   Installing Istio Base..."
  helm upgrade --install istio-base istio/base -n istio-system --create-namespace --wait

  Write-Host "   Installing Istiod..."
  helm upgrade --install istiod istio/istiod -n istio-system --wait
}
catch {
  Write-Warning "Helm installation failed: $_"
  # Don't stop deployment, as they might be already there or partial failure?
  # Actually, we should probably stop if CRDs are missing.
  throw
}

Write-Host ">>> API: Deploying to EKS..." -ForegroundColor Cyan
kubectl apply -k k8s/overlays/aws-dev

Write-Host ">>> Deployment Complete!" -ForegroundColor Green
Write-Host "Check status: kubectl get pods -n neobank"
