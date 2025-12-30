# ============================================================================
# Centralized Logging Module
# Bank-grade audit trail and log aggregation
# ============================================================================

terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# ============================================================================
# Variables
# ============================================================================

variable "organization_name" {
  description = "Name of the organization"
  type        = string
  default     = "neobank"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "prod"
}

variable "organization_id" {
  description = "The AWS Organization ID"
  type        = string
}

variable "log_archive_account_id" {
  description = "The AWS Account ID for the Log Archive account"
  type        = string
}

variable "cloudtrail_retention_days" {
  description = "Number of days to retain CloudTrail logs"
  type        = number
  default     = 2555 # 7 years for compliance
}

variable "config_retention_days" {
  description = "Number of days to retain AWS Config snapshots"
  type        = number
  default     = 2555 # 7 years for compliance
}

variable "vpc_flow_log_retention_days" {
  description = "Number of days to retain VPC Flow Logs"
  type        = number
  default     = 365
}

variable "enable_s3_object_lock" {
  description = "Enable S3 Object Lock for immutable logs"
  type        = bool
  default     = true
}

variable "object_lock_retention_years" {
  description = "Number of years to lock objects (GOVERNANCE mode)"
  type        = number
  default     = 7
}

variable "enable_access_logging" {
  description = "Enable access logging for log buckets"
  type        = bool
  default     = true
}

variable "kms_key_deletion_window" {
  description = "KMS key deletion window in days"
  type        = number
  default     = 30
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}

# ============================================================================
# Data Sources
# ============================================================================

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# ============================================================================
# KMS Key for Log Encryption
# ============================================================================

resource "aws_kms_key" "log_encryption" {
  description             = "KMS key for centralized log encryption"
  deletion_window_in_days = var.kms_key_deletion_window
  enable_key_rotation     = true
  multi_region            = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow CloudTrail to encrypt logs"
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action = [
          "kms:GenerateDataKey*",
          "kms:Decrypt"
        ]
        Resource = "*"
        Condition = {
          StringEquals = {
            "aws:SourceOrgID" = var.organization_id
          }
        }
      },
      {
        Sid    = "Allow Config to encrypt data"
        Effect = "Allow"
        Principal = {
          Service = "config.amazonaws.com"
        }
        Action = [
          "kms:GenerateDataKey*",
          "kms:Decrypt"
        ]
        Resource = "*"
      },
      {
        Sid    = "Allow VPC Flow Logs to encrypt"
        Effect = "Allow"
        Principal = {
          Service = "logs.amazonaws.com"
        }
        Action = [
          "kms:Encrypt*",
          "kms:Decrypt*",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:Describe*"
        ]
        Resource = "*"
        Condition = {
          ArnLike = {
            "kms:EncryptionContext:aws:logs:arn" = "arn:aws:logs:*:*:log-group:/aws/vpc-flow-logs/*"
          }
        }
      },
      {
        Sid    = "Allow S3 to use key for encryption"
        Effect = "Allow"
        Principal = {
          Service = "s3.amazonaws.com"
        }
        Action = [
          "kms:GenerateDataKey*",
          "kms:Decrypt"
        ]
        Resource = "*"
      },
      {
        Sid    = "Allow Log Archive Account Access"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${var.log_archive_account_id}:root"
        }
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:DescribeKey"
        ]
        Resource = "*"
      }
    ]
  })

  tags = merge(var.tags, {
    Name    = "${var.organization_name}-log-encryption-key"
    Purpose = "CentralizedLogging"
  })
}

resource "aws_kms_alias" "log_encryption" {
  name          = "alias/${var.organization_name}-log-encryption"
  target_key_id = aws_kms_key.log_encryption.key_id
}

# ============================================================================
# S3 Bucket for Access Logging
# ============================================================================

resource "aws_s3_bucket" "access_logs" {
  count  = var.enable_access_logging ? 1 : 0
  bucket = "${var.organization_name}-access-logs-${data.aws_caller_identity.current.account_id}"

  tags = merge(var.tags, {
    Name    = "${var.organization_name}-access-logs"
    Purpose = "S3AccessLogging"
  })
}

