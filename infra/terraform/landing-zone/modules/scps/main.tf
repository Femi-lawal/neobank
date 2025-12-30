# ============================================================================
# Service Control Policies (SCPs) Module
# Bank-grade guardrails for AWS Organization
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

variable "allowed_regions" {
  description = "List of allowed AWS regions"
  type        = list(string)
  default     = ["us-east-1", "us-west-2", "eu-west-1"]
}

variable "required_tags" {
  description = "Tags required on all resources"
  type        = list(string)
  default     = ["Environment", "Project", "Owner", "CostCenter", "DataClassification"]
}

variable "root_ou_id" {
  description = "The ID of the root OU"
  type        = string
}

variable "security_ou_id" {
  description = "The ID of the Security OU"
  type        = string
}

variable "prod_ou_id" {
  description = "The ID of the Production OU"
  type        = string
}

variable "nonprod_ou_id" {
  description = "The ID of the Non-Production OU"
  type        = string
}

variable "sandbox_ou_id" {
  description = "The ID of the Sandbox OU"
  type        = string
}

variable "pci_cde_ou_id" {
  description = "The ID of the PCI-CDE OU"
  type        = string
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}

# ============================================================================
# SCP 1: Prevent Disabling Security Services
# Attached to: Root OU
# ============================================================================

resource "aws_organizations_policy" "deny_disable_security" {
  name        = "${var.organization_name}-deny-disable-security"
  description = "Prevents disabling critical security and audit services"
  type        = "SERVICE_CONTROL_POLICY"

  content = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DenyDisableCloudTrail"
        Effect = "Deny"
        Action = [
          "cloudtrail:DeleteTrail",
          "cloudtrail:StopLogging",
          "cloudtrail:UpdateTrail",
          "cloudtrail:PutEventSelectors"
        ]
        Resource = "*"
        Condition = {
          StringNotLike = {
            "aws:PrincipalArn" = [
              "arn:aws:iam::*:role/AWSControlTowerExecution",
              "arn:aws:iam::*:role/OrganizationAccountAccessRole"
            ]
          }
        }
      },
      {
        Sid    = "DenyDisableConfig"
        Effect = "Deny"
        Action = [
          "config:DeleteConfigRule",
          "config:DeleteConfigurationRecorder",
          "config:DeleteDeliveryChannel",
          "config:StopConfigurationRecorder"
        ]
        Resource = "*"
        Condition = {
          StringNotLike = {
            "aws:PrincipalArn" = [
              "arn:aws:iam::*:role/AWSControlTowerExecution",
              "arn:aws:iam::*:role/OrganizationAccountAccessRole"
            ]
          }
        }
      },
      {
        Sid    = "DenyDisableGuardDuty"
        Effect = "Deny"
        Action = [
          "guardduty:DeleteDetector",
          "guardduty:DeleteMembers",
          "guardduty:DisassociateFromMasterAccount",
          "guardduty:DisassociateMembers",
          "guardduty:StopMonitoringMembers",
          "guardduty:UpdateDetector"
        ]
        Resource = "*"
        Condition = {
          StringNotLike = {
            "aws:PrincipalArn" = [
              "arn:aws:iam::*:role/AWSControlTowerExecution",
              "arn:aws:iam::*:role/OrganizationAccountAccessRole"
            ]
          }
        }
      },
      {
        Sid    = "DenyDisableSecurityHub"
        Effect = "Deny"
        Action = [
          "securityhub:DeleteMembers",
          "securityhub:DisableSecurityHub",
          "securityhub:DisassociateFromMasterAccount",
          "securityhub:DisassociateMembers"
        ]
        Resource = "*"
        Condition = {
          StringNotLike = {
            "aws:PrincipalArn" = [
              "arn:aws:iam::*:role/AWSControlTowerExecution",
              "arn:aws:iam::*:role/OrganizationAccountAccessRole"
            ]
          }
        }
      },
      {
        Sid    = "DenyDisableAccessAnalyzer"
        Effect = "Deny"
        Action = [
          "access-analyzer:DeleteAnalyzer"
        ]
        Resource = "*"
        Condition = {
          StringNotLike = {
            "aws:PrincipalArn" = [
              "arn:aws:iam::*:role/AWSControlTowerExecution",
              "arn:aws:iam::*:role/OrganizationAccountAccessRole"
            ]
          }
        }
      },
      {
        Sid    = "DenyModifyingOrgTrail"
        Effect = "Deny"
        Action = [
          "organizations:LeaveOrganization"
        ]
        Resource = "*"
      }
    ]
  })

  tags = merge(var.tags, {
    Name     = "${var.organization_name}-deny-disable-security"
    SCPType  = "Security"
    Attached = "Root"
  })
}

