# ============================================================================
# AWS Organizations Module
# Creates the multi-account structure for NeoBank Landing Zone
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
  default     = "production"
}

variable "aws_service_access_principals" {
  description = "List of AWS service principal names for which you want to enable integration with your organization"
  type        = list(string)
  default = [
    "cloudtrail.amazonaws.com",
    "config.amazonaws.com",
    "config-multiaccountsetup.amazonaws.com",
    "guardduty.amazonaws.com",
    "securityhub.amazonaws.com",
    "sso.amazonaws.com",
    "tagpolicies.tag.amazonaws.com",
    "servicecatalog.amazonaws.com",
    "ram.amazonaws.com",
    "account.amazonaws.com",
    "member.org.stacksets.cloudformation.amazonaws.com"
  ]
}

variable "enabled_policy_types" {
  description = "List of Organizations policy types to enable"
  type        = list(string)
  default = [
    "SERVICE_CONTROL_POLICY",
    "TAG_POLICY",
    "BACKUP_POLICY"
  ]
}

variable "feature_set" {
  description = "Feature set of the organization (ALL or CONSOLIDATED_BILLING)"
  type        = string
  default     = "ALL"
}

variable "create_accounts" {
  description = "Whether to create member accounts (set to false if accounts already exist)"
  type        = bool
  default     = false
}

variable "account_email_prefix" {
  description = "Email prefix for account root emails using Gmail + aliases (format: prefix+alias@gmail.com)"
  type        = string
  # No default - must be provided via terraform.tfvars or -var flag
  # Example: femilawal76

  validation {
    condition     = length(var.account_email_prefix) > 0
    error_message = "account_email_prefix must be provided and cannot be empty."
  }
}

