# ============================================================================
# Control Tower Variables
# ============================================================================

variable "aws_region" {
  description = "Primary AWS region for Control Tower (becomes home region - cannot be changed later)"
  type        = string
  default     = "us-east-1"
}

variable "governed_regions" {
  description = "Regions governed by Control Tower"
  type        = list(string)
  default     = ["us-east-1"]
}

variable "control_tower_version" {
  description = "Control Tower Landing Zone version (4.0 is latest)"
  type        = string
  default     = "4.0"
}

variable "log_archive_account_id" {
  description = "Log Archive account ID (required - must exist before landing zone creation)"
  type        = string

  validation {
    condition     = can(regex("^[0-9]{12}$", var.log_archive_account_id))
    error_message = "Log Archive account ID must be a 12-digit AWS account ID."
  }
}

variable "audit_account_id" {
  description = "Audit account ID (required - must exist before landing zone creation)"
  type        = string

  validation {
    condition     = can(regex("^[0-9]{12}$", var.audit_account_id))
    error_message = "Audit account ID must be a 12-digit AWS account ID."
  }
}

variable "log_retention_days" {
  description = "Retention days for logging bucket"
  type        = number
  default     = 365
}

variable "access_log_retention_days" {
  description = "Retention days for access logging bucket"
  type        = number
  default     = 3653 # 10 years for compliance
}