resource "aws_organizations_policy_attachment" "deny_disable_security_root" {
  policy_id = aws_organizations_policy.deny_disable_security.id
  target_id = var.root_ou_id
}

# ============================================================================
# SCP 2: Region Restriction
# Attached to: Production OU, Non-Production OU
# ============================================================================

resource "aws_organizations_policy" "region_restriction" {
  name        = "${var.organization_name}-region-restriction"
  description = "Restricts AWS service usage to approved regions only"
  type        = "SERVICE_CONTROL_POLICY"

  content = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DenyUnapprovedRegions"
        Effect = "Deny"
        NotAction = [
          # Global services that don't support region restriction
          "a4b:*",
          "budgets:*",
          "ce:*",
          "chime:*",
          "cloudfront:*",
          "config:*",
          "cur:*",
          "globalaccelerator:*",
          "health:*",
          "iam:*",
          "importexport:*",
          "kms:*",
          "mobileanalytics:*",
          "networkmanager:*",
          "organizations:*",
          "pricing:*",
          "route53:*",
          "route53domains:*",
          "s3:GetAccountPublic*",
          "s3:ListAllMyBuckets",
          "s3:PutAccountPublic*",
          "shield:*",
          "sts:*",
          "support:*",
          "trustedadvisor:*",
          "waf:*",
          "waf-regional:*",
          "wafv2:*",
          "wellarchitected:*"
        ]
        Resource = "*"
        Condition = {
          StringNotEquals = {
            "aws:RequestedRegion" = var.allowed_regions
          }
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name     = "${var.organization_name}-region-restriction"
    SCPType  = "Compliance"
    Attached = "Workloads"
  })
}

resource "aws_organizations_policy_attachment" "region_restriction_prod" {
  policy_id = aws_organizations_policy.region_restriction.id
  target_id = var.prod_ou_id
}

resource "aws_organizations_policy_attachment" "region_restriction_nonprod" {
  policy_id = aws_organizations_policy.region_restriction.id
  target_id = var.nonprod_ou_id
}

# ============================================================================
# SCP 3: Encryption Guardrails
# Attached to: Root OU
# ============================================================================

resource "aws_organizations_policy" "require_encryption" {
  name        = "${var.organization_name}-require-encryption"
  description = "Enforces encryption at rest for storage services"
  type        = "SERVICE_CONTROL_POLICY"

  content = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DenyUnencryptedS3Objects"
        Effect = "Deny"
        Action = [
          "s3:PutObject"
        ]
        Resource = "*"
        Condition = {
          Null = {
            "s3:x-amz-server-side-encryption" = "true"
          }
        }
      },
      {
        Sid    = "DenyUnencryptedEBSVolumes"
        Effect = "Deny"
        Action = [
          "ec2:CreateVolume"
        ]
        Resource = "*"
        Condition = {
          Bool = {
            "ec2:Encrypted" = "false"
          }
        }
      },
      {
        Sid    = "DenyUnencryptedRDS"
        Effect = "Deny"
        Action = [
          "rds:CreateDBInstance",
          "rds:CreateDBCluster"
        ]
        Resource = "*"
        Condition = {
          Bool = {
            "rds:StorageEncrypted" = "false"
          }
        }
      },
      {
        Sid    = "DenyDeletingCriticalKMSKeys"
        Effect = "Deny"
        Action = [
          "kms:ScheduleKeyDeletion",
          "kms:Delete*"
        ]
        Resource = "*"
        Condition = {
          StringNotLike = {
            "aws:PrincipalArn" = [
              "arn:aws:iam::*:role/BreakGlassAdmin",
              "arn:aws:iam::*:role/OrganizationAccountAccessRole"
            ]
          }
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name     = "${var.organization_name}-require-encryption"
    SCPType  = "Security"
    Attached = "Root"
  })
}

resource "aws_organizations_policy_attachment" "require_encryption_root" {
  policy_id = aws_organizations_policy.require_encryption.id
  target_id = var.root_ou_id
}

