# ============================================================================
# Monitoring Module - CloudWatch & Container Insights
# ============================================================================

variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "eks_cluster_name" {
  description = "EKS cluster name"
  type        = string
}

variable "rds_instance_id" {
  description = "RDS instance identifier"
  type        = string
}

variable "elasticache_cluster_id" {
  description = "ElastiCache cluster identifier"
  type        = string
}

variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 30
}

variable "alarm_email" {
  description = "Email address for alarm notifications"
  type        = string
  default     = ""
}

variable "enable_dashboard" {
  description = "Enable CloudWatch dashboard"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}

# -----------------------------------------------------------------------------
# Data Sources
# -----------------------------------------------------------------------------
data "aws_region" "current" {}

# -----------------------------------------------------------------------------
# SNS Topic for Alarms
# -----------------------------------------------------------------------------
resource "aws_sns_topic" "alarms" {
  name = "${var.name_prefix}-alarms"

  tags = var.tags
}

resource "aws_sns_topic_subscription" "email" {
  count = var.alarm_email != "" ? 1 : 0

  topic_arn = aws_sns_topic.alarms.arn
  protocol  = "email"
  endpoint  = var.alarm_email
}

# -----------------------------------------------------------------------------
# CloudWatch Log Groups
# -----------------------------------------------------------------------------
resource "aws_cloudwatch_log_group" "application" {
  name              = "/neobank/${var.name_prefix}/application"
  retention_in_days = var.log_retention_days

  tags = var.tags
}

resource "aws_cloudwatch_log_group" "audit" {
  name              = "/neobank/${var.name_prefix}/audit"
  retention_in_days = 365 # Audit logs retained for 1 year

  tags = var.tags
}

resource "aws_cloudwatch_log_group" "security" {
  name              = "/neobank/${var.name_prefix}/security"
  retention_in_days = 365 # Security logs retained for 1 year

  tags = var.tags
}

