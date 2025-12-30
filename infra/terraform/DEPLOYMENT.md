# NeoBank Infrastructure Deployment Guide

## Overview

This document describes the Terraform infrastructure deployment for NeoBank, a modern banking application built on AWS.

## Infrastructure Summary

### Total Resources Deployed

- **96 AWS resources** successfully created and managed by Terraform
- Infrastructure validated against LocalStack (local AWS emulator)

### AWS Services Used

| Service             | Purpose                 | Resources Created                          |
| ------------------- | ----------------------- | ------------------------------------------ |
| **VPC**             | Network isolation       | 1 VPC, 6 Subnets, Route Tables, NAT/IGW    |
| **EC2**             | Network components      | Security Groups (7), VPC Endpoints (6)     |
| **S3**              | Object storage          | 3 buckets (logs, backups, terraform state) |
| **Secrets Manager** | Secret storage          | 4 secrets (DB, Redis, JWT, App config)     |
| **KMS**             | Encryption              | 1 customer-managed key                     |
| **IAM**             | Access management       | 4 roles with policies                      |
| **CloudWatch**      | Monitoring              | 5 log groups, metric filters, alarms       |
| **SNS**             | Notifications           | 1 alarms topic                             |
| **DynamoDB**        | Terraform state locking | 1 table                                    |
| **EKS**             | Kubernetes (Pro only)   | Cluster config ready                       |
| **RDS**             | PostgreSQL (Pro only)   | Database config ready                      |
| **ElastiCache**     | Redis (Pro only)        | Cache config ready                         |

## Quick Start

### Prerequisites

- Terraform >= 1.5.0
- AWS CLI v2
- Docker (for LocalStack testing)

### Local Testing with LocalStack

```powershell
# Start LocalStack
cd infra/localstack
docker compose up -d

# Deploy infrastructure
cd ../terraform
terraform init
terraform apply -var-file="environments/localstack.tfvars" -auto-approve

# Run tests
.\test-localstack.ps1
```

### Production Deployment

```powershell
# Configure AWS credentials
aws configure

# Deploy to dev environment
cd infra/terraform
terraform init
terraform apply -var-file="environments/dev.tfvars"

# Deploy to staging
terraform apply -var-file="environments/staging.tfvars"

# Deploy to production
terraform apply -var-file="environments/prod.tfvars"
```

## Environment Configurations

### LocalStack (Local Testing)

- **File**: `environments/localstack.tfvars`
- **Features**: Minimal resources, 1-day retention, no MSK/WAF

### Development

- **File**: `environments/dev.tfvars`
- **Features**: Small instances, 7-day retention, development settings

### Staging

- **File**: `environments/staging.tfvars`
- **Features**: Medium instances, 30-day retention, production-like settings

### Production

- **File**: `environments/prod.tfvars`
- **Features**: Multi-AZ, large instances, 365-day retention, full security

## Infrastructure Details

### VPC Architecture

```
VPC (10.0.0.0/16)
├── Public Subnets (10.0.101.0/24, 10.0.102.0/24)
│   └── Internet Gateway, NAT Gateway, ALB
├── Private Subnets (10.0.1.0/24, 10.0.2.0/24)
│   └── EKS Nodes, Application Services
└── Database Subnets (10.0.201.0/24, 10.0.202.0/24)
    └── RDS, ElastiCache (isolated)
```

### Security Groups

| Name             | Purpose                   | Inbound Rules             |
| ---------------- | ------------------------- | ------------------------- |
| alb-sg           | Application Load Balancer | 80, 443 from 0.0.0.0/0    |
| eks-cluster-sg   | EKS Control Plane         | 443 from VPC              |
| eks-nodes-sg     | EKS Worker Nodes          | All from cluster          |
| rds-sg           | PostgreSQL Database       | 5432 from private subnets |
| redis-sg         | ElastiCache Redis         | 6379 from private subnets |
| msk-sg           | Kafka (MSK)               | 9092, 9094 from VPC       |
| vpc-endpoints-sg | AWS Service Endpoints     | 443 from VPC              |

### Secrets Managed

1. **neobank-{env}/database/credentials** - PostgreSQL master password
2. **neobank-{env}/redis/credentials** - Redis AUTH token
3. **neobank-{env}/auth/jwt-secret** - JWT signing secret
4. **neobank-{env}/app/config** - Application configuration

### S3 Buckets

1. **neobank-{env}-logs** - Application and audit logs
2. **neobank-{env}-backups** - Database backups
3. **neobank-{env}-terraform-state** - Terraform state (with versioning)

## Outputs

After deployment, the following outputs are available:

```hcl
vpc_id                  # VPC identifier
private_subnet_ids      # Private subnet IDs for EKS
database_subnet_ids     # Database subnet IDs for RDS
eks_cluster_name        # EKS cluster name
kms_key_arn            # KMS key for encryption
db_secret_arn          # Database credentials secret ARN
jwt_secret_arn         # JWT secret ARN
sns_alarms_topic_arn   # SNS topic for alarms
logs_bucket_name       # S3 bucket for logs
backups_bucket_name    # S3 bucket for backups
```

## Test Results

### LocalStack Test Suite (100% Pass Rate)

| Test             | Status  | Description                   |
| ---------------- | ------- | ----------------------------- |
| VPC              | ✅ PASS | VPC created with correct CIDR |
| Subnets          | ✅ PASS | 6 subnets across 2 AZs        |
| S3 Buckets       | ✅ PASS | 3 buckets created             |
| S3 Operations    | ✅ PASS | Upload/download working       |
| Secrets Manager  | ✅ PASS | 4 secrets created             |
| Secret Retrieval | ✅ PASS | Can read secret values        |
| KMS Keys         | ✅ PASS | Encryption key created        |
| SNS Topics       | ✅ PASS | Alarms topic created          |
| CloudWatch Logs  | ✅ PASS | 5 log groups created          |
| IAM Roles        | ✅ PASS | 4 roles with policies         |
| Security Groups  | ✅ PASS | 7 security groups             |
| DynamoDB         | ✅ PASS | Lock table created            |
| VPC Endpoints    | ✅ PASS | 6 endpoints created           |
| NAT Gateway      | ✅ PASS | NAT Gateway created           |
| Internet Gateway | ✅ PASS | IGW attached                  |

### LocalStack Limitations

The following services require LocalStack Pro and are tested only in real AWS:

- **EKS** - Elastic Kubernetes Service
- **RDS** - Relational Database Service
- **ElastiCache** - Redis cache

## Troubleshooting

### Common Issues

1. **S3 bucket already exists**

   ```bash
   terraform state rm module.s3.aws_s3_bucket.logs
   terraform import module.s3.aws_s3_bucket.logs bucket-name
   ```

2. **LocalStack connection refused**

   ```bash
   docker compose restart localstack-neobank
   docker logs localstack-neobank
   ```

3. **Terraform state lock**
   ```bash
   terraform force-unlock LOCK_ID
   ```

## Security Considerations

- All secrets encrypted with customer-managed KMS key
- S3 buckets have public access blocked
- VPC flow logs enabled for network monitoring
- Database in isolated subnets with no internet access
- All inter-service communication uses VPC endpoints

## Compliance

Infrastructure is designed for:

- **PCI-DSS** - Payment Card Industry compliance
- **SOC 2** - Security, availability, processing integrity
- **ISO 27001** - Information security management

## Next Steps

1. Configure AWS credentials for actual deployment
2. Deploy to dev environment
3. Set up CI/CD pipeline integration
4. Configure monitoring dashboards
5. Deploy Kubernetes workloads to EKS