# ============================================================================
# SCP 4: Network Exposure Protection
# Attached to: Production OU, PCI-CDE OU
# ============================================================================

resource "aws_organizations_policy" "deny_public_exposure" {
  name        = "${var.organization_name}-deny-public-exposure"
  description = "Prevents public exposure of resources in sensitive environments"
  type        = "SERVICE_CONTROL_POLICY"

  content = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DenyPublicS3Buckets"
        Effect = "Deny"
        Action = [
          "s3:PutBucketPublicAccessBlock"
        ]
        Resource = "*"
        Condition = {
          Bool = {
            "s3:PublicAccessBlockConfiguration.BlockPublicAcls"       = "false"
            "s3:PublicAccessBlockConfiguration.BlockPublicPolicy"     = "false"
            "s3:PublicAccessBlockConfiguration.IgnorePublicAcls"      = "false"
            "s3:PublicAccessBlockConfiguration.RestrictPublicBuckets" = "false"
          }
        }
      },
      {
        Sid    = "DenyRemovingS3PublicAccessBlock"
        Effect = "Deny"
        Action = [
          "s3:DeleteBucketPublicAccessBlock"
        ]
        Resource = "*"
      },
      {
        Sid    = "DenyPublicRDSInstances"
        Effect = "Deny"
        Action = [
          "rds:CreateDBInstance",
          "rds:ModifyDBInstance"
        ]
        Resource = "*"
        Condition = {
          Bool = {
            "rds:PubliclyAccessible" = "true"
          }
        }
      },
      {
        Sid    = "DenyPublicSubnets"
        Effect = "Deny"
        Action = [
          "ec2:CreateInternetGateway",
          "ec2:AttachInternetGateway"
        ]
        Resource = "*"
        Condition = {
          StringNotLike = {
            "aws:PrincipalArn" = [
              "arn:aws:iam::*:role/NetworkAdmin",
              "arn:aws:iam::*:role/OrganizationAccountAccessRole"
            ]
          }
        }
      },
      {
        Sid    = "DenyOpenSecurityGroups"
        Effect = "Deny"
        Action = [
          "ec2:AuthorizeSecurityGroupIngress",
          "ec2:AuthorizeSecurityGroupEgress"
        ]
        Resource = "*"
        Condition = {
          StringEquals = {
            "ec2:AuthorizedService" = "0.0.0.0/0"
          }
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name     = "${var.organization_name}-deny-public-exposure"
    SCPType  = "Network"
    Attached = "Production"
  })
}

resource "aws_organizations_policy_attachment" "deny_public_exposure_prod" {
  policy_id = aws_organizations_policy.deny_public_exposure.id
  target_id = var.prod_ou_id
}

resource "aws_organizations_policy_attachment" "deny_public_exposure_pci" {
  policy_id = aws_organizations_policy.deny_public_exposure.id
  target_id = var.pci_cde_ou_id
}

# ============================================================================
# SCP 5: Identity Hardening
# Attached to: Root OU
# ============================================================================

resource "aws_organizations_policy" "identity_hardening" {
  name        = "${var.organization_name}-identity-hardening"
  description = "Enforces identity best practices and prevents privilege escalation"
  type        = "SERVICE_CONTROL_POLICY"

  content = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DenyIAMUserCreation"
        Effect = "Deny"
        Action = [
          "iam:CreateUser",
          "iam:CreateAccessKey"
        ]
        Resource = "*"
        Condition = {
          StringNotLike = {
            "aws:PrincipalArn" = [
              "arn:aws:iam::*:role/BreakGlassAdmin",
              "arn:aws:iam::*:role/OrganizationAccountAccessRole",
              "arn:aws:iam::*:role/AWSControlTowerExecution"
            ]
          }
        }
      },
      {
        Sid      = "DenyRootUserActions"
        Effect   = "Deny"
        Action   = "*"
        Resource = "*"
        Condition = {
          StringLike = {
            "aws:PrincipalArn" = "arn:aws:iam::*:root"
          }
        }
      },
      {
        Sid    = "DenyChangingCriticalRoles"
        Effect = "Deny"
        Action = [
          "iam:AttachRolePolicy",
          "iam:DeleteRole",
          "iam:DeleteRolePermissionsBoundary",
          "iam:DeleteRolePolicy",
          "iam:DetachRolePolicy",
          "iam:PutRolePermissionsBoundary",
          "iam:PutRolePolicy",
          "iam:UpdateAssumeRolePolicy",
          "iam:UpdateRole"
        ]
        Resource = [
          "arn:aws:iam::*:role/AWSControlTowerExecution",
          "arn:aws:iam::*:role/OrganizationAccountAccessRole",
          "arn:aws:iam::*:role/BreakGlassAdmin",
          "arn:aws:iam::*:role/aws-reserved/*"
        ]
      },
      {
        Sid    = "RequireMFAForSensitiveActions"
        Effect = "Deny"
        Action = [
          "iam:DeactivateMFADevice",
          "iam:DeleteVirtualMFADevice"
        ]
        Resource = "*"
        Condition = {
          BoolIfExists = {
            "aws:MultiFactorAuthPresent" = "false"
          }
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name     = "${var.organization_name}-identity-hardening"
    SCPType  = "Identity"
    Attached = "Root"
  })
}

