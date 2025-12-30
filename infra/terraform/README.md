# NeoBank AWS Infrastructure

This directory contains Terraform configurations for deploying the NeoBank platform to AWS with production-grade security and compliance.

## Architecture Overview

The infrastructure follows best practices from PCI-DSS, SOC 2, and NIST 800-53 frameworks:

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                    AWS Cloud                                     │
│  ┌─────────────────────────────────────────────────────────────────────────────┐│
│  │                              VPC (10.0.0.0/16)                              ││
│  │  ┌─────────────────────────────────────────────────────────────────────────┐││
│  │  │                        Public Subnets (3 AZs)                           │││
│  │  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                     │││
│  │  │  │ NAT Gateway │  │ NAT Gateway │  │ NAT Gateway │  (prod only)        │││
│  │  │  └─────────────┘  └─────────────┘  └─────────────┘                     │││
│  │  │                         ↓ Internet Gateway                              │││
│  │  └─────────────────────────────────────────────────────────────────────────┘││
│  │  ┌─────────────────────────────────────────────────────────────────────────┐││
│  │  │                       Private Subnets (3 AZs)                           │││
│  │  │  ┌─────────────────────────────────────────────────────────────────────┐│││
│  │  │  │                      EKS Cluster                                    ││││
│  │  │  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐       ││││
│  │  │  │  │ Identity   │ │ Ledger     │ │ Payment    │ │ Card       │       ││││
│  │  │  │  │ Service    │ │ Service    │ │ Service    │ │ Service    │       ││││
│  │  │  │  └────────────┘ └────────────┘ └────────────┘ └────────────┘       ││││
│  │  │  └─────────────────────────────────────────────────────────────────────┘│││
│  │  └─────────────────────────────────────────────────────────────────────────┘││
│  │  ┌─────────────────────────────────────────────────────────────────────────┐││
│  │  │                      Database Subnets (3 AZs)                           │││
│  │  │  ┌─────────────────────┐  ┌─────────────────────┐                       │││
│  │  │  │    RDS PostgreSQL   │  │  ElastiCache Redis  │                       │││
│  │  │  │    (Multi-AZ)       │  │  (Cluster Mode)     │                       │││
│  │  │  └─────────────────────┘  └─────────────────────┘                       │││
│  │  └─────────────────────────────────────────────────────────────────────────┘││
│  └─────────────────────────────────────────────────────────────────────────────┘│
│                                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  WAF         │  │  KMS         │  │  Secrets     │  │  CloudWatch  │         │
│  │  (Prod)      │  │  Encryption  │  │  Manager     │  │  Monitoring  │         │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Prerequisites

1. **AWS CLI** configured with appropriate credentials
2. **Terraform** >= 1.5.0
3. **kubectl** for Kubernetes management
4. **kustomize** for K8s manifest generation

## Quick Start

### 1. Initialize Terraform

```bash
cd infra/terraform
terraform init
```

### 2. Plan Infrastructure (Development)

```bash
terraform plan -var-file="environments/dev.tfvars" -out="tfplan-dev"
```

### 3. Apply Infrastructure

```bash
terraform apply "tfplan-dev"
```

### 4. Configure kubectl

```bash
aws eks update-kubeconfig --region us-east-1 --name neobank-dev-eks
```

### 5. Deploy Kubernetes Resources

```bash
cd ../k8s/overlays/aws
kustomize build . | kubectl apply -f -
```

## Module Structure

```
terraform/
├── main.tf              # Root module orchestration
├── variables.tf         # Input variables
├── outputs.tf          # Output values
├── versions.tf         # Provider versions
├── environments/
│   ├── dev.tfvars      # Development configuration
│   └── prod.tfvars     # Production configuration
└── modules/
    ├── vpc/            # VPC, subnets, NAT, VPC endpoints
    ├── security/       # KMS keys, security groups
    ├── secrets/        # AWS Secrets Manager
    ├── rds/            # PostgreSQL database
    ├── elasticache/    # Redis cache cluster
    ├── msk/            # Managed Kafka (optional)
    ├── eks/            # Kubernetes cluster
    ├── waf/            # Web Application Firewall
    ├── irsa/           # IAM Roles for Service Accounts
    ├── monitoring/     # CloudWatch, alarms, dashboards
    └── s3/             # Logs and backup buckets
```

