# ============================================================================
# AFT Account Request: NeoBank Staging Account
# ============================================================================

module "neobank_staging_account" {
  source = "./modules/aft-account-request"

  control_tower_parameters = {
    AccountEmail              = "neobank-staging@example.com" # TODO: Update with valid email
    AccountName               = "neobank-staging"
    ManagedOrganizationalUnit = "NonProd"

    SSOUserEmail     = "admin@example.com"
    SSOUserFirstName = "Admin"
    SSOUserLastName  = "User"
  }

  account_tags = {
    Environment        = "staging"
    AccountType        = "Workload"
    ApplicationName    = "neobank"
    CostCenter         = "Engineering"
    DataClassification = "Confidential"
    Compliance         = "PCI-DSS"
  }

  account_customizations_name = "neobank-staging"

  change_management_parameters = {
    change_requested_by = "Platform Team"
    change_reason       = "Staging environment for pre-production testing"
  }

  custom_fields = {
    application = "neobank"
    owner       = "platform-team"
  }
}