resource "aws_s3_bucket_versioning" "access_logs" {
  count  = var.enable_access_logging ? 1 : 0
  bucket = aws_s3_bucket.access_logs[0].id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "access_logs" {
  count  = var.enable_access_logging ? 1 : 0
  bucket = aws_s3_bucket.access_logs[0].id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.log_encryption.arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "access_logs" {
  count  = var.enable_access_logging ? 1 : 0
  bucket = aws_s3_bucket.access_logs[0].id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ============================================================================
# CloudTrail S3 Bucket
# ============================================================================

resource "aws_s3_bucket" "cloudtrail" {
  bucket = "${var.organization_name}-cloudtrail-${data.aws_caller_identity.current.account_id}"

  # Object Lock must be enabled at bucket creation time
  object_lock_enabled = var.enable_s3_object_lock

  tags = merge(var.tags, {
    Name    = "${var.organization_name}-cloudtrail"
    Purpose = "CloudTrailLogs"
  })
}

resource "aws_s3_bucket_versioning" "cloudtrail" {
  bucket = aws_s3_bucket.cloudtrail.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_object_lock_configuration" "cloudtrail" {
  count  = var.enable_s3_object_lock ? 1 : 0
  bucket = aws_s3_bucket.cloudtrail.id

  rule {
    default_retention {
      mode  = "GOVERNANCE"
      years = var.object_lock_retention_years
    }
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "cloudtrail" {
  bucket = aws_s3_bucket.cloudtrail.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.log_encryption.arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "cloudtrail" {
  bucket = aws_s3_bucket.cloudtrail.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_logging" "cloudtrail" {
  count         = var.enable_access_logging ? 1 : 0
  bucket        = aws_s3_bucket.cloudtrail.id
  target_bucket = aws_s3_bucket.access_logs[0].id
  target_prefix = "cloudtrail/"
}

resource "aws_s3_bucket_lifecycle_configuration" "cloudtrail" {
  bucket = aws_s3_bucket.cloudtrail.id

  rule {
    id     = "transition-to-glacier"
    status = "Enabled"

    filter {}

    transition {
      days          = 90
      storage_class = "GLACIER"
    }

    transition {
      days          = 365
      storage_class = "DEEP_ARCHIVE"
    }

    expiration {
      days = var.cloudtrail_retention_days
    }
  }
}

resource "aws_s3_bucket_policy" "cloudtrail" {
  bucket = aws_s3_bucket.cloudtrail.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AWSCloudTrailAclCheck"
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action   = "s3:GetBucketAcl"
        Resource = aws_s3_bucket.cloudtrail.arn
        Condition = {
          StringEquals = {
            "aws:SourceOrgID" = var.organization_id
          }
        }
      },
      {
        Sid    = "AWSCloudTrailWrite"
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action   = "s3:PutObject"
        Resource = "${aws_s3_bucket.cloudtrail.arn}/*"
        Condition = {
          StringEquals = {
            "s3:x-amz-acl"    = "bucket-owner-full-control"
            "aws:SourceOrgID" = var.organization_id
          }
        }
      },
      {
        Sid       = "AllowSSLRequestsOnly"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:*"
        Resource = [
          aws_s3_bucket.cloudtrail.arn,
          "${aws_s3_bucket.cloudtrail.arn}/*"
        ]
        Condition = {
          Bool = {
            "aws:SecureTransport" = "false"
          }
        }
      }
    ]
  })
}

# ============================================================================
# Organization CloudTrail
# ============================================================================

resource "aws_cloudtrail" "organization" {
  name                          = "${var.organization_name}-org-trail"
  s3_bucket_name                = aws_s3_bucket.cloudtrail.id
  s3_key_prefix                 = "cloudtrail"
  include_global_service_events = true
  is_multi_region_trail         = true
  is_organization_trail         = true
  enable_log_file_validation    = true
  kms_key_id                    = aws_kms_key.log_encryption.arn

  event_selector {
    read_write_type           = "All"
    include_management_events = true

    data_resource {
      type   = "AWS::S3::Object"
      values = ["arn:aws:s3"]
    }

    data_resource {
      type   = "AWS::Lambda::Function"
      values = ["arn:aws:lambda"]
    }
  }

  event_selector {
    read_write_type           = "All"
    include_management_events = true

    data_resource {
      type   = "AWS::DynamoDB::Table"
      values = ["arn:aws:dynamodb"]
    }
  }

  insight_selector {
    insight_type = "ApiCallRateInsight"
  }

  insight_selector {
    insight_type = "ApiErrorRateInsight"
  }

  tags = merge(var.tags, {
    Name    = "${var.organization_name}-org-trail"
    Purpose = "OrganizationCloudTrail"
  })

  depends_on = [aws_s3_bucket_policy.cloudtrail]
}

# ============================================================================
# AWS Config S3 Bucket
# ============================================================================

resource "aws_s3_bucket" "config" {
  bucket = "${var.organization_name}-config-${data.aws_caller_identity.current.account_id}"

  tags = merge(var.tags, {
    Name    = "${var.organization_name}-config"
    Purpose = "AWSConfigSnapshots"
  })
}

resource "aws_s3_bucket_versioning" "config" {
  bucket = aws_s3_bucket.config.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "config" {
  bucket = aws_s3_bucket.config.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.log_encryption.arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "config" {
  bucket = aws_s3_bucket.config.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_logging" "config" {
  count         = var.enable_access_logging ? 1 : 0
  bucket        = aws_s3_bucket.config.id
  target_bucket = aws_s3_bucket.access_logs[0].id
  target_prefix = "config/"
}

resource "aws_s3_bucket_lifecycle_configuration" "config" {
  bucket = aws_s3_bucket.config.id

  rule {
    id     = "transition-to-glacier"
    status = "Enabled"

    filter {}

    transition {
      days          = 90
      storage_class = "GLACIER"
    }

    expiration {
      days = var.config_retention_days
    }
  }
}

resource "aws_s3_bucket_policy" "config" {
  bucket = aws_s3_bucket.config.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AWSConfigBucketPermissionsCheck"
        Effect = "Allow"
        Principal = {
          Service = "config.amazonaws.com"
        }
        Action   = "s3:GetBucketAcl"
        Resource = aws_s3_bucket.config.arn
      },
      {
        Sid    = "AWSConfigBucketExistenceCheck"
        Effect = "Allow"
        Principal = {
          Service = "config.amazonaws.com"
        }
        Action   = "s3:ListBucket"
        Resource = aws_s3_bucket.config.arn
      },
      {
        Sid    = "AWSConfigBucketDelivery"
        Effect = "Allow"
        Principal = {
          Service = "config.amazonaws.com"
        }
        Action   = "s3:PutObject"
        Resource = "${aws_s3_bucket.config.arn}/*"
        Condition = {
          StringEquals = {
            "s3:x-amz-acl" = "bucket-owner-full-control"
          }
        }
      },
      {
        Sid       = "AllowSSLRequestsOnly"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:*"
        Resource = [
          aws_s3_bucket.config.arn,
          "${aws_s3_bucket.config.arn}/*"
        ]
        Condition = {
          Bool = {
            "aws:SecureTransport" = "false"
          }
        }
      }
    ]
  })
}

