# Account Factory for Terraform (AFT) Deployment Guide

## Overview

This guide walks through deploying AWS Control Tower with Account Factory for Terraform (AFT) - the enterprise-grade "managed path" for multi-account AWS architectures.

## Architecture

```
Management Account
├── Control Tower
│   ├── Landing Zone
│   ├── Guardrails (SCPs)
│   └── Account Factory
├── AFT Infrastructure
│   ├── AFT Management VPC (10.100.0.0/16)
│   ├── CodePipeline for account provisioning
│   ├── Step Functions for orchestration
│   └── CodeCommit repositories
└── Created Accounts
    ├── Log Archive Account (auto-created)
    ├── Audit Account (auto-created)
    ├── NeoBank Dev Account (AFT-provisioned)
    ├── NeoBank Staging Account (AFT-provisioned)
    └── NeoBank Prod Account (AFT-provisioned)
```

## Prerequisites

### 1. AWS Requirements

- ✅ AWS Organizations enabled (already done)
- ✅ Root account access or AdministratorAccess
- ❌ Control Tower NOT yet deployed (we'll do this)
- ✅ No existing OUs named "Security" or "Sandbox" (we already have these)

### 2. Email Addresses Needed

You'll need unique email addresses for:

- Log Archive account: `neobank-log-archive@yourdomain.com`
- Audit account: `neobank-audit@yourdomain.com`
- Dev account: `neobank-dev@yourdomain.com`
- Staging account: `neobank-staging@yourdomain.com`
- Prod account: `neobank-prod@yourdomain.com`

## Deployment Steps

### Phase 1: Deploy Control Tower (Console-Based)

Control Tower must be deployed through the AWS Console (no Terraform option).

1. **Navigate to Control Tower Console**

   ```
   https://us-east-1.console.aws.amazon.com/controltower/home
   ```

2. **Click "Set up landing zone"**

3. **Configure Landing Zone Settings:**
   - **Region selection:**
     - Home Region: `us-east-1`
     - Additional governed regions: `us-west-2`
   - **Organizational units:**
     - Security OU: ✅ (already exists)
     - Sandbox OU: ✅ (already exists)
   - **Log Archive account email:** `neobank-log-archive@yourdomain.com`
   - **Audit account email:** `neobank-audit@yourdomain.com`
   - **Logging configuration:**
     - Enable CloudTrail: ✅
     - Enable AWS Config: ✅
     - Log retention: 1 year minimum

4. **Review and Deploy**
   - Deployment takes 45-60 minutes
   - Do NOT interrupt the process
   - Monitor the "Set up" progress page

5. **Post-Deployment Verification**

   ```powershell
   # Verify Control Tower is deployed
   aws controltower list-landing-zones

   # Get Log Archive account ID
   $LogArchiveId = aws organizations list-accounts --query 'Accounts[?Name==`Log Archive`].Id' --output text
   echo "Log Archive: $LogArchiveId"

   # Get Audit account ID
   $AuditId = aws organizations list-accounts --query 'Accounts[?Name==`Audit`].Id' --output text
   echo "Audit: $AuditId"
   ```

### Phase 2: Deploy AFT Infrastructure

1. **Update AFT Configuration**

   ```powershell
   cd C:\Users\femil\Documents\PersonalProjects\new_bank\infra\terraform\control-tower-aft

   # Edit terraform.tfvars with actual account IDs
   notepad terraform.tfvars
   ```

   Update these values:

   ```hcl
   log_archive_account_id = "YOUR_LOG_ARCHIVE_ACCOUNT_ID"
   audit_account_id       = "YOUR_AUDIT_ACCOUNT_ID"
   ```

2. **Initialize and Deploy AFT**

   ```powershell
   terraform init
   terraform plan -out=tfplan-aft
   terraform apply tfplan-aft
   ```

   Deployment takes 30-45 minutes and creates:
   - AFT VPC and networking
   - CodePipeline for account provisioning
   - Step Functions for orchestration
   - 4 CodeCommit repositories
   - IAM roles and policies
   - DynamoDB tables for state tracking

3. **Verify AFT Deployment**

   ```powershell
   # Get AFT outputs
   terraform output

   # Verify Step Function
   aws stepfunctions list-state-machines --query 'stateMachines[?contains(name, `aft`)]'

   # Verify CodeCommit repos
   aws codecommit list-repositories
   ```

### Phase 3: Provision Accounts via AFT

AFT uses a GitOps workflow. Account requests are Terraform files committed to Git.

1. **Clone AFT Account Request Repository**

   ```powershell
   cd C:\Users\femil\Documents\PersonalProjects
   $RepoUrl = terraform output -raw account_request_repo_url
   git clone $RepoUrl aft-account-request
   cd aft-account-request
   ```

2. **Copy Account Request Files**

   ```powershell
   # Copy our pre-created account requests
   Copy-Item ..\new_bank\infra\terraform\aft-account-requests\*.tf .

   # Update email addresses in the files
   notepad dev-account.tf
   notepad staging-account.tf
   notepad prod-account.tf
   ```

   Replace `@example.com` with your actual domain.

3. **Commit and Push to Trigger Provisioning**

   ```powershell
   git add *.tf
   git commit -m "Add neobank dev, staging, and prod account requests"
   git push origin main
   ```

   This automatically triggers:
   - CodePipeline execution
   - Step Functions orchestration
   - Account creation via Control Tower
   - Baseline configuration application

4. **Monitor Account Provisioning**

   ```powershell
   # Watch CodePipeline
   aws codepipeline list-pipeline-executions --pipeline-name ct-aft-account-provisioning-pipeline

   # Check Step Function executions
   aws stepfunctions list-executions --state-machine-arn <AFT_SFN_ARN>

   # List newly created accounts
   aws organizations list-accounts
   ```

   Each account takes 20-30 minutes to provision.

### Phase 4: Deploy Global Customizations

Global customizations are applied to ALL AFT-provisioned accounts.

1. **Clone Global Customizations Repository**

   ```powershell
   cd C:\Users\femil\Documents\PersonalProjects
   $RepoUrl = terraform output -raw global_customizations_repo_url
   git clone $RepoUrl aft-global-customizations
   cd aft-global-customizations
   ```

2. **Copy Baseline Configuration**

   ```powershell
   Copy-Item ..\new_bank\infra\terraform\aft-global-customizations\*.tf .
   ```

3. **Commit and Push**

   ```powershell
   git add *.tf
   git commit -m "Add global security baselines"
   git push origin main
   ```

   This applies to all accounts:
   - CloudWatch log encryption
   - AWS Config rules
   - GuardDuty enablement
   - Security Hub standards
   - S3 public access block
   - EBS encryption by default

### Phase 5: Deploy Account-Specific Customizations

Account customizations are specific to each account.

1. **Clone Account Customizations Repository**

   ```powershell
   cd C:\Users\femil\Documents\PersonalProjects
   $RepoUrl = terraform output -raw account_customizations_repo_url
   git clone $RepoUrl aft-account-customizations
   cd aft-account-customizations
   ```

2. **Copy Dev Account Customizations**

   ```powershell
   mkdir neobank-dev
   Copy-Item ..\new_bank\infra\terraform\aft-account-customizations\neobank-dev\*.tf .\neobank-dev\
   ```

3. **Commit and Push**

   ```powershell
   git add .
   git commit -m "Add neobank-dev account customizations"
   git push origin main
   ```

   This deploys in the dev account:
   - VPC with public/private/database subnets
   - KMS keys for RDS and EKS
   - IAM roles for EKS IRSA
   - Secrets Manager secrets

### Phase 6: Migrate Applications to Dev Account

Once the dev account is provisioned, migrate infrastructure.

1. **Assume Role into Dev Account**

   ```powershell
   # Get dev account ID
   $DevAccountId = aws organizations list-accounts --query 'Accounts[?Name==`neobank-dev`].Id' --output text

   # Assume AWSControlTowerExecution role
   $Creds = aws sts assume-role --role-arn "arn:aws:iam::${DevAccountId}:role/AWSControlTowerExecution" --role-session-name "migration"

   # Export credentials
   $Env:AWS_ACCESS_KEY_ID = ($Creds | ConvertFrom-Json).Credentials.AccessKeyId
   $Env:AWS_SECRET_ACCESS_KEY = ($Creds | ConvertFrom-Json).Credentials.SecretAccessKey
   $Env:AWS_SESSION_TOKEN = ($Creds | ConvertFrom-Json).Credentials.SessionToken
   ```

2. **Deploy Infrastructure in Dev Account**

   ```powershell
   cd C:\Users\femil\Documents\PersonalProjects\new_bank\infra\terraform

   # Initialize new backend for dev account
   terraform init -backend-config="bucket=neobank-terraform-state-${DevAccountId}"

   # Deploy
   terraform apply
   ```

3. **Update Kubernetes Configurations**
   - Update EKS cluster name in kubeconfig
   - Update External Secrets to point to dev account
   - Update IAM role ARNs for service accounts
   - Redeploy applications

## Cost Considerations

- **Control Tower:** ~$350/month (CloudTrail, Config, GuardDuty)
- **AFT Infrastructure:** ~$100/month (CodePipeline, Step Functions, VPC)
- **Per Account:** ~$50/month (Config rules, GuardDuty)
- **Total for 5 accounts:** ~$700/month

## Security Features

AFT-provisioned accounts automatically have:

- ✅ CloudTrail enabled in all regions
- ✅ AWS Config rules for compliance
- ✅ GuardDuty threat detection
- ✅ Security Hub with CIS, PCI-DSS, AWS Foundational standards
- ✅ S3 public access blocked
- ✅ EBS encryption by default
- ✅ VPC Flow Logs enabled
- ✅ SCPs from Control Tower guardrails

## Troubleshooting

### AFT Pipeline Failures

```powershell
# Check pipeline logs
aws codepipeline get-pipeline-execution --pipeline-name ct-aft-account-provisioning-pipeline --execution-id <ID>

# Check Step Function logs
aws stepfunctions describe-execution --execution-arn <ARN>
```

### Account Provisioning Stuck

```powershell
# Check Control Tower status
aws controltower list-landing-zones

# Check account status
aws organizations describe-account --account-id <ACCOUNT_ID>
```

### Global Customizations Not Applying

```powershell
# Check CodeBuild logs
aws codebuild list-builds-for-project --project-name aft-global-customizations

# View specific build logs
aws codebuild batch-get-builds --ids <BUILD_ID>
```

## Next Steps

After AFT is fully deployed:

1. **Migrate Workloads:**
   - Move EKS cluster to dev account
   - Move RDS to dev account
   - Move Redis to dev account
   - Update application configs

2. **Set Up CI/CD:**
   - GitOps workflow for account requests
   - Automated baseline updates
   - Account vending via API

3. **Enable Additional Guardrails:**
   - Detective controls via Control Tower
   - Preventive controls via SCPs
   - Budget alerts per account

## References

- [AFT Workshop](https://catalog.workshops.aws/control-tower/en-US/customization/aft)
- [AFT GitHub](https://github.com/aws-ia/terraform-aws-control_tower_account_factory)
- [Control Tower User Guide](https://docs.aws.amazon.com/controltower/latest/userguide/)