# -----------------------------------------------------------------------------
# CloudWatch Alarms - EKS
# -----------------------------------------------------------------------------
resource "aws_cloudwatch_metric_alarm" "eks_cpu_high" {
  alarm_name          = "${var.name_prefix}-eks-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "node_cpu_utilization"
  namespace           = "ContainerInsights"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "EKS cluster CPU utilization is above 80%"
  alarm_actions       = [aws_sns_topic.alarms.arn]
  ok_actions          = [aws_sns_topic.alarms.arn]

  dimensions = {
    ClusterName = var.eks_cluster_name
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "eks_memory_high" {
  alarm_name          = "${var.name_prefix}-eks-memory-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "node_memory_utilization"
  namespace           = "ContainerInsights"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "EKS cluster memory utilization is above 80%"
  alarm_actions       = [aws_sns_topic.alarms.arn]
  ok_actions          = [aws_sns_topic.alarms.arn]

  dimensions = {
    ClusterName = var.eks_cluster_name
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "eks_pod_restarts" {
  alarm_name          = "${var.name_prefix}-eks-pod-restarts"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "pod_number_of_container_restarts"
  namespace           = "ContainerInsights"
  period              = 300
  statistic           = "Sum"
  threshold           = 5
  alarm_description   = "EKS pods are restarting frequently"
  alarm_actions       = [aws_sns_topic.alarms.arn]
  ok_actions          = [aws_sns_topic.alarms.arn]

  dimensions = {
    ClusterName = var.eks_cluster_name
  }

  tags = var.tags
}

# -----------------------------------------------------------------------------
# CloudWatch Alarms - RDS
# -----------------------------------------------------------------------------
resource "aws_cloudwatch_metric_alarm" "rds_cpu_high" {
  alarm_name          = "${var.name_prefix}-rds-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "CPUUtilization"
  namespace           = "AWS/RDS"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "RDS CPU utilization is above 80%"
  alarm_actions       = [aws_sns_topic.alarms.arn]
  ok_actions          = [aws_sns_topic.alarms.arn]

  dimensions = {
    DBInstanceIdentifier = var.rds_instance_id
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "rds_connections_high" {
  alarm_name          = "${var.name_prefix}-rds-connections-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "DatabaseConnections"
  namespace           = "AWS/RDS"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "RDS connection count is high"
  alarm_actions       = [aws_sns_topic.alarms.arn]
  ok_actions          = [aws_sns_topic.alarms.arn]

  dimensions = {
    DBInstanceIdentifier = var.rds_instance_id
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "rds_free_storage_low" {
  alarm_name          = "${var.name_prefix}-rds-storage-low"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 3
  metric_name         = "FreeStorageSpace"
  namespace           = "AWS/RDS"
  period              = 300
  statistic           = "Average"
  threshold           = 5368709120 # 5GB in bytes
  alarm_description   = "RDS free storage is below 5GB"
  alarm_actions       = [aws_sns_topic.alarms.arn]
  ok_actions          = [aws_sns_topic.alarms.arn]

  dimensions = {
    DBInstanceIdentifier = var.rds_instance_id
  }

  tags = var.tags
}

# -----------------------------------------------------------------------------
# CloudWatch Alarms - ElastiCache
# -----------------------------------------------------------------------------
resource "aws_cloudwatch_metric_alarm" "redis_cpu_high" {
  alarm_name          = "${var.name_prefix}-redis-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "CPUUtilization"
  namespace           = "AWS/ElastiCache"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "ElastiCache CPU utilization is above 80%"
  alarm_actions       = [aws_sns_topic.alarms.arn]
  ok_actions          = [aws_sns_topic.alarms.arn]

  dimensions = {
    CacheClusterId = var.elasticache_cluster_id
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "redis_memory_high" {
  alarm_name          = "${var.name_prefix}-redis-memory-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "DatabaseMemoryUsagePercentage"
  namespace           = "AWS/ElastiCache"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "ElastiCache memory usage is above 80%"
  alarm_actions       = [aws_sns_topic.alarms.arn]
  ok_actions          = [aws_sns_topic.alarms.arn]

  dimensions = {
    CacheClusterId = var.elasticache_cluster_id
  }

  tags = var.tags
}

# -----------------------------------------------------------------------------
# CloudWatch Dashboard
# -----------------------------------------------------------------------------
resource "aws_cloudwatch_dashboard" "main" {
  count = var.enable_dashboard ? 1 : 0

  dashboard_name = "${var.name_prefix}-dashboard"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "text"
        x      = 0
        y      = 0
        width  = 24
        height = 1
        properties = {
          markdown = "# NeoBank Infrastructure Dashboard - ${var.name_prefix}"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 1
        width  = 8
        height = 6
        properties = {
          title  = "EKS CPU Utilization"
          region = data.aws_region.current.name
          metrics = [
            ["ContainerInsights", "node_cpu_utilization", "ClusterName", var.eks_cluster_name]
          ]
          period = 300
          stat   = "Average"
          view   = "timeSeries"
        }
      },
      {
        type   = "metric"
        x      = 8
        y      = 1
        width  = 8
        height = 6
        properties = {
          title  = "EKS Memory Utilization"
          region = data.aws_region.current.name
          metrics = [
            ["ContainerInsights", "node_memory_utilization", "ClusterName", var.eks_cluster_name]
          ]
          period = 300
          stat   = "Average"
          view   = "timeSeries"
        }
      },
      {
        type   = "metric"
        x      = 16
        y      = 1
        width  = 8
        height = 6
        properties = {
          title  = "EKS Pod Count"
          region = data.aws_region.current.name
          metrics = [
            ["ContainerInsights", "pod_number_of_running_containers", "ClusterName", var.eks_cluster_name]
          ]
          period = 300
          stat   = "Sum"
          view   = "timeSeries"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 7
        width  = 12
        height = 6
        properties = {
          title  = "RDS CPU Utilization"
          region = data.aws_region.current.name
          metrics = [
            ["AWS/RDS", "CPUUtilization", "DBInstanceIdentifier", var.rds_instance_id]
          ]
          period = 300
          stat   = "Average"
          view   = "timeSeries"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 7
        width  = 12
        height = 6
        properties = {
          title  = "RDS Database Connections"
          region = data.aws_region.current.name
          metrics = [
            ["AWS/RDS", "DatabaseConnections", "DBInstanceIdentifier", var.rds_instance_id]
          ]
          period = 300
          stat   = "Sum"
          view   = "timeSeries"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 13
        width  = 12
        height = 6
        properties = {
          title  = "ElastiCache CPU"
          region = data.aws_region.current.name
          metrics = [
            ["AWS/ElastiCache", "CPUUtilization", "CacheClusterId", var.elasticache_cluster_id]
          ]
          period = 300
          stat   = "Average"
          view   = "timeSeries"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 13
        width  = 12
        height = 6
        properties = {
          title  = "ElastiCache Memory"
          region = data.aws_region.current.name
          metrics = [
            ["AWS/ElastiCache", "DatabaseMemoryUsagePercentage", "CacheClusterId", var.elasticache_cluster_id]
          ]
          period = 300
          stat   = "Average"
          view   = "timeSeries"
        }
      }
    ]
  })
}

# -----------------------------------------------------------------------------
# Metric Filters for Application Logs
# -----------------------------------------------------------------------------
resource "aws_cloudwatch_log_metric_filter" "error_count" {
  name           = "${var.name_prefix}-error-count"
  pattern        = "ERROR"
  log_group_name = aws_cloudwatch_log_group.application.name

  metric_transformation {
    name          = "ErrorCount"
    namespace     = "NeoBank/${var.name_prefix}"
    value         = "1"
    default_value = "0"
  }
}

resource "aws_cloudwatch_log_metric_filter" "auth_failures" {
  name           = "${var.name_prefix}-auth-failures"
  pattern        = "[timestamp, level=ERROR, ..., msg=\"*authentication*failed*\"]"
  log_group_name = aws_cloudwatch_log_group.security.name

  metric_transformation {
    name          = "AuthFailures"
    namespace     = "NeoBank/${var.name_prefix}"
    value         = "1"
    default_value = "0"
  }
}

resource "aws_cloudwatch_metric_alarm" "error_rate_high" {
  alarm_name          = "${var.name_prefix}-error-rate-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "ErrorCount"
  namespace           = "NeoBank/${var.name_prefix}"
  period              = 300
  statistic           = "Sum"
  threshold           = 50
  alarm_description   = "Application error rate is high"
  alarm_actions       = [aws_sns_topic.alarms.arn]
  ok_actions          = [aws_sns_topic.alarms.arn]

  tags = var.tags
}

# -----------------------------------------------------------------------------
# Outputs
# -----------------------------------------------------------------------------
output "sns_topic_arn" {
  description = "SNS topic ARN for alarms"
  value       = aws_sns_topic.alarms.arn
}

output "application_log_group_name" {
  description = "Application log group name"
  value       = aws_cloudwatch_log_group.application.name
}

output "audit_log_group_name" {
  description = "Audit log group name"
  value       = aws_cloudwatch_log_group.audit.name
}

output "security_log_group_name" {
  description = "Security log group name"
  value       = aws_cloudwatch_log_group.security.name
}

output "dashboard_name" {
  description = "CloudWatch dashboard name"
  value       = var.enable_dashboard ? aws_cloudwatch_dashboard.main[0].dashboard_name : ""
}
