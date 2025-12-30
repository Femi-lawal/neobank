# ============================================================================
# Production Environment Configuration
# ============================================================================

environment  = "prod"
project_name = "neobank"
aws_region   = "us-east-1"
dr_region    = "us-west-2"

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
eks_node_instance_types     = ["m5.large", "m5.xlarge"]
eks_node_desired_size       = 3
eks_node_min_size           = 2
eks_node_max_size           = 10
enable_eks_private_endpoint = true
enable_eks_public_endpoint  = false # Disabled for production security

# -----------------------------------------------------------------------------
# RDS Configuration
# -----------------------------------------------------------------------------
rds_instance_class          = "db.r5.large"
rds_allocated_storage       = 100
rds_max_allocated_storage   = 1000
rds_engine_version          = "16.1"
rds_multi_az                = true
rds_deletion_protection     = true
rds_backup_retention_period = 30

# -----------------------------------------------------------------------------
# ElastiCache Configuration
# -----------------------------------------------------------------------------
redis_node_type              = "cache.r5.large"
redis_num_cache_nodes        = 3
redis_engine_version         = "7.1"
redis_parameter_group_family = "redis7"

# -----------------------------------------------------------------------------
# MSK Configuration (Enabled for prod)
# -----------------------------------------------------------------------------
enable_msk                 = true
msk_instance_type          = "kafka.m5.large"
msk_number_of_broker_nodes = 3
msk_ebs_volume_size        = 500
msk_kafka_version          = "3.5.1"

# -----------------------------------------------------------------------------
# Security Configuration
# -----------------------------------------------------------------------------
enable_waf          = true
allowed_cidr_blocks = [] # No public access in production

# -----------------------------------------------------------------------------
# Monitoring Configuration
# -----------------------------------------------------------------------------
enable_cloudwatch_logs        = true
cloudwatch_log_retention_days = 365
enable_container_insights     = true

# -----------------------------------------------------------------------------
# Tags
# -----------------------------------------------------------------------------
additional_tags = {
  CostCenter = "production"
  ManagedBy  = "terraform"
  Compliance = "PCI-DSS"
}