variable "account_email_domain" {
  description = "Email domain for account root emails"
  type        = string
  default     = "gmail.com"
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
# AWS Organization
# ============================================================================

resource "aws_organizations_organization" "main" {
  aws_service_access_principals = var.aws_service_access_principals
  enabled_policy_types          = var.enabled_policy_types
  feature_set                   = var.feature_set
}

# ============================================================================
# Organizational Units (OUs)
# ============================================================================

# Security OU - Contains Log Archive and Audit accounts
resource "aws_organizations_organizational_unit" "security" {
  name      = "Security"
  parent_id = aws_organizations_organization.main.roots[0].id

  tags = merge(var.tags, {
    Name        = "${var.organization_name}-security-ou"
    Purpose     = "Security and compliance accounts"
    Criticality = "Critical"
  })
}

# Infrastructure OU - Contains Network, Shared Services, and Observability accounts
resource "aws_organizations_organizational_unit" "infrastructure" {
  name      = "Infrastructure"
  parent_id = aws_organizations_organization.main.roots[0].id

  tags = merge(var.tags, {
    Name        = "${var.organization_name}-infrastructure-ou"
    Purpose     = "Infrastructure and shared services accounts"
    Criticality = "High"
  })
}

# Workloads OU - Parent for NonProd and Prod OUs
resource "aws_organizations_organizational_unit" "workloads" {
  name      = "Workloads"
  parent_id = aws_organizations_organization.main.roots[0].id

  tags = merge(var.tags, {
    Name        = "${var.organization_name}-workloads-ou"
    Purpose     = "Application workload accounts"
    Criticality = "High"
  })
}

# NonProd OU - Under Workloads
resource "aws_organizations_organizational_unit" "nonprod" {
  name      = "NonProd"
  parent_id = aws_organizations_organizational_unit.workloads.id

  tags = merge(var.tags, {
    Name        = "${var.organization_name}-nonprod-ou"
    Purpose     = "Non-production workload accounts"
    Criticality = "Medium"
  })
}

# Prod OU - Under Workloads
resource "aws_organizations_organizational_unit" "prod" {
  name      = "Prod"
  parent_id = aws_organizations_organizational_unit.workloads.id

  tags = merge(var.tags, {
    Name        = "${var.organization_name}-prod-ou"
    Purpose     = "Production workload accounts"
    Criticality = "Critical"
  })
}

# PCI/CDE OU - Under Workloads for card data isolation
resource "aws_organizations_organizational_unit" "pci_cde" {
  name      = "PCI-CDE"
  parent_id = aws_organizations_organizational_unit.workloads.id

  tags = merge(var.tags, {
    Name        = "${var.organization_name}-pci-cde-ou"
    Purpose     = "PCI DSS Cardholder Data Environment"
    Criticality = "Critical"
    Compliance  = "PCI-DSS"
  })
}

# Sandbox OU - Heavily restricted experimentation environment
resource "aws_organizations_organizational_unit" "sandbox" {
  name      = "Sandbox"
  parent_id = aws_organizations_organization.main.roots[0].id

  tags = merge(var.tags, {
    Name        = "${var.organization_name}-sandbox-ou"
    Purpose     = "Experimentation and sandbox accounts"
    Criticality = "Low"
  })
}

# ============================================================================
# Member Accounts (Optional - only create if var.create_accounts is true)
# ============================================================================

# Log Archive Account
resource "aws_organizations_account" "log_archive" {
  count = var.create_accounts ? 1 : 0

  name      = "${var.organization_name}-log-archive"
  email     = "${var.account_email_prefix}+${var.organization_name}-log-archive@${var.account_email_domain}"
  parent_id = aws_organizations_organizational_unit.security.id
  role_name = "OrganizationAccountAccessRole"

  # Prevent accidental deletion
  close_on_deletion = false

  tags = merge(var.tags, {
    Name        = "${var.organization_name}-log-archive"
    AccountType = "LogArchive"
    Criticality = "Critical"
  })

  lifecycle {
    ignore_changes = [role_name]
  }
}

# Audit Account
resource "aws_organizations_account" "audit" {
  count = var.create_accounts ? 1 : 0

  name      = "${var.organization_name}-audit"
  email     = "${var.account_email_prefix}+${var.organization_name}-audit@${var.account_email_domain}"
  parent_id = aws_organizations_organizational_unit.security.id
  role_name = "OrganizationAccountAccessRole"

  close_on_deletion = false

  tags = merge(var.tags, {
    Name        = "${var.organization_name}-audit"
    AccountType = "Audit"
    Criticality = "Critical"
  })

  lifecycle {
    ignore_changes = [role_name]
  }
}

# Network Account
resource "aws_organizations_account" "network" {
  count = var.create_accounts ? 1 : 0

  name      = "${var.organization_name}-network"
  email     = "${var.account_email_prefix}+${var.organization_name}-network@${var.account_email_domain}"
  parent_id = aws_organizations_organizational_unit.infrastructure.id
  role_name = "OrganizationAccountAccessRole"

  close_on_deletion = false

  tags = merge(var.tags, {
    Name        = "${var.organization_name}-network"
    AccountType = "Network"
    Criticality = "High"
  })

  lifecycle {
    ignore_changes = [role_name]
  }
}

# Shared Services Account
resource "aws_organizations_account" "shared_services" {
  count = var.create_accounts ? 1 : 0

  name      = "${var.organization_name}-shared-services"
  email     = "${var.account_email_prefix}+${var.organization_name}-shared-services@${var.account_email_domain}"
  parent_id = aws_organizations_organizational_unit.infrastructure.id
  role_name = "OrganizationAccountAccessRole"

  close_on_deletion = false

  tags = merge(var.tags, {
    Name        = "${var.organization_name}-shared-services"
    AccountType = "SharedServices"
    Criticality = "High"
  })

  lifecycle {
    ignore_changes = [role_name]
  }
}

# Dev Account
resource "aws_organizations_account" "dev" {
  count = var.create_accounts ? 1 : 0

  name      = "${var.organization_name}-dev"
  email     = "${var.account_email_prefix}+${var.organization_name}-dev@${var.account_email_domain}"
  parent_id = aws_organizations_organizational_unit.nonprod.id
  role_name = "OrganizationAccountAccessRole"

  close_on_deletion = false

  tags = merge(var.tags, {
    Name        = "${var.organization_name}-dev"
    AccountType = "Workload"
    Environment = "dev"
    Criticality = "Medium"
  })

  lifecycle {
    ignore_changes = [role_name]
  }
}

# Stage Account
resource "aws_organizations_account" "stage" {
  count = var.create_accounts ? 1 : 0

  name      = "${var.organization_name}-stage"
  email     = "${var.account_email_prefix}+${var.organization_name}-stage@${var.account_email_domain}"
  parent_id = aws_organizations_organizational_unit.nonprod.id
  role_name = "OrganizationAccountAccessRole"

  close_on_deletion = false

  tags = merge(var.tags, {
    Name        = "${var.organization_name}-stage"
    AccountType = "Workload"
    Environment = "stage"
    Criticality = "Medium"
  })

  lifecycle {
    ignore_changes = [role_name]
  }
}

# Production Account
resource "aws_organizations_account" "prod" {
  count = var.create_accounts ? 1 : 0

  name      = "${var.organization_name}-prod"
  email     = "${var.account_email_prefix}+${var.organization_name}-prod@${var.account_email_domain}"
  parent_id = aws_organizations_organizational_unit.prod.id
  role_name = "OrganizationAccountAccessRole"

  close_on_deletion = false

  tags = merge(var.tags, {
    Name        = "${var.organization_name}-prod"
    AccountType = "Workload"
    Environment = "prod"
    Criticality = "Critical"
  })

  lifecycle {
    ignore_changes = [role_name]
  }
}

# ============================================================================
# Outputs
# ============================================================================

output "organization_id" {
  description = "The ID of the AWS Organization"
  value       = aws_organizations_organization.main.id
}

output "organization_arn" {
  description = "The ARN of the AWS Organization"
  value       = aws_organizations_organization.main.arn
}

output "organization_master_account_id" {
  description = "The ID of the management account"
  value       = aws_organizations_organization.main.master_account_id
}

output "root_ou_id" {
  description = "The ID of the root OU"
  value       = aws_organizations_organization.main.roots[0].id
}

output "security_ou_id" {
  description = "The ID of the Security OU"
  value       = aws_organizations_organizational_unit.security.id
}

output "infrastructure_ou_id" {
  description = "The ID of the Infrastructure OU"
  value       = aws_organizations_organizational_unit.infrastructure.id
}

output "workloads_ou_id" {
  description = "The ID of the Workloads OU"
  value       = aws_organizations_organizational_unit.workloads.id
}

output "nonprod_ou_id" {
  description = "The ID of the NonProd OU"
  value       = aws_organizations_organizational_unit.nonprod.id
}

output "prod_ou_id" {
  description = "The ID of the Prod OU"
  value       = aws_organizations_organizational_unit.prod.id
}

output "pci_cde_ou_id" {
  description = "The ID of the PCI-CDE OU"
  value       = aws_organizations_organizational_unit.pci_cde.id
}

output "sandbox_ou_id" {
  description = "The ID of the Sandbox OU"
  value       = aws_organizations_organizational_unit.sandbox.id
}

output "log_archive_account_id" {
  description = "The ID of the Log Archive account"
  value       = var.create_accounts ? aws_organizations_account.log_archive[0].id : null
}

output "audit_account_id" {
  description = "The ID of the Audit account"
  value       = var.create_accounts ? aws_organizations_account.audit[0].id : null
}

output "ou_ids" {
  description = "Map of all OU IDs"
  value = {
    root           = aws_organizations_organization.main.roots[0].id
    security       = aws_organizations_organizational_unit.security.id
    infrastructure = aws_organizations_organizational_unit.infrastructure.id
    workloads      = aws_organizations_organizational_unit.workloads.id
    nonprod        = aws_organizations_organizational_unit.nonprod.id
    prod           = aws_organizations_organizational_unit.prod.id
    pci_cde        = aws_organizations_organizational_unit.pci_cde.id
    sandbox        = aws_organizations_organizational_unit.sandbox.id
  }
}
