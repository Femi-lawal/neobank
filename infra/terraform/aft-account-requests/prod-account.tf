# ============================================================================
# AFT Account Request: NeoBank Production Account
# ============================================================================

module "neobank_prod_account" {
  source = "./modules/aft-account-request"

  control_tower_parameters = {
    AccountEmail              = "neobank-prod@example.com" # TODO: Update with valid email
    AccountName               = "neobank-prod"
    ManagedOrganizationalUnit = "Prod" # Production OU

    SSOUserEmail     = "admin@example.com"
    SSOUserFirstName = "Admin"
    SSOUserLastName  = "User"
  }

  account_tags = {
    Environment        = "prod"
    AccountType        = "Workload"
    ApplicationName    = "neobank"
    CostCenter         = "Engineering"
    DataClassification = "Restricted"
    Compliance         = "PCI-DSS"
    Criticality        = "Critical"
  }

  account_customizations_name = "neobank-prod"

  change_management_parameters = {
    change_requested_by = "Platform Team"
    change_reason       = "Production environment for neobank application"
  }

  custom_fields = {
    application = "neobank"
    owner       = "platform-team"
    backup      = "enabled"
  }
}
