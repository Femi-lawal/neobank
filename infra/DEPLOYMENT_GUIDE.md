# NeoBank AWS Infrastructure Deployment Guide

## Prerequisites

### Required Tools

1. **Terraform** (>= 1.5.0)

   ```powershell
   # Windows (using winget)
   winget install Hashicorp.Terraform

   # Windows (using Chocolatey)
   choco install terraform

   # Verify installation
   terraform --version
   ```

2. **AWS CLI** (>= 2.0)

   ```powershell
   # Windows (using winget)
   winget install Amazon.AWSCLI

   # Windows (using Chocolatey)
   choco install awscli

   # Verify installation
   aws --version
   ```

3. **kubectl**

   ```powershell
   # Windows (using winget)
   winget install Kubernetes.kubectl

   # Windows (using Chocolatey)
   choco install kubernetes-cli

   # Verify installation
   kubectl version --client
   ```

### AWS Account Setup

1. **Create an IAM User** with the following permissions:

   - `AdministratorAccess` (for initial setup) OR
   - Custom policy with:
     - VPC Full Access
     - EKS Full Access
     - RDS Full Access
     - ElastiCache Full Access
     - Secrets Manager Full Access
     - KMS Full Access
     - S3 Full Access
     - IAM Limited Access (for IRSA)
     - CloudWatch Full Access
     - WAF Full Access (if using WAF)
     - MSK Full Access (if using Kafka)

2. **Generate Access Keys**:

   - Go to IAM Console → Users → Your User → Security Credentials
   - Create Access Key → Download the CSV file

3. **Configure AWS CLI**:

   ```powershell
   aws configure
   # Enter your Access Key ID
   # Enter your Secret Access Key
   # Enter default region: us-east-1
   # Enter default output format: json

   # Verify configuration
   aws sts get-caller-identity
   ```

## Deployment Steps

### Step 1: Initialize Terraform

```powershell
cd infra\terraform
terraform init
```

### Step 2: Review the Plan (Development Environment)

```powershell
terraform plan -var-file="environments/dev.tfvars"
```

This will show:

- ~60+ resources to be created
- VPC with 3 AZs, 9 subnets
- EKS cluster with managed node group
- RDS PostgreSQL (Multi-AZ optional)
- ElastiCache Redis
- Secrets Manager secrets
- Security groups and IAM roles

### Step 3: Apply the Configuration

```powershell
# Using the deployment script
.\deploy.ps1 -Action apply -Environment dev

# Or manually
terraform apply -var-file="environments/dev.tfvars" -auto-approve
```

**Note**: First deployment takes approximately 25-35 minutes due to:

- EKS cluster creation (~15 min)
- RDS instance creation (~10 min)
- ElastiCache cluster creation (~5 min)

### Step 4: Configure kubectl

After deployment, configure kubectl to connect to the EKS cluster:

```powershell
# Get the cluster name from Terraform output
$clusterName = terraform output -raw eks_cluster_name

# Update kubeconfig
aws eks update-kubeconfig --name $clusterName --region us-east-1

# Verify connection
kubectl get nodes
```

### Step 5: Deploy Application to Kubernetes

```powershell
# Apply AWS-specific Kubernetes resources
kubectl apply -k k8s/overlays/aws

# Verify deployments
kubectl get pods -n neobank
kubectl get services -n neobank
```

## Environment Configurations

### Development (`dev.tfvars`)

- Single NAT Gateway (cost optimization)
- Smaller instance sizes (t3.small/medium)
- No MSK (Kafka disabled)
- No WAF
- 7-day log retention
- Single-AZ RDS

### Production (`prod.tfvars`)

- Multi-AZ NAT Gateways
- Production instance sizes (r5.large)
- MSK enabled for event streaming
- WAF enabled for security
- 365-day log retention
- Multi-AZ RDS with deletion protection

## Cost Estimation

### Development Environment (~$200-300/month)

- EKS Cluster: $73/month
- EKS Nodes (2x t3.medium): ~$60/month
- RDS (db.t3.small): ~$25/month
- ElastiCache (cache.t3.micro): ~$13/month
- NAT Gateway: ~$32/month
- S3/CloudWatch: ~$10/month

### Production Environment (~$1,500-2,500/month)

- EKS Cluster: $73/month
- EKS Nodes (3x m5.large): ~$200/month
- RDS (db.r5.large Multi-AZ): ~$400/month
- ElastiCache (cache.r5.large, 3 nodes): ~$350/month
- MSK (3x kafka.m5.large): ~$500/month
- NAT Gateways (3x): ~$100/month
- WAF: ~$10/month + per request
- S3/CloudWatch: ~$50/month

## Important Outputs

After deployment, retrieve important values:

```powershell
# Get all outputs
terraform output

# Specific outputs
terraform output eks_cluster_endpoint
terraform output rds_endpoint
terraform output redis_endpoint
terraform output -json irsa_role_arns
```

## Cleanup

To destroy all resources:

```powershell
# Using the deployment script
.\deploy.ps1 -Action destroy -Environment dev

# Or manually
terraform destroy -var-file="environments/dev.tfvars"
```

**Warning**: This will delete all data! Make sure to backup any important data before destroying.

## Troubleshooting

### Common Issues

1. **"No valid credential sources found"**

   ```powershell
   # Verify AWS credentials
   aws sts get-caller-identity

   # If not configured, run
   aws configure
   ```

2. **"Error creating EKS Cluster: AccessDeniedException"**

   - Ensure your IAM user has EKS permissions
   - Check if service-linked role exists:
     ```powershell
     aws iam get-role --role-name AWSServiceRoleForAmazonEKS
     ```

3. **"Error creating DB Instance: DBSubnetGroupDoesNotCoverEnoughAZs"**

   - Ensure you have at least 2 AZs in your subnet configuration
   - Check database_subnet_cidrs in your tfvars

4. **kubectl cannot connect to cluster**

   ```powershell
   # Update kubeconfig
   aws eks update-kubeconfig --name neobank-dev-eks --region us-east-1

   # Check cluster status
   aws eks describe-cluster --name neobank-dev-eks --query cluster.status
   ```

5. **Pods cannot pull from ECR**
   - Verify IRSA is configured correctly
   - Check service account annotations:
     ```powershell
     kubectl describe sa identity-service -n neobank
     ```

## Security Notes

1. **Never commit AWS credentials** to version control
2. **Use environment-specific tfvars** files
3. **Enable deletion protection** in production
4. **Review security group rules** before deployment
5. **Enable MFA** on your AWS account
6. **Rotate secrets** regularly using Secrets Manager rotation

## Support

For issues or questions:

1. Check the [Terraform documentation](https://developer.hashicorp.com/terraform/docs)
2. Review [AWS EKS best practices](https://aws.github.io/aws-eks-best-practices/)
3. Open an issue in the repository
