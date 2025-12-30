# ============================================================================
# Security Hub Module
# Centralized security findings aggregation
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

variable "delegated_admin_account_id" {
  description = "The AWS Account ID for the delegated administrator"
  type        = string
}

variable "enable_cis_standard" {
  description = "Enable CIS AWS Foundations Benchmark"
  type        = bool
  default     = true
}

variable "enable_pci_dss_standard" {
  description = "Enable PCI DSS v3.2.1 standard"
  type        = bool
  default     = true
}

variable "enable_aws_foundational_standard" {
  description = "Enable AWS Foundational Security Best Practices"
  type        = bool
  default     = true
}

variable "enable_nist_standard" {
  description = "Enable NIST SP 800-53 Rev. 5"
  type        = bool
  default     = true
}

variable "auto_enable_new_accounts" {
  description = "Auto-enable Security Hub for new member accounts"
  type        = bool
  default     = true
}

variable "aggregator_regions" {
  description = "Regions to aggregate findings from"
  type        = list(string)
  default     = ["us-east-1", "us-west-2", "eu-west-1"]
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
# Security Hub
# ============================================================================

resource "aws_securityhub_account" "main" {
  enable_default_standards  = false
  auto_enable_controls      = true
  control_finding_generator = "SECURITY_CONTROL"
}

# ============================================================================
# Security Hub Organization Admin
# ============================================================================

resource "aws_securityhub_organization_admin_account" "main" {
  admin_account_id = var.delegated_admin_account_id

  depends_on = [aws_securityhub_account.main]
}

# ============================================================================
# Security Hub Organization Configuration
# ============================================================================

resource "aws_securityhub_organization_configuration" "main" {
  auto_enable           = var.auto_enable_new_accounts
  auto_enable_standards = "DEFAULT"

  organization_configuration {
    configuration_type = "LOCAL"
  }

  depends_on = [
    aws_securityhub_organization_admin_account.main,
    aws_securityhub_finding_aggregator.main
  ]
}

# ============================================================================
# Security Hub Finding Aggregator (Cross-Region)
# ============================================================================

resource "aws_securityhub_finding_aggregator" "main" {
  linking_mode      = "SPECIFIED_REGIONS"
  specified_regions = var.aggregator_regions

  depends_on = [aws_securityhub_account.main]
}

# ============================================================================
# Security Standards
# ============================================================================

# CIS AWS Foundations Benchmark
resource "aws_securityhub_standards_subscription" "cis" {
  count         = var.enable_cis_standard ? 1 : 0
  standards_arn = "arn:aws:securityhub:${data.aws_region.current.name}::standards/cis-aws-foundations-benchmark/v/1.4.0"

  depends_on = [aws_securityhub_account.main]
}

# PCI DSS v3.2.1
resource "aws_securityhub_standards_subscription" "pci_dss" {
  count         = var.enable_pci_dss_standard ? 1 : 0
  standards_arn = "arn:aws:securityhub:${data.aws_region.current.name}::standards/pci-dss/v/3.2.1"

  depends_on = [aws_securityhub_account.main]
}

# AWS Foundational Security Best Practices
resource "aws_securityhub_standards_subscription" "aws_foundational" {
  count         = var.enable_aws_foundational_standard ? 1 : 0
  standards_arn = "arn:aws:securityhub:${data.aws_region.current.name}::standards/aws-foundational-security-best-practices/v/1.0.0"

  depends_on = [aws_securityhub_account.main]
}

# NIST SP 800-53 Rev. 5
resource "aws_securityhub_standards_subscription" "nist" {
  count         = var.enable_nist_standard ? 1 : 0
  standards_arn = "arn:aws:securityhub:${data.aws_region.current.name}::standards/nist-800-53/v/5.0.0"

  depends_on = [aws_securityhub_account.main]
}

# ============================================================================
# GuardDuty Organization Admin
# ============================================================================

resource "aws_guardduty_detector" "main" {
  enable                       = true
  finding_publishing_frequency = "FIFTEEN_MINUTES"

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

  tags = merge(var.tags, {
    Name = "${var.organization_name}-guardduty"
  })
}

resource "aws_guardduty_organization_admin_account" "main" {
  admin_account_id = var.delegated_admin_account_id

  depends_on = [aws_guardduty_detector.main]
}

resource "aws_guardduty_organization_configuration" "main" {
  auto_enable_organization_members = "ALL"
  detector_id                      = aws_guardduty_detector.main.id

  datasources {
    s3_logs {
      auto_enable = true
    }
    kubernetes {
      audit_logs {
        enable = true
      }
    }
    malware_protection {
      scan_ec2_instance_with_findings {
        ebs_volumes {
          auto_enable = true
        }
      }
    }
  }

  depends_on = [aws_guardduty_organization_admin_account.main]
}

# ============================================================================
# AWS IAM Access Analyzer (Organization)
# ============================================================================

resource "aws_accessanalyzer_analyzer" "organization" {
  analyzer_name = "${var.organization_name}-org-analyzer"
  type          = "ORGANIZATION"

  tags = merge(var.tags, {
    Name = "${var.organization_name}-access-analyzer"
  })
}

# ============================================================================
# Amazon Inspector
# ============================================================================

resource "aws_inspector2_enabler" "main" {
  account_ids    = [data.aws_caller_identity.current.account_id]
  resource_types = ["EC2", "ECR", "LAMBDA", "LAMBDA_CODE"]
}

resource "aws_inspector2_delegated_admin_account" "main" {
  account_id = var.delegated_admin_account_id

  depends_on = [aws_inspector2_enabler.main]
}

resource "aws_inspector2_organization_configuration" "main" {
  auto_enable {
    ec2         = true
    ecr         = true
    lambda      = true
    lambda_code = true
  }

  depends_on = [aws_inspector2_delegated_admin_account.main]
}

# ============================================================================
# Amazon Macie (Optional - for S3 data discovery)
# ============================================================================

resource "aws_macie2_account" "main" {
  finding_publishing_frequency = "FIFTEEN_MINUTES"
  status                       = "ENABLED"
}

# Macie admin account temporarily commented out due to state sync issues
# resource "aws_macie2_organization_admin_account" "main" {
#   admin_account_id = var.delegated_admin_account_id
# 
#   lifecycle {
#     ignore_changes = all
#   }
# 
#   depends_on = [aws_macie2_account.main]
# }

# ============================================================================
# SNS Topic for Security Alerts
# ============================================================================

resource "aws_sns_topic" "security_alerts" {
  name              = "${var.organization_name}-security-alerts"
  kms_master_key_id = "alias/aws/sns"

  tags = merge(var.tags, {
    Name = "${var.organization_name}-security-alerts"
  })
}

resource "aws_sns_topic_policy" "security_alerts" {
  arn = aws_sns_topic.security_alerts.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowSecurityHubPublish"
        Effect = "Allow"
        Principal = {
          Service = "securityhub.amazonaws.com"
        }
        Action   = "sns:Publish"
        Resource = aws_sns_topic.security_alerts.arn
      },
      {
        Sid    = "AllowGuardDutyPublish"
        Effect = "Allow"
        Principal = {
          Service = "guardduty.amazonaws.com"
        }
        Action   = "sns:Publish"
        Resource = aws_sns_topic.security_alerts.arn
      }
    ]
  })
}

