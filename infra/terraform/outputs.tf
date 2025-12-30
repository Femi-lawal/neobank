# ============================================================================
# Terraform Outputs
# ============================================================================

# -----------------------------------------------------------------------------
# VPC Outputs
# -----------------------------------------------------------------------------
output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = module.vpc.private_subnet_ids
}

output "database_subnet_ids" {
  description = "Database subnet IDs"
  value       = module.vpc.database_subnet_ids
}

# -----------------------------------------------------------------------------
# EKS Outputs
# -----------------------------------------------------------------------------
output "eks_cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name
}

output "eks_cluster_endpoint" {
  description = "EKS cluster API endpoint"
  value       = module.eks.cluster_endpoint
}

output "eks_cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data for cluster authentication"
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
}

output "eks_oidc_provider_arn" {
  description = "OIDC provider ARN for IRSA"
  value       = module.eks.oidc_provider_arn
}

# -----------------------------------------------------------------------------
# Database Outputs
# -----------------------------------------------------------------------------
output "rds_endpoint" {
  description = "RDS instance endpoint"
  value       = module.rds.endpoint
}

output "rds_port" {
  description = "RDS instance port"
  value       = module.rds.port
}

output "rds_database_name" {
  description = "RDS database name"
  value       = module.rds.database_name
}

# -----------------------------------------------------------------------------
# ElastiCache Outputs
# -----------------------------------------------------------------------------
output "redis_endpoint" {
  description = "ElastiCache Redis endpoint"
  value       = module.elasticache.endpoint
}

output "redis_port" {
  description = "ElastiCache Redis port"
  value       = module.elasticache.port
}

# -----------------------------------------------------------------------------
# MSK Outputs (Conditional)
# -----------------------------------------------------------------------------
output "msk_bootstrap_brokers" {
  description = "MSK bootstrap brokers"
  value       = var.enable_msk ? module.msk[0].bootstrap_brokers_tls : ""
}

output "msk_bootstrap_brokers_iam" {
  description = "MSK bootstrap brokers for IAM auth"
  value       = var.enable_msk ? module.msk[0].bootstrap_brokers_iam : ""
}

# -----------------------------------------------------------------------------
# Secrets Outputs
# -----------------------------------------------------------------------------
output "db_secret_arn" {
  description = "ARN of the database credentials secret"
  value       = module.secrets.db_credentials_arn
  sensitive   = true
}

output "redis_secret_arn" {
  description = "ARN of the Redis credentials secret"
  value       = module.secrets.redis_credentials_arn
  sensitive   = true
}

output "jwt_secret_arn" {
  description = "ARN of the JWT secret"
  value       = module.secrets.jwt_secret_arn
  sensitive   = true
}

output "app_config_secret_arn" {
  description = "ARN of the application config secret"
  value       = module.secrets.app_config_arn
  sensitive   = true
}

# -----------------------------------------------------------------------------
# IRSA Outputs
# -----------------------------------------------------------------------------
output "identity_service_role_arn" {
  description = "IAM Role ARN for identity-service"
  value       = module.irsa.identity_service_role_arn
}

output "ledger_service_role_arn" {
  description = "IAM Role ARN for ledger-service"
  value       = module.irsa.ledger_service_role_arn
}

output "payment_service_role_arn" {
  description = "IAM Role ARN for payment-service"
  value       = module.irsa.payment_service_role_arn
}

output "product_service_role_arn" {
  description = "IAM Role ARN for product-service"
  value       = module.irsa.product_service_role_arn
}

output "card_service_role_arn" {
  description = "IAM Role ARN for card-service"
  value       = module.irsa.card_service_role_arn
}

# -----------------------------------------------------------------------------
# S3 Outputs
# -----------------------------------------------------------------------------
output "logs_bucket_name" {
  description = "S3 bucket name for logs"
  value       = module.s3.logs_bucket_id
}

output "backups_bucket_name" {
  description = "S3 bucket name for backups"
  value       = module.s3.backups_bucket_id
}

# -----------------------------------------------------------------------------
# Security Outputs
# -----------------------------------------------------------------------------
output "kms_key_arn" {
  description = "KMS key ARN for encryption"
  value       = module.security.kms_key_arn
}

output "kms_key_id" {
  description = "KMS key ID"
  value       = module.security.kms_key_id
}

# -----------------------------------------------------------------------------
# WAF Outputs (Conditional)
# -----------------------------------------------------------------------------
output "waf_web_acl_arn" {
  description = "WAF Web ACL ARN"
  value       = var.enable_waf ? module.waf[0].web_acl_arn : ""
}

# -----------------------------------------------------------------------------
# Monitoring Outputs
# -----------------------------------------------------------------------------
output "sns_alarms_topic_arn" {
  description = "SNS topic ARN for alarms"
  value       = module.monitoring.sns_topic_arn
}

output "cloudwatch_dashboard_name" {
  description = "CloudWatch dashboard name"
  value       = module.monitoring.dashboard_name
}

# -----------------------------------------------------------------------------
# Kubeconfig Command
# -----------------------------------------------------------------------------
output "configure_kubectl" {
  description = "Command to configure kubectl"
  value       = "aws eks update-kubeconfig --region ${var.aws_region} --name ${module.eks.cluster_name}"
}

# -----------------------------------------------------------------------------
# ECR Outputs
# -----------------------------------------------------------------------------
output "ecr_repository_urls" {
  description = "Map of ECR repository URLs"
  value       = module.ecr.repository_urls
}
