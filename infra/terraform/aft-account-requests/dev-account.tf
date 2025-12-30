# ============================================================================
# AFT Account Request: NeoBank Dev Account
# This file will be placed in the aft-account-request CodeCommit repository
# ============================================================================

module "neobank_dev_account" {
  source = "./modules/aft-account-request"

  control_tower_parameters = {
    # Account Details
    AccountEmail = "neobank-dev@example.com" # TODO: Update with valid email
    AccountName  = "neobank-dev"

    # Organizational Unit (OU) Path
    ManagedOrganizationalUnit = "NonProd" # Places account in NonProd OU

    # SSO Configuration
    SSOUserEmail     = "admin@example.com" # TODO: Update
    SSOUserFirstName = "Admin"
    SSOUserLastName  = "User"
  }

  # Account Tags
  account_tags = {
    Environment        = "dev"
    AccountType        = "Workload"
    ApplicationName    = "neobank"
    CostCenter         = "Engineering"
    DataClassification = "Confidential"
    Compliance         = "PCI-DSS"
  }

  # Terraform Configuration for Account Customizations
  account_customizations_name = "neobank-dev"

  # Change Management Metadata
  change_management_parameters = {
    change_requested_by = "Platform Team"
    change_reason       = "Initial dev account for neobank application"
  }

  # Custom Fields (optional)
  custom_fields = {
    application = "neobank"
    owner       = "platform-team"
  }
}
