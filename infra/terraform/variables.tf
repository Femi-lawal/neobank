# ============================================================================
# NeoBank AWS Infrastructure - Variables
# ============================================================================

# -----------------------------------------------------------------------------
# General Configuration
# -----------------------------------------------------------------------------
variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "neobank"
}

variable "environment" {
  description = "Environment name (dev, staging, prod, localstack)"
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "staging", "prod", "localstack"], var.environment)
    error_message = "Environment must be dev, staging, prod, or localstack."
  }
}

variable "aws_region" {
  description = "Primary AWS region for deployment"
  type        = string
  default     = "us-east-1"
}

variable "target_account_id" {
  description = "Target AWS account ID for cross-account deployment (leave empty for management account)"
  type        = string
  default     = ""
}

variable "assume_role_enabled" {
  description = "Enable role assumption for cross-account deployment (set to false when already in target account)"
  type        = bool
  default     = false
}

variable "dr_region" {
  description = "Disaster recovery AWS region"
  type        = string
  default     = "us-west-2"
}

# -----------------------------------------------------------------------------
# Network Configuration
# -----------------------------------------------------------------------------
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones to use"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets (one per AZ)"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets (one per AZ)"
  type        = list(string)
  default     = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
}

variable "database_subnet_cidrs" {
  description = "CIDR blocks for database subnets (isolated, one per AZ)"
  type        = list(string)
  default     = ["10.0.201.0/24", "10.0.202.0/24", "10.0.203.0/24"]
}

# -----------------------------------------------------------------------------
# EKS Configuration
# -----------------------------------------------------------------------------
variable "eks_cluster_version" {
  description = "Kubernetes version for EKS cluster"
  type        = string
  default     = "1.29"
}

variable "eks_node_instance_types" {
  description = "Instance types for EKS managed node groups"
  type        = list(string)
  default     = ["t3.medium"]
}

variable "eks_node_desired_size" {
  description = "Desired number of nodes in EKS node group"
  type        = number
  default     = 2
}

variable "eks_node_min_size" {
  description = "Minimum number of nodes in EKS node group"
  type        = number
  default     = 1
}

variable "eks_node_max_size" {
  description = "Maximum number of nodes in EKS node group"
  type        = number
  default     = 5
}

variable "enable_eks_private_endpoint" {
  description = "Enable private API server endpoint for EKS"
  type        = bool
  default     = true
}

variable "enable_eks_public_endpoint" {
  description = "Enable public API server endpoint for EKS"
  type        = bool
  default     = true
}

# -----------------------------------------------------------------------------
# Database Configuration (RDS PostgreSQL)
# -----------------------------------------------------------------------------
variable "rds_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "rds_allocated_storage" {
  description = "Allocated storage for RDS in GB"
  type        = number
  default     = 20
}

variable "rds_max_allocated_storage" {
  description = "Maximum allocated storage for RDS autoscaling in GB"
  type        = number
  default     = 100
}

variable "rds_engine_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "16.6"
}

variable "rds_multi_az" {
  description = "Enable Multi-AZ deployment for RDS"
  type        = bool
  default     = false
}

variable "rds_backup_retention_period" {
  description = "Number of days to retain RDS backups"
  type        = number
  default     = 7
}

variable "rds_deletion_protection" {
  description = "Enable deletion protection for RDS"
  type        = bool
  default     = false
}

# -----------------------------------------------------------------------------
# ElastiCache (Redis) Configuration
# -----------------------------------------------------------------------------
variable "redis_node_type" {
  description = "ElastiCache Redis node type"
  type        = string
  default     = "cache.t3.micro"
}

variable "redis_num_cache_nodes" {
  description = "Number of cache nodes in the Redis cluster"
  type        = number
  default     = 1
}

variable "redis_engine_version" {
  description = "Redis engine version"
  type        = string
  default     = "7.1"
}

variable "redis_parameter_group_family" {
  description = "Redis parameter group family"
  type        = string
  default     = "redis7"
}

# -----------------------------------------------------------------------------
# MSK (Kafka) Configuration
# -----------------------------------------------------------------------------
variable "enable_msk" {
  description = "Enable Amazon MSK for Kafka"
  type        = bool
  default     = false
}

variable "msk_kafka_version" {
  description = "Apache Kafka version for MSK"
  type        = string
  default     = "3.5.1"
}

variable "msk_instance_type" {
  description = "MSK broker instance type"
  type        = string
  default     = "kafka.t3.small"
}

variable "msk_number_of_broker_nodes" {
  description = "Number of broker nodes in MSK cluster"
  type        = number
  default     = 3
}

variable "msk_ebs_volume_size" {
  description = "EBS volume size for MSK brokers in GB"
  type        = number
  default     = 100
}

# -----------------------------------------------------------------------------
# Security Configuration
# -----------------------------------------------------------------------------
variable "enable_waf" {
  description = "Enable AWS WAF for ALB"
  type        = bool
  default     = true
}

variable "allowed_cidr_blocks" {
  description = "CIDR blocks allowed to access the cluster"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

# -----------------------------------------------------------------------------
# Monitoring Configuration
# -----------------------------------------------------------------------------
variable "enable_cloudwatch_logs" {
  description = "Enable CloudWatch logs for EKS control plane"
  type        = bool
  default     = true
}

variable "cloudwatch_log_retention_days" {
  description = "Number of days to retain CloudWatch logs"
  type        = number
  default     = 30
}

variable "enable_container_insights" {
  description = "Enable CloudWatch Container Insights for EKS"
  type        = bool
  default     = true
}

# -----------------------------------------------------------------------------
# Tags
# -----------------------------------------------------------------------------
variable "additional_tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}

# -----------------------------------------------------------------------------
# ECR Configuration
# -----------------------------------------------------------------------------
variable "ecr_repository_names" {
  description = "List of ECR repository names to create"
  type        = list(string)
  default     = []
}
