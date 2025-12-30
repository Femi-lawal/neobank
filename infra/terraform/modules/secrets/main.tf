# ============================================================================
# Secrets Module - AWS Secrets Manager
# ============================================================================
# Manages all application secrets with automatic rotation support
# ============================================================================

variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "kms_key_arn" {
  description = "ARN of KMS key for encryption"
  type        = string
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}

# -----------------------------------------------------------------------------
# Generate secure random passwords
# -----------------------------------------------------------------------------
resource "random_password" "db_master" {
  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}:?"
}

resource "random_password" "redis_auth" {
  length  = 64
  special = false
}

resource "random_password" "jwt_secret" {
  length  = 64
  special = false
}

# -----------------------------------------------------------------------------
# Database Credentials Secret
# -----------------------------------------------------------------------------
resource "aws_secretsmanager_secret" "db_credentials" {
  name        = "${var.name_prefix}/database/credentials"
  description = "PostgreSQL database credentials for NeoBank"
  kms_key_id  = var.kms_key_arn

  tags = merge(var.tags, {
    Name = "${var.name_prefix}-db-credentials"
  })
}

resource "aws_secretsmanager_secret_version" "db_credentials" {
  secret_id = aws_secretsmanager_secret.db_credentials.id
  secret_string = jsonencode({
    username = "neobank_admin"
    password = random_password.db_master.result
    database = "newbank_core"
    port     = 5432
  })
}

# -----------------------------------------------------------------------------
# Redis Credentials Secret
# -----------------------------------------------------------------------------
resource "aws_secretsmanager_secret" "redis_credentials" {
  name        = "${var.name_prefix}/redis/credentials"
  description = "Redis authentication credentials for NeoBank"
  kms_key_id  = var.kms_key_arn

  tags = merge(var.tags, {
    Name = "${var.name_prefix}-redis-credentials"
  })
}

resource "aws_secretsmanager_secret_version" "redis_credentials" {
  secret_id = aws_secretsmanager_secret.redis_credentials.id
  secret_string = jsonencode({
    auth_token = random_password.redis_auth.result
  })
}

# -----------------------------------------------------------------------------
# JWT Secret
# -----------------------------------------------------------------------------
resource "aws_secretsmanager_secret" "jwt_secret" {
  name        = "${var.name_prefix}/auth/jwt-secret"
  description = "JWT signing secret for NeoBank authentication"
  kms_key_id  = var.kms_key_arn

  tags = merge(var.tags, {
    Name = "${var.name_prefix}-jwt-secret"
  })
}

resource "aws_secretsmanager_secret_version" "jwt_secret" {
  secret_id = aws_secretsmanager_secret.jwt_secret.id
  secret_string = jsonencode({
    secret = random_password.jwt_secret.result
  })
}

# -----------------------------------------------------------------------------
# Application Configuration Secret (non-sensitive configs)
# -----------------------------------------------------------------------------
resource "aws_secretsmanager_secret" "app_config" {
  name        = "${var.name_prefix}/app/config"
  description = "Application configuration for NeoBank services"
  kms_key_id  = var.kms_key_arn

  tags = merge(var.tags, {
    Name = "${var.name_prefix}-app-config"
  })
}

resource "aws_secretsmanager_secret_version" "app_config" {
  secret_id = aws_secretsmanager_secret.app_config.id
  secret_string = jsonencode({
    log_level    = "info"
    log_format   = "json"
    server_mode  = "release"
    jwt_expiry   = "24h"
    bcrypt_cost  = 12
    cors_origins = ["https://neobank.example.com"]
  })
}

# -----------------------------------------------------------------------------
# Outputs
# -----------------------------------------------------------------------------
output "db_credentials_arn" {
  description = "ARN of the database credentials secret"
  value       = aws_secretsmanager_secret.db_credentials.arn
}

output "db_credentials_name" {
  description = "Name of the database credentials secret"
  value       = aws_secretsmanager_secret.db_credentials.name
}

output "db_master_password" {
  description = "Database master password"
  value       = random_password.db_master.result
  sensitive   = true
}

output "redis_credentials_arn" {
  description = "ARN of the Redis credentials secret"
  value       = aws_secretsmanager_secret.redis_credentials.arn
}

output "redis_credentials_name" {
  description = "Name of the Redis credentials secret"
  value       = aws_secretsmanager_secret.redis_credentials.name
}

output "redis_auth_token" {
  description = "Redis auth token"
  value       = random_password.redis_auth.result
  sensitive   = true
}

output "jwt_secret_arn" {
  description = "ARN of the JWT secret"
  value       = aws_secretsmanager_secret.jwt_secret.arn
}

output "jwt_secret_name" {
  description = "Name of the JWT secret"
  value       = aws_secretsmanager_secret.jwt_secret.name
}

output "app_config_arn" {
  description = "ARN of the application config secret"
  value       = aws_secretsmanager_secret.app_config.arn
}
