# ============================================================================
# AFT Variables
# ============================================================================

variable "aws_region" {
  description = "Primary AWS region for AFT deployment"
  type        = string
  default     = "us-east-1"
}

variable "ct_secondary_regions" {
  description = "Secondary regions for Control Tower"
  type        = list(string)
  default     = ["us-west-2"]
}

variable "log_archive_account_id" {
  description = "Log Archive account ID created by Control Tower"
  type        = string
  default     = "" # Will be populated after Control Tower setup
}

variable "audit_account_id" {
  description = "Audit account ID created by Control Tower"
  type        = string
  default     = "" # Will be populated after Control Tower setup
}

variable "aft_management_account_id" {
  description = "AFT Management account ID (can be same as CT management)"
  type        = string
  default     = "" # Leave empty to use current account
}

variable "terraform_version" {
  description = "Terraform version for AFT to use (minimum 1.6.0 required)"
  type        = string
  default     = "1.6.6"

  validation {
    condition     = can(regex("^1\\.(6|[7-9]|[1-9][0-9])\\.", var.terraform_version))
    error_message = "AFT requires Terraform version 1.6.0 or later."
  }
}

variable "aft_vpc_cidr" {
  description = "CIDR block for AFT VPC"
  type        = string
  default     = "10.100.0.0/16"
}

variable "vcs_provider" {
  description = "Version control system provider (codecommit, github, gitlab, bitbucket)"
  type        = string
  default     = "codecommit"
}

variable "account_request_repo_name" {
  description = "Name of the account request repository"
  type        = string
  default     = "aft-account-request"
}

variable "global_customizations_repo_name" {
  description = "Name of the global customizations repository"
  type        = string
  default     = "aft-global-customizations"
}

variable "account_customizations_repo_name" {
  description = "Name of the account customizations repository"
  type        = string
  default     = "aft-account-customizations"
}

variable "account_provisioning_customizations_repo_name" {
  description = "Name of the account provisioning customizations repository"
  type        = string
  default     = "aft-account-provisioning-customizations"
}

variable "enable_enterprise_support" {
  description = "Enable AWS Enterprise Support for accounts"
  type        = bool
  default     = false # Set to true if you have Enterprise Support
}
