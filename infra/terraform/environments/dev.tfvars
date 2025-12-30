# ============================================================================
# Development Environment Configuration
# ============================================================================

environment       = "dev"
project_name      = "neobank"
aws_region        = "us-east-1"
target_account_id = "243221828782" # neobank-dev account

# -----------------------------------------------------------------------------
# Network Configuration
# -----------------------------------------------------------------------------
vpc_cidr              = "10.0.0.0/16"
availability_zones    = ["us-east-1a", "us-east-1b", "us-east-1c"]
private_subnet_cidrs  = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
public_subnet_cidrs   = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
database_subnet_cidrs = ["10.0.201.0/24", "10.0.202.0/24", "10.0.203.0/24"]

# -----------------------------------------------------------------------------
# EKS Configuration
# -----------------------------------------------------------------------------
eks_cluster_version         = "1.29"
eks_node_instance_types     = ["t3.medium"]
eks_node_desired_size       = 2
eks_node_min_size           = 1
eks_node_max_size           = 4
enable_eks_private_endpoint = true
enable_eks_public_endpoint  = true

# -----------------------------------------------------------------------------
# RDS Configuration
# -----------------------------------------------------------------------------
rds_instance_class          = "db.t3.small"
rds_allocated_storage       = 20
rds_max_allocated_storage   = 100
rds_engine_version          = "16.6" # Updated to match current version
rds_multi_az                = false
rds_deletion_protection     = false
rds_backup_retention_period = 7

# -----------------------------------------------------------------------------
# ElastiCache Configuration
# -----------------------------------------------------------------------------
redis_node_type              = "cache.t3.micro"
redis_num_cache_nodes        = 1
redis_engine_version         = "7.1"
redis_parameter_group_family = "redis7"

# -----------------------------------------------------------------------------
# MSK Configuration (Disabled for dev)
# -----------------------------------------------------------------------------
enable_msk                 = false
msk_instance_type          = "kafka.t3.small"
msk_number_of_broker_nodes = 3
msk_ebs_volume_size        = 100
msk_kafka_version          = "3.5.1"

# -----------------------------------------------------------------------------
# Security Configuration
# -----------------------------------------------------------------------------
enable_waf          = false
allowed_cidr_blocks = ["0.0.0.0/0"]

# -----------------------------------------------------------------------------
# Monitoring Configuration
# -----------------------------------------------------------------------------
enable_cloudwatch_logs        = true
cloudwatch_log_retention_days = 7
enable_container_insights     = true

# -----------------------------------------------------------------------------
# Tags
# -----------------------------------------------------------------------------
additional_tags = {
  CostCenter = "development"
  ManagedBy  = "terraform"
}

# -----------------------------------------------------------------------------
# ECR Configuration
# -----------------------------------------------------------------------------
ecr_repository_names = [
  "neobank/identity-service",
  "neobank/ledger-service",
  "neobank/payment-service",
  "neobank/product-service",
  "neobank/card-service",
  "neobank/frontend"
]
