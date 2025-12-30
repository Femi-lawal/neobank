# ============================================================================
# NeoBank AWS Infrastructure - Main Configuration
# ============================================================================
# This creates a production-ready AWS infrastructure for the NeoBank platform
# following PCI-DSS, SOC 2, and NIST best practices.
# ============================================================================

locals {
  name_prefix = "${var.project_name}-${var.environment}"

  common_tags = merge(var.additional_tags, {
    Project             = var.project_name
    Environment         = var.environment
    ManagedBy           = "Terraform"
    Owner               = "platform-team"
    CostCenter          = "engineering"
    DataClassification  = "confidential"
    ComplianceFramework = "PCI-DSS-SOC2"
  })
}

# -----------------------------------------------------------------------------
# Data Sources
# -----------------------------------------------------------------------------
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"
}

# -----------------------------------------------------------------------------
# VPC Module - Network Foundation
# -----------------------------------------------------------------------------
module "vpc" {
  source = "./modules/vpc"

  name_prefix        = local.name_prefix
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones

  private_subnet_cidrs  = var.private_subnet_cidrs
  public_subnet_cidrs   = var.public_subnet_cidrs
  database_subnet_cidrs = var.database_subnet_cidrs

  enable_nat_gateway = true
  single_nat_gateway = var.environment != "prod"

  enable_vpn_gateway = false
  enable_flow_logs   = true

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# Security Module - KMS Keys, Security Groups
# -----------------------------------------------------------------------------
module "security" {
  source = "./modules/security"

  name_prefix = local.name_prefix
  vpc_id      = module.vpc.vpc_id
  vpc_cidr    = var.vpc_cidr

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# Secrets Manager - Database & Application Secrets
# -----------------------------------------------------------------------------
module "secrets" {
  source = "./modules/secrets"

  name_prefix = local.name_prefix
  kms_key_arn = module.security.kms_key_arn

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# RDS PostgreSQL - Primary Database
# -----------------------------------------------------------------------------
module "rds" {
  source = "./modules/rds"

  name_prefix = local.name_prefix

  vpc_id             = module.vpc.vpc_id
  subnet_ids         = module.vpc.database_subnet_ids
  security_group_ids = [module.security.rds_security_group_id]

  instance_class        = var.rds_instance_class
  allocated_storage     = var.rds_allocated_storage
  max_allocated_storage = var.rds_max_allocated_storage
  engine_version        = var.rds_engine_version

  database_name   = "newbank_core"
  master_username = "neobank_admin"
  master_password = module.secrets.db_master_password

  multi_az                = var.rds_multi_az
  backup_retention_period = var.rds_backup_retention_period
  deletion_protection     = var.rds_deletion_protection

  kms_key_arn                  = module.security.kms_key_arn
  performance_insights_enabled = var.environment == "prod"

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# ElastiCache Redis - Caching & Session Store
# -----------------------------------------------------------------------------
module "elasticache" {
  source = "./modules/elasticache"

  name_prefix = local.name_prefix

  vpc_id             = module.vpc.vpc_id
  subnet_ids         = module.vpc.private_subnet_ids
  security_group_ids = [module.security.redis_security_group_id]

  node_type              = var.redis_node_type
  num_cache_nodes        = var.redis_num_cache_nodes
  engine_version         = var.redis_engine_version
  parameter_group_family = var.redis_parameter_group_family

  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  auth_token                 = module.secrets.redis_auth_token

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# MSK Kafka - Event Streaming (Optional)
# -----------------------------------------------------------------------------
module "msk" {
  source = "./modules/msk"
  count  = var.enable_msk ? 1 : 0

  name_prefix = local.name_prefix

  vpc_id             = module.vpc.vpc_id
  subnet_ids         = module.vpc.private_subnet_ids
  security_group_ids = [module.security.msk_security_group_id]

  kafka_version          = var.msk_kafka_version
  instance_type          = var.msk_instance_type
  number_of_broker_nodes = var.msk_number_of_broker_nodes
  ebs_volume_size        = var.msk_ebs_volume_size

  kms_key_arn = module.security.kms_key_arn

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# EKS Cluster - Kubernetes Platform
# -----------------------------------------------------------------------------
module "eks" {
  source = "./modules/eks"

  name_prefix = local.name_prefix

  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids

  cluster_security_group_id = module.security.eks_cluster_security_group_id
  node_security_group_id    = module.security.eks_nodes_security_group_id

  kubernetes_version = var.eks_cluster_version

  # Node groups
  node_instance_types = var.eks_node_instance_types
  node_desired_size   = var.eks_node_desired_size
  node_min_size       = var.eks_node_min_size
  node_max_size       = var.eks_node_max_size

  # Encryption
  kms_key_arn = module.security.kms_key_arn

  enable_container_insights = var.enable_container_insights
  environment               = var.environment

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# WAF - Web Application Firewall
# -----------------------------------------------------------------------------
module "waf" {
  source = "./modules/waf"
  count  = var.enable_waf ? 1 : 0

  name_prefix = local.name_prefix

  # Associate with ALB created by EKS Ingress Controller
  enable_logging     = true
  log_retention_days = var.cloudwatch_log_retention_days

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# IAM Roles for Kubernetes Service Accounts (IRSA)
# -----------------------------------------------------------------------------
module "irsa" {
  source = "./modules/irsa"

  name_prefix = local.name_prefix
  namespace   = "neobank"

  oidc_provider_arn = module.eks.oidc_provider_arn
  oidc_provider_url = module.eks.oidc_provider_url

  # Secrets Manager access for each service
  db_secret_arn         = module.secrets.db_credentials_arn
  redis_secret_arn      = module.secrets.redis_credentials_arn
  jwt_secret_arn        = module.secrets.jwt_secret_arn
  app_config_secret_arn = module.secrets.app_config_arn
  kms_key_arn           = module.security.kms_key_arn

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# Monitoring & Observability
# -----------------------------------------------------------------------------
module "monitoring" {
  source = "./modules/monitoring"

  name_prefix = local.name_prefix
  vpc_id      = module.vpc.vpc_id

  eks_cluster_name       = module.eks.cluster_name
  rds_instance_id        = module.rds.instance_id
  elasticache_cluster_id = module.elasticache.cluster_id

  log_retention_days = var.cloudwatch_log_retention_days
  enable_dashboard   = true

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# S3 Buckets - Logs, Backups
# -----------------------------------------------------------------------------
module "s3" {
  source = "./modules/s3"

  name_prefix = local.name_prefix
  environment = var.environment

  kms_key_arn = module.security.kms_key_arn

  # Enable Object Lock for compliance (immutable audit logs)
  enable_object_lock = var.environment == "prod"

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# ECR Repositories - Container Registry
# -----------------------------------------------------------------------------
module "ecr" {
  source = "./modules/ecr"

  repository_names = var.ecr_repository_names

  tags = local.common_tags
}
