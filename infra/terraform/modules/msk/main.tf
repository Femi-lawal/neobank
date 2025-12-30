# ============================================================================
# MSK Module - Amazon Managed Streaming for Kafka
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
  description = "List of subnet IDs (must be at least 2 for MSK)"
  type        = list(string)
}

variable "security_group_ids" {
  description = "List of security group IDs"
  type        = list(string)
}

variable "kafka_version" {
  description = "Apache Kafka version"
  type        = string
  default     = "3.5.1"
}

variable "instance_type" {
  description = "MSK broker instance type"
  type        = string
  default     = "kafka.t3.small"
}

variable "number_of_broker_nodes" {
  description = "Number of broker nodes"
  type        = number
  default     = 3
}

variable "ebs_volume_size" {
  description = "EBS volume size in GB"
  type        = number
  default     = 100
}

variable "kms_key_arn" {
  description = "ARN of KMS key for encryption"
  type        = string
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}

# -----------------------------------------------------------------------------
# MSK Configuration
# -----------------------------------------------------------------------------
resource "aws_msk_configuration" "main" {
  name           = "${var.name_prefix}-msk-config"
  kafka_versions = [var.kafka_version]
  description    = "MSK configuration for ${var.name_prefix}"

  server_properties = <<PROPERTIES
auto.create.topics.enable=false
delete.topic.enable=true
default.replication.factor=3
min.insync.replicas=2
num.partitions=3
log.retention.hours=168
log.segment.bytes=1073741824
PROPERTIES
}

# -----------------------------------------------------------------------------
# MSK Cluster
# -----------------------------------------------------------------------------
resource "aws_msk_cluster" "main" {
  cluster_name           = "${var.name_prefix}-msk"
  kafka_version          = var.kafka_version
  number_of_broker_nodes = var.number_of_broker_nodes

  broker_node_group_info {
    instance_type   = var.instance_type
    client_subnets  = var.subnet_ids
    security_groups = var.security_group_ids

    storage_info {
      ebs_storage_info {
        volume_size = var.ebs_volume_size
      }
    }
  }

  configuration_info {
    arn      = aws_msk_configuration.main.arn
    revision = aws_msk_configuration.main.latest_revision
  }

  encryption_info {
    encryption_at_rest_kms_key_arn = var.kms_key_arn

    encryption_in_transit {
      client_broker = "TLS"
      in_cluster    = true
    }
  }

  client_authentication {
    sasl {
      iam   = true
      scram = false
    }
    unauthenticated = false
  }

  logging_info {
    broker_logs {
      cloudwatch_logs {
        enabled   = true
        log_group = aws_cloudwatch_log_group.msk.name
      }
    }
  }

  tags = merge(var.tags, {
    Name = "${var.name_prefix}-msk"
  })
}

# -----------------------------------------------------------------------------
# CloudWatch Log Group for MSK
# -----------------------------------------------------------------------------
resource "aws_cloudwatch_log_group" "msk" {
  name              = "/aws/msk/${var.name_prefix}"
  retention_in_days = 30

  tags = var.tags
}

# -----------------------------------------------------------------------------
# Outputs
# -----------------------------------------------------------------------------
output "cluster_arn" {
  description = "MSK cluster ARN"
  value       = aws_msk_cluster.main.arn
}

output "bootstrap_brokers_tls" {
  description = "TLS connection string for Kafka brokers"
  value       = aws_msk_cluster.main.bootstrap_brokers_tls
}

output "bootstrap_brokers_iam" {
  description = "IAM connection string for Kafka brokers"
  value       = aws_msk_cluster.main.bootstrap_brokers_sasl_iam
}

output "zookeeper_connect_string" {
  description = "Zookeeper connection string"
  value       = aws_msk_cluster.main.zookeeper_connect_string
}
