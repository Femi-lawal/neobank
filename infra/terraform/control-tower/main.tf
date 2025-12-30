# ============================================================================
# AWS Control Tower Landing Zone - Terraform First Approach
# Based on AWS APIs-only walkthrough for Control Tower v4.0
# ============================================================================

terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # S3 backend for Terraform state management
  # NOTE: Replace TERRAFORM_STATE_BUCKET with your S3 bucket name
  backend "s3" {
    bucket         = "neobank-terraform-state-${AWS_ACCOUNT_ID}"
    key            = "control-tower/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      ManagedBy   = "Terraform"
      Project     = "neobank"
      Module      = "ControlTower"
      Environment = "prod"
    }
  }
}

# ============================================================================
# Data Sources
# ============================================================================

data "aws_caller_identity" "current" {}
data "aws_organizations_organization" "current" {}

# ============================================================================
# Local Variables - Landing Zone Manifest (v4.0 format)
# v4.0 removes organizationStructure from manifest
# ============================================================================

locals {
  # Sort governed regions to avoid diff noise
  governed_regions = sort(var.governed_regions)

  # Landing Zone Manifest for v4.0
  # Reference: https://docs.aws.amazon.com/controltower/latest/userguide/lz-api-launch.html
  manifest = {
    # Access Management (IAM Identity Center)
    accessManagement = {
      enabled = true
    }

    # Centralized Logging Configuration
    centralizedLogging = {
      accountId = var.log_archive_account_id
      enabled   = true
      configurations = {
        loggingBucket = {
          retentionDays = var.log_retention_days
        }
        accessLoggingBucket = {
          retentionDays = var.access_log_retention_days
        }
      }
    }



    # Security Roles Configuration (v4.0 requires enabled flag)
    securityRoles = {
      accountId = var.audit_account_id
      enabled   = true
    }

    # AWS Config Configuration (required when securityRoles is enabled)
    config = {
      accountId = var.audit_account_id
      enabled   = true
    }

    # Governed Regions
    governedRegions = local.governed_regions
  }
}

# ============================================================================
# Control Tower Landing Zone Resource
# ============================================================================

resource "aws_controltower_landing_zone" "main" {
  manifest_json = jsonencode(local.manifest)
  version       = var.control_tower_version

  timeouts {
    create = "120m"
    update = "120m"
    delete = "120m"
  }
}

# ============================================================================
# Outputs
# ============================================================================

output "landing_zone_id" {
  description = "Control Tower Landing Zone ID"
  value       = aws_controltower_landing_zone.main.id
}

output "landing_zone_arn" {
  description = "Control Tower Landing Zone ARN"
  value       = aws_controltower_landing_zone.main.arn
}

output "landing_zone_version" {
  description = "Control Tower Landing Zone version"
  value       = aws_controltower_landing_zone.main.version
}

output "landing_zone_drift_status" {
  description = "Control Tower Landing Zone drift status"
  value       = aws_controltower_landing_zone.main.drift_status
}

output "organization_id" {
  description = "AWS Organization ID"
  value       = data.aws_organizations_organization.current.id
}

output "management_account_id" {
  description = "Management Account ID"
  value       = data.aws_caller_identity.current.account_id
}