# ============================================================================
# EventBridge Rule for Critical Findings
# ============================================================================

resource "aws_cloudwatch_event_rule" "critical_findings" {
  name        = "${var.organization_name}-critical-findings"
  description = "Capture critical security findings from Security Hub"

  event_pattern = jsonencode({
    source      = ["aws.securityhub"]
    detail_type = ["Security Hub Findings - Imported"]
    detail = {
      findings = {
        Severity = {
          Label = ["CRITICAL", "HIGH"]
        }
        Workflow = {
          Status = ["NEW"]
        }
      }
    }
  })

  tags = var.tags
}

resource "aws_cloudwatch_event_target" "critical_findings_sns" {
  rule      = aws_cloudwatch_event_rule.critical_findings.name
  target_id = "send-to-sns"
  arn       = aws_sns_topic.security_alerts.arn
}

resource "aws_cloudwatch_event_rule" "guardduty_findings" {
  name        = "${var.organization_name}-guardduty-findings"
  description = "Capture GuardDuty findings"

  event_pattern = jsonencode({
    source      = ["aws.guardduty"]
    detail_type = ["GuardDuty Finding"]
    detail = {
      severity = [
        { "numeric" = [">=", 7] }
      ]
    }
  })

  tags = var.tags
}

resource "aws_cloudwatch_event_target" "guardduty_findings_sns" {
  rule      = aws_cloudwatch_event_rule.guardduty_findings.name
  target_id = "send-to-sns"
  arn       = aws_sns_topic.security_alerts.arn
}

# ============================================================================
# Outputs
# ============================================================================

output "security_hub_arn" {
  description = "ARN of the Security Hub account"
  value       = aws_securityhub_account.main.id
}

output "guardduty_detector_id" {
  description = "ID of the GuardDuty detector"
  value       = aws_guardduty_detector.main.id
}

output "access_analyzer_arn" {
  description = "ARN of the IAM Access Analyzer"
  value       = aws_accessanalyzer_analyzer.organization.arn
}

output "security_alerts_topic_arn" {
  description = "ARN of the security alerts SNS topic"
  value       = aws_sns_topic.security_alerts.arn
}

output "enabled_standards" {
  description = "List of enabled security standards"
  value = {
    cis_benchmark    = var.enable_cis_standard
    pci_dss          = var.enable_pci_dss_standard
    aws_foundational = var.enable_aws_foundational_standard
    nist_800_53      = var.enable_nist_standard
  }
}
