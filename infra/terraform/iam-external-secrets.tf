# IAM Role for External Secrets Operator
resource "aws_iam_role" "external_secrets" {
  name               = "${local.name_prefix}-external-secrets-role"
  assume_role_policy = data.aws_iam_policy_document.external_secrets_trust.json

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-external-secrets-role"
    }
  )
}

# Trust policy for External Secrets ServiceAccount
data "aws_iam_policy_document" "external_secrets_trust" {
  statement {
    effect = "Allow"
    principals {
      type        = "Federated"
      identifiers = [module.eks.oidc_provider_arn]
    }
    actions = ["sts:AssumeRoleWithWebIdentity"]
    condition {
      test     = "StringEquals"
      variable = "${replace(module.eks.oidc_provider_url, "https://", "")}:sub"
      values   = ["system:serviceaccount:neobank:external-secrets-sa"]
    }
    condition {
      test     = "StringEquals"
      variable = "${replace(module.eks.oidc_provider_url, "https://", "")}:aud"
      values   = ["sts.amazonaws.com"]
    }
  }
}

# IAM policy for External Secrets to read from Secrets Manager
resource "aws_iam_policy" "external_secrets" {
  name        = "${local.name_prefix}-external-secrets-policy"
  description = "Policy for External Secrets Operator to read secrets from Secrets Manager"

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
          module.secrets.db_credentials_arn,
          module.secrets.jwt_secret_arn,
          module.secrets.redis_credentials_arn,
          module.secrets.app_config_arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
          "kms:DescribeKey"
        ]
        Resource = [
          module.security.kms_key_arn
        ]
      }
    ]
  })

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-external-secrets-policy"
    }
  )
}

# Attach policy to role
resource "aws_iam_role_policy_attachment" "external_secrets" {
  role       = aws_iam_role.external_secrets.name
  policy_arn = aws_iam_policy.external_secrets.arn
}

# Output the role ARN
output "external_secrets_role_arn" {
  description = "ARN of the IAM role for External Secrets Operator"
  value       = aws_iam_role.external_secrets.arn
}
