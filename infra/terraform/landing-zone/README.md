# AWS Landing Zone for NeoBank

Bank-grade multi-account governance infrastructure following AWS best practices and the Control Tower model.

## Quick Start

```bash
# Navigate to the landing-zone directory
cd infra/terraform/landing-zone

# Copy and configure variables
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your settings

# Initialize Terraform
terraform init

# Preview changes
terraform plan

# Apply (for single-account deployment)
terraform apply
```

## Architecture Overview

```
Root
├── Security OU
│   ├── Log Archive Account (immutable log sink)
│   └── Audit Account (delegated admin for security services)
├── Infrastructure OU
│   ├── Network Account (central VPC networking, shared DNS, egress controls)
│   ├── Shared Services Account (CI/CD, artifact registries, shared tooling)
│   └── Observability Account (optional telemetry plane)
├── Workloads OU
│   ├── NonProd OU
│   │   ├── neobank-dev
│   │   └── neobank-stage
│   ├── Prod OU
│   │   └── neobank-prod
│   └── PCI-CDE OU
│       └── (cardholder data environment)
└── Sandbox OU (restricted, short retention, no peering to prod)
```

## Modules

| Module               | Description                                                             |
| -------------------- | ----------------------------------------------------------------------- |
| **organizations/**   | AWS Organizations, OUs, and account structure                           |
| **scps/**            | Service Control Policies (8 policies covering all guardrails)           |
| **logging/**         | Centralized logging (CloudTrail, Config, VPC Flow Logs, S3 Object Lock) |
| **security-hub/**    | GuardDuty, Security Hub, Inspector, Macie, IAM Access Analyzer          |
| **alb/**             | AWS Load Balancer Controller for EKS                                    |
| **service-catalog/** | Golden path products and portfolios (RDS, ElastiCache, etc.)            |

## Service Control Policies (SCPs)

| SCP                   | Target        | Description                                                    |
| --------------------- | ------------- | -------------------------------------------------------------- |
| deny-disable-security | Root OU       | Prevents disabling CloudTrail, Config, GuardDuty, Security Hub |
| region-restriction    | Prod, NonProd | Restricts workloads to approved AWS regions                    |
| require-encryption    | Root OU       | Enforces encryption at rest for S3, EBS, RDS                   |
| deny-public-exposure  | Prod, PCI-CDE | Prevents public S3 buckets and RDS instances                   |
| identity-hardening    | Root OU       | Prevents IAM user creation, enforces MFA                       |
| sandbox-restrictions  | Sandbox OU    | Limits expensive services and instance types                   |
| pci-cde-controls      | PCI-CDE OU    | Strict PCI DSS compliance controls                             |
| mandatory-tags        | Root OU       | Enforces Environment, Project, DataClassification tags         |

## Security Standards Enabled

- CIS AWS Foundations Benchmark v1.4.0
- PCI DSS v3.2.1
- AWS Foundational Security Best Practices v1.0.0
- NIST SP 800-53 Rev. 5

## Service Catalog Portfolios

| Portfolio  | Products                          |
| ---------- | --------------------------------- |
| Platform   | Infrastructure components         |
| Data       | PostgreSQL RDS, ElastiCache Redis |
| Kubernetes | EKS namespaces, workloads         |
| Messaging  | SQS, SNS, EventBridge             |

## Usage Scenarios

### Single Account Deployment (Development)

For testing in a single account:

```hcl
# terraform.tfvars
create_organization = false
create_accounts     = false
log_archive_account_id = ""  # Uses current account
audit_account_id       = ""  # Uses current account
```

### Multi-Account Deployment (Production)

For production with separate accounts:

```hcl
# terraform.tfvars
create_organization    = true   # Only in management account
create_accounts        = true   # Creates Log Archive, Audit, etc.
account_email_domain   = "company.com"
enable_s3_object_lock  = true   # Immutable audit logs
```

## Prerequisites

- AWS CLI configured with appropriate credentials
- Terraform >= 1.0
- For multi-account: Management account access with OrganizationsFullAccess
- For single-account: AdministratorAccess or equivalent

## Integration with Existing EKS

To deploy the ALB module with an existing EKS cluster:

```hcl
# terraform.tfvars
eks_cluster_name   = "neobank-dev-eks"
vpc_id             = "vpc-xxx"
public_subnet_ids  = ["subnet-xxx", "subnet-yyy"]
private_subnet_ids = ["subnet-aaa", "subnet-bbb"]
enable_waf         = true
waf_web_acl_arn    = "arn:aws:wafv2:..."
```

## Outputs

After deployment, key outputs include:

- Organization and OU IDs
- CloudTrail, Config, and VPC Flow Logs bucket names
- KMS key ARN for log encryption
- Security Hub and GuardDuty detector IDs
- SNS topic for security alerts
- Service Catalog portfolio IDs

## Cost Considerations

- CloudTrail: ~$2/100,000 events + S3 storage
- Config: ~$0.003/configuration item
- GuardDuty: ~$4/GB for CloudTrail analysis
- Security Hub: ~$0.0010/finding ingested
- S3 Glacier: ~$0.004/GB/month (after 90 days)

## Related Documentation

- [AWS Control Tower](https://docs.aws.amazon.com/controltower/)
- [AWS Organizations](https://docs.aws.amazon.com/organizations/)
- [Service Control Policies](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_scps.html)
- [Security Hub Standards](https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-standards.html)