resource "aws_organizations_policy_attachment" "identity_hardening_root" {
  policy_id = aws_organizations_policy.identity_hardening.id
  target_id = var.root_ou_id
}

# ============================================================================
# SCP 6: Sandbox Restrictions
# Attached to: Sandbox OU
# ============================================================================

resource "aws_organizations_policy" "sandbox_restrictions" {
  name        = "${var.organization_name}-sandbox-restrictions"
  description = "Heavy restrictions for sandbox/experimentation environments"
  type        = "SERVICE_CONTROL_POLICY"

  content = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DenyExpensiveServices"
        Effect = "Deny"
        Action = [
          "redshift:*",
          "emr:*",
          "sagemaker:Create*",
          "glue:*",
          "dms:*",
          "msk:*",
          "es:*",
          "opensearch:*"
        ]
        Resource = "*"
      },
      {
        Sid    = "DenyLargeEC2Instances"
        Effect = "Deny"
        Action = [
          "ec2:RunInstances"
        ]
        Resource = "arn:aws:ec2:*:*:instance/*"
      },
      {
        Sid    = "DenyLargeRDSInstances"
        Effect = "Deny"
        Action = [
          "rds:CreateDBInstance"
        ]
        Resource = "*"
      },
      {
        Sid    = "DenyVPCPeering"
        Effect = "Deny"
        Action = [
          "ec2:CreateVpcPeeringConnection",
          "ec2:AcceptVpcPeeringConnection"
        ]
        Resource = "*"
      },
      {
        Sid    = "DenyDirectConnect"
        Effect = "Deny"
        Action = [
          "directconnect:*"
        ]
        Resource = "*"
      },
      {
        Sid    = "RestrictToSandboxRegion"
        Effect = "Deny"
        NotAction = [
          "iam:*",
          "sts:*",
          "organizations:*",
          "support:*"
        ]
        Resource = "*"
        Condition = {
          StringNotEquals = {
            "aws:RequestedRegion" = ["us-east-1"]
          }
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name     = "${var.organization_name}-sandbox-restrictions"
    SCPType  = "CostControl"
    Attached = "Sandbox"
  })
}

resource "aws_organizations_policy_attachment" "sandbox_restrictions" {
  policy_id = aws_organizations_policy.sandbox_restrictions.id
  target_id = var.sandbox_ou_id
}

# ============================================================================
# SCP 7: PCI-CDE Strict Controls
# Attached to: PCI-CDE OU
# ============================================================================

resource "aws_organizations_policy" "pci_cde_controls" {
  name        = "${var.organization_name}-pci-cde-controls"
  description = "Strict controls for PCI DSS Cardholder Data Environment"
  type        = "SERVICE_CONTROL_POLICY"

  content = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "RequireVPCEndpoints"
        Effect = "Deny"
        Action = [
          "ec2:CreateVpcEndpoint"
        ]
        Resource = "*"
        Condition = {
          StringNotEquals = {
            "ec2:VpcEndpointType" = "Interface"
          }
        }
      },
      {
        Sid    = "DenyNonPrivateLambda"
        Effect = "Deny"
        Action = [
          "lambda:CreateFunction",
          "lambda:UpdateFunctionConfiguration"
        ]
        Resource = "*"
        Condition = {
          Null = {
            "lambda:VpcIds" = "true"
          }
        }
      },
      {
        Sid    = "EnforceSSLOnS3"
        Effect = "Deny"
        Action = [
          "s3:*"
        ]
        Resource = "*"
        Condition = {
          Bool = {
            "aws:SecureTransport" = "false"
          }
        }
      },
      {
        Sid    = "DenyUnencryptedConnections"
        Effect = "Deny"
        Action = [
          "elasticache:CreateCacheCluster",
          "elasticache:CreateReplicationGroup"
        ]
        Resource = "*"
        Condition = {
          Bool = {
            "elasticache:TransitEncryptionEnabled" = "false"
          }
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name       = "${var.organization_name}-pci-cde-controls"
    SCPType    = "Compliance"
    Compliance = "PCI-DSS"
    Attached   = "PCI-CDE"
  })
}