## Security Features

### Encryption

- **At Rest**: All data encrypted with customer-managed KMS keys
- **In Transit**: TLS 1.2+ for all communications
- **Secrets**: AWS Secrets Manager with automatic rotation

### Network Security

- **VPC Isolation**: Dedicated VPC with private subnets
- **Security Groups**: Least-privilege access rules
- **VPC Endpoints**: Private AWS API access without internet traversal
- **Network Policies**: Kubernetes NetworkPolicies for pod isolation

### Access Control

- **IRSA**: IAM Roles for Service Accounts (no long-lived credentials)
- **Pod Security Standards**: Restricted security context
- **RBAC**: Kubernetes role-based access control

### Compliance

- **Audit Logging**: CloudWatch, CloudTrail, EKS audit logs
- **Object Lock**: Immutable audit logs for compliance
- **WAF**: Protection against OWASP Top 10

## Environment Differences

| Feature             | Development | Production  |
| ------------------- | ----------- | ----------- |
| RDS Instance        | db.t3.small | db.r5.large |
| RDS Multi-AZ        | No          | Yes         |
| Redis Nodes         | 1           | 3           |
| EKS Nodes           | 2-4         | 2-10        |
| MSK Kafka           | Disabled    | Enabled     |
| WAF                 | Disabled    | Enabled     |
| NAT Gateways        | 1           | 3 (per AZ)  |
| Deletion Protection | No          | Yes         |
| Log Retention       | 7 days      | 365 days    |

## Cost Optimization

Development environment is optimized for cost:

- Single NAT gateway
- Smaller instance sizes
- Optional services disabled
- Shorter log retention

## Disaster Recovery

Production includes:

- Multi-AZ deployment for all components
- Automated backups with 30-day retention
- Cross-region replication capability (DR region: us-west-2)
- Point-in-time recovery for RDS

## Deployment Script

Use the deployment script for automated deployments:

```bash
# Plan changes
./deploy.sh plan

# Apply changes
./deploy.sh apply

# Deploy to Kubernetes
./deploy.sh deploy-k8s

# Full deployment
ENVIRONMENT=prod ./deploy.sh all

# Check status
./deploy.sh status

# Destroy (use with caution!)
./deploy.sh destroy
```

## Outputs

After deployment, retrieve important outputs:

```bash
# Get all outputs
terraform output

# Get specific output
terraform output eks_cluster_endpoint
terraform output configure_kubectl

# Get connection info
terraform output database_connection_info
terraform output redis_connection_info
```

## Local Development

For local development without AWS:

1. Copy `secrets.example.json` to `secrets.json`
2. Set `USE_LOCAL_SECRETS=true` environment variable
3. Use docker-compose for local services

```bash
cd infra
docker-compose up -d
```

## Testing

Run infrastructure tests:

```bash
cd tests
go test -v ./...

# With actual AWS deployment (requires credentials)
RUN_TERRAFORM_TESTS=true go test -v ./...
```

## Troubleshooting

### Common Issues

1. **Terraform state lock**: Delete the DynamoDB lock entry
2. **EKS authentication**: Update kubeconfig
3. **Secret access denied**: Check IRSA role annotations
4. **VPC endpoint errors**: Verify security group rules

### Useful Commands

```bash
# Check EKS cluster status
kubectl get nodes
kubectl get pods -n neobank

# View pod logs
kubectl logs -n neobank deployment/identity-service

# Check secrets access
kubectl exec -it <pod> -- aws secretsmanager get-secret-value --secret-id <arn>

# View CloudWatch logs
aws logs tail /neobank/dev/application --follow
```

## Contributing

1. Create a feature branch
2. Run `terraform fmt` and `terraform validate`
3. Test changes in dev environment
4. Create pull request with detailed description

## License

MIT License - see LICENSE file