# ============================================================================
# AWS Config Aggregator
# ============================================================================

resource "aws_config_configuration_aggregator" "organization" {
  name = "${var.organization_name}-config-aggregator"

  organization_aggregation_source {
    all_regions = true
    role_arn    = aws_iam_role.config_aggregator.arn
  }

  tags = merge(var.tags, {
    Name = "${var.organization_name}-config-aggregator"
  })
}

resource "aws_iam_role" "config_aggregator" {
  name = "${var.organization_name}-config-aggregator-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "config.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "config_aggregator" {
  role       = aws_iam_role.config_aggregator.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSConfigRoleForOrganizations"
}

# ============================================================================
# VPC Flow Logs S3 Bucket
# ============================================================================

resource "aws_s3_bucket" "vpc_flow_logs" {
  bucket = "${var.organization_name}-vpc-flow-logs-${data.aws_caller_identity.current.account_id}"

  tags = merge(var.tags, {
    Name    = "${var.organization_name}-vpc-flow-logs"
    Purpose = "VPCFlowLogs"
  })
}

resource "aws_s3_bucket_versioning" "vpc_flow_logs" {
  bucket = aws_s3_bucket.vpc_flow_logs.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "vpc_flow_logs" {
  bucket = aws_s3_bucket.vpc_flow_logs.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.log_encryption.arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "vpc_flow_logs" {
  bucket = aws_s3_bucket.vpc_flow_logs.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_logging" "vpc_flow_logs" {
  count         = var.enable_access_logging ? 1 : 0
  bucket        = aws_s3_bucket.vpc_flow_logs.id
  target_bucket = aws_s3_bucket.access_logs[0].id
  target_prefix = "vpc-flow-logs/"
}

resource "aws_s3_bucket_lifecycle_configuration" "vpc_flow_logs" {
  bucket = aws_s3_bucket.vpc_flow_logs.id

  rule {
    id     = "transition-to-glacier"
    status = "Enabled"

    filter {}

    transition {
      days          = 90
      storage_class = "GLACIER"
    }

    expiration {
      days = var.vpc_flow_log_retention_days
    }
  }
}

resource "aws_s3_bucket_policy" "vpc_flow_logs" {
  bucket = aws_s3_bucket.vpc_flow_logs.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AWSLogDeliveryAclCheck"
        Effect = "Allow"
        Principal = {
          Service = "delivery.logs.amazonaws.com"
        }
        Action   = "s3:GetBucketAcl"
        Resource = aws_s3_bucket.vpc_flow_logs.arn
      },
      {
        Sid    = "AWSLogDeliveryWrite"
        Effect = "Allow"
        Principal = {
          Service = "delivery.logs.amazonaws.com"
        }
        Action   = "s3:PutObject"
        Resource = "${aws_s3_bucket.vpc_flow_logs.arn}/*"
        Condition = {
          StringEquals = {
            "s3:x-amz-acl" = "bucket-owner-full-control"
          }
        }
      },
      {
        Sid       = "AllowSSLRequestsOnly"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:*"
        Resource = [
          aws_s3_bucket.vpc_flow_logs.arn,
          "${aws_s3_bucket.vpc_flow_logs.arn}/*"
        ]
        Condition = {
          Bool = {
            "aws:SecureTransport" = "false"
          }
        }
      }
    ]
  })
}

