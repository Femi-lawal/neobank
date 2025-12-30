# ============================================================================
# AWS Landing Zone - Main Configuration
# Bank-grade multi-account governance infrastructure
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
  backend "s3" {
    bucket         = "neobank-terraform-state-${var.environment}"
    key            = "landing-zone/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
  }
}

# ============================================================================
# Provider Configuration
# ============================================================================

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "Terraform"
      Repository  = "neobank"
    }
  }
}

# ============================================================================
# Variables
# ============================================================================

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "neobank"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "prod"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "allowed_regions" {
  description = "List of allowed AWS regions"
  type        = list(string)
  default     = ["us-east-1", "us-west-2", "eu-west-1"]
}

variable "create_organization" {
  description = "Whether to create a new AWS Organization (set to false if one already exists)"
  type        = bool
  default     = false
}

variable "create_accounts" {
  description = "Whether to create new AWS accounts"
  type        = bool
  default     = false
}

variable "account_email_domain" {
  description = "Domain for account email addresses"
  type        = string
  default     = "example.com"
}

variable "log_archive_account_id" {
  description = "AWS Account ID for Log Archive (if not creating new accounts)"
  type        = string
  default     = ""
}

variable "audit_account_id" {
  description = "AWS Account ID for Audit/Security (if not creating new accounts)"
  type        = string
  default     = ""
}

variable "enable_s3_object_lock" {
  description = "Enable S3 Object Lock for immutable logs"
  type        = bool
  default     = false # Requires bucket to be created with object lock enabled
}

variable "admin_role_arn" {
  description = "ARN of the admin role for Service Catalog"
  type        = string
  default     = ""
}

variable "developer_role_arn" {
  description = "ARN of the developer role for Service Catalog"
  type        = string
  default     = ""
}

# ============================================================================
# Data Sources
# ============================================================================

data "aws_caller_identity" "current" {}

data "aws_organizations_organization" "current" {
  count = var.create_organization ? 0 : 1
}

# ============================================================================
# Module: AWS Organizations
# ============================================================================

module "organizations" {
  source = "./modules/organizations"
  count  = var.create_organization ? 1 : 0

  organization_name    = var.project_name
  create_accounts      = var.create_accounts
  account_email_domain = var.account_email_domain

  tags = {
    Module = "Organizations"
  }
}

# ============================================================================
# Module: Service Control Policies (SCPs)
# ============================================================================

module "scps" {
  source = "./modules/scps"
  count  = var.create_organization ? 1 : 0

  organization_name = var.project_name
  allowed_regions   = var.allowed_regions

  root_ou_id     = module.organizations[0].root_ou_id
  security_ou_id = module.organizations[0].security_ou_id
  prod_ou_id     = module.organizations[0].prod_ou_id
  nonprod_ou_id  = module.organizations[0].nonprod_ou_id
  sandbox_ou_id  = module.organizations[0].sandbox_ou_id
  pci_cde_ou_id  = module.organizations[0].pci_cde_ou_id

  tags = {
    Module = "SCPs"
  }

  depends_on = [module.organizations]
}

# ============================================================================
# Module: Centralized Logging
# ============================================================================

module "logging" {
  source = "./modules/logging"

  organization_name      = var.project_name
  environment            = var.environment
  organization_id        = var.create_organization ? module.organizations[0].organization_id : data.aws_organizations_organization.current[0].id
  log_archive_account_id = var.log_archive_account_id != "" ? var.log_archive_account_id : data.aws_caller_identity.current.account_id
  enable_s3_object_lock  = var.enable_s3_object_lock

  cloudtrail_retention_days   = 2555 # 7 years
  config_retention_days       = 2555 # 7 years
  vpc_flow_log_retention_days = 365

  tags = {
    Module = "Logging"
  }
}

# ============================================================================
# Module: Security Hub & GuardDuty
# ============================================================================

module "security_hub" {
  source = "./modules/security-hub"

  organization_name          = var.project_name
  environment                = var.environment
  delegated_admin_account_id = var.audit_account_id != "" ? var.audit_account_id : data.aws_caller_identity.current.account_id

  enable_cis_standard              = true
  enable_pci_dss_standard          = true
  enable_aws_foundational_standard = true
  enable_nist_standard             = true
  auto_enable_new_accounts         = true
  aggregator_regions               = ["us-west-2"] # Exclude current region (us-east-1)

  tags = {
    Module = "SecurityHub"
  }
}

# ============================================================================
# Module: Service Catalog
# ============================================================================
# Disabled temporarily due to provider resource type issues
# module "service_catalog" {
#   source = "./modules/service-catalog"
#   count  = var.admin_role_arn != "" && var.developer_role_arn != "" ? 1 : 0
# 
#   organization_name  = var.project_name
#   environment        = var.environment
#   admin_role_arn     = var.admin_role_arn
#   developer_role_arn = var.developer_role_arn
# 
#   tags = {
#     Module = "ServiceCatalog"
#   }
# }

# ============================================================================
# Outputs
# ============================================================================

output "organization_id" {
  description = "AWS Organization ID"
  value       = var.create_organization ? module.organizations[0].organization_id : data.aws_organizations_organization.current[0].id
}

output "logging_kms_key_arn" {
  description = "ARN of the logging KMS key"
  value       = module.logging.kms_key_arn
}

output "cloudtrail_bucket" {
  description = "CloudTrail S3 bucket name"
  value       = module.logging.cloudtrail_bucket_name
}

output "config_bucket" {
  description = "AWS Config S3 bucket name"
  value       = module.logging.config_bucket_name
}

output "vpc_flow_logs_bucket" {
  description = "VPC Flow Logs S3 bucket name"
  value       = module.logging.vpc_flow_logs_bucket_name
}

output "security_alerts_topic_arn" {
  description = "ARN of the security alerts SNS topic"
  value       = module.security_hub.security_alerts_topic_arn
}

output "guardduty_detector_id" {
  description = "GuardDuty detector ID"
  value       = module.security_hub.guardduty_detector_id
}

output "service_catalog_portfolios" {
  description = "Service Catalog portfolio IDs"
  value       = null
  # value = var.admin_role_arn != "" && var.developer_role_arn != "" ? {
  #   platform   = module.service_catalog[0].platform_portfolio_id
  #   data       = module.service_catalog[0].data_portfolio_id
  #   kubernetes = module.service_catalog[0].kubernetes_portfolio_id
  #   messaging  = module.service_catalog[0].messaging_portfolio_id
  # } : null
}
