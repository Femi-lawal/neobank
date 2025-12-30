# ============================================================================
# AFT Global Customizations
# Applied to ALL accounts provisioned through AFT
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
    # Don't modify these values
  }
}

provider "aws" {
  region = var.aws_region
}

# ============================================================================
# Data Sources
# ============================================================================

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# ============================================================================
# CloudWatch Log Encryption
# Enable encryption for all CloudWatch log groups
# ============================================================================

resource "aws_kms_key" "cloudwatch_logs" {
  description             = "KMS key for CloudWatch Logs encryption"
  deletion_window_in_days = 30
  enable_key_rotation     = true

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
        Sid    = "Allow CloudWatch Logs"
        Effect = "Allow"
        Principal = {
          Service = "logs.${data.aws_region.current.name}.amazonaws.com"
        }
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:CreateGrant",
          "kms:DescribeKey"
        ]
        Resource = "*"
        Condition = {
          ArnLike = {
            "kms:EncryptionContext:aws:logs:arn" = "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
          }
        }
      }
    ]
  })

  tags = {
    Name    = "cloudwatch-logs-encryption"
    Purpose = "CloudWatch Logs Encryption"
  }
}

resource "aws_kms_alias" "cloudwatch_logs" {
  name          = "alias/cloudwatch-logs"
  target_key_id = aws_kms_key.cloudwatch_logs.key_id
}

# ============================================================================
# AWS Config Baseline
# Enable AWS Config in all regions
# ============================================================================

module "aws_config" {
  source = "terraform-aws-modules/config/aws"

  create_sns_topic = true
  sns_topic_name   = "aws-config-notifications"

  s3_bucket_name               = "aws-config-${data.aws_caller_identity.current.account_id}"
  s3_bucket_versioning         = true
  s3_bucket_encryption_enabled = true

  # Recorder Configuration
  is_aggregator_account     = false
  is_organization_supported = true

  # Config Rules - Common compliance checks
  config_rules = {
    encrypted-volumes = {
      description       = "Checks whether EBS volumes are encrypted"
      source_owner      = "AWS"
      source_identifier = "ENCRYPTED_VOLUMES"
    }
    rds-encryption-enabled = {
      description       = "Checks whether RDS DB instances are encrypted"
      source_owner      = "AWS"
      source_identifier = "RDS_STORAGE_ENCRYPTED"
    }
    s3-bucket-public-read-prohibited = {
      description       = "Checks that S3 buckets do not allow public read access"
      source_owner      = "AWS"
      source_identifier = "S3_BUCKET_PUBLIC_READ_PROHIBITED"
    }
    s3-bucket-public-write-prohibited = {
      description       = "Checks that S3 buckets do not allow public write access"
      source_owner      = "AWS"
      source_identifier = "S3_BUCKET_PUBLIC_WRITE_PROHIBITED"
    }
  }

  tags = {
    Name    = "aws-config"
    Purpose = "Compliance Monitoring"
  }
}

# ============================================================================
# GuardDuty Baseline
# Enable GuardDuty for threat detection
# ============================================================================

resource "aws_guardduty_detector" "main" {
  enable = true

  datasources {
    s3_logs {
      enable = true
    }
    kubernetes {
      audit_logs {
        enable = true
      }
    }
    malware_protection {
      scan_ec2_instance_with_findings {
        ebs_volumes {
          enable = true
        }
      }
    }
  }

  finding_publishing_frequency = "FIFTEEN_MINUTES"

  tags = {
    Name    = "guardduty-detector"
    Purpose = "Threat Detection"
  }
}

# ============================================================================
# Security Hub Baseline
# Enable Security Hub with relevant standards
# ============================================================================

resource "aws_securityhub_account" "main" {}

resource "aws_securityhub_standards_subscription" "cis" {
  depends_on    = [aws_securityhub_account.main]
  standards_arn = "arn:aws:securityhub:${data.aws_region.current.name}::standards/cis-aws-foundations-benchmark/v/1.4.0"
}

resource "aws_securityhub_standards_subscription" "pci_dss" {
  depends_on    = [aws_securityhub_account.main]
  standards_arn = "arn:aws:securityhub:${data.aws_region.current.name}::standards/pci-dss/v/3.2.1"
}

resource "aws_securityhub_standards_subscription" "aws_foundational" {
  depends_on    = [aws_securityhub_account.main]
  standards_arn = "arn:aws:securityhub:${data.aws_region.current.name}::standards/aws-foundational-security-best-practices/v/1.0.0"
}

# ============================================================================
# S3 Block Public Access
# Enable account-level S3 public access block
# ============================================================================

resource "aws_s3_account_public_access_block" "main" {
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ============================================================================
# EBS Encryption by Default
# Enable default EBS volume encryption
# ============================================================================

resource "aws_ebs_encryption_by_default" "main" {
  enabled = true
}

# ============================================================================
# Outputs
# ============================================================================

output "cloudwatch_logs_kms_key_arn" {
  description = "KMS key ARN for CloudWatch Logs encryption"
  value       = aws_kms_key.cloudwatch_logs.arn
}

output "guardduty_detector_id" {
  description = "GuardDuty Detector ID"
  value       = aws_guardduty_detector.main.id
}

output "security_hub_arn" {
  description = "Security Hub ARN"
  value       = aws_securityhub_account.main.arn
}
