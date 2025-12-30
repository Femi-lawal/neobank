# ============================================================================
# LocalStack Test Environment Configuration
# ============================================================================
# This configuration is used for testing Terraform with LocalStack
# LocalStack emulates AWS services locally for development and testing
# ============================================================================

environment  = "localstack"
project_name = "neobank"
aws_region   = "us-east-1"

# -----------------------------------------------------------------------------
# Network Configuration (simplified for LocalStack)
# -----------------------------------------------------------------------------
vpc_cidr              = "10.0.0.0/16"
availability_zones    = ["us-east-1a", "us-east-1b"]
private_subnet_cidrs  = ["10.0.1.0/24", "10.0.2.0/24"]
public_subnet_cidrs   = ["10.0.101.0/24", "10.0.102.0/24"]
database_subnet_cidrs = ["10.0.201.0/24", "10.0.202.0/24"]

# -----------------------------------------------------------------------------
# EKS Configuration (minimal for testing)
# -----------------------------------------------------------------------------
eks_cluster_version         = "1.29"
eks_node_instance_types     = ["t3.small"]
eks_node_desired_size       = 1
eks_node_min_size           = 1
eks_node_max_size           = 2
enable_eks_private_endpoint = true
enable_eks_public_endpoint  = true

# -----------------------------------------------------------------------------
# RDS Configuration (minimal for testing)
# -----------------------------------------------------------------------------
rds_instance_class          = "db.t3.micro"
rds_allocated_storage       = 20
rds_max_allocated_storage   = 50
rds_engine_version          = "16.1"
rds_multi_az                = false
rds_deletion_protection     = false
rds_backup_retention_period = 1

# -----------------------------------------------------------------------------
# ElastiCache Configuration (minimal for testing)
# -----------------------------------------------------------------------------
redis_node_type              = "cache.t3.micro"
redis_num_cache_nodes        = 1
redis_engine_version         = "7.1"
redis_parameter_group_family = "redis7"

# -----------------------------------------------------------------------------
# MSK and WAF disabled for LocalStack testing
# -----------------------------------------------------------------------------
enable_msk                 = false
msk_instance_type          = "kafka.t3.small"
msk_number_of_broker_nodes = 2
msk_ebs_volume_size        = 50
msk_kafka_version          = "3.5.1"

enable_waf          = false
allowed_cidr_blocks = ["0.0.0.0/0"]

# -----------------------------------------------------------------------------
# Monitoring Configuration
# -----------------------------------------------------------------------------
enable_cloudwatch_logs        = true
cloudwatch_log_retention_days = 1
enable_container_insights     = false

# -----------------------------------------------------------------------------
# Tags
# -----------------------------------------------------------------------------
additional_tags = {
  Environment = "localstack"
  Testing     = "true"
  ManagedBy   = "terraform"
}
