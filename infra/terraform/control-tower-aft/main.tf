# ============================================================================
# Account Factory for Terraform (AFT) Deployment
# This deploys AWS Control Tower's Account Factory for Terraform
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
    key            = "control-tower-aft/terraform.tfstate"
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
      Repository  = "neobank"
      Module      = "AFT"
      Environment = "prod"
    }
  }
}

# ============================================================================
# Data Sources
# ============================================================================

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# ============================================================================
# AFT Module
# Deploys Account Factory for Terraform infrastructure
# ============================================================================

module "aft" {
  source  = "aws-ia/control_tower_account_factory/aws"
  version = "~> 1.18.0" # Use latest stable version from Terraform Registry

  # CT Management Account Configuration
  ct_management_account_id = data.aws_caller_identity.current.account_id
  ct_home_region           = var.aws_region

  # Log Archive Account (will be created by Control Tower)
  log_archive_account_id = var.log_archive_account_id

  # Audit Account (will be created by Control Tower)
  audit_account_id = var.audit_account_id

  # AFT Management Account (can be same as CT management or separate)
  aft_management_account_id = var.aft_management_account_id != "" ? var.aft_management_account_id : data.aws_caller_identity.current.account_id

  # Control Tower Regions
  ct_secondary_regions = var.ct_secondary_regions

  # Terraform Configuration
  terraform_version      = var.terraform_version
  terraform_distribution = "oss" # or "tfc" for Terraform Cloud

  # VPC Configuration for AFT
  aft_vpc_cidr                   = var.aft_vpc_cidr
  aft_vpc_private_subnet_01_cidr = cidrsubnet(var.aft_vpc_cidr, 8, 1)
  aft_vpc_private_subnet_02_cidr = cidrsubnet(var.aft_vpc_cidr, 8, 2)
  aft_vpc_public_subnet_01_cidr  = cidrsubnet(var.aft_vpc_cidr, 8, 101)
  aft_vpc_public_subnet_02_cidr  = cidrsubnet(var.aft_vpc_cidr, 8, 102)

  # Git Repository Configuration
  # These repositories will contain account requests and customizations
  vcs_provider                                  = "codecommit" # or "github", "gitlab", "bitbucket"
  account_request_repo_name                     = var.account_request_repo_name
  global_customizations_repo_name               = var.global_customizations_repo_name
  account_customizations_repo_name              = var.account_customizations_repo_name
  account_provisioning_customizations_repo_name = var.account_provisioning_customizations_repo_name

  # Account Customization Configuration
  aft_feature_cloudtrail_data_events      = true
  aft_feature_enterprise_support          = var.enable_enterprise_support
  aft_feature_delete_default_vpcs_enabled = true

  # CloudWatch Log Retention
  cloudwatch_log_group_retention = 365 # 1 year

  # Maximum concurrent account customizations
  maximum_concurrent_customizations = 5

  # AFT Framework Configuration
  aft_enable_vpc = true

  # Tags
  tags = {
    Component = "AFT"
    Purpose   = "Account Factory"
  }
}

# ============================================================================
# CodeCommit Repositories (if using CodeCommit)
# ============================================================================

resource "aws_codecommit_repository" "account_requests" {
  count           = var.vcs_provider == "codecommit" ? 1 : 0
  repository_name = var.account_request_repo_name
  description     = "AFT Account Requests - Contains Terraform configs for account provisioning"

  tags = {
    Name    = var.account_request_repo_name
    Purpose = "AFT Account Requests"
  }
}

resource "aws_codecommit_repository" "global_customizations" {
  count           = var.vcs_provider == "codecommit" ? 1 : 0
  repository_name = var.global_customizations_repo_name
  description     = "AFT Global Customizations - Applied to all accounts"

  tags = {
    Name    = var.global_customizations_repo_name
    Purpose = "AFT Global Customizations"
  }
}

resource "aws_codecommit_repository" "account_customizations" {
  count           = var.vcs_provider == "codecommit" ? 1 : 0
  repository_name = var.account_customizations_repo_name
  description     = "AFT Account-Specific Customizations"

  tags = {
    Name    = var.account_customizations_repo_name
    Purpose = "AFT Account Customizations"
  }
}

resource "aws_codecommit_repository" "account_provisioning_customizations" {
  count           = var.vcs_provider == "codecommit" ? 1 : 0
  repository_name = var.account_provisioning_customizations_repo_name
  description     = "AFT Account Provisioning Customizations - Runs during account creation"

  tags = {
    Name    = var.account_provisioning_customizations_repo_name
    Purpose = "AFT Provisioning Customizations"
  }
}

# ============================================================================
# Outputs
# ============================================================================

output "aft_management_account_id" {
  description = "AFT Management Account ID"
  value       = module.aft.aft_management_account_id
}

output "aft_vpc_id" {
  description = "AFT VPC ID"
  value       = module.aft.aft_vpc_id
}

output "aft_step_function_arn" {
  description = "AFT Account Provisioning Step Function ARN"
  value       = module.aft.aft_account_provisioning_framework_sfn_arn
}

output "account_request_repo_url" {
  description = "Account Request Repository URL"
  value       = var.vcs_provider == "codecommit" ? aws_codecommit_repository.account_requests[0].clone_url_http : "Configure your external Git provider"
}

output "global_customizations_repo_url" {
  description = "Global Customizations Repository URL"
  value       = var.vcs_provider == "codecommit" ? aws_codecommit_repository.global_customizations[0].clone_url_http : "Configure your external Git provider"
}

output "account_customizations_repo_url" {
  description = "Account Customizations Repository URL"
  value       = var.vcs_provider == "codecommit" ? aws_codecommit_repository.account_customizations[0].clone_url_http : "Configure your external Git provider"
}
