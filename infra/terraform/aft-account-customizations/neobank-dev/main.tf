# ============================================================================
# NeoBank Dev Account Customizations
# Infrastructure baselines specific to the development account
# ============================================================================

terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    # AFT automatically configures the backend
  }
}

provider "aws" {
  region = var.aws_region
}

# ============================================================================
# VPC for Development Environment
# ============================================================================

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "neobank-dev-vpc"
  cidr = "10.0.0.0/16"

  azs              = ["${var.aws_region}a", "${var.aws_region}b", "${var.aws_region}c"]
  private_subnets  = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets   = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
  database_subnets = ["10.0.201.0/24", "10.0.202.0/24", "10.0.203.0/24"]

  enable_nat_gateway   = true
  enable_vpn_gateway   = false
  enable_dns_hostnames = true
  enable_dns_support   = true

  # Enable VPC Flow Logs
  enable_flow_log                      = true
  create_flow_log_cloudwatch_iam_role  = true
  create_flow_log_cloudwatch_log_group = true

  # Tags for EKS integration
  public_subnet_tags = {
    "kubernetes.io/role/elb"                = "1"
    "kubernetes.io/cluster/neobank-dev-eks" = "shared"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb"       = "1"
    "kubernetes.io/cluster/neobank-dev-eks" = "shared"
  }

  tags = {
    Environment = "dev"
    Application = "neobank"
    ManagedBy   = "AFT"
  }
}

# ============================================================================
# KMS Keys for Data Encryption
# ============================================================================

resource "aws_kms_key" "rds" {
  description             = "KMS key for RDS encryption - Dev"
  deletion_window_in_days = 30
  enable_key_rotation     = true

  tags = {
    Name        = "rds-encryption-dev"
    Environment = "dev"
  }
}

resource "aws_kms_alias" "rds" {
  name          = "alias/rds-dev"
  target_key_id = aws_kms_key.rds.key_id
}

resource "aws_kms_key" "eks" {
  description             = "KMS key for EKS encryption - Dev"
  deletion_window_in_days = 30
  enable_key_rotation     = true

  tags = {
    Name        = "eks-encryption-dev"
    Environment = "dev"
  }
}

resource "aws_kms_alias" "eks" {
  name          = "alias/eks-dev"
  target_key_id = aws_kms_key.eks.key_id
}

# ============================================================================
# IAM Roles for EKS Service Accounts (IRSA)
# ============================================================================

# OIDC Provider for EKS (will be created with EKS cluster)
# This is a placeholder - actual OIDC provider URL will come from EKS

data "aws_iam_policy_document" "eks_oidc_assume_role" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"

    condition {
      test     = "StringEquals"
      variable = "oidc.eks.${var.aws_region}.amazonaws.com/id/OIDC_ID:sub"
      values   = ["system:serviceaccount:neobank:*"]
    }

    principals {
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/oidc.eks.${var.aws_region}.amazonaws.com/id/OIDC_ID"]
      type        = "Federated"
    }
  }
}

# External Secrets IAM Role
resource "aws_iam_role" "external_secrets" {
  name               = "neobank-dev-external-secrets"
  assume_role_policy = data.aws_iam_policy_document.eks_oidc_assume_role.json

  tags = {
    Name        = "external-secrets-role"
    Environment = "dev"
  }
}

resource "aws_iam_role_policy_attachment" "external_secrets_secrets_manager" {
  role       = aws_iam_role.external_secrets.name
  policy_arn = "arn:aws:iam::aws:policy/SecretsManagerReadWrite"
}

# ============================================================================
# Secrets Manager Secrets Placeholders
# ============================================================================

resource "aws_secretsmanager_secret" "db_credentials" {
  name        = "neobank-dev/db-credentials"
  description = "Database credentials for neobank dev"

  kms_key_id = aws_kms_key.rds.id

  tags = {
    Environment = "dev"
    Application = "neobank"
  }
}

resource "aws_secretsmanager_secret" "jwt_secret" {
  name        = "neobank-dev/jwt-secret"
  description = "JWT secret for neobank dev"

  kms_key_id = aws_kms_key.rds.id

  tags = {
    Environment = "dev"
    Application = "neobank"
  }
}

resource "aws_secretsmanager_secret" "redis_credentials" {
  name        = "neobank-dev/redis-credentials"
  description = "Redis credentials for neobank dev"

  kms_key_id = aws_kms_key.rds.id

  tags = {
    Environment = "dev"
    Application = "neobank"
  }
}

# ============================================================================
# Data Sources
# ============================================================================

data "aws_caller_identity" "current" {}

# ============================================================================
# Outputs
# ============================================================================

output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = module.vpc.private_subnets
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = module.vpc.public_subnets
}

output "database_subnet_ids" {
  description = "Database subnet IDs"
  value       = module.vpc.database_subnets
}

output "rds_kms_key_arn" {
  description = "RDS KMS Key ARN"
  value       = aws_kms_key.rds.arn
}

output "eks_kms_key_arn" {
  description = "EKS KMS Key ARN"
  value       = aws_kms_key.eks.arn
}

output "external_secrets_role_arn" {
  description = "External Secrets IAM Role ARN"
  value       = aws_iam_role.external_secrets.arn
}