# ============================================================================
# Outputs
# ============================================================================

output "kms_key_arn" {
  description = "ARN of the log encryption KMS key"
  value       = aws_kms_key.log_encryption.arn
}

output "kms_key_id" {
  description = "ID of the log encryption KMS key"
  value       = aws_kms_key.log_encryption.key_id
}

output "cloudtrail_bucket_name" {
  description = "Name of the CloudTrail S3 bucket"
  value       = aws_s3_bucket.cloudtrail.id
}

output "cloudtrail_bucket_arn" {
  description = "ARN of the CloudTrail S3 bucket"
  value       = aws_s3_bucket.cloudtrail.arn
}

output "cloudtrail_arn" {
  description = "ARN of the organization CloudTrail"
  value       = aws_cloudtrail.organization.arn
}

output "config_bucket_name" {
  description = "Name of the Config S3 bucket"
  value       = aws_s3_bucket.config.id
}

output "config_bucket_arn" {
  description = "ARN of the Config S3 bucket"
  value       = aws_s3_bucket.config.arn
}

output "config_aggregator_arn" {
  description = "ARN of the Config aggregator"
  value       = aws_config_configuration_aggregator.organization.arn
}

output "vpc_flow_logs_bucket_name" {
  description = "Name of the VPC Flow Logs S3 bucket"
  value       = aws_s3_bucket.vpc_flow_logs.id
}

output "vpc_flow_logs_bucket_arn" {
  description = "ARN of the VPC Flow Logs S3 bucket"
  value       = aws_s3_bucket.vpc_flow_logs.arn
}

output "access_logs_bucket_name" {
  description = "Name of the access logs S3 bucket"
  value       = var.enable_access_logging ? aws_s3_bucket.access_logs[0].id : null
}