resource "aws_organizations_policy_attachment" "pci_cde_controls" {
  policy_id = aws_organizations_policy.pci_cde_controls.id
  target_id = var.pci_cde_ou_id
}

# ============================================================================
# Tag Policy for Mandatory Tagging
# ============================================================================

# Tag Policy - Temporarily commented out due to policy document format issues
# Will be re-enabled after validation
/*
resource "aws_organizations_policy" "mandatory_tags" {
  name        = "${var.organization_name}-mandatory-tags"
  description = "Enforces mandatory tags on all resources"
  type        = "TAG_POLICY"

  content = jsonencode({
    tags = {
      Environment = {
        tag_key = {
          "@@assign" = "Environment"
        }
        tag_value = {
          "@@assign" = ["dev", "stage", "prod", "sandbox"]
        }
        enforced_for = {
          "@@assign" = [
            "ec2:instance",
            "ec2:volume",
            "rds:db",
            "s3:bucket",
            "lambda:function",
            "eks:cluster"
          ]
        }
      }
      Project = {
        tag_key = {
          "@@assign" = "Project"
        }
        enforced_for = {
          "@@assign" = [
            "ec2:instance",
            "ec2:volume",
            "rds:db",
            "s3:bucket",
            "lambda:function",
            "eks:cluster"
          ]
        }
      }
      DataClassification = {
        tag_key = {
          "@@assign" = "DataClassification"
        }
        tag_value = {
          "@@assign" = ["Public", "Internal", "Confidential", "Restricted"]
        }
        enforced_for = {
          "@@assign" = [
            "rds:db",
            "s3:bucket",
            "dynamodb:table"
          ]
        }
      }
      CostCenter = {
        tag_key = {
          "@@assign" = "CostCenter"
        }
        enforced_for = {
          "@@assign" = [
            "ec2:instance",
            "rds:db",
            "eks:cluster"
          ]
        }
      }
    }
  })

  tags = merge(var.tags, {
    Name     = "${var.organization_name}-mandatory-tags"
    SCPType  = "Tagging"
    Attached = "Root"
  })
}

resource "aws_organizations_policy_attachment" "mandatory_tags_root" {
  policy_id = aws_organizations_policy.mandatory_tags.id
  target_id = var.root_ou_id
}
*/

# ============================================================================
# Outputs
# ============================================================================

output "scp_ids" {
  description = "Map of all SCP IDs"
  value = {
    deny_disable_security = aws_organizations_policy.deny_disable_security.id
    region_restriction    = aws_organizations_policy.region_restriction.id
    require_encryption    = aws_organizations_policy.require_encryption.id
    deny_public_exposure  = aws_organizations_policy.deny_public_exposure.id
    identity_hardening    = aws_organizations_policy.identity_hardening.id
    sandbox_restrictions  = aws_organizations_policy.sandbox_restrictions.id
    pci_cde_controls      = aws_organizations_policy.pci_cde_controls.id
    # mandatory_tags        = aws_organizations_policy.mandatory_tags.id
  }
}

output "scp_arns" {
  description = "Map of all SCP ARNs"
  value = {
    deny_disable_security = aws_organizations_policy.deny_disable_security.arn
    region_restriction    = aws_organizations_policy.region_restriction.arn
    require_encryption    = aws_organizations_policy.require_encryption.arn
    deny_public_exposure  = aws_organizations_policy.deny_public_exposure.arn
    identity_hardening    = aws_organizations_policy.identity_hardening.arn
    sandbox_restrictions  = aws_organizations_policy.sandbox_restrictions.arn
    pci_cde_controls      = aws_organizations_policy.pci_cde_controls.arn
    # mandatory_tags        = aws_organizations_policy.mandatory_tags.arn
  }
}
