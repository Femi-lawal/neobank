# ============================================================================
# IRSA Module - IAM Roles for Service Accounts
# ============================================================================

variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "oidc_provider_arn" {
  description = "OIDC provider ARN from EKS"
  type        = string
}

variable "oidc_provider_url" {
  description = "OIDC provider URL from EKS"
  type        = string
}

variable "namespace" {
  description = "Kubernetes namespace for service accounts"
  type        = string
  default     = "default"
}

variable "db_secret_arn" {
  description = "ARN of the database credentials secret"
  type        = string
}

variable "redis_secret_arn" {
  description = "ARN of the Redis credentials secret"
  type        = string
}

variable "jwt_secret_arn" {
  description = "ARN of the JWT secret"
  type        = string
}

variable "app_config_secret_arn" {
  description = "ARN of the app config secret"
  type        = string
}

variable "kms_key_arn" {
  description = "ARN of the KMS key"
  type        = string
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}

# -----------------------------------------------------------------------------
# Identity Service IAM Role
# -----------------------------------------------------------------------------
resource "aws_iam_role" "identity_service" {
  name = "${var.name_prefix}-identity-service-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = var.oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:${var.namespace}:identity-service"
            "${replace(var.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy" "identity_service" {
  name = "${var.name_prefix}-identity-service-policy"
  role = aws_iam_role.identity_service.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = [
          var.db_secret_arn,
          var.redis_secret_arn,
          var.jwt_secret_arn,
          var.app_config_secret_arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
          "kms:GenerateDataKey"
        ]
        Resource = var.kms_key_arn
      }
    ]
  })
}

# -----------------------------------------------------------------------------
# Ledger Service IAM Role
# -----------------------------------------------------------------------------
resource "aws_iam_role" "ledger_service" {
  name = "${var.name_prefix}-ledger-service-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = var.oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:${var.namespace}:ledger-service"
            "${replace(var.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy" "ledger_service" {
  name = "${var.name_prefix}-ledger-service-policy"
  role = aws_iam_role.ledger_service.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = [
          var.db_secret_arn,
          var.redis_secret_arn,
          var.app_config_secret_arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
          "kms:GenerateDataKey"
        ]
        Resource = var.kms_key_arn
      }
    ]
  })
}

# -----------------------------------------------------------------------------
# Payment Service IAM Role
# -----------------------------------------------------------------------------
resource "aws_iam_role" "payment_service" {
  name = "${var.name_prefix}-payment-service-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = var.oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:${var.namespace}:payment-service"
            "${replace(var.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy" "payment_service" {
  name = "${var.name_prefix}-payment-service-policy"
  role = aws_iam_role.payment_service.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = [
          var.db_secret_arn,
          var.redis_secret_arn,
          var.app_config_secret_arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
          "kms:GenerateDataKey"
        ]
        Resource = var.kms_key_arn
      }
    ]
  })
}

# -----------------------------------------------------------------------------
# Product Service IAM Role
# -----------------------------------------------------------------------------
resource "aws_iam_role" "product_service" {
  name = "${var.name_prefix}-product-service-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = var.oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:${var.namespace}:product-service"
            "${replace(var.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy" "product_service" {
  name = "${var.name_prefix}-product-service-policy"
  role = aws_iam_role.product_service.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = [
          var.db_secret_arn,
          var.redis_secret_arn,
          var.app_config_secret_arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
          "kms:GenerateDataKey"
        ]
        Resource = var.kms_key_arn
      }
    ]
  })
}

# -----------------------------------------------------------------------------
# Card Service IAM Role
# -----------------------------------------------------------------------------
resource "aws_iam_role" "card_service" {
  name = "${var.name_prefix}-card-service-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = var.oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:${var.namespace}:card-service"
            "${replace(var.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy" "card_service" {
  name = "${var.name_prefix}-card-service-policy"
  role = aws_iam_role.card_service.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = [
          var.db_secret_arn,
          var.redis_secret_arn,
          var.app_config_secret_arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
          "kms:GenerateDataKey"
        ]
        Resource = var.kms_key_arn
      }
    ]
  })
}

# -----------------------------------------------------------------------------
# Outputs
# -----------------------------------------------------------------------------
output "identity_service_role_arn" {
  description = "IAM Role ARN for identity-service"
  value       = aws_iam_role.identity_service.arn
}

output "ledger_service_role_arn" {
  description = "IAM Role ARN for ledger-service"
  value       = aws_iam_role.ledger_service.arn
}

output "payment_service_role_arn" {
  description = "IAM Role ARN for payment-service"
  value       = aws_iam_role.payment_service.arn
}

output "product_service_role_arn" {
  description = "IAM Role ARN for product-service"
  value       = aws_iam_role.product_service.arn
}

output "card_service_role_arn" {
  description = "IAM Role ARN for card-service"
  value       = aws_iam_role.card_service.arn
}

output "service_account_annotations" {
  description = "Service account annotations for each service"
  value = {
    identity-service = {
      "eks.amazonaws.com/role-arn" = aws_iam_role.identity_service.arn
    }
    ledger-service = {
      "eks.amazonaws.com/role-arn" = aws_iam_role.ledger_service.arn
    }
    payment-service = {
      "eks.amazonaws.com/role-arn" = aws_iam_role.payment_service.arn
    }
    product-service = {
      "eks.amazonaws.com/role-arn" = aws_iam_role.product_service.arn
    }
    card-service = {
      "eks.amazonaws.com/role-arn" = aws_iam_role.card_service.arn
    }
  }
}
