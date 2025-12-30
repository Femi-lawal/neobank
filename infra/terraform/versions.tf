# ============================================================================
# NeoBank AWS Infrastructure - Terraform Versions & Providers
# ============================================================================
# Following AWS best practices for banking/financial services:
# - PCI-DSS compliance requirements
# - SOC 2 controls
# - NIST 800-53 guidelines
# ============================================================================

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.25"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }

}

# Configure the AWS Provider
# Note: When running from management account, set assume_role_enabled = true
# When already in target account (assumed role), set assume_role_enabled = false
provider "aws" {
  region = var.aws_region

  # Assume role only when explicitly enabled (running from management account)
  dynamic "assume_role" {
    for_each = var.assume_role_enabled && var.target_account_id != "" ? [1] : []
    content {
      role_arn     = "arn:aws:iam::${var.target_account_id}:role/OrganizationAccountAccessRole"
      session_name = "TerraformDeployment"
    }
  }

  default_tags {
    tags = {
      Project             = "NeoBank"
      Environment         = var.environment
      ManagedBy           = "Terraform"
      DataClassification  = "Confidential"
      ComplianceFramework = "PCI-DSS_SOC2_ISO27001"
    }
  }
}

# Secondary provider for disaster recovery region
provider "aws" {
  alias  = "dr"
  region = var.dr_region

  default_tags {
    tags = {
      Project             = "NeoBank"
      Environment         = var.environment
      ManagedBy           = "Terraform"
      DataClassification  = "Confidential"
      ComplianceFramework = "PCI-DSS,SOC2,ISO27001"
      Purpose             = "DisasterRecovery"
    }
  }
}

# Kubernetes provider - configured after EKS cluster is created
provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
  }
}

# Helm provider for deploying charts
provider "helm" {
  kubernetes {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)

    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
    }
  }
}
