# ============================================================================
# ElastiCache Module - Redis Cluster
# ============================================================================

variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "subnet_ids" {
  description = "List of subnet IDs"
  type        = list(string)
}

variable "security_group_ids" {
  description = "List of security group IDs"
  type        = list(string)
}

variable "node_type" {
  description = "ElastiCache node type"
  type        = string
  default     = "cache.t3.micro"
}

variable "num_cache_nodes" {
  description = "Number of cache nodes"
  type        = number
  default     = 1
}

variable "engine_version" {
  description = "Redis engine version"
  type        = string
  default     = "7.1"
}

variable "parameter_group_family" {
  description = "Redis parameter group family"
  type        = string
  default     = "redis7"
}

variable "at_rest_encryption_enabled" {
  description = "Enable encryption at rest"
  type        = bool
  default     = true
}

variable "transit_encryption_enabled" {
  description = "Enable encryption in transit"
  type        = bool
  default     = true
}

variable "auth_token" {
  description = "Auth token for Redis"
  type        = string
  sensitive   = true
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}

# -----------------------------------------------------------------------------
# Subnet Group
# -----------------------------------------------------------------------------
resource "aws_elasticache_subnet_group" "main" {
  name        = "${var.name_prefix}-redis-subnet-group"
  description = "Redis subnet group for ${var.name_prefix}"
  subnet_ids  = var.subnet_ids

  tags = var.tags
}

# -----------------------------------------------------------------------------
# Parameter Group
# -----------------------------------------------------------------------------
resource "aws_elasticache_parameter_group" "main" {
  name        = "${var.name_prefix}-redis-params"
  family      = var.parameter_group_family
  description = "Redis parameters for ${var.name_prefix}"

  # Enable keyspace notifications for cache invalidation
  parameter {
    name  = "notify-keyspace-events"
    value = "Ex"
  }

  # Timeout for idle connections
  parameter {
    name  = "timeout"
    value = "300"
  }

  tags = var.tags
}

# -----------------------------------------------------------------------------
# Redis Cluster
# -----------------------------------------------------------------------------
resource "aws_elasticache_cluster" "main" {
  cluster_id           = "${var.name_prefix}-redis"
  engine               = "redis"
  engine_version       = var.engine_version
  node_type            = var.node_type
  num_cache_nodes      = var.num_cache_nodes
  parameter_group_name = aws_elasticache_parameter_group.main.name
  port                 = 6379

  subnet_group_name  = aws_elasticache_subnet_group.main.name
  security_group_ids = var.security_group_ids

  # Maintenance
  maintenance_window = "sun:05:00-sun:06:00"

  # Snapshots
  snapshot_retention_limit = 7
  snapshot_window          = "03:00-04:00"

  # Auto minor version upgrade
  auto_minor_version_upgrade = true

  tags = merge(var.tags, {
    Name = "${var.name_prefix}-redis"
  })
}

# For production with auth token and encryption, use replication group instead
resource "aws_elasticache_replication_group" "main" {
  count = var.transit_encryption_enabled ? 1 : 0

  replication_group_id = "${var.name_prefix}-redis-rg"
  description          = "Redis replication group for ${var.name_prefix}"

  engine               = "redis"
  engine_version       = var.engine_version
  node_type            = var.node_type
  num_cache_clusters   = var.num_cache_nodes
  parameter_group_name = aws_elasticache_parameter_group.main.name
  port                 = 6379

  subnet_group_name  = aws_elasticache_subnet_group.main.name
  security_group_ids = var.security_group_ids

  # Encryption
  at_rest_encryption_enabled = var.at_rest_encryption_enabled
  transit_encryption_enabled = var.transit_encryption_enabled
  auth_token                 = var.auth_token

  # Maintenance
  maintenance_window = "sun:05:00-sun:06:00"

  # Snapshots
  snapshot_retention_limit = 7
  snapshot_window          = "03:00-04:00"

  # Auto minor version upgrade
  auto_minor_version_upgrade = true

  tags = merge(var.tags, {
    Name = "${var.name_prefix}-redis-rg"
  })
}

# -----------------------------------------------------------------------------
# Outputs
# -----------------------------------------------------------------------------
output "endpoint" {
  description = "Redis endpoint"
  value       = var.transit_encryption_enabled ? aws_elasticache_replication_group.main[0].primary_endpoint_address : aws_elasticache_cluster.main.cache_nodes[0].address
}

output "port" {
  description = "Redis port"
  value       = 6379
}

output "configuration_endpoint" {
  description = "Redis configuration endpoint (for cluster mode)"
  value       = var.transit_encryption_enabled ? aws_elasticache_replication_group.main[0].configuration_endpoint_address : null
}

output "cluster_id" {
  description = "ElastiCache cluster identifier"
  value       = var.transit_encryption_enabled ? aws_elasticache_replication_group.main[0].id : aws_elasticache_cluster.main.cluster_id
}
